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

// +build !js

package game

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/hajimehoshi/ebiten"
)

var (
	cpuProfile     = flag.String("cpuprofile", "", "write cpu profile to file")
	cpuProfileFile *os.File

	traceOut     = flag.String("trace", "", "write trace output to file")
	traceOutFile *os.File
)

func takeCPUProfileIfAvailable() {
	if ebiten.IsKeyPressed(ebiten.KeyP) {
		if *cpuProfile != "" && cpuProfileFile == nil {
			f, err := os.Create(*cpuProfile)
			if err != nil {
				panic(err)
			}
			cpuProfileFile = f
			pprof.StartCPUProfile(f)
			log.Print("Start CPU Profiling")
		}
		if *traceOut != "" && traceOutFile == nil {
			f, err := os.Create(*traceOut)
			if err != nil {
				panic(err)
			}
			traceOutFile = f
			trace.Start(f)
			log.Print("Start Tracing")
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		if cpuProfileFile != nil {
			pprof.StopCPUProfile()
			cpuProfileFile.Close()
			cpuProfileFile = nil
			log.Print("Stop CPU Profiling")
		}
		if traceOutFile != nil {
			trace.Stop()
			traceOutFile.Close()
			traceOutFile = nil
			log.Print("Stop Tracing")
		}
	}
}
