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

package lang

import (
	"golang.org/x/text/language"
)

var currentLang language.Tag

func Get() language.Tag {
	return currentLang
}

func Set(lang language.Tag) {
	currentLang = lang
}

func Normalize(lang language.Tag) language.Tag {
	base, _ := lang.Base()
	newLang, _ := language.Compose(base)
	if newLang == language.Chinese {
		// If the language is Chinese use zh-Hans or zh-Hant.
		s, _ := lang.Script()
		if s.String() != "Hans" && s.String() != "Hant" {
			// If the language is just "zh" or other Chinese, use Hans (simplified).
			// There is no strong reason why Hans is preferred.
			s = language.MustParseScript("Hans")
		}
		var err error
		newLang, err = language.Compose(base, s)
		if err != nil {
			panic(err)
		}
	}
	return newLang
}
