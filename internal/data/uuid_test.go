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
	"testing"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

func TestUUID(t *testing.T) {
	cases := []struct {
		str   string
		valid bool
	}{
		{
			str:   "7eb9c8dc-c9a6-4ae2-8ddd-45b9a50baeb4",
			valid: true,
		},
		{
			// Nil
			str:   "00000000-0000-0000-0000-000000000000",
			valid: true,
		},
		{
			// Not version 4
			str:   "dfe7edf6-c217-11e6-a4a6-cec0c932ce01",
			valid: false,
		},
		{
			// Invalid bits in clock_seq_hi_and_reserved part
			// See RFC 4122.
			str:   "7eb9c8dc-c9a6-4ae2-0ddd-45b9a50baeb4",
			valid: true,
		},
	}
	for _, c := range cases {
		var id UUID
		if err := id.UnmarshalText([]uint8(c.str)); err != nil && c.valid {
			t.Fatal(err)
		} else if err == nil && !c.valid {
			t.Errorf("UnmarshalText with %s should return error but not", c.str)
		}
		out, err := id.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != c.str {
			t.Errorf("output: %s, want %s", string(out), c.str)
		}
	}
}
