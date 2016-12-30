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
	"os"
	"runtime/pprof"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/game"
)

var theGame *game.Game

func update(screen *ebiten.Image) error {
	if err := theGame.Update(); err != nil {
		return err
	}
	if ebiten.IsRunningSlowly() {
		return nil
	}
	if err := theGame.Draw(screen); err != nil {
		return err
	}
	return nil
}

var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}
	g, err := game.New()
	if err != nil {
		panic(err)
	}
	theGame = g
	w, h := game.Size()
	title := game.Title()
	if err := ebiten.Run(update, w, h, game.Scale(), title); err != nil {
		panic(err)
	}
}
