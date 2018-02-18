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

type fontLoadingResult struct {
	body []byte
	err  error
}

var (
	scCh = make(chan fontLoadingResult)
	tcCh = make(chan fontLoadingResult)
)

func init() {
	const (
		scURL = "https://rpgsnack-e85d3.appspot.com/static/fonts/scregular.subset.ttf"
		tcURL = "https://rpgsnack-e85d3.appspot.com/static/fonts/tcregular.subset.ttf"
	)

	go func() {
		resp, err := http.Get(scURL)
		if err != nil {
			scCh <- fontLoadingResult{nil, err}
			return
		}
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			scCh <- fontLoadingResult{nil, err}
			return
		}
		scCh <- fontLoadingResult{bs, nil}
	}()
	go func() {
		resp, err := http.Get(tcURL)
		if err != nil {
			tcCh <- fontLoadingResult{nil, err}
			return
		}
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tcCh <- fontLoadingResult{nil, err}
			return
		}
		tcCh <- fontLoadingResult{bs, nil}
	}()
}

func getSCTTF() ([]byte, error) {
	r := <-scCh
	return r.body, r.err
}

func getTCTTF() ([]byte, error) {
	r := <-tcCh
	return r.body, r.err
}
