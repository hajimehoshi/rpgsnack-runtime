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

package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	basepath = flag.String("basepath", ".", "base filepath")
	port     = flag.Int("port", 8000, "port number")
)

type ManifestBody struct {
	Manifest map[string][]string `json:"manifest"`
}

type ManifestResponse struct {
	Body *ManifestBody `json:"body"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// Serve regular files
	if r.URL.Path != "/" {
		// Don't use http.ServeFile due to directory traversal attack.
		http.FileServer(http.Dir(*basepath)).ServeHTTP(w, r)
		return
	}

	// Serve the manifest file
	m := map[string][]string{}
	if err := filepath.Walk(*basepath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() != "." && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		relpath, err := filepath.Rel(*basepath, path)
		if err != nil {
			return err
		}
		m["http://"+r.Host+"/"+filepath.ToSlash(relpath)] = []string{filepath.ToSlash(relpath)}
		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/json")
	if err := json.NewEncoder(w).Encode(&ManifestResponse{
		Body: &ManifestBody{
			Manifest: m,
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		panic(err)
	}
}
