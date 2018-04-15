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
	"github.com/vmihailenco/msgpack"
	languagepkg "golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/sort"
)

type Language languagepkg.Tag

func (l *Language) String() string {
	return (*languagepkg.Tag)(l).String()
}

func (l *Language) DecodeMsgpack(dec *msgpack.Decoder) error {
	str, err := dec.DecodeString()
	if err != nil {
		return err
	}
	lang, err := languagepkg.Parse(str)
	if err != nil {
		return err
	}
	*l = Language(lang)
	return nil
}

type Texts struct {
	data      map[UUID]map[Language]string
	languages []Language
}

func sortLanguages(languages []Language) {
	sort.Slice(languages, func(i, j int) bool {
		// English first
		l := languages
		if l[i] == Language(languagepkg.English) {
			return true
		}
		if l[j] == Language(languagepkg.English) {
			return false
		}
		return l[i].String() < l[j].String()
	})
}

func (t *Texts) DecodeMsgpack(dec *msgpack.Decoder) error {
	type TextData struct {
		Data map[Language]string `msgpack:"data"`
		// ignore "meta" key.
	}
	var data map[UUID]TextData
	if err := dec.Decode(&data); err != nil {
		return err
	}
	t.data = map[UUID]map[Language]string{}
	for uuid, textdata := range data {
		t.data[uuid] = textdata.Data
	}

	langs := map[Language]struct{}{}
	for _, d := range t.data {
		for l := range d {
			langs[l] = struct{}{}
		}
	}

	t.languages = []Language{}
	for l := range langs {
		t.languages = append(t.languages, l)
	}
	sortLanguages(t.languages)

	return nil
}

func (t *Texts) UnmarshalJSON(data []uint8) error {
	type TextData struct {
		Data map[string]string `json:"data"`
		// ignore "meta" key.
	}
	orig := map[UUID]TextData{}
	if err := unmarshalJSON(data, &orig); err != nil {
		return err
	}
	langs := map[Language]struct{}{}
	t.languages = []Language{}
	t.data = map[UUID]map[Language]string{}
	for id, textData := range orig {
		t.data[id] = map[Language]string{}
		for langStr, text := range textData.Data {
			lang, err := languagepkg.Parse(langStr)
			if err != nil {
				return err
			}
			l := Language(lang)
			if _, ok := langs[l]; !ok {
				t.languages = append(t.languages, l)
				langs[l] = struct{}{}
			}
			t.data[id][l] = text
		}
	}
	sortLanguages(t.languages)
	return nil
}

func (t *Texts) Languages() []languagepkg.Tag {
	ls := make([]languagepkg.Tag, len(t.languages))
	for i, l := range t.languages {
		ls[i] = languagepkg.Tag(l)
	}
	return ls
}

func (t *Texts) Get(lang languagepkg.Tag, uuid UUID) string {
	return t.data[uuid][Language(lang)]
}
