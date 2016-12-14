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
	"encoding/json"

	"golang.org/x/text/language"
)

type Texts struct {
	data map[language.Tag]map[UUID]string
}

func (t *Texts) UnmarshalJSON(data []uint8) error {
	orig := map[string]map[UUID]string{}
	if err := json.Unmarshal(data, &orig); err != nil {
		return err
	}
	t.data = map[language.Tag]map[UUID]string{}
	for langStr, text := range orig {
		lang, err := language.Parse(langStr)
		if err != nil {
			return err
		}
		t.data[lang] = text
	}
	return nil
}

func (t *Texts) Get(lang language.Tag, uuid UUID) string {
	return t.data[lang][uuid]
}
