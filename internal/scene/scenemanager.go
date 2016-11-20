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

type scene interface {
	Update(sceneManager *SceneManager) error
	Draw(screen *ebiten.Image) error
}

type SceneManager struct {
	current scene
	next    scene
}

func NewSceneManager(initScene scene) *SceneManager {
	return &SceneManager{
		current: initScene,
	}
}

func (s *SceneManager) Update() error {
	updated, err := task.Update()
	if err != nil {
		return err
	}
	if updated {
		return nil
	}
	if s.next != nil {
		s.current = s.next
		s.next = nil
	}
	if err := s.current.Update(s); err != nil {
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
