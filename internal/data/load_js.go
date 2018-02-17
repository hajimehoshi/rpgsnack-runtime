// Copyright 2016 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build js

package data

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/vmihailenco/msgpack"

	"github.com/gopherjs/gopherjs/js"
)

// TODO: We should remove this hardcoded value in the future
const storageUrl = "https://storage.googleapis.com/rpgsnack-e85d3.appspot.com"

type manifestBody struct {
	Manifest map[string][]string `json:"manifest"`
}

type manifestResponse struct {
	Body *manifestBody `json:"body"`
}

func fetch(path string) <-chan []byte {
	// TODO: Use fetch API in the future.
	ch := make(chan []byte)
	xhr := js.Global.Get("XMLHttpRequest").New()
	xhr.Set("responseType", "arraybuffer")
	xhr.Call("addEventListener", "load", func() {
		res := xhr.Get("response")
		ch <- js.Global.Get("Uint8Array").New(res).Interface().([]byte)
		close(ch)
	})
	xhr.Call("open", "GET", path)
	xhr.Call("send")
	return ch
}

func fetchProgress() <-chan []byte {
	ch := make(chan []byte)
	go func() {
		data := js.Global.Get("localStorage").Call("getItem", "progress")
		if data == nil {
			close(ch)
			return
		}
		b, err := base64.StdEncoding.DecodeString(data.String())
		if err != nil {
			log.Printf("localStroge's progress is invalid: %v", err)
			close(ch)
			return
		}
		ch <- b
		close(ch)
	}()
	return ch
}

func loadAssets(path string) ([]byte, []byte, error) {
	mBinary := <-fetch(path)

	mr := manifestResponse{}
	if err := unmarshalJSON(mBinary, &mr); err != nil {
		return nil, nil, fmt.Errorf("unmarshalJSON Error %s", err)
	}

	var projectData []byte
	assetData := make(map[string][]byte, len(mr.Body.Manifest))

	var wg sync.WaitGroup
	for key, paths := range mr.Body.Manifest {
		for _, value := range paths {
			wg.Add(1)
			go func(key, value string) {
				defer wg.Done()
				if value == "project.json" {
					projectData = <-fetch(fmt.Sprintf("%s/%s", storageUrl, key))
				} else {
					localPath := strings.Replace(value, "assets/", "", -1)
					assetData[localPath] = <-fetch(fmt.Sprintf("%s/%s", storageUrl, key))
				}
			}(key, value)
		}
	}

	wg.Wait()

	b, err := msgpack.Marshal(assetData)
	if err != nil {
		return nil, nil, fmt.Errorf("MsgPack Error %s", err)
	}

	return projectData, b, nil
}

// TODO: Change the API from `web`.
var gameVersionUrlRegexp = regexp.MustCompile(`\A/web/([0-9]+)\z`)

func versionFromURL() (string, error) {
	href := js.Global.Get("window").Get("location").Get("href").String()
	u, err := url.Parse(href)
	if err != nil {
		panic(err)
	}

	v := u.Query().Get("version")
	if v != "" {
		return v, nil
	}
	arr := gameVersionUrlRegexp.FindStringSubmatch(u.Path)
	if len(arr) == 2 {
		return arr[1], nil
	}
	return "", fmt.Errorf("data: invalid URL: version is not specified?: %s", u)
}

func isLoopback() bool {
	href := js.Global.Get("window").Get("location").Get("href").String()
	u, err := url.Parse(href)
	if err != nil {
		panic(err)
	}

	if u.Hostname() == "localhost" {
		return true
	}
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		if ip.IsLoopback() {
			return true
		}
		if ip.IsGlobalUnicast() {
			return true
		}
	}

	return false
}

func loadRawData(projectPath string) (*rawData, error) {
	// projectPath is ignored so far.

	gameVersion, err := versionFromURL()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/games/%s", gameVersion)
	// TODO: This is a dirty hack to do tests on local machines.
	// useDefaultURL should be specificed in another way e.g. from clients.
	if isLoopback() {
		// TODO: Stop hard-coding URLs.
		const defaultURL = "https://rpgsnack-e85d3.appspot.com"
		path = defaultURL + path
	}

	project, assets, err := loadAssets(path)
	if err != nil {
		return nil, err
	}

	l := js.Global.Get("navigator").Get("language").String()

	return &rawData{
		Project:   project,
		Assets:    assets,
		Progress:  <-fetchProgress(),
		Purchases: nil, // TODO: Implement this
		Language:  []byte(fmt.Sprintf(`"%s"`, l)),
	}, nil
}
