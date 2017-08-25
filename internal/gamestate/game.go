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
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/items"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/variables"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

type Rand interface {
	Intn(n int) int
}

type Game struct {
	hints                *Hints
	items                *items.Items
	variables            *variables.Variables
	screen               *Screen
	windows              *window.Windows
	currentMap           *Map
	lastInterpreterID    int
	autoSaveEnabled      bool
	playerControlEnabled bool
	cleared              bool

	lastPlayingBGMName   string
	lastPlayingBGMVolume float64

	// Fields that are not dumped
	rand             Rand
	waitingRequestID int
	prices           map[string]string // TODO: We want to use https://godoc.org/golang.org/x/text/currency
}

func generateDefaultRand() Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewGame() *Game {
	g := &Game{
		hints:                &Hints{},
		items:                &items.Items{},
		variables:            &variables.Variables{},
		screen:               &Screen{},
		windows:              &window.Windows{},
		rand:                 generateDefaultRand(),
		autoSaveEnabled:      true,
		playerControlEnabled: true,
	}
	g.currentMap = NewMap(g)
	return g
}

func (g *Game) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("hints")
	e.EncodeInterface(g.hints)

	e.EncodeString("items")
	e.EncodeInterface(g.items)

	e.EncodeString("variables")
	e.EncodeInterface(g.variables)

	e.EncodeString("screen")
	e.EncodeInterface(g.screen)

	e.EncodeString("windows")
	e.EncodeInterface(g.windows)

	e.EncodeString("currentMap")
	e.EncodeInterface(g.currentMap)

	e.EncodeString("lastInterpreterId")
	e.EncodeInt(g.lastInterpreterID)

	e.EncodeString("autoSaveEnabled")
	e.EncodeBool(g.autoSaveEnabled)

	e.EncodeString("playerControlEnabled")
	e.EncodeBool(g.playerControlEnabled)

	e.EncodeString("cleared")
	e.EncodeBool(g.cleared)

	e.EncodeString("lastPlayingBGMName")
	e.EncodeString(audio.PlayingBGMName())

	e.EncodeString("lastPlayingBGMVolume")
	e.EncodeFloat64(audio.PlayingBGMVolume())

	e.EndMap()
	return e.Flush()
}

func (g *Game) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		k := d.DecodeString()
		switch k {
		case "hints":
			if !d.SkipCodeIfNil() {
				g.hints = &Hints{}
				d.DecodeInterface(g.hints)
			}
		case "items":
			if !d.SkipCodeIfNil() {
				g.items = &items.Items{}
				d.DecodeInterface(g.items)
			}
		case "variables":
			if !d.SkipCodeIfNil() {
				g.variables = &variables.Variables{}
				d.DecodeInterface(g.variables)
			}
		case "screen":
			if !d.SkipCodeIfNil() {
				g.screen = &Screen{}
				d.DecodeInterface(g.screen)
			}
		case "windows":
			if !d.SkipCodeIfNil() {
				g.windows = &window.Windows{}
				d.DecodeInterface(g.windows)
			}
		case "currentMap":
			if !d.SkipCodeIfNil() {
				g.currentMap = &Map{}
				d.DecodeInterface(g.currentMap)
				g.currentMap.setGame(g)
			}
		case "lastInterpreterId":
			g.lastInterpreterID = d.DecodeInt()
		case "autoSaveEnabled":
			g.autoSaveEnabled = d.DecodeBool()
		case "playerControlEnabled":
			g.playerControlEnabled = d.DecodeBool()
		case "cleared":
			g.cleared = d.DecodeBool()
		case "lastPlayingBGMName":
			g.lastPlayingBGMName = d.DecodeString()
		case "lastPlayingBGMVolume":
			g.lastPlayingBGMVolume = d.DecodeFloat64()
		default:
			if err := d.Error(); err != nil {
				return err
			}
			return fmt.Errorf("gamestate: Game.DecodeMsgpack failed: unknown key: %s", k)
		}
	}
	g.rand = generateDefaultRand()
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Game.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (g *Game) UpdateScreen() {
	g.screen.Update()
}

func (g *Game) Items() *items.Items {
	return g.items
}

func (g *Game) UpdateWindows(sceneManager *scene.Manager) {
	playerY := 0
	if g.currentMap.player != nil {
		_, playerY = g.currentMap.player.Position()
	}
	g.windows.Update(playerY, sceneManager)
}

func (g *Game) Map() *Map {
	return g.currentMap
}

func (g *Game) Update(sceneManager *scene.Manager) {
	if g.lastPlayingBGMName != "" {
		audio.PlayBGM(g.lastPlayingBGMName, g.lastPlayingBGMVolume)
		g.lastPlayingBGMName = ""
		g.lastPlayingBGMVolume = 0
	}
	if g.waitingRequestID != 0 {
		if sceneManager.ReceiveResultIfExists(g.waitingRequestID) != nil {
			g.waitingRequestID = 0
		}
	}
}

func (g *Game) SetBGM(bgm data.BGM) {
	if bgm.Name == "" {
		audio.StopBGM()
	} else {
		audio.PlayBGM(bgm.Name, float64(bgm.Volume)/100)
	}
}

func (g *Game) setAutoSaveEnabled(enabled bool) {
	g.autoSaveEnabled = enabled
}

func (g *Game) IsAutoSaveEnabled() bool {
	return g.autoSaveEnabled
}

func (g *Game) setPlayerControlEnabled(enabled bool) {
	g.playerControlEnabled = enabled
}

func (g *Game) IsPlayerControlEnabled() bool {
	return g.playerControlEnabled
}

func (g *Game) RequestSave(sceneManager *scene.Manager) bool {
	// If there is an unfinished request, stop saving the progress.
	if g.waitingRequestID != 0 {
		return false
	}
	if g.currentMap.waitingRequestResponse() {
		return false
	}
	id := sceneManager.GenerateRequestID()
	g.waitingRequestID = id
	m, err := msgpack.Marshal(g)
	if err != nil {
		panic(fmt.Sprintf("gamestate: JSON encoding error: %v", err))
	}
	sceneManager.Requester().RequestSaveProgress(id, m)
	sceneManager.SetProgress(m)
	return true
}

var reMessage = regexp.MustCompile(`\\([a-zA-Z])\[([^\]]+)\]`)

func (g *Game) parseMessageSyntax(str string) string {
	return reMessage.ReplaceAllStringFunc(str, func(part string) string {
		name := strings.ToLower(part[1:2])
		args := part[3 : len(part)-1]

		switch name {
		case "p":
			return g.price(args)
		case "v":
			id, err := strconv.Atoi(args)
			if err != nil {
				panic(fmt.Sprintf("not reach: %s", err))
			}
			return strconv.Itoa(g.variables.VariableValue(id))
		}
		return str
	})
}

const (
	specialConditionEventExistsAtPlayer = "event_exists_at_player"
)

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
		rhs := int(cond.Value.(float64))
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = g.variables.VariableValue(rhs)
		default:
			return false, fmt.Errorf("gamestate: invalid value type: %s eventId %d", cond, eventID)
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
			return false, fmt.Errorf("gamestate: invalid comp: %s eventId %d", cond.Comp, eventID)
		}
	case data.ConditionTypeItem:
		id := cond.ID
		itemValue := data.ConditionItemValue(cond.Value.(string))

		switch itemValue {
		case data.ConditionItemOwn:
			if id == 0 {
				return len(g.items.Items()) > 0, nil
			} else {
				return g.items.Includes(id), nil
			}
		case data.ConditionItemNotOwn:
			if id == 0 {
				return len(g.items.Items()) == 0, nil
			} else {
				return !g.items.Includes(id), nil
			}
		case data.ConditionItemActive:
			if id == 0 {
				return g.items.ActiveItem() > 0, nil
			} else {
				return id == g.items.ActiveItem(), nil
			}

		default:
			return false, fmt.Errorf("gamestate: invalid item value: %s eventId %d", itemValue, eventID)
		}
	case data.ConditionTypeSpecial:
		switch cond.Value.(string) {
		case specialConditionEventExistsAtPlayer:
			e := g.currentMap.executableEventAt(g.currentMap.player.Position())
			return e != nil, nil
		default:
			return false, fmt.Errorf("gamestate: ConditionTypeSpecial: invalid value: %s eventId %d", cond, eventID)
		}
	default:
		return false, fmt.Errorf("gamestate: invalid condition: %s eventId %d", cond, eventID)
	}
	return false, nil
}

func (g *Game) generateInterpreterID() int {
	g.lastInterpreterID++
	return g.lastInterpreterID
}

func (g *Game) SetRandomForTesting(r Rand) {
	g.rand = r
}

func (g *Game) RandomValue(min, max int) int {
	return min + g.rand.Intn(max-min)
}

func (g *Game) DrawScreen(screen *ebiten.Image, tilesImage *ebiten.Image, op *ebiten.DrawImageOptions) {
	g.screen.Draw(screen, tilesImage, op)
}

func (g *Game) DrawWindows(screen *ebiten.Image) {
	cs := []*character.Character{}
	cs = append(cs, g.currentMap.player)
	cs = append(cs, g.currentMap.events...)
	g.windows.Draw(screen, cs)
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

func (g *Game) price(productID string) string {
	if _, ok := g.prices[productID]; ok {
		return g.prices[productID]
	}
	return ""
}

func (g *Game) updatePrices(p map[string]string) {
	g.prices = p
}

func (g *Game) SetSwitchValue(id int, value bool) {
	g.variables.SetSwitchValue(id, value)
}

func (g *Game) SetSelfSwitchValue(eventID int, id int, value bool) {
	m, r := g.currentMap.mapID, g.currentMap.roomID
	g.variables.SetSelfSwitchValue(m, r, eventID, id, value)
}

func (g *Game) VariableValue(id int) int {
	return g.variables.VariableValue(id)
}

func (g *Game) SetVariableValue(id int, value int) {
	g.variables.SetVariableValue(id, value)
}

func (g *Game) StartItemCommands() {
	g.currentMap.StartItemCommands(g.Items().EventItem())
}

func (g *Game) ExecutingItemCommands() bool {
	return g.currentMap.ExecutingItemCommands()
}
