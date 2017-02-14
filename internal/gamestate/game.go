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
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

type Rand interface {
	Intn(n int) int
}

type Game struct {
	variables         *Variables
	screen            *Screen
	windows           *window.Windows
	currentMap        *Map
	lastInterpreterID int

	// Fields that are not dumped
	rand             Rand
	waitingRequestID int
}

func generateDefaultRand() Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewGame() (*Game, error) {
	g := &Game{
		variables: &Variables{},
		screen:    &Screen{},
		windows:   &window.Windows{},
		rand:      generateDefaultRand(),
	}
	m, err := NewMap(g)
	if err != nil {
		return nil, err
	}
	g.currentMap = m
	return g, nil
}

type tmpGame struct {
	Variables         *Variables      `json:"variables"`
	Screen            *Screen         `json:"screen"`
	Windows           *window.Windows `json:"windows"`
	Map               *Map            `json:"map"`
	LastInterpreterID int             `json:"lastInterpreterId"`
}

func (g *Game) MarshalJSON() ([]uint8, error) {
	tmp := &tmpGame{
		Variables:         g.variables,
		Screen:            g.screen,
		Windows:           g.windows,
		Map:               g.currentMap,
		LastInterpreterID: g.lastInterpreterID,
	}
	return json.Marshal(tmp)
}

func (g *Game) UnmarshalJSON(data []uint8) error {
	var tmp *tmpGame
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	g.variables = tmp.Variables
	g.screen = tmp.Screen
	g.windows = tmp.Windows
	g.currentMap = tmp.Map
	g.currentMap.setGame(g)
	g.lastInterpreterID = tmp.LastInterpreterID
	g.rand = generateDefaultRand()
	return nil
}

func (g *Game) Screen() *Screen {
	return g.screen
}

func (g *Game) Windows() *window.Windows {
	return g.windows
}

func (g *Game) Map() *Map {
	return g.currentMap
}

func (g *Game) Update(sceneManager *scene.Manager) error {
	if g.waitingRequestID != 0 {
		if sceneManager.HasFinishedRequestID(g.waitingRequestID) {
			sceneManager.FinishRequestID(g.waitingRequestID)
			g.waitingRequestID = 0
		}
	}
	return nil
}

func (g *Game) RequestSave(sceneManager *scene.Manager) (bool, error) {
	// If there is an unfinished request, stop saving the progress.
	if g.waitingRequestID != 0 {
		return false, nil
	}
	if g.currentMap.waitingRequestResponse() {
		return false, nil
	}
	id := sceneManager.GenerateRequestID()
	g.waitingRequestID = id
	j, err := json.Marshal(g)
	if err != nil {
		return false, err
	}
	sceneManager.Requester().RequestSaveProgress(id, j)
	return true, nil
}

var reMessage = regexp.MustCompile(`\\([a-zA-Z])\[(\d+)\]`)

func (g *Game) parseMessageSyntax(str string) string {
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

func (g *Game) meetsCondition(cond *data.Condition, eventID int) (bool, error) {
	// TODO: Is it OK to allow null conditions?
	if cond == nil {
		return true, nil
	}
	switch cond.Type {
	case data.ConditionTypeSwitch:
		id := cond.ID
		v := g.variables.SwitchValue(id)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeSelfSwitch:
		m, r := g.currentMap.mapID, g.currentMap.roomID
		v := g.variables.SelfSwitchValue(m, r, eventID, cond.ID)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeVariable:
		id := cond.ID
		v := g.variables.VariableValue(id)
		rhs := cond.Value.(int)
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = g.variables.VariableValue(rhs)
		default:
			return false, fmt.Errorf("mapscene: invalid value type: %s", cond.ValueType)
		}
		switch cond.Comp {
		case data.ConditionCompEqualTo:
			return v == rhs, nil
		case data.ConditionCompNotEqualTo:
			return v != rhs, nil
		case data.ConditionCompGreaterThanOrEqualTo:
			return v >= rhs, nil
		case data.ConditionCompGreaterThan:
			return v > rhs, nil
		case data.ConditionCompLessThanOrEqualTo:
			return v <= rhs, nil
		case data.ConditionCompLessThan:
			return v < rhs, nil
		default:
			return false, fmt.Errorf("mapscene: invalid comp: %s", cond.Comp)
		}
	default:
		return false, fmt.Errorf("mapscene: invalid condition: %s", cond)
	}
	return false, nil
}

func (g *Game) generateInterpreterID() int {
	g.lastInterpreterID++
	return g.lastInterpreterID
}

func (g *Game) SetRandom(r Rand) {
	g.rand = r
}

func (g *Game) randomValue(min, max int) int {
	return min + g.rand.Intn(max-min)
}

func (g *Game) DrawWindows(screen *ebiten.Image) error {
	cs := []*character.Character{}
	cs = append(cs, g.currentMap.player)
	cs = append(cs, g.currentMap.events...)
	if err := g.windows.Draw(screen, cs); err != nil {
		return err
	}
	return nil
}

func (g *Game) character(mapID, roomID, eventID int) *character.Character {
	if eventID == character.PlayerEventID {
		return g.currentMap.player
	}
	if g.currentMap.mapID != mapID {
		return nil
	}
	if g.currentMap.roomID != roomID {
		return nil
	}
	for _, e := range g.currentMap.events {
		if eventID == e.EventID() {
			return e
		}
	}
	return nil
}
