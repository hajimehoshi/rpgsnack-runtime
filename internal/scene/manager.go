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

package scene

import (
	"github.com/hajimehoshi/ebiten"
)

const (
	TileSize      = 16
	TileXNum      = 10
	TileYNum      = 10
	TileScale     = 3
	GameMarginTop = 2 * TileSize * TileScale
	TextScale     = 2
)

type scene interface {
	Update(manager *Manager) error
	Draw(screen *ebiten.Image) error
}

type Manager struct {
	width   int
	height  int
	current scene
	next    scene
}

func NewManager(width, height int, initScene scene) *Manager {
	return &Manager{
		width:   width,
		height:  height,
		current: initScene,
	}
}

func (m *Manager) Size() (int, int) {
	return m.width, m.height
}

func (m *Manager) Update() error {
	if m.next != nil {
		m.current = m.next
		m.next = nil
	}
	if err := m.current.Update(m); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) error {
	if err := m.current.Draw(screen); err != nil {
		return err
	}
	return nil
}

func (m *Manager) GoTo(next scene) {
	m.next = next
}
