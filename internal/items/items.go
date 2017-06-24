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
	"encoding/json"
)

type Items struct {
	items      []int
	activeItem int
}

type tmpItems struct {
	Items      []int `json:"items"`
	ActiveItem int   `json:"activeItem"`
}

func NewItems(items []int, activeItem int) *Items {
	return &Items{
		items:      items,
		activeItem: activeItem,
	}
}

func (i *Items) MarshalJSON() ([]uint8, error) {
	tmp := &tmpItems{
		Items:      i.items,
		ActiveItem: i.activeItem,
	}
	return json.Marshal(tmp)
}

func (i *Items) UnmarshalJSON(data []uint8) error {
	var tmp *tmpItems
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	i.items = tmp.Items
	i.activeItem = tmp.ActiveItem
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

func (i *Items) Items() []int {
	return i.items
}

func (i *Items) Add(id int) {
	if i.items == nil {
		i.items = []int{}
	}

	if i.index(id) < 0 {
		i.items = append(i.items, id)
	}
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
}

func (i *Items) Activate(id int) {
	i.activeItem = id
}

func (i *Items) Deactivate() {
	i.activeItem = 0
}
