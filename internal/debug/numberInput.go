// Copyright 2019 The RPGSnack Authors
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

package debug

import (
	"fmt"
	"strconv"

	"github.com/hajimehoshi/ebiten"
)

type NumberInput struct {
	value           int
	editingText     string
	waitForNextChar bool
}

func NewNumberInput() *NumberInput {
	return &NumberInput{
		value:       0,
		editingText: "0",
	}
}

func (n *NumberInput) SetValue(value int) {
	n.value = value
	n.editingText = strconv.Itoa(value)
	n.waitForNextChar = false
}

func (n *NumberInput) Value() int {
	return n.value
}

func (n *NumberInput) Text() string {
	return n.editingText
}

func (n *NumberInput) Update() {
	chars := ebiten.InputChars()
	for c := range chars {
		if c == '-' {
			n.editingText = "-"
			n.waitForNextChar = true
		}

		if '0' <= c && c <= '9' {
			v := strconv.Itoa(int(c - '0'))
			if n.waitForNextChar {
				n.editingText += v
			} else {
				n.editingText = v
				n.waitForNextChar = true
			}
		}
	}

	i, err := strconv.Atoi(n.editingText)
	if err != nil {
		panic(fmt.Sprintf("failed to convert editingText %s : %s", n.editingText, err))
	}
	n.value = i
}
