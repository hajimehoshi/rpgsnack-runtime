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

package scene

import (
	"github.com/hajimehoshi/ebiten"
)

const (
	TileSize      = 16
	TileXNum      = 10
	TileYNum      = 10
	TileScale     = 3
	GameMarginTop = 2 * TileSize * TileScale
	TextScale     = 2
)

type scene interface {
	Update(manager *Manager) error
	Draw(screen *ebiten.Image) error
}

type Manager struct {
	width              int
	height             int
	requester          Requester
	current            scene
	next               scene
	lastRequestID      int
	requestFinisher    chan func() int
	finishedRequestIDs map[int]struct{}
}

type Requester interface {
	RequestUnlockAchievement(requestID int, achievementID int)
	RequestSaveProgress(requestID int, data []uint8)
	RequestPurchase(requestID int, productID string)
	RequestInterstitialAds(requestID int)
	RequestRewardedAds(requestID int)
	RequestOpenLink(requestID int, linkType string, data string)
	RequestShareImage(requestID int, title string, message string, image string)
}

func NewManager(width, height int, requester Requester, initScene scene) *Manager {
	return &Manager{
		width:              width,
		height:             height,
		requester:          requester,
		current:            initScene,
		requestFinisher:    make(chan func() int, 1),
		finishedRequestIDs: map[int]struct{}{},
	}
}

func (m *Manager) Size() (int, int) {
	return m.width, m.height
}

func (m *Manager) Requester() Requester {
	return m.requester
}

func (m *Manager) MapOffsetX() int {
	return (m.width - TileXNum*TileSize*TileScale) / 2
}

func (m *Manager) Update() error {
	select {
	case f := <-m.requestFinisher:
		id := f()
		m.finishedRequestIDs[id] = struct{}{}
	default:
	}
	if m.next != nil {
		m.current = m.next
		m.next = nil
	}
	if err := m.current.Update(m); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) error {
	if err := m.current.Draw(screen); err != nil {
		return err
	}
	return nil
}

func (m *Manager) GoTo(next scene) {
	m.next = next
}

func (m *Manager) GenerateRequestID() int {
	m.lastRequestID++
	return m.lastRequestID
}

func (m *Manager) HasFinishedRequestID(id int) bool {
	_, ok := m.finishedRequestIDs[id]
	return ok
}

func (m *Manager) FinishRequestID(id int) {
	delete(m.finishedRequestIDs, id)
}

func (m *Manager) FinishUnlockAchievement(id int) {
	m.requestFinisher <- func() int {
		// TODO: Implement this
		return id
	}
}

func (m *Manager) FinishSaveProgress(id int) {
	m.requestFinisher <- func() int {
		return id
	}
}
