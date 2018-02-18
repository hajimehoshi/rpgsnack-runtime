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

// +build js

package font

import (
	"io/ioutil"
	"net/http"
)

const (
	scURL = "https://rpgsnack-e85d3.appspot.com/static/fonts/scregular.subset.ttf"
	tcURL = "https://rpgsnack-e85d3.appspot.com/static/fonts/tcregular.subset.ttf"
)

func getSCTTF() ([]byte, error) {
	resp, err := http.Get(scURL)
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func getTCTTF() ([]byte, error) {
	resp, err := http.Get(tcURL)
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
