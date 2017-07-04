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

package gamestate_test

import (
	"testing"

	"github.com/vmihailenco/msgpack"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
)

type pseudoRand struct {
	values []int
	index  int
}

func (p *pseudoRand) Intn(n int) int {
	v := p.values[p.index]
	p.index++
	return int(uint(v) % uint(n))
}

func TestRandomValue(t *testing.T) {
	values := []int{-1, 0, 3, 4}
	g := &Game{}
	g.SetRandomForTesting(&pseudoRand{values, 0})
	for range values {
		got := g.RandomValue(1, 4)
		if got <= 0 || got >= 4 {
			t.Errorf("RandomValue(1, 4) out of range: got: %v", got)
		}
	}
}

func TestMarshalGame(t *testing.T) {
	g := &Game{}
	b, err := msgpack.Marshal(g)
	if err != nil {
		t.Error(err)
	}
	var g2 *Game
	if err := msgpack.Unmarshal(b, &g2); err != nil {
		t.Error(err)
	}
}
