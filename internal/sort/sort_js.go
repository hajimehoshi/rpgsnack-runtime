// Copyright 2017 Hajime Hoshi
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

package sort

import (
	"github.com/gopherjs/gopherjs/js"
)

func Slice(slice interface{}, comp func(i, j int) bool) {
	// The standard package's sort.Sort is slow on browsers.
	// Let's use native Array.prototype.sort for performance.
	a := js.InternalObject(slice).Get("$array")
	o := js.InternalObject(slice).Get("$offset").Int()
	l := js.InternalObject(slice).Get("$length").Int()
	if l == 0 {
		return
	}
	orig := a.Call("slice")
	indices := js.Global.Get("Array").New(l)
	for i := 0; i < indices.Length(); i++ {
		indices.SetIndex(i, i)
	}
	indices.Call("sort", func(i, j *js.Object) bool {
		return comp(i.Int(), j.Int())
	})
	for i := 0; i < indices.Length(); i++ {
		a.SetIndex(i+o, orig.Index(indices.Index(i).Int()+o))
	}
}
