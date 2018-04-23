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

// +build !js

package font

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"github.com/hajimehoshi/chinesegamefonts/scregular"
	"github.com/hajimehoshi/chinesegamefonts/tcregular"
)

func startLoadingChineseFonts() {
	// Do nothing
}

func getSCTTF() ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(scregular.CompressedTTF))
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func getTCTTF() ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(tcregular.CompressedTTF))
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
