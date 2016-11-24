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

package task

import (
	"errors"
)

var Terminated = errors.New("task terminated")

type Task interface {
	Update() error
}

func Sleep(frames int) Task {
	c := frames
	return taskFunc(func() error {
		c--
		if c == 0 {
			return Terminated
		}
		return nil
	})
}

func Sub(f func(sub *TaskLine) error) Task {
	sub := &TaskLine{}
	terminated := false
	return taskFunc(func() error {
		if updated, err := sub.Update(); err != nil {
			return err
		} else if updated {
			return nil
		}
		if terminated {
			return Terminated
		}
		if err := f(sub); err == Terminated {
			terminated = true
			if len(sub.tasks) == 0 {
				return Terminated
			}
		} else if err != nil {
			return err
		}
		return nil
	})
}

type taskFunc func() error

func (t taskFunc) Update() error {
	return t()
}

type TaskLine struct {
	tasks []Task
}

func (t *TaskLine) Push(task Task) {
	t.tasks = append(t.tasks, task)
}

func (t *TaskLine) PushFunc(task func() error) {
	t.tasks = append(t.tasks, taskFunc(task))
}

func (t *TaskLine) Update() (bool, error) {
	if len(t.tasks) == 0 {
		return false, nil
	}
	task := t.tasks[0]
	if err := task.Update(); err == Terminated {
		t.tasks = t.tasks[1:]
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (t *TaskLine) ToTask() Task {
	return Parallel(t)
}

func Parallel(taskLines ...*TaskLine) Task {
	return taskFunc(func() error {
		done := []int{}
		for i, t := range taskLines {
			if t == nil {
				continue
			}
			if updated, err := t.Update(); err != nil {
				return err
			} else if !updated {
				done = append(done, i)
			}
		}
		for _, i := range done {
			taskLines[i] = nil
		}
		for _, t := range taskLines {
			if t != nil {
				return nil
			}
		}
		return Terminated
	})
}
