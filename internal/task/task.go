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

type task func() error

var theTasks = &tasks{}

type tasks struct {
	tasks []task
}

func (t *tasks) push(task task)  {
	t.tasks = append(t.tasks, task)
}

func (t *tasks) update() (bool, error) {
	if len(t.tasks) == 0 {
		return false, nil
	}
	task := t.tasks[0]
	if err := task(); err == Terminated {
		t.tasks = t.tasks[1:]
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func Push(task task) {
	theTasks.push(task)
}

func Update() (bool, error) {
	return theTasks.update()
}
