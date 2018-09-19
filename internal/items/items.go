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

package items

import (
	"fmt"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type Items struct {
	items       []int
	activeItem  int
	combineItem int
	eventItem   int
	activeGroup int
	dataItems   []*data.Item // Do not save
	activeItems []*data.Item // Do not save
}

func NewItems(items []int, activeItem int) *Items {
	return &Items{
		items:       items,
		activeItem:  activeItem,
		activeGroup: 0,
		dataItems:   []*data.Item{},
	}
}

func (i *Items) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()
	e.EncodeString("items")
	e.BeginArray()
	for _, item := range i.items {
		e.EncodeInt(item)
	}
	e.EndArray()
	e.EncodeString("activeItem")
	e.EncodeInt(i.activeItem)
	e.EncodeString("combineItem")
	e.EncodeInt(i.combineItem)
	e.EncodeString("eventItem")
	e.EncodeInt(i.eventItem)
	e.EncodeString("activeGroup")
	e.EncodeInt(i.activeGroup)
	e.EndMap()
	return e.Flush()
}

func (i *Items) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for j := 0; j < n; j++ {
		switch d.DecodeString() {
		case "items":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				i.items = make([]int, n)
				for j := 0; j < n; j++ {
					i.items[j] = d.DecodeInt()
				}
			}
		case "activeItem":
			i.activeItem = d.DecodeInt()
		case "combineItem":
			i.combineItem = d.DecodeInt()
		case "eventItem":
			i.eventItem = d.DecodeInt()
		case "activeGroup":
			i.activeGroup = d.DecodeInt()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("items: Items.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (i *Items) index(id int) int {
	if i.items == nil {
		return -1
	}

	for index, itemId := range i.items {
		if itemId == id {
			return index
		}
	}
	return -1
}

func (i *Items) Includes(id int) bool {
	return i.index(id) >= 0
}

func (i *Items) ActiveItem() int {
	return i.activeItem
}

func (i *Items) EventItem() int {
	return i.eventItem
}

func (i *Items) SetEventItem(id int) {
	i.eventItem = id
}

func (i *Items) Items() []*data.Item {
	if i.activeItems == nil {
		idToItem := map[int]*data.Item{}
		for _, i := range i.dataItems {
			idToItem[i.ID] = i
		}
		is := []*data.Item{}
		for _, id := range i.items {
			item := idToItem[id]
			if item.Group == i.activeGroup {
				is = append(is, item)
			}
		}
		i.activeItems = is
	}
	return i.activeItems
}

func (i *Items) ItemIDAt(index int) int {
	return i.Items()[index].ID
}

func (i *Items) ItemNum() int {
	return len(i.Items())
}

func (i *Items) Add(id int) {
	if i.items == nil {
		i.items = []int{}
	}

	if i.index(id) < 0 {
		i.items = append(i.items, id)
	}
	i.activeItems = nil
}

func (i *Items) InsertBefore(targetItemID int, insertItemID int) {
	index := i.index(targetItemID)
	// if the targetItem does not exist, fail this ops
	if index < 0 {
		return
	}

	// Only insert the item if it does not exist
	if i.index(insertItemID) < 0 {
		i.items = append(i.items, 0)
		copy(i.items[index+1:], i.items[index:])
		i.items[index] = insertItemID
	}
	i.activeItems = nil
}

func (i *Items) Remove(id int) {
	if i.items == nil {
		i.items = []int{}
	}

	index := i.index(id)
	if index >= 0 {
		i.items = append(i.items[:index], i.items[(index+1):]...)
	}

	if id == i.activeItem {
		i.activeItem = 0
	}
	i.activeItems = nil
}

func (i *Items) Activate(id int) {
	i.activeItem = id
}

func (i *Items) CombineItem() int {
	return i.combineItem
}

func (i *Items) SetCombineItem(id int) {
	i.combineItem = id
}

func (i *Items) Deactivate() {
	i.activeItem = 0
}

func (i *Items) SetDataItems(dataItems []*data.Item) {
	i.dataItems = dataItems
}

func (i *Items) SetActiveItemGroup(group int) {
	if i.activeGroup != group {
		i.activeItem = 0
		i.activeGroup = group
		i.activeItems = nil
	}
}
