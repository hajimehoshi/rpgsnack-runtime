// Copyright 2016 Hajime Hoshi
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
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/game"
)

var (
	cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")
	screenSize = flag.String("screensize", "480x720", "screen size like 480x720")
)

func main() {
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	sp := strings.Split(*screenSize, "x")
	sw, err := strconv.Atoi(sp[0])
	if err != nil {
		log.Fatal(err)
	}
	sh, err := strconv.Atoi(sp[1])
	if err != nil {
		log.Fatal(err)
	}
	g, err := game.NewWithDefaultRequester(sw, sh)
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.Run(g.Update, sw, sh, game.Scale(), game.Title()); err != nil {
		log.Fatal(err)
	}
}
