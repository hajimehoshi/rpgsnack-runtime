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

// +build js gomobilebuild

package data

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/vmihailenco/msgpack"

	"github.com/gopherjs/gopherjs/js"
)

type manifestBody struct {
	Manifest map[string][]string `json:"manifest"`
}

type manifestResponse struct {
	Body *manifestBody `json:"body"`
}

type fetchResult struct {
	Body []byte
	Err  error
}

func fetch(path string) <-chan fetchResult {
	ch := make(chan fetchResult)
	go func() {
		defer close(ch)

		res, err := http.Get(path)
		if err != nil {
			ch <- fetchResult{
				nil, err,
			}
			return
		}
		bs, err := ioutil.ReadAll(res.Body)
		if err != nil {
			ch <- fetchResult{
				nil, err,
			}
			return
		}
		ch <- fetchResult{
			bs, nil,
		}
	}()
	return ch
}

func fetchProgress() <-chan []byte {
	ch := make(chan []byte)
	if js.Global == nil {
		go func() {
			ch <- nil
			close(ch)
		}()
		return ch
	}
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

func loadManifest(path string) (map[string][]string, error) {
	res := <-fetch(path)
	if res.Err != nil {
		return nil, res.Err
	}

	mr := manifestResponse{}
	if err := unmarshalJSON(res.Body, &mr); err != nil {
		return nil, fmt.Errorf("unmarshalJSON Error %s", err)
	}
	return mr.Body.Manifest, nil
}

func loadAssetsFromManifest(manifest map[string][]string, progress chan<- float64) ([]byte, []byte, error) {
	// TODO: We should remove this hardcoded value in the future
	const storageUrl = "https://storage.googleapis.com/rpgsnack-e85d3.appspot.com"

	var projectData []byte

	var wg sync.WaitGroup
	loadedCh := make(chan map[string][]byte)
	errCh := make(chan error)
	for key, paths := range manifest {
		wg.Add(1)
		go func(key string, paths []string) {
			defer wg.Done()

			url := key
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = fmt.Sprintf("%s/%s", storageUrl, key)
			}
			res := <-fetch(url)
			if res.Err != nil {
				// TODO: Use context.Context?
				errCh <- res.Err
				return
			}

			data := map[string][]byte{}
			for _, value := range paths {
				switch {
				case value == "project.json":
					projectData = res.Body
				case strings.HasPrefix(value, "assets/"):
					localPath := strings.Replace(value, "assets/", "", 1)
					data[localPath] = res.Body
				}
			}
			loadedCh <- data
		}(key, paths)
	}

	go func() {
		wg.Wait()
		close(loadedCh)
	}()

	assetData := map[string][]byte{}
	nloaded := 0
loadLoop:
	for {
		select {
		case data, ok := <-loadedCh:
			if !ok {
				break loadLoop
			}
			for k, v := range data {
				assetData[k] = v
			}
			nloaded++
			progress <- float64(nloaded) / float64(len(manifest))
		case err := <-errCh:
			return nil, nil, err
		}
	}

	b, err := msgpack.Marshal(assetData)
	if err != nil {
		return nil, nil, fmt.Errorf("MsgPack Error %s", err)
	}
	progress <- 1

	return projectData, b, nil
}

// TODO: Change the API from `web`.
var gameIDUrlRegexp = regexp.MustCompile(`\A/web/([0-9]+)\z`)

func gameIDFromURL() (string, error) {
	if js.Global == nil {
		panic("not reached")
	}

	href := js.Global.Get("window").Get("location").Get("href").String()
	u, err := url.Parse(href)
	if err != nil {
		panic(err)
	}

	v := u.Query().Get("version")
	if v != "" {
		return v, nil
	}
	arr := gameIDUrlRegexp.FindStringSubmatch(u.Path)
	if len(arr) == 2 {
		return arr[1], nil
	}
	return "", fmt.Errorf("data: invalid URL: version is not specified?: %s", u)
}

func isLoopback() bool {
	if js.Global == nil {
		panic("not reached")
	}

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

func loadRawData(projectLocation string, progress chan<- float64) (*rawData, error) {
	defer close(progress)

	// If a project path is not specified from the URL query,
	// get the game ID from the URL path.
	if projectLocation == "" && js.Global != nil {
		gameID, err := gameIDFromURL()
		if err != nil {
			return nil, err
		}

		projectLocation = fmt.Sprintf("/games/%s", gameID)
		// TODO: This is a dirty hack to do tests on local machines.
		// useDefaultURL should be specificed in another way e.g. from clients.
		if isLoopback() {
			// TODO: Stop hard-coding URLs.
			const defaultURL = "https://rpgsnack-e85d3.appspot.com"
			projectLocation = defaultURL + projectLocation
		}
	}

	manifest, err := loadManifest(projectLocation)
	if err != nil {
		return nil, err
	}

	// TODO: manifest might be nil on local server.

	project, assets, err := loadAssetsFromManifest(manifest, progress)
	if err != nil {
		return nil, err
	}

	l := "en"
	if js.Global != nil {
		l = js.Global.Get("navigator").Get("language").String()
	}

	return &rawData{
		Project:   project,
		Assets:    assets,
		Progress:  <-fetchProgress(),
		Purchases: nil, // TODO: Implement this
		Language:  []byte(fmt.Sprintf(`"%s"`, l)),
	}, nil
}
