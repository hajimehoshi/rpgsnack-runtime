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
	"path"

	"github.com/gopherjs/gopherjs/js"
)

func fetch(path string) <-chan []uint8 {
	ch := make(chan []uint8)
	js.Global.Call("fetch", path).Call("then", func(res *js.Object) *js.Object {
		return res.Call("arrayBuffer")
	}).Call("then", func(buf *js.Object) {
		ch <- js.Global.Get("Uint8Array").New(buf).Interface().([]uint8)
		close(ch)
	})
	return ch
}

func loadRawData(projectPath string) (*rawData, error) {
	return &rawData{
		Project:   <-fetch(path.Join(projectPath, "project.json")),
		Assets:    <-fetch(path.Join(projectPath, "assets.msgpack")),
		Progress:  nil, // TODO: Implement this
		Purchases: nil, // TODO: Implement this
		Language:  []uint8(`"en"`),
	}, nil
}
