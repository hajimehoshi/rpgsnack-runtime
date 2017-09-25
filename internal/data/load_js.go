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
	"log"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

func fetch(path string) <-chan []uint8 {
	// TODO: Use fetch API in the future.
	ch := make(chan []uint8)
	xhr := js.Global.Get("XMLHttpRequest").New()
	xhr.Set("responseType", "arraybuffer")
	xhr.Call("addEventListener", "load", func() {
		res := xhr.Get("response")
		ch <- js.Global.Get("Uint8Array").New(res).Interface().([]uint8)
		close(ch)
	})
	xhr.Call("open", "GET", path)
	xhr.Call("send")
	return ch
}

func fetchProgress() <-chan []uint8 {
	ch := make(chan []uint8)
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

func loadRawData(projectPath string) (*rawData, error) {
	// projectPath might be an absolute path, and in this case path.Join doesn't work.
	for strings.HasSuffix(projectPath, "/") {
		projectPath = projectPath[:len(projectPath)-1]
	}
	return &rawData{
		Project:   <-fetch(projectPath + "/project.json"),
		Assets:    <-fetch(projectPath + "/assets.msgpack"),
		Progress:  <-fetchProgress(),
		Purchases: nil, // TODO: Implement this
		Language:  []uint8(`"en"`),
	}, nil
}
