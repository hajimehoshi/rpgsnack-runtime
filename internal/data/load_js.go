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
	"github.com/gopherjs/gopherjs/js"
)

func loadJSONData() (*jsonData, error) {
	if js.Global.Get("_data") == nil {
		ch := make(chan struct{})
		js.Global.Set("_dataNotify", func() {
			close(ch)
		})
		<-ch
	}
	println(js.Global.Get("_data"))
	dataJsonStr := js.Global.Get("JSON").Call("stringify", js.Global.Get("_data"))
	dataJson := ([]uint8)(dataJsonStr.String())
	return &jsonData{
		Game:      dataJson,
		Progress:  nil, // TODO: Implement this
		Purchases: nil, // TODO: Implement this
		Language:  "en",
	}, nil
}
