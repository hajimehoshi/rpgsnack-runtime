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
	"sort"

	"golang.org/x/text/language"
)

type languagesByAlphabet []language.Tag

func (l languagesByAlphabet) Len() int {
	return len(l)
}

func (l languagesByAlphabet) Less(i, j int) bool {
	// English first
	if l[i] == language.English {
		return true
	}
	if l[j] == language.English {
		return false
	}
	return l[i].String() < l[j].String()
}

func (l languagesByAlphabet) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type Texts struct {
	data      map[language.Tag]map[UUID]string
	languages []language.Tag
}

func (t *Texts) UnmarshalJSON(data []uint8) error {
	orig := map[string]map[UUID]string{}
	if err := unmarshalJSON(data, &orig); err != nil {
		return err
	}
	langs := map[language.Tag]struct{}{}
	t.languages = []language.Tag{}
	t.data = map[language.Tag]map[UUID]string{}
	for langStr, text := range orig {
		lang, err := language.Parse(langStr)
		if err != nil {
			return err
		}
		if _, ok := langs[lang]; !ok {
			t.languages = append(t.languages, lang)
			langs[lang] = struct{}{}
		}
		t.data[lang] = text
	}
	sort.Sort(languagesByAlphabet(t.languages))
	return nil
}

func (t *Texts) Languages() []language.Tag {
	return t.languages
}

func (t *Texts) Get(lang language.Tag, uuid UUID) string {
	return t.data[lang][uuid]
}
