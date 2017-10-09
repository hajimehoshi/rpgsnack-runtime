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

// +build ignore

package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var jsDir = ""

func init() {
	var err error
	jsDir, err = ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
}

func createJSIfNeeded() (string, error) {
	const target = "github.com/hajimehoshi/rpgsnack-runtime"

	out := filepath.Join(jsDir, "main.js")
	stat, err := os.Stat(out)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if (err != nil && os.IsNotExist(err)) || time.Now().Sub(stat.ModTime()) > 5*time.Second {
		out, err := exec.Command("gopherjs", "build", "-o", out, target).CombinedOutput()
		if err != nil {
			if string(out) == "" {
				log.Print("gopherjs command execution failed: gopherjs not command?")
			} else {
				log.Printf("gopherjs command execution failed: %s", string(out))
			}
			return "", errors.New(string(out))
		}
	}
	return out, nil
}

func serveMainJS(w http.ResponseWriter, r *http.Request) {
	out, err := createJSIfNeeded()
	if err != nil {
		t := template.JSEscapeString(template.HTMLEscapeString(err.Error()))
		js := `
window.addEventListener('load', () => {
  document.body.innerHTML="<pre style='white-space: pre-wrap;'><code>` + t + `</code></pre>";
}`
		w.Header().Set("Content-Type", "text/javascript")
		io.WriteString(w, js)
		return
	}

	f, err := os.Open(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/javascript")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}

func handler(w http.ResponseWriter, r *http.Request) {
	const indexHtml = `<!DOCTYPE html>
<script src="main.js"></script>
`

	if strings.HasSuffix(r.URL.Path, "/main.js") {
		serveMainJS(w, r)
		return
	}

	if strings.HasSuffix(r.URL.Path, "/main.js.map") {
		http.NotFound(w, r)
		return
	}

	if strings.HasSuffix(r.URL.Path, "/favicon.ico") {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/web/") {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, indexHtml)
		return
	}

	http.NotFound(w, r)
}

var (
	host = flag.String("host", "", "host name")
	port = flag.Int("port", 8000, "port number")
)

func init() {
	flag.Parse()
}

func main() {
	http.HandleFunc("/", handler)
	if *host == "" {
		fmt.Printf("http://127.0.0.1:%d/\n", *port)
	} else {
		fmt.Printf("http://%s:%d/\n", *host, *port)
	}
	log.Fatal(http.ListenAndServe(*host+":"+strconv.Itoa(*port), nil))
}
