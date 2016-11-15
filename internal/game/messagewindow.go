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

package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

const windowMaxCount = 20

type messageWindow struct {
	content string
	count   int
}

func (m *messageWindow) show(message string) {
	m.content = message
	m.count = windowMaxCount
	task.Push(func() error {
		m.count--
		if m.count == windowMaxCount/2 {
			return task.Terminated
		}
		return nil
	})
	task.Push(func() error {
		if input.Triggered() {
			return task.Terminated
		}
		return nil
	})
	task.Push(func() error {
		m.count--
		if m.count == 0 {
			return task.Terminated
		}
		return nil
	})
}

func (m *messageWindow) draw(screen *ebiten.Image) error {
	if m.count == windowMaxCount/2 {
		if err := font.DrawText(screen, m.content, 0, 32, textScale, color.White); err != nil {
			return err
		}
	}
	return nil
}
