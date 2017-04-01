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

package gamestate

import (
	"encoding/json"
)

type IAPProducts struct {
	purchased map[string]bool
}

type tmpIAPProducts struct {
	Purchased map[string]bool `json:"purchased"`
}

func (i *IAPProducts) MarshalJSON() ([]uint8, error) {
	tmp := &tmpIAPProducts{
		Purchased: i.purchased,
	}
	return json.Marshal(tmp)
}

func (i *IAPProducts) UnmarshalJSON(data []uint8) error {
	var tmp *tmpIAPProducts
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	i.purchased = tmp.Purchased
	return nil
}

func (i *IAPProducts) isPurhcased(key string) bool {
	return i.purchased[key]
}

func (i *IAPProducts) Purchase(key string) {
	i.purchased[key] = true
}
