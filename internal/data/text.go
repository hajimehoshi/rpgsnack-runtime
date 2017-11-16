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

package data

import (
	languagepkg "golang.org/x/text/language"

	"github.com/google/uuid"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/sort"
)

type Texts struct {
	data      map[uuid.UUID]map[languagepkg.Tag]string
	languages []languagepkg.Tag
}

func (t *Texts) UnmarshalJSON(data []uint8) error {
	type TextData struct {
		Data map[string]string `json:data`
		// ignore "meta" key.
	}
	orig := map[uuid.UUID]TextData{}
	if err := unmarshalJSON(data, &orig); err != nil {
		return err
	}
	langs := map[languagepkg.Tag]struct{}{}
	t.languages = []languagepkg.Tag{}
	t.data = map[uuid.UUID]map[languagepkg.Tag]string{}
	for id, textData := range orig {
		t.data[id] = map[languagepkg.Tag]string{}
		for langStr, text := range textData.Data {
			lang, err := languagepkg.Parse(langStr)
			if err != nil {
				return err
			}
			if _, ok := langs[lang]; !ok {
				t.languages = append(t.languages, lang)
				langs[lang] = struct{}{}
			}
			t.data[id][lang] = text
		}
	}
	sort.Slice(t.languages, func(i, j int) bool {
		// English first
		l := t.languages
		if l[i] == languagepkg.English {
			return true
		}
		if l[j] == languagepkg.English {
			return false
		}
		return l[i].String() < l[j].String()
	})
	return nil
}

func (t *Texts) Languages() []languagepkg.Tag {
	return t.languages
}

func (t *Texts) Get(lang languagepkg.Tag, uuid uuid.UUID) string {
	return t.data[uuid][lang]
}
