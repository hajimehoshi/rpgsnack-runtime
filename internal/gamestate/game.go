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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Game struct {
	variables       *Variables
	screen          *Screen
	windows         *window.Windows
	player          *character.Player
	mapID           int
	roomID          int
	events          []*character.Event
	continuingEvent *character.Event // TODO: This should be an Interpreter?
}

func NewGame() (*Game, error) {
	pos := data.Current().System.InitialPosition
	x, y, roomID := 0, 0, 1
	if pos != nil {
		x, y, roomID = pos.X, pos.Y, pos.RoomID
	}
	player, err := character.NewPlayer(x, y)
	if err != nil {
		return nil, err
	}
	g := &Game{
		variables: &Variables{},
		screen:    &Screen{},
		windows:   &window.Windows{},
		player:    player,
		mapID:     1,
	}
	g.setRoomID(roomID)
	return g, nil
}

func (g *Game) setRoomID(id int) error {
	g.roomID = id
	g.events = nil
	for _, e := range g.CurrentRoom().Events {
		i := NewInterpreter(g)
		event, err := character.NewEvent(e, i)
		if err != nil {
			return err
		}
		g.events = append(g.events, event)
	}
	return nil
}

func (g *Game) Events() []*character.Event {
	return g.events
}

func (g *Game) executingEvent() *character.Event {
	if g.continuingEvent != nil && g.continuingEvent.IsExecutingCommands() {
		return g.continuingEvent
	}
	for _, e := range g.events {
		if e.IsExecutingCommands() {
			return e
		}
	}
	return nil
}

func (g *Game) IsEventExecuting() bool {
	return g.executingEvent() != nil
}

func (g *Game) UpdateEvents() error {
	if e := g.executingEvent(); e != nil {
		if err := e.Update(); err != nil {
			return err
		}
		return nil
	}
	for _, e := range g.events {
		if err := e.Update(); err != nil {
			return err
		}
	}
	if g.continuingEvent != nil {
		if err := g.continuingEvent.Update(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) transferPlayerImmediately(roomID, x, y int, e *character.Event) {
	g.player.TransferImmediately(x, y)
	g.setRoomID(roomID)
	g.continuingEvent = e
}

func (g *Game) CurrentMap() *data.Map {
	for _, d := range data.Current().Maps {
		if d.ID == g.mapID {
			return d
		}
	}
	return nil
}

func (g *Game) CurrentRoom() *data.Room {
	for _, r := range g.CurrentMap().Rooms {
		if r.ID == g.roomID {
			return r
		}
	}
	return nil
}

func (g *Game) Screen() *Screen {
	return g.screen
}

func (g *Game) Windows() *window.Windows {
	return g.windows
}

func (g *Game) Player() *character.Player {
	return g.player
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
			return strconv.Itoa(g.variables.VariableValue(id))
		}
		return str
	})
}
