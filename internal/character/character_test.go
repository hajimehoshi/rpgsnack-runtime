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

package character_test

import (
	"testing"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/character"
)

func TestCharacterSize(t *testing.T) {
	character := Character{}
	character.SetImage("test_24_32")
	w, h := character.Size()
	if w != 24 || h != 32 {
		t.Errorf("output: %dx%d, want 24x32", w, h)
	}
}

func TestCharacterSizeInvalid(t *testing.T) {
	character := Character{}
	character.SetImage("test")
	w, h := character.Size()
	if w != 0 || h != 0 {
		t.Errorf("output: %dx%d, want 0x0", w, h)
	}
}

func TestCharacterFrameCount(t *testing.T) {
	character := Character{}
	character.SetImage("test_24_32")
	character.SetSizeForTesting(48, 128)

	frameCount := character.FrameCount()
	if frameCount != 2 {
		t.Errorf("output: %d, want %d", frameCount, 2)
	}

	frameCount = character.FrameCount()
	if frameCount != 2 {
		t.Errorf("output: %d, want %d", frameCount, 2)
	}
}

func TestCharacterDirCount(t *testing.T) {
	character := Character{}
	character.SetImage("test_24_32")
	character.SetSizeForTesting(48, 128)

	frameCount := character.DirCount()
	if frameCount != 4 {
		t.Errorf("output: %d, want %d", frameCount, 4)
	}

	frameCount = character.DirCount()
	if frameCount != 4 {
		t.Errorf("output: %d, want %d", frameCount, 4)
	}
}
