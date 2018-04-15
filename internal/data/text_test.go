// Copyright 2018 Hajime Hoshi
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
	"testing"

	"github.com/vmihailenco/msgpack"
	"golang.org/x/text/language"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

func TestTexts(t *testing.T) {
	uuid1 := NewUUID()
	uuid2 := NewUUID()
	tmp := map[UUID]map[string]map[string]string{
		uuid1: {
			"data": {
				"en": "Hello",
				"fr": "Bonjour",
				"ja": "こんにちは",
			},
		},
		uuid2: {
			"data": {
				"en": "Good Bye",
				"ja": "さようなら",
			},
		},
	}
	encoded, err := msgpack.Marshal(tmp)
	if err != nil {
		t.Fatal(err)
	}
	var texts *Texts
	if err := msgpack.Unmarshal(encoded, &texts); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		lang language.Tag
		uuid UUID
		text string
	}{
		{
			lang: language.English,
			uuid: uuid1,
			text: "Hello",
		},
		{
			lang: language.Japanese,
			uuid: uuid2,
			text: "さようなら",
		},
	}
	for _, c := range cases {
		got := texts.Get(c.lang, c.uuid)
		want := c.text
		if got != want {
			t.Errorf("texts.Get(%v, %v): got %s, want: %s", c.lang, c.uuid, got, want)
		}
	}

	if texts.Languages()[0] != language.English {
		t.Errorf("texts.Languages[0]: got: %v, want: %v", texts.Languages()[0], language.English)
	}
	if texts.Languages()[1] != language.French {
		t.Errorf("texts.Languages[1]: got: %v, want: %v", texts.Languages()[1], language.French)
	}
	if texts.Languages()[2] != language.Japanese {
		t.Errorf("texts.Languages[2]: got: %v, want: %v", texts.Languages()[2], language.Japanese)
	}
}
