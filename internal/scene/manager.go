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
	"image/color"
	"log"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
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
	Draw(screen *ebiten.Image)
}

type Manager struct {
	width                 int
	height                int
	requester             Requester
	current               scene
	next                  scene
	fadingCountMax        int
	fadingCount           int
	lastRequestID         int
	resultCh              chan RequestResult
	results               map[int]*RequestResult
	language              language.Tag
	interstitialAdsLoaded bool
	rewardedAdsLoaded     bool
	blackImage            *ebiten.Image
}

type Requester interface {
	RequestUnlockAchievement(requestID int, achievementID int)
	RequestSaveProgress(requestID int, data []uint8)
	RequestPurchase(requestID int, productID string)
	RequestRestorePurchases(requestID int)
	RequestInterstitialAds(requestID int)
	RequestRewardedAds(requestID int)
	RequestOpenLink(requestID int, linkType string, data string)
	RequestShareImage(requestID int, title string, message string, image string)
}

type RequestType int

const (
	RequestTypeUnlockAchievement RequestType = iota
	RequestTypeSaveProgress
	RequestTypePurchase
	RequestTypeRestorePurchases
	RequestTypeInterstitialAds
	RequestTypeRewardedAds
	RequestTypeOpenLink
	RequestTypeShareImage
)

type RequestResult struct {
	ID        int
	Type      RequestType
	Succeeded bool
	Data      []uint8
}

type PlatformDataKey string

const (
	PlatformDataKeyInterstitialAdsLoaded PlatformDataKey = "interstitial_ads_loaded"
	PlatformDataKeyRewardedAdsLoaded     PlatformDataKey = "rewarded_ads_loaded"
)

func NewManager(width, height int, requester Requester) *Manager {
	m := &Manager{
		width:     width,
		height:    height,
		requester: requester,
		resultCh:  make(chan RequestResult, 1),
		results:   map[int]*RequestResult{},
		language:  language.English,
	}
	m.blackImage, _ = ebiten.NewImage(16, 16, ebiten.FilterNearest)
	m.blackImage.Fill(color.Black)
	return m
}

func (m *Manager) InitScene(scene scene) {
	if m.current != nil {
		panic("not reach")
	}
	m.current = scene
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
	case r := <-m.resultCh:
		m.results[r.ID] = &r
	default:
	}
	if m.next != nil {
		if m.fadingCount > 0 {
			if m.fadingCount <= m.fadingCountMax/2 {
				m.current = m.next
				m.next = nil
			}
		} else {
			m.current = m.next
			m.next = nil
		}
	}
	if err := m.current.Update(m); err != nil {
		return err
	}
	if 0 < m.fadingCount {
		m.fadingCount--
	}
	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) {
	m.current.Draw(screen)
	if 0 < m.fadingCount {
		alpha := 0.0
		if m.fadingCount > m.fadingCountMax/2 {
			alpha = 1 - float64(m.fadingCount-m.fadingCountMax/2)/float64(m.fadingCountMax/2)
		} else {
			alpha = float64(m.fadingCount) / float64(m.fadingCountMax/2)
		}
		sw, sh := screen.Size()
		w, h := m.blackImage.Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(sw)/float64(w), float64(sh)/float64(h))
		op.ColorM.Scale(1, 1, 1, alpha)
		screen.DrawImage(m.blackImage, op)
	}
}

func (m *Manager) Language() language.Tag {
	return m.language
}

func (m *Manager) SetLanguage(language language.Tag) {
	for _, l := range data.Current().Texts.Languages() {
		if l == language {
			m.language = language
			return
		}
	}
	m.language = data.Current().Texts.Languages()[0]
}

func (m *Manager) GoTo(next scene) {
	m.GoToWithFading(next, 0)
}

func (m *Manager) GoToWithFading(next scene, frames int) {
	if 0 < m.fadingCount {
		// TODO: Should panic here?
		return
	}
	m.next = next
	m.fadingCount = frames
	m.fadingCountMax = frames
}

func (m *Manager) GenerateRequestID() int {
	m.lastRequestID++
	return m.lastRequestID
}

func (m *Manager) ReceiveResultIfExists(id int) *RequestResult {
	if r, ok := m.results[id]; ok {
		delete(m.results, id)
		return r
	}
	return nil
}

func (m *Manager) FinishUnlockAchievement(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeUnlockAchievement,
	}
}

func (m *Manager) FinishSaveProgress(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeSaveProgress,
	}
}

func (m *Manager) FinishPurchase(id int, success bool, purchases []uint8) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypePurchase,
		Succeeded: success,
		Data:      purchases,
	}
}

func (m *Manager) FinishRestorePurchases(id int, success bool, purchases []uint8) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeRestorePurchases,
		Succeeded: success,
		Data:      purchases,
	}
}

func (m *Manager) FinishInterstitialAds(id int) {
	m.interstitialAdsLoaded = false
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeInterstitialAds,
	}
}

func (m *Manager) FinishRewardedAds(id int, success bool) {
	m.rewardedAdsLoaded = false
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeRewardedAds,
		Succeeded: success,
	}
}

func (m *Manager) FinishOpenLink(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeOpenLink,
	}
}

func (m *Manager) FinishShareImage(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeShareImage,
	}
}

func (m *Manager) SetPlatformData(key PlatformDataKey, value int) {
	switch key {
	case PlatformDataKeyInterstitialAdsLoaded:
		m.interstitialAdsLoaded = true
	case PlatformDataKeyRewardedAdsLoaded:
		m.rewardedAdsLoaded = true
	default:
		log.Printf("platform data key not implemented: %s", key)
	}
}

func (m *Manager) InterstitialAdsLoaded() bool {
	return m.interstitialAdsLoaded
}

func (m *Manager) RewardedAdsLoaded() bool {
	return m.rewardedAdsLoaded
}
