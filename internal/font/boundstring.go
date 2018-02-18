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

package font

import (
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type boundStringCacheEntry struct {
	bounds  *fixed.Rectangle26_6
	advance fixed.Int26_6
}

var boundStringCache = map[font.Face]map[string]*boundStringCacheEntry{}

func boundString(face font.Face, str string) (*fixed.Rectangle26_6, fixed.Int26_6) {
	m, ok := boundStringCache[face]
	if !ok {
		m = map[string]*boundStringCacheEntry{}
		boundStringCache[face] = m
	}

	entry, ok := m[str]
	if !ok {
		// Delete all entries if the capacity exceeds the limit.
		if len(m) >= 256 {
			for k := range m {
				delete(m, k)
			}
		}

		b, a := font.BoundString(face, str)
		entry = &boundStringCacheEntry{
			bounds:  &b,
			advance: a,
		}
		m[str] = entry
	}

	return entry.bounds, entry.advance
}
