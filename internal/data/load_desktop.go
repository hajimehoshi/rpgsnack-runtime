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

// +build !android
// +build !ios
// +build !js

package data

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	dataPath      = flag.String("data", "./data.json", "data path")
	resourcesPath = flag.String("resources", "./internal/assets", "resources directory path")
	purchasesPath = flag.String("purchases", "./purchases.json", "purchases path")
	savePath      = flag.String("save", "./save.json", "save path")
	languagePath  = flag.String("language", "./language.json", "language path")
)

func PurchasesPath() string {
	return *purchasesPath
}

func LanguagePath() string {
	return *languagePath
}

func SavePath() string {
	return *savePath
}

func loadResources() ([]uint8, error) {
	resources := map[string][]uint8{}
	for _, dir := range []string{"images"} {
		images, err := ioutil.ReadDir(filepath.Join(*resourcesPath, dir))
		if err != nil {
			return nil, err
		}
		for _, i := range images {
			if strings.HasPrefix(i.Name(), ".") {
				continue
			}
			k := filepath.Join(dir, i.Name())
			b, err := ioutil.ReadFile(filepath.Join(*resourcesPath, k))
			if err != nil {
				return nil, err
			}
			resources[k] = b
		}
	}
	b, err := msgpack.Marshal(resources)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func loadRawData() (*rawData, error) {
	game, err := ioutil.ReadFile(*dataPath)
	if err != nil {
		return nil, err
	}
	progress, err := ioutil.ReadFile(*savePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		progress = nil
	}
	resources, err := loadResources()
	if err != nil {
		return nil, err
	}
	purchases, err := ioutil.ReadFile(*purchasesPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		purchases = nil
	}

	langData, err := ioutil.ReadFile(*languagePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		langData = []uint8(`"en"`)
	}

	return &rawData{
		Game:      game,
		Resources: resources,
		Progress:  progress,
		Purchases: purchases,
		Language:  langData,
	}, nil
}
