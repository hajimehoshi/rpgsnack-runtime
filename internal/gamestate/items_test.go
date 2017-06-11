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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
)

func TestItemDefaultIncludes(t *testing.T) {
	items := gamestate.Items{}
	if items.Includes(0) {
		t.Errorf("output: %b, want %b", true, false)
	}
}

func TestItemActiveItem(t *testing.T) {
	items := gamestate.Items{}
	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemDefaultCount(t *testing.T) {
	items := gamestate.Items{}
	if len(items.Items()) != 0 {
		t.Errorf("output: %d, want %d", len(items.Items()), 0)
	}
}

func TestItemDefaultAdd(t *testing.T) {
	items := gamestate.Items{}
	items.Add(1)
	if len(items.Items()) != 1 {
		t.Errorf("output: %d, want %d", len(items.Items()), 1)
	}
	if !items.Includes(1) {
		t.Errorf("output: %b, want %b", false, true)
	}

	i := items.Items()
	if i[0] != 1 {
		t.Errorf("output: %d, want %d", i[0], 1)
	}
}

func TestItemDefaultRemove(t *testing.T) {
	items := gamestate.Items{}
	items.Remove(0)
	if len(items.Items()) != 0 {
		t.Errorf("output: %d, want %d", len(items.Items()), 0)
	}
}

func TestItemDefaultActiveItem(t *testing.T) {
	items := gamestate.Items{}
	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemAddTwiceUnique(t *testing.T) {
	items := gamestate.Items{}
	items.Add(1)
	items.Add(2)
	if len(items.Items()) != 2 {
		t.Errorf("output: %d, want %d", len(items.Items()), 2)
	}
}

func TestItemAddTwiceDupe(t *testing.T) {
	items := gamestate.Items{}
	items.Add(1)
	items.Add(1)
	i := items.Items()
	if i[0] != 1 {
		t.Errorf("output: %d, want %d", i[0], 1)
	}
}

func TestItemRemove(t *testing.T) {
	items := gamestate.Items{}
	items.Add(1)
	items.Add(3)
	items.Add(2)
	items.Remove(1)
	if len(items.Items()) != 2 {
		t.Errorf("output: %d, want %d", len(items.Items()), 2)
	}

	i := items.Items()
	if i[0] != 3 {
		t.Errorf("output: %d, want %d", i[0], 3)
	}
	if i[1] != 2 {
		t.Errorf("output: %d, want %d", i[1], 2)
	}
}

func TestItemRemoveActiveItem(t *testing.T) {
	items := gamestate.Items{}
	items.Add(1)
	items.Activate(1)
	items.Remove(1)

	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemActivate(t *testing.T) {
	items := gamestate.Items{}
	items.Add(1)
	items.Activate(1)
	if items.ActiveItem() != 1 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 1)
	}
	items.Deactivate()
	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemMarshalAndUnmarshal(t *testing.T) {
	items := gamestate.NewItems([]int{1, 2, 3}, 2)
	out, err := items.MarshalJSON()
	if err != nil {
		t.Errorf("error %s", err)
	}

	newItems := gamestate.NewItems([]int{1}, 1)
	err = newItems.UnmarshalJSON(out)
	if err != nil {
		t.Errorf("error %s", err)
	}

	if newItems.ActiveItem() != 2 {
		t.Errorf("output: %d, want %d", newItems.ActiveItem(), 2)
	}
	if len(newItems.Items()) != 3 {
		t.Errorf("output: %d, want %d", len(newItems.Items()), 3)
	}
}
