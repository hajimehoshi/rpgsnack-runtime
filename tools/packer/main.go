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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"
)

func showUsage() {
	fmt.Fprintf(os.Stderr, "packer PROJECT_PATH\n")
}

var resourceRe = regexp.MustCompile(`^[a-zA-Z1-9_-]+(\.(mp3|ogg|png|wav))?$`)

func isResourceFile(path string) bool {
	return resourceRe.MatchString(filepath.Base(path))
}

func run(in, out string) error {
	resources := []string{}
	if err := filepath.Walk(filepath.Join(in, "assets"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isResourceFile(path) {
			resources = append(resources, path)
		}
		return nil
	}); err != nil {
		return err
	}

	m := map[string][]uint8{}
	for _, r := range resources {
		rel, err := filepath.Rel(filepath.Join(in, "assets"), r)
		if err != nil {
			return err
		}
		key := strings.Join(strings.Split(rel, string(filepath.Separator)), "/")
		data, err := ioutil.ReadFile(r)
		if err != nil {
			return err
		}
		m[key] = data
	}
	b, err := msgpack.Marshal(m)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(out, b, 0644); err != nil {
		return err
	}

	return nil
}

func main() {
	in := flag.String("in", "", "input project path")
	out := flag.String("out", "", "output msgpack path")
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		os.Exit(1)
	}
	if err := run(*in, *out); err != nil {
		panic(err)
	}
}
