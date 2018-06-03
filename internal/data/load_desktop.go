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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vmihailenco/msgpack"
)

var (
	purchasesPath = flag.String("purchases-json-path", filepath.Join(".", "purchases.json"), "purchases path")
	savePath      = flag.String("save-msgpack-path", filepath.Join(".", "save.msgpack"), "save path")
	languagePath  = flag.String("language-json-path", filepath.Join(".", "language.json"), "language path")
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

func loadAssets(projectionLocation string) ([]byte, error) {
	assets := map[string][]byte{}
	for _, dir := range assetDirs {
		images, err := ioutil.ReadDir(filepath.Join(projectionLocation, "assets", dir))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, i := range images {
			if strings.HasPrefix(i.Name(), ".") {
				continue
			}
			iPath := filepath.Join(projectionLocation, "assets", dir, i.Name())
			if isDir(iPath) {
				continue
			}
			b, err := ioutil.ReadFile(iPath)
			if err != nil {
				return nil, err
			}
			l := strings.Split(dir, string(filepath.Separator))
			l = append(l, i.Name())
			assets[path.Join(l...)] = b
		}
	}
	b, err := msgpack.Marshal(assets)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func isDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		fmt.Errorf("check isDir error: %s", err)
	}

	mode := f.Mode()
	if mode.IsDir() {
		return true
	}
	return false
}

func loadRawData(projectionLocation string, progressCh chan<- float64) (*rawData, error) {
	defer close(progressCh)

	project, err := ioutil.ReadFile(filepath.Join(projectionLocation, "project.msgpack"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var projectJSON []byte
	if project == nil {
		projectJSON, err = ioutil.ReadFile(filepath.Join(projectionLocation, "project.json"))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	progress, err := ioutil.ReadFile(*savePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		progress = nil
	}
	assets, err := loadAssets(projectionLocation)
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
		langData = nil
	}

	return &rawData{
		Project:     project,
		ProjectJSON: projectJSON,
		Assets:      [][]byte{assets},
		Progress:    progress,
		Purchases:   purchases,
		Language:    langData,
	}, nil
}

func SetData(project []byte, assets []byte, progress []byte, purchases []byte, language string) {
	// Not implemented
}
