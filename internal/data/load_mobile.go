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

// +build android ios
// +build !gomobilebuild
// +build !js

package data

import (
	"encoding/json"
)

var (
	dataCh = make(chan *rawData, 1)
)

func PurchasesPath() string {
	return ""
}

func LanguagePath() string {
	return ""
}

func SavePath() string {
	return ""
}

func PermanentPath() string {
	return ""
}

func CreditsPath() string {
	return ""
}

func loadRawData(projectionLocation string, progress chan<- float64) (*rawData, error) {
	defer close(progress)

	return <-dataCh, nil
}

func SetData(project []byte, assets [][]byte, progress []byte, permanent []byte, purchases []byte, language string) {
	l, err := json.Marshal(language)
	if err != nil {
		panic(err)
	}
	dataCh <- &rawData{
		Project:     nil, // TODO: Implement msgpack version
		ProjectJSON: project,
		Assets:      assets,
		Progress:    progress,
		Permanent:   permanent,
		Purchases:   purchases,
		Language:    l,
	}
}
