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

package data_test

import (
	"encoding/json"
	"testing"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

func TestPurchases(t *testing.T) {
	testString := "[\"itemA\", \"itemB\"]"
	testData := []byte(testString)

	data.UpdatePurchases(testData)

	var purchases []string
	json.Unmarshal(data.Purchases(), &purchases)

	if purchases[0] != "itemA" {
		t.Errorf("itemA not found")
	}

	if purchases[1] != "itemB" {
		t.Errorf("itemB not found")
	}
}
