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
	"github.com/hajimehoshi/ebiten"
)

type scene interface {
	Update(sceneManager *sceneManager) error
	Draw(screen *ebiten.Image) error
}

type sceneManager struct {
	current scene
	next    scene
	tasks   []task
}

func newSceneManager(initScene scene) *sceneManager {
	return &sceneManager{
		current: initScene,
	}
}

func (s *sceneManager) pushTask(task task) {
	s.tasks = append(s.tasks, task)
}

func (s *sceneManager) Update() error {
	if len(s.tasks) > 0 {
		t := s.tasks[0]
		if err := t(); err == taskTerminated {
			s.tasks = s.tasks[1:]
		} else if err != nil {
			return err
		}
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

func (s *sceneManager) Draw(screen *ebiten.Image) error {
	if err := s.current.Draw(screen); err != nil {
		return err
	}
	return nil
}

func (s *sceneManager) GoTo(next scene) {
	s.next = next
}
