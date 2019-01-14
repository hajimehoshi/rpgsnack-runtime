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
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

var (
	cpuProfile     = flag.String("cpuprofile", "", "write cpu profile to file")
	cpuProfileFile *os.File

	memProfile = flag.String("memprofile", "", "write memory profile to file")

	traceOut     = flag.String("trace", "", "write trace output to file")
	traceOutFile *os.File
)

func takeCPUProfileIfAvailable() {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
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
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if cpuProfileFile != nil {
			pprof.StopCPUProfile()
			cpuProfileFile.Close()
			cpuProfileFile = nil
			log.Print("Stop CPU Profiling")
		}
		if *memProfile != "" {
			f, err := os.Create(*memProfile)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			runtime.GC()
			if err := pprof.WriteHeapProfile(f); err != nil {
				panic(err)
			}
			log.Print("Memory Dumped")
		}
		if traceOutFile != nil {
			trace.Stop()
			traceOutFile.Close()
			traceOutFile = nil
			log.Print("Stop Tracing")
		}
	}
}
