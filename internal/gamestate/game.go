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
	"image/color"
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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/hints"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/items"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/picture"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/variables"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/weather"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

type Rand interface {
	Intn(n int) int
}

type Game struct {
	hints                *hints.Hints
	items                *items.Items
	variables            *variables.Variables
	screen               *Screen
	windows              *window.Windows
	pictures             *picture.Pictures
	currentMap           *Map
	lastInterpreterID    int
	autoSaveEnabled      bool
	playerControlEnabled bool
	inventoryVisible     bool
	weatherType          data.WeatherType
	cleared              bool

	lastPlayingBGMName   string
	lastPlayingBGMVolume float64

	backgrounds map[int]map[int]string
	foregrounds map[int]map[int]string
	playerSpeed data.Speed

	// Fields that are not dumped
	rand             Rand
	waitingRequestID int
	prices           map[string]string // TODO: We want to use https://godoc.org/golang.org/x/text/currency
	weather          *weather.Weather
}

func generateDefaultRand() Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewGame() *Game {
	g := &Game{
		currentMap:           NewMap(),
		hints:                &hints.Hints{},
		items:                &items.Items{},
		variables:            &variables.Variables{},
		screen:               &Screen{},
		windows:              &window.Windows{},
		pictures:             &picture.Pictures{},
		rand:                 generateDefaultRand(),
		autoSaveEnabled:      true,
		playerControlEnabled: true,
		inventoryVisible:     false,
		playerSpeed:          data.Speed5,
	}
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

	e.EncodeString("pictures")
	e.EncodeInterface(g.pictures)

	e.EncodeString("currentMap")
	e.EncodeInterface(g.currentMap)

	e.EncodeString("lastInterpreterId")
	e.EncodeInt(g.lastInterpreterID)

	e.EncodeString("autoSaveEnabled")
	e.EncodeBool(g.autoSaveEnabled)

	e.EncodeString("playerControlEnabled")
	e.EncodeBool(g.playerControlEnabled)

	e.EncodeString("inventoryVisible")
	e.EncodeBool(g.inventoryVisible)

	e.EncodeString("weatherType")
	e.EncodeString(string(g.weatherType))

	e.EncodeString("cleared")
	e.EncodeBool(g.cleared)

	e.EncodeString("lastPlayingBGMName")
	e.EncodeString(audio.PlayingBGMName())

	e.EncodeString("lastPlayingBGMVolume")
	e.EncodeFloat64(audio.PlayingBGMVolume())

	e.EncodeString("playerSpeed")
	e.EncodeInt(int(g.playerSpeed))

	e.EncodeString("backgrounds")
	e.BeginMap()
	for id, m := range g.backgrounds {
		e.EncodeInt(id)
		e.BeginMap()
		for id, r := range m {
			e.EncodeInt(id)
			e.EncodeString(r)
		}
		e.EndMap()
	}
	e.EndMap()

	e.EncodeString("foregrounds")
	e.BeginMap()
	for id, m := range g.foregrounds {
		e.EncodeInt(id)
		e.BeginMap()
		for id, r := range m {
			e.EncodeInt(id)
			e.EncodeString(r)
		}
		e.EndMap()
	}
	e.EndMap()

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
				g.hints = &hints.Hints{}
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
		case "pictures":
			if !d.SkipCodeIfNil() {
				g.pictures = &picture.Pictures{}
				d.DecodeInterface(g.pictures)
			}
		case "currentMap":
			if !d.SkipCodeIfNil() {
				g.currentMap = &Map{}
				d.DecodeInterface(g.currentMap)
			}
		case "lastInterpreterId":
			g.lastInterpreterID = d.DecodeInt()
		case "autoSaveEnabled":
			g.autoSaveEnabled = d.DecodeBool()
		case "playerControlEnabled":
			g.playerControlEnabled = d.DecodeBool()
		case "inventoryVisible":
			g.inventoryVisible = d.DecodeBool()
		case "weatherType":
			g.SetWeather(data.WeatherType(d.DecodeString()))
		case "cleared":
			g.cleared = d.DecodeBool()
		case "lastPlayingBGMName":
			g.lastPlayingBGMName = d.DecodeString()
		case "lastPlayingBGMVolume":
			g.lastPlayingBGMVolume = d.DecodeFloat64()
		case "playerSpeed":
			g.playerSpeed = data.Speed(d.DecodeInt())
		case "backgrounds":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				g.backgrounds = map[int]map[int]string{}
				for i := 0; i < n; i++ {
					id := d.DecodeInt()
					g.backgrounds[id] = map[int]string{}
					n2 := d.DecodeMapLen()
					for j := 0; j < n2; j++ {
						id2 := d.DecodeInt()
						r := d.DecodeString()
						g.backgrounds[id][id2] = r
					}
				}
			}
		case "foregrounds":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				g.foregrounds = map[int]map[int]string{}
				for i := 0; i < n; i++ {
					id := d.DecodeInt()
					g.foregrounds[id] = map[int]string{}
					n2 := d.DecodeMapLen()
					for j := 0; j < n2; j++ {
						id2 := d.DecodeInt()
						r := d.DecodeString()
						g.foregrounds[id][id2] = r
					}
				}
			}
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

func (g *Game) Items() *items.Items {
	return g.items
}

// TODO: Remove this
func (g *Game) Map() *Map {
	return g.currentMap
}

func (g *Game) MapPassableAt(through bool, x, y int, ignoreCharacters bool) bool {
	return g.currentMap.Passable(through, x, y, ignoreCharacters)
}

func (g *Game) Update(sceneManager *scene.Manager) error {
	g.items.SetDataItems(sceneManager.Game().Items)
	if g.lastPlayingBGMName != "" {
		audio.PlayBGM(g.lastPlayingBGMName, g.lastPlayingBGMVolume, 0)
		g.lastPlayingBGMName = ""
		g.lastPlayingBGMVolume = 0
	}
	if g.waitingRequestID != 0 {
		if sceneManager.ReceiveResultIfExists(g.waitingRequestID) != nil {
			g.waitingRequestID = 0
		}
	}
	g.weather.Update()
	g.screen.Update()
	playerY := 0
	if g.currentMap.player != nil {
		_, playerY = g.currentMap.player.DrawPosition()
	}
	g.windows.Update(playerY, sceneManager, g.createCharacterList())
	g.pictures.Update()
	if err := g.currentMap.Update(sceneManager, g); err != nil {
		return err
	}
	return nil
}

func (g *Game) Clear() {
	g.cleared = true
}

func (g *Game) SetBGM(bgm data.BGM) {
	if bgm.Name == "" {
		audio.StopBGM(0)
	} else {
		audio.PlayBGM(bgm.Name, float64(bgm.Volume)/100, 0)
	}
}

func (g *Game) ShowInventory(group int) {
	g.inventoryVisible = true
	g.items.SetActiveItemGroup(group)
}

func (g *Game) HideInventory() {
	g.inventoryVisible = false
}

func (g *Game) InventoryVisible() bool {
	return g.inventoryVisible
}

func (g *Game) SetAutoSaveEnabled(enabled bool) {
	g.autoSaveEnabled = enabled
}

func (g *Game) IsAutoSaveEnabled() bool {
	return g.autoSaveEnabled
}

func (g *Game) SetPlayerControlEnabled(enabled bool) {
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
		panic(fmt.Sprintf("gamestate: msgpack encoding error: %v", err))
	}
	sceneManager.Requester().RequestSaveProgress(id, m)
	sceneManager.SetProgress(m)
	return true
}

var reMessage = regexp.MustCompile(`\\([a-zA-Z])\[([^\]]+)\]`)

func (g *Game) ParseMessageSyntax(str string) string {
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

func (g *Game) MeetsCondition(cond *data.Condition, eventID int) (bool, error) {
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
		rhs := 0
		// TODO: This is redundant: can we refactor them?
		switch value := cond.Value.(type) {
		case float32:
			rhs = int(value)
		case float64:
			rhs = int(value)
		case int:
			rhs = value
		case int8:
			rhs = int(value)
		case int16:
			rhs = int(value)
		case int32:
			rhs = int(value)
		case int64:
			rhs = int(value)
		case uint8:
			rhs = int(value)
		case uint16:
			rhs = int(value)
		case uint32:
			rhs = int(value)
		case uint64:
			rhs = int(value)
		}
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = g.variables.VariableValue(rhs)
		default:
			return false, fmt.Errorf("gamestate: invalid value type: %v eventID %d", cond, eventID)
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
			return false, fmt.Errorf("gamestate: invalid comp: %s eventID %d", cond.Comp, eventID)
		}
	case data.ConditionTypeItem:
		id := cond.ID
		itemValue := data.ConditionItemValue(cond.Value.(string))

		switch itemValue {
		case data.ConditionItemOwn:
			if id == 0 {
				return g.items.ItemNum() > 0, nil
			} else {
				return g.items.Includes(id), nil
			}
		case data.ConditionItemNotOwn:
			if id == 0 {
				return g.items.ItemNum() == 0, nil
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
			return false, fmt.Errorf("gamestate: invalid item value: %s eventID %d", itemValue, eventID)
		}
	case data.ConditionTypeSpecial:
		switch cond.Value.(string) {
		case specialConditionEventExistsAtPlayer:
			e := g.currentMap.executableEventAt(g.currentMap.player.Position())
			return e != nil, nil
		default:
			return false, fmt.Errorf("gamestate: ConditionTypeSpecial: invalid value: %v eventID %d", cond, eventID)
		}
	default:
		return false, fmt.Errorf("gamestate: invalid condition: %v eventID %d", cond, eventID)
	}
	return false, nil
}

func (g *Game) GenerateInterpreterID() int {
	g.lastInterpreterID++
	return g.lastInterpreterID
}

func (g *Game) SetRandomForTesting(r Rand) {
	g.rand = r
}

func (g *Game) RandomValue(min, max int) int {
	return min + g.rand.Intn(max-min)
}

func (g *Game) DrawWeather(screen *ebiten.Image) {
	g.weather.Draw(screen)
}

func (g *Game) ApplyTintColor(c *ebiten.ColorM) {
	g.screen.ApplyTintColor(c)
}

func (g *Game) ApplyShake(geo *ebiten.GeoM) {
	g.screen.ApplyShake(geo)
}

func (g *Game) DrawScreen(screenImage *ebiten.Image) {
	g.screen.Draw(screenImage)
}

func (g *Game) DrawWindows(screen *ebiten.Image, offsetX, offsetY, windowOffsetY int) {
	g.windows.Draw(screen, g.createCharacterList(), offsetX, offsetY, windowOffsetY)
}

func (g *Game) createCharacterList() []*character.Character {
	cs := []*character.Character{}
	cs = append(cs, g.currentMap.player)
	cs = append(cs, g.currentMap.events...)
	return cs
}

func (g *Game) DrawPictures(screen *ebiten.Image, offsetX, offsetY int, priority data.PicturePriorityType) {
	g.pictures.Draw(screen, offsetX, offsetY, priority)
}

func (g *Game) Character(mapID, roomID, eventID int) *character.Character {
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

func (g *Game) SetPrices(p map[string]string) {
	g.prices = p
}

func (g *Game) CanWindowProceed(interpreterID int) bool {
	return g.windows.CanProceed(interpreterID)
}

func (g *Game) IsWindowAnimating(interpreterID int) bool {
	return g.windows.IsAnimating(interpreterID)
}

func (g *Game) CloseAllWindows() {
	g.windows.CloseAll()
}

func (g *Game) HasChosenWindowIndex() bool {
	return g.windows.HasChosenIndex()
}

func (g *Game) ChosenWindowIndex() int {
	return g.windows.ChosenIndex()
}

func (g *Game) ShowBalloon(sceneManager *scene.Manager, interpreterID, mapID, roomID, eventID int, contentID data.UUID, balloonType data.BalloonType, messageStyle *data.MessageStyle) bool {
	ch := g.Character(mapID, roomID, eventID)
	if ch == nil {
		return false
	}

	content := sceneManager.Game().Texts.Get(lang.Get(), contentID)
	content = g.ParseMessageSyntax(content)
	g.windows.ShowBalloon(contentID, content, balloonType, eventID, interpreterID, messageStyle)
	return true
}

func (g *Game) ShowMessage(sceneManager *scene.Manager, interpreterID, eventID int, contentID data.UUID, background data.MessageBackground, positionType data.MessagePositionType, textAlign data.TextAlign, messageStyle *data.MessageStyle) {
	content := sceneManager.Game().Texts.Get(lang.Get(), contentID)
	content = g.ParseMessageSyntax(content)
	g.windows.ShowMessage(contentID, content, eventID, background, positionType, textAlign, interpreterID, messageStyle)
}

func (g *Game) ShowChoices(sceneManager *scene.Manager, interpreterID int, choiceIDs []data.UUID) {
	choices := []string{}
	for _, id := range choiceIDs {
		choice := sceneManager.Game().Texts.Get(lang.Get(), id)
		choice = g.ParseMessageSyntax(choice)
		choices = append(choices, choice)
	}
	g.windows.ShowChoices(sceneManager, choiceIDs, choices, interpreterID)
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

func (g *Game) StartCombineCommands(combine *data.Combine) {
	g.currentMap.StartCombineCommands(g, combine)
}

func (g *Game) StartItemCommands(itemID int) {
	g.currentMap.StartItemCommands(g, itemID)
}

func (g *Game) SetPlayerDir(dir data.Dir) {
	g.currentMap.player.SetDir(dir)
}

func (g *Game) SetWeather(weatherType data.WeatherType) {
	if g.weatherType == weatherType {
		return
	}
	g.weatherType = weatherType
	if weatherType == data.WeatherTypeNone {
		g.weather = nil
		return
	}
	g.weather = weather.New(weatherType)
}

func (g *Game) TransferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	g.currentMap.transferPlayerImmediately(g, roomID, x, y, interpreter)
}

func (g *Game) ExecutableEventAtPlayer() *character.Character {
	p := g.currentMap.player
	return g.currentMap.executableEventAt(p.Position())
}

func (g *Game) CurrentEvents() []*data.Event {
	return g.currentMap.CurrentRoom().Events
}

func (g *Game) SetFadeColor(clr color.Color) {
	g.screen.setFadeColor(clr)
}

func (g *Game) IsScreenFadedOut() bool {
	return g.screen.isFadedOut()
}

func (g *Game) IsScreenFading() bool {
	return g.screen.isFading()
}

func (g *Game) FadeIn(time int) {
	g.screen.fadeIn(time)
}

func (g *Game) FadeOut(time int) {
	g.screen.fadeOut(time)
}

func (g *Game) StartShaking(power, speed, count int, dir data.ShakeDirection) {
	g.screen.startShaking(power, speed, count, dir)
}

func (g *Game) StopShaking() {
	g.screen.stopShaking()
}

func (g *Game) IsShaking() bool {
	return g.screen.isShaking()
}

func (g *Game) StartTint(red, green, blue, gray float64, time int) {
	g.screen.startTint(red, green, blue, gray, time)
}

func (g *Game) IsChangingTint() bool {
	return g.screen.isChangingTint()
}

func (g *Game) RefreshEvents() error {
	return g.currentMap.refreshEvents(g)
}

func (g *Game) SetVariable(sceneManager *scene.Manager, variableID int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}, mapID, roomID, eventID int) error {
	rhs := 0
	switch valueType {
	case data.SetVariableValueTypeConstant:
		rhs = value.(int)
	case data.SetVariableValueTypeVariable:
		rhs = g.VariableValue(value.(int))
	case data.SetVariableValueTypeRandom:
		v := value.(*data.SetVariableValueRandom)
		rhs = g.RandomValue(v.Begin, v.End+1)
	case data.SetVariableValueTypeCharacter:
		args := value.(*data.SetVariableCharacterArgs)
		id := args.EventID
		if id == 0 {
			id = eventID
		}
		ch := g.Character(mapID, roomID, id)
		if ch == nil {
			// TODO: return error?
			return nil
		}
		dir := ch.Dir()
		switch args.Type {
		case data.SetVariableCharacterTypeDirection:
			switch dir {
			case data.DirUp:
				rhs = 0
			case data.DirRight:
				rhs = 1
			case data.DirDown:
				rhs = 2
			case data.DirLeft:
				rhs = 3
			default:
				panic("not reach")
			}
		case data.SetVariableCharacterTypeRoomX:
			rhs, _ = ch.Position()
		case data.SetVariableCharacterTypeRoomY:
			_, rhs = ch.Position()
		case data.SetVariableCharacterTypeScreenX:
			rhs, _ = ch.DrawPosition()
		case data.SetVariableCharacterTypeScreenY:
			_, rhs = ch.DrawPosition()
		case data.SetVariableCharacterTypeIsPressed:
			x, y := ch.Position()
			pressX, pressY := g.currentMap.GetPressedPosition()
			if x == pressX && y == pressY {
				rhs = 1
			}
		default:
			return fmt.Errorf("gamestate: not implemented yet (set_variable): type %s", args.Type)
		}
	case data.SetVariableValueTypeIAPProduct:
		rhs = 0
		id := value.(int)
		rhs = 0
		if sceneManager.IsUnlocked(id) {
			rhs = 1
		}
	case data.SetVariableValueTypeSystem:
		systemVariableType := value.(data.SystemVariableType)
		switch systemVariableType {
		case data.SystemVariableHintCount:
			rhs = g.hints.ActiveHintCount()
		case data.SystemVariableInterstitialAdsLoaded:
			if sceneManager.InterstitialAdsLoaded() {
				rhs = 1
			}
		case data.SystemVariableRewardedAdsLoaded:
			if sceneManager.RewardedAdsLoaded() {
				rhs = 1
			}
		case data.SystemVariableRoomID:
			rhs = roomID
		default:
			return fmt.Errorf("gamestate: not implemented yet (set_variable): systemVariableType %s", systemVariableType)
		}
	}
	switch op {
	case data.SetVariableOpAssign:
	case data.SetVariableOpAdd:
		rhs = g.VariableValue(variableID) + rhs
	case data.SetVariableOpSub:
		rhs = g.VariableValue(variableID) - rhs
	case data.SetVariableOpMul:
		rhs = g.VariableValue(variableID) * rhs
	case data.SetVariableOpDiv:
		rhs = g.VariableValue(variableID) / rhs
	case data.SetVariableOpMod:
		rhs = g.VariableValue(variableID) % rhs
	default:
		return fmt.Errorf("gamestate: not implemented yet (set_variable): SetVariableOp %s", op)
	}
	g.variables.SetVariableValue(variableID, rhs)
	return nil
}

func (g *Game) PauseHint(id int) {
	g.hints.Pause(id)
}

func (g *Game) ActivateHint(id int) {
	g.hints.Activate(id)
}

func (g *Game) CompleteHint(id int) {
	g.hints.Complete(id)
}

func (g *Game) AddItem(id int) {
	g.items.Add(id)
}

func (g *Game) RemoveItem(id int) {
	g.items.Remove(id)
}

func (g *Game) SetEventItem(id int) {
	g.items.SetEventItem(id)
}

func (g *Game) InsertItemBefore(targetItemID int, insertItemID int) {
	g.items.InsertBefore(targetItemID, insertItemID)
}

func (g *Game) SetBackground(mapID, roomID int, image string) {
	if g.backgrounds == nil {
		g.backgrounds = map[int]map[int]string{}
	}
	if _, ok := g.backgrounds[mapID]; !ok {
		g.backgrounds[mapID] = map[int]string{}
	}
	g.backgrounds[mapID][roomID] = image
}

func (g *Game) SetForeground(mapID, roomID int, image string) {
	if g.foregrounds == nil {
		g.foregrounds = map[int]map[int]string{}
	}
	if _, ok := g.foregrounds[mapID]; !ok {
		g.foregrounds[mapID] = map[int]string{}
	}
	g.foregrounds[mapID][roomID] = image
}

func (g *Game) Background(mapID, roomID int) (string, bool) {
	if g.backgrounds != nil {
		if r, ok := g.backgrounds[mapID]; ok {
			if img, ok := r[roomID]; ok {
				return img, true
			}
		}
	}
	return "", false
}

func (g *Game) Foreground(mapID, roomID int) (string, bool) {
	if g.foregrounds != nil {
		if r, ok := g.foregrounds[mapID]; ok {
			if img, ok := r[roomID]; ok {
				return img, true
			}
		}
	}
	return "", false
}

func (g *Game) PlayerSpeed() data.Speed {
	return g.playerSpeed
}

func (g *Game) SetPlayerSpeed(value data.Speed) {
	g.playerSpeed = value
}
