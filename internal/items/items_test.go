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

package items_test

import (
	"testing"

	"github.com/vmihailenco/msgpack"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/items"
)

func TestItemDefaultIncludes(t *testing.T) {
	items := Items{}
	if items.Includes(0) {
		t.Errorf("output: %t, want %t", true, false)
	}
}

func TestItemActiveItem(t *testing.T) {
	items := Items{}
	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemDefaultCount(t *testing.T) {
	items := Items{}
	if items.ItemNum() != 0 {
		t.Errorf("output: %d, want %d", items.ItemNum(), 0)
	}
}

func TestItemDefaultAdd(t *testing.T) {
	items := Items{}
	items.Add(1)
	if items.ItemNum() != 1 {
		t.Errorf("output: %d, want %d", items.ItemNum(), 1)
	}
	if !items.Includes(1) {
		t.Errorf("output: %t, want %t", false, true)
	}

	if items.ItemIDAt(0) != 1 {
		t.Errorf("output: %d, want %d", items.ItemIDAt(0), 1)
	}
}

func TestItemDefaultRemove(t *testing.T) {
	items := Items{}
	items.Remove(0)
	if items.ItemNum() != 0 {
		t.Errorf("output: %d, want %d", items.ItemNum(), 0)
	}
}

func TestItemDefaultActiveItem(t *testing.T) {
	items := Items{}
	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemAddTwiceUnique(t *testing.T) {
	items := Items{}
	items.Add(1)
	items.Add(2)
	if items.ItemNum() != 2 {
		t.Errorf("output: %d, want %d", items.ItemNum(), 2)
	}
}

func TestItemAddTwiceDupe(t *testing.T) {
	items := Items{}
	items.Add(1)
	items.Add(1)
	if items.ItemIDAt(0) != 1 {
		t.Errorf("output: %d, want %d", items.ItemIDAt(0), 1)
	}
}

func TestItemRemove(t *testing.T) {
	items := Items{}
	items.Add(1)
	items.Add(3)
	items.Add(2)
	items.Remove(1)
	if items.ItemNum() != 2 {
		t.Errorf("output: %d, want %d", items.ItemNum(), 2)
	}

	if items.ItemIDAt(0) != 3 {
		t.Errorf("output: %d, want %d", items.ItemIDAt(0), 3)
	}
	if items.ItemIDAt(1) != 2 {
		t.Errorf("output: %d, want %d", items.ItemIDAt(1), 2)
	}
}

func TestItemRemoveActiveItem(t *testing.T) {
	items := Items{}
	items.Add(1)
	items.Activate(1)
	items.Remove(1)

	if items.ActiveItem() != 0 {
		t.Errorf("output: %d, want %d", items.ActiveItem(), 0)
	}
}

func TestItemActivate(t *testing.T) {
	items := Items{}
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
	items := NewItems([]int{1, 2, 3}, 2)
	out, err := msgpack.Marshal(items)
	if err != nil {
		t.Errorf("error %s", err)
	}

	var newItems *Items
	if err = msgpack.Unmarshal(out, &newItems); err != nil {
		t.Errorf("error %s", err)
	}

	if newItems.ActiveItem() != 2 {
		t.Errorf("output: %d, want %d", newItems.ActiveItem(), 2)
	}
	if newItems.ItemNum() != 3 {
		t.Errorf("output: %d, want %d", newItems.ItemNum(), 3)
	}
}
