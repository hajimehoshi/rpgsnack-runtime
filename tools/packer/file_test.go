// Copyright 2018 Hajime Hoshi
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

package main_test

import (
	"testing"

	. "github.com/hajimehoshi/rpgsnack-runtime/tools/packer"
)

func TestIsResourceFile(t *testing.T) {
	cases := []struct {
		In  string
		Out bool
	}{
		{
			In:  "foo.png",
			Out: true,
		},
		{
			In:  "foo@ja.png",
			Out: true,
		},
		{
			In:  "foo-bar@ja.png",
			Out: true,
		},
		{
			In:  "foo.bar@ja.png",
			Out: true,
		},
		{
			In:  "foo.unknown",
			Out: false,
		},
		{
			In:  "日本語",
			Out: false,
		},
	}
	for _, c := range cases {
		got := IsResourceFile(c.In)
		want := c.Out
		if got != want {
			t.Errorf("IsResourceFile(%q): got: %t, want: %t", c.In, got, want)
		}
	}
}
