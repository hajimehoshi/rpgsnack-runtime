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

package window

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

const (
	choiceBalloonHeight        = 20
	chosenBalloonWaitingFrames = 5
)

type Windows struct {
	nextBalloon    *balloon
	nextBanner     *banner
	balloons       []*balloon // TODO: Rename?
	choiceBalloons []*balloon
	banner         *banner

	chosenIndex               int
	choosing                  bool
	choosingInterpreterID     int
	chosenBalloonWaitingCount int
	hasChosenIndex            bool

	// Not dump
	lastLang language.Tag
}

type Choice struct {
	ID      data.UUID
	Text    string
	Checked bool
}

func (w *Windows) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("nextBalloon")
	e.EncodeInterface(w.nextBalloon)

	e.EncodeString("nextBanner")
	e.EncodeInterface(w.nextBanner)

	e.EncodeString("balloons")
	e.BeginArray()
	for _, b := range w.balloons {
		e.EncodeInterface(b)
	}
	e.EndArray()

	e.EncodeString("choiceBalloons")
	e.BeginArray()
	for _, b := range w.choiceBalloons {
		e.EncodeInterface(b)
	}
	e.EndArray()

	e.EncodeString("banner")
	e.EncodeInterface(w.banner)

	e.EncodeString("chosenIndex")
	e.EncodeInt(w.chosenIndex)

	e.EncodeString("choosing")
	e.EncodeBool(w.choosing)

	e.EncodeString("choosingInterpreterId")
	e.EncodeInt(w.choosingInterpreterID)

	e.EncodeString("chosenBalloonWaitingCount")
	e.EncodeInt(w.chosenBalloonWaitingCount)

	e.EncodeString("hasChosenIndex")
	e.EncodeBool(w.hasChosenIndex)

	e.EndMap()
	return e.Flush()
}

func (w *Windows) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		k := d.DecodeString()
		switch k {
		case "nextBalloon":
			if !d.SkipCodeIfNil() {
				w.nextBalloon = &balloon{}
				d.DecodeInterface(w.nextBalloon)
			}
		case "nextBanner":
			if !d.SkipCodeIfNil() {
				w.nextBanner = &banner{}
				d.DecodeInterface(w.nextBanner)
			}
		case "balloons":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				w.balloons = make([]*balloon, n)
				for i := 0; i < n; i++ {
					if !d.SkipCodeIfNil() {
						w.balloons[i] = &balloon{}
						d.DecodeInterface(w.balloons[i])
					}
				}
			}
		case "choiceBalloons":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				w.choiceBalloons = make([]*balloon, n)
				for i := 0; i < n; i++ {
					if !d.SkipCodeIfNil() {
						w.choiceBalloons[i] = &balloon{}
						d.DecodeInterface(w.choiceBalloons[i])
					}
				}
			}
		case "banner":
			if !d.SkipCodeIfNil() {
				w.banner = &banner{}
				d.DecodeInterface(w.banner)
			}
		case "chosenIndex":
			w.chosenIndex = d.DecodeInt()
		case "choosing":
			w.choosing = d.DecodeBool()
		case "choosingInterpreterId":
			w.choosingInterpreterID = d.DecodeInt()
		case "chosenBalloonWaitingCount":
			w.chosenBalloonWaitingCount = d.DecodeInt()
		case "hasChosenIndex":
			w.hasChosenIndex = d.DecodeBool()
		default:
			if err := d.Error(); err != nil {
				return err
			}
			return fmt.Errorf("window: Windows.DecodeMsgpack failed: unknown key: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: Windows.DecodeMsgpack failed: %v", err)
	}
	return nil
}

// ChosenIndex returns the chosen index that are chosen lastly.
// This value is valid even after the choosing windows are closed.
// This value is invalidated when the new choosing windows are shown.
func (w *Windows) ChosenIndex() int {
	return w.chosenIndex
}

func (w *Windows) HasChosenIndex() bool {
	return w.hasChosenIndex
}

func (w *Windows) ShowBalloon(contentID data.UUID, content string, balloonType data.BalloonType, eventID int, interpreterID int, messageStyle *data.MessageStyle) {
	if w.nextBalloon != nil {
		panic("window: nextBalloon must be nil at ShowBalloon")
	}
	// TODO: How to call newBalloonCenter?
	w.nextBalloon = newBalloonWithArrow(contentID, content, balloonType, eventID, interpreterID, messageStyle)
}

func (w *Windows) ShowMessage(contentID data.UUID, content string, eventID int, background data.MessageBackground, positionType data.MessagePositionType, textAlign data.TextAlign, interpreterID int, messageStyle *data.MessageStyle) {
	if w.nextBanner != nil {
		panic("window: nextBalloon must be nil at ShowMessage")
	}
	w.nextBanner = newBanner(contentID, content, eventID, background, positionType, textAlign, interpreterID, messageStyle)
}

func (w *Windows) ShowChoices(sceneManager *scene.Manager, choices []*Choice, interpreterID int) {
	// TODO: w.chosenBalloonWaitingCount should be 0 here!
	if w.chosenBalloonWaitingCount > 0 {
		panic("windows: chosenBalloonWaitingCount must be > 0 at ShowChoices")
	}
	w.choiceBalloons = nil
	for i, choice := range choices {
		x := 0
		y := i * choiceBalloonHeight
		width := consts.MapWidth
		balloon := newBalloon(x, y, width, choiceBalloonHeight, choice.ID, choice.Text, data.BalloonTypeNormal, interpreterID, sceneManager.Game().CreateChoicesMessageStyle(), choice.Checked)
		w.choiceBalloons = append(w.choiceBalloons, balloon)
		balloon.open()
	}
	w.chosenIndex = 0
	w.choosing = true
	w.choosingInterpreterID = interpreterID
	w.hasChosenIndex = false
}

func (w *Windows) CloseAll() {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		b.close()
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		b.close()
	}
	if w.banner != nil {
		w.banner.close()
	}
}

func (w *Windows) IsBusyWithChoosing() bool {
	return w.choosing || w.chosenBalloonWaitingCount > 0
}

// IsAnimating reports whether some of windows are busy (animating or opened).
// If interpreterID is non 0, checking the window related to the interpreter is skipped.
func (w *Windows) IsBusy(interpreterID int) bool {
	if w.IsAnimating(interpreterID) {
		return true
	}
	if w.choosingInterpreterID == interpreterID {
		if w.IsBusyWithChoosing() {
			return true
		}
	}
	if w.isOpened(interpreterID) {
		return true
	}
	if w.nextBalloon != nil && (interpreterID == 0 || w.nextBalloon.interpreterID == interpreterID) {
		return true
	}
	if w.nextBanner != nil && (interpreterID == 0 || w.nextBanner.interpreterID == interpreterID) {
		return true
	}
	return false
}

func (w *Windows) CanProceed(interpreterID int) bool {
	if !w.IsBusy(interpreterID) {
		return true
	}
	if !w.isOpened(interpreterID) {
		return false
	}
	if !input.Triggered() {
		return false
	}
	return true
}

func (w *Windows) isOpened(interpreterID int) bool {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isOpened() {
			return true
		}
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isOpened() {
			return true
		}
	}
	if w.banner != nil && (interpreterID == 0 || w.banner.interpreterID == interpreterID) {
		return true
	}
	return false
}

// IsAnimating reports whether some of windows are animating.
// If interpreterID is non 0, checking the window related to the interpreter is skipped.
func (w *Windows) IsAnimating(interpreterID int) bool {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isAnimating() {
			return true
		}
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isAnimating() {
			return true
		}
	}
	if w.banner != nil {
		if interpreterID > 0 && w.banner.interpreterID != interpreterID {
			return false
		}
		if w.banner.isAnimating() {
			return true
		}
	}
	return false
}

func (w *Windows) findCharacterByEventID(characters []*character.Character, eventID int) *character.Character {
	var c *character.Character
	for _, cc := range characters {
		if cc.EventID() == eventID {
			c = cc
			break
		}
	}
	return c
}

func (w *Windows) Update(playerY int, sceneManager *scene.Manager, characters []*character.Character) {
	if w.lastLang == language.Und {
		w.lastLang = lang.Get()
	}

	if w.lastLang != lang.Get() {
		for _, b := range w.balloons {
			if b == nil {
				continue
			}
			content := sceneManager.Game().Texts.Get(lang.Get(), b.contentID)
			b.setContent(content)
		}
		for _, b := range w.choiceBalloons {
			if b == nil {
				continue
			}
			content := sceneManager.Game().Texts.Get(lang.Get(), b.contentID)
			b.setContent(content)
		}
		if w.banner != nil {
			content := sceneManager.Game().Texts.Get(lang.Get(), w.banner.contentID)
			w.banner.setContent(content)
		}
		w.lastLang = lang.Get()
	}

	if !w.choosing {
		// 0 means to check all balloons.
		// TODO: Don't use magic numbers.
		if w.nextBalloon != nil && !w.IsAnimating(0) && !w.isOpened(0) {
			w.balloons = []*balloon{w.nextBalloon}
			w.balloons[0].open()
			w.nextBalloon = nil
		}
		if w.nextBanner != nil && !w.IsAnimating(0) && !w.isOpened(0) {
			w.banner = w.nextBanner
			w.banner.open()
			w.nextBanner = nil
		}
	}
	if w.chosenBalloonWaitingCount > 0 {
		w.chosenBalloonWaitingCount--
		if w.chosenBalloonWaitingCount == 0 {
			w.choiceBalloons[w.chosenIndex].close()
			for _, b := range w.balloons {
				if b == nil {
					continue
				}
				b.close()
			}
			if w.banner != nil {
				w.banner.close()
			}
		}
	} else if w.choosing && w.isOpened(0) && input.Triggered() {
		_, h := sceneManager.Size()
		ymax := h / consts.TileScale
		ymin := ymax - len(w.choiceBalloons)*choiceBalloonHeight
		_, y := input.Position()
		y += sceneManager.BottomOffset()
		y /= consts.TileScale
		if y < ymin || ymax <= y {
			return
		}
		// Close regular balloons
		w.chosenIndex = (y - ymin) / choiceBalloonHeight
		for i, b := range w.choiceBalloons {
			if i == w.chosenIndex {
				continue
			}
			b.close()
		}
		w.chosenBalloonWaitingCount = chosenBalloonWaitingFrames
		w.choosing = false
		w.choosingInterpreterID = 0
		w.hasChosenIndex = true
	}
	for i, b := range w.balloons {
		if b == nil {
			continue
		}
		b.update(w.findCharacterByEventID(characters, b.eventID))
		if b.isAnimating() && input.Triggered() {
			b.trySkipTypingAnim()
		} else if b.isClosed() {
			w.balloons[i] = nil
		}
	}
	for i, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		b.update(nil)
		if b.isClosed() {
			w.choiceBalloons[i] = nil
		}
	}
	if w.banner != nil {
		w.banner.update(playerY, w.findCharacterByEventID(characters, w.banner.eventID))
		if w.banner.isAnimating() && input.Triggered() {
			w.banner.trySkipTypingAnim()
		} else if w.banner.isClosed() {
			w.banner = nil
		}
	}
}

func (w *Windows) Draw(screen *ebiten.Image, characters []*character.Character, offsetX, offsetY, windowOffsetY int) {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		b.draw(screen, w.findCharacterByEventID(characters, b.eventID), offsetX, offsetY)
	}
	if w.banner != nil {
		w.banner.draw(screen, offsetX, 0)
	}
	_, sh := screen.Size()
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		b.draw(screen, nil, offsetX, sh/consts.TileScale-windowOffsetY-len(w.choiceBalloons)*choiceBalloonHeight)
	}

}
