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

package gamestate

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Game struct {
	variables *Variables
	screen    *Screen
	windows   *window.Windows
}

func NewGame() *Game {
	return &Game{
		variables: &Variables{},
		screen:    &Screen{},
		windows:   &window.Windows{},
	}
}

func (g *Game) Variables() *Variables {
	return g.variables
}

func (g *Game) Screen() *Screen {
	return g.screen
}

func (g *Game) Windows() *window.Windows {
	return g.windows
}

var reMessage = regexp.MustCompile(`\\([a-zA-Z])\[(\d+)\]`)

func (g *Game) ParseMessageSyntax(str string) string {
	return reMessage.ReplaceAllStringFunc(str, func(part string) string {
		name := strings.ToLower(part[1:2])
		id, err := strconv.Atoi(part[3 : len(part)-1])
		if err != nil {
			panic(fmt.Sprintf("not reach: %s", err))
		}
		switch name {
		case "v":
			return strconv.Itoa(g.Variables().VariableValue(id))
		}
		return str
	})
}
