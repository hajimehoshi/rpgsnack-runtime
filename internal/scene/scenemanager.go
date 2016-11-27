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

	"github.com/hajimehoshi/tsugunai/internal/task"
)

const (
	TileSize         = 16
	TileXNum         = 10
	TileYNum         = 10
	TileScale        = 3
	GameMarginX      = 0
	GameMarginTop    = 2 * TileSize * TileScale
	GameMarginBottom = 3 * TileSize * TileScale
	TextScale        = 2
)

type scene interface {
	Update(subTasksUpdated bool, taskLine *task.TaskLine, sceneManager *SceneManager) error
	Draw(screen *ebiten.Image) error
}

type SceneManager struct {
	current  scene
	next     scene
	taskLine *task.TaskLine
}

func NewSceneManager(initScene scene) *SceneManager {
	return &SceneManager{
		current:  initScene,
		taskLine: &task.TaskLine{},
	}
}

func (s *SceneManager) Update() error {
	updated, err := s.taskLine.Update()
	if err != nil {
		return err
	}
	if !updated {
		if s.next != nil {
			s.current = s.next
			s.next = nil
		}
	}
	if err := s.current.Update(updated, s.taskLine, s); err != nil {
		return err
	}
	return nil
}

func (s *SceneManager) Draw(screen *ebiten.Image) error {
	if err := s.current.Draw(screen); err != nil {
		return err
	}
	return nil
}

func (s *SceneManager) GoTo(next scene) {
	s.next = next
}
