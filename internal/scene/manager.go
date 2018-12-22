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
	"encoding/json"
	"image/color"
	"log"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

var TierTypes = [...]string{"tier1_donation", "tier2_donation", "tier3_donation"}

type scene interface {
	Update(manager *Manager) error
	Draw(screen *ebiten.Image)
	Resize()
}

type setPlatformDataArgs struct {
	key   PlatformDataKey
	value string
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
	setPlatformDataCh     chan setPlatformDataArgs
	game                  *data.Game
	progress              []byte
	purchases             []string
	interstitialAdsLoaded bool
	rewardedAdsLoaded     bool
	blackImage            *ebiten.Image
	turbo                 bool
	offscreen             *ebiten.Image
}

type PlatformDataKey string

const (
	PlatformDataKeyInterstitialAdsLoaded PlatformDataKey = "interstitial_ads_loaded"
	PlatformDataKeyRewardedAdsLoaded     PlatformDataKey = "rewarded_ads_loaded"
	PlatformDataKeyBackButton            PlatformDataKey = "backbutton"
)

func NewManager(width, height int, requester Requester, game *data.Game, progress []byte, purchases []string) *Manager {
	m := &Manager{
		width:             width,
		height:            height,
		requester:         &requesterImpl{Requester: requester},
		resultCh:          make(chan RequestResult, 1),
		results:           map[int]*RequestResult{},
		setPlatformDataCh: make(chan setPlatformDataArgs, 1),
		game:              game,
		progress:          progress,
		purchases:         purchases,
	}
	m.blackImage, _ = ebiten.NewImage(16, 16, ebiten.FilterDefault)
	m.blackImage.Fill(color.Black)
	w, h := m.Size()
	m.offscreen, _ = ebiten.NewImage(w, h, ebiten.FilterDefault)
	return m
}

func (m *Manager) InitScene(scene scene) {
	if m.current != nil {
		panic("not reach")
	}
	m.current = scene
}

func (m *Manager) Size() (int, int) {
	// Logical width is always a constant value.
	return consts.MapScaledWidth, m.height
}

func (m *Manager) SetScreenSize(width, height int) {
	if m.width != width || m.height != height {
		m.width = width
		m.height = height
		m.offscreen.Dispose()
		w, h := m.Size()
		m.offscreen, _ = ebiten.NewImage(w, h, ebiten.FilterDefault)
		m.current.Resize()
	}
}

func (m *Manager) BottomOffset() int {
	if m.height > consts.SuperLargeScreenHeight {
		return consts.TileSize * consts.TileScale
	}
	return 0
}

func (m *Manager) HasExtraBottomGrid() bool {
	return m.BottomOffset() > 0
}

func (m *Manager) Requester() Requester {
	return m.requester
}

func (m *Manager) Update() error {
	triggerBack := false
	select {
	case r := <-m.resultCh:
		m.results[r.ID] = &r
		switch r.Type {
		case RequestTypeInterstitialAds:
			m.interstitialAdsLoaded = false
		case RequestTypeRewardedAds:
			m.rewardedAdsLoaded = false
		case RequestTypePurchase, RequestTypeRestorePurchases, RequestTypeShowShop:
			if r.Succeeded {
				var purchases []string
				if err := json.Unmarshal(r.Data, &purchases); err != nil {
					return err
				}
				m.purchases = purchases
			}
		default:
			// There is no action here. It's ok to ignore.
		}
	case a := <-m.setPlatformDataCh:
		switch a.key {
		case PlatformDataKeyInterstitialAdsLoaded:
			m.interstitialAdsLoaded = true
		case PlatformDataKeyRewardedAdsLoaded:
			m.rewardedAdsLoaded = true
		case PlatformDataKeyBackButton:
			triggerBack = true
		default:
			log.Printf("platform data key not implemented: %s", a.key)
		}
	default:
	}

	if input.IsTurboButtonTriggered() {
		m.turbo = !m.turbo
	}
	n := 1
	if m.turbo {
		n = 5
	}
	for i := 0; i < n; i++ {
		input.Update(m.widthScale(), 1)
		if triggerBack {
			input.TriggerBackButton()
			triggerBack = false
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
	}
	return nil
}

func (m *Manager) widthScale() float64 {
	ow, _ := m.offscreen.Size()
	return float64(m.width) / float64(ow)
}

func (m *Manager) Draw(screen *ebiten.Image) {
	s := m.widthScale()
	if s == 1 {
		m.drawImpl(screen)
		return
	}

	m.offscreen.Clear()
	m.drawImpl(m.offscreen)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s, 1)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(m.offscreen, op)
}

func (m *Manager) drawImpl(screen *ebiten.Image) {
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

func (m *Manager) Game() *data.Game {
	return m.game
}

func (m *Manager) HasProgress() bool {
	return m.progress != nil
}

func (m *Manager) Progress() []byte {
	return m.progress
}

func (m *Manager) SetProgress(progress []byte) {
	m.progress = progress
}

func (m *Manager) IsPurchased(key string) bool {
	for _, p := range m.purchases {
		if p == key {
			return true
		}
	}
	return false
}

func (m *Manager) IsUnlocked(id int) bool {
	ip := m.Game().IAPProductByID(id)
	if m.IsPurchased(ip.Key) {
		return true
	}
	for _, bi := range ip.Bundles {
		bip := m.Game().IAPProductByID(bi)
		if m.IsPurchased(bip.Key) {
			return true
		}
	}

	return false
}

func (m *Manager) IsAvailable(id int) bool {
	ip := m.Game().IAPProductByID(id)
	if m.IsPurchased(ip.Key) {
		return true
	}

	for _, tip := range m.game.IAPProducts {
		if m.IsPurchased(tip.Key) {
			for _, bi := range tip.Bundles {
				if bi == id {
					return false
				}
			}
		}

	}

	return true
}

func (m *Manager) ShopProductsDataByShop(name data.ShopType) []byte {
	shop := m.Game().GetShop(name)
	if shop == nil {
		return nil
	}
	return m.GetShopProductsData(shop.Products)
}

func (m *Manager) GetShopProductsData(products []int) []byte {
	data := []*data.ShopProduct{}
	shopProducts := m.Game().GetShopProducts(products)
	for _, shopProduct := range shopProducts {
		if m.IsAvailable(shopProduct.ID) {
			shopProduct.Unlocked = m.IsUnlocked(shopProduct.ID)
			data = append(data, shopProduct)
		}
	}

	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *Manager) MaxPurchaseTier() int {
	maxTier := 0
	for _, p := range m.purchases {
		iap := m.Game().GetIAPProduct(p)
		if iap == nil {
			continue
		}
		for i, tierType := range TierTypes {
			if iap.Type == tierType {
				if maxTier < i+1 {
					maxTier = i + 1
				}
			}
		}
	}
	return maxTier
}

func (m *Manager) IsAdsRemovable() bool {
	return m.game.GetIAPProductByType("ads_removal") != nil
}

func (m *Manager) IsAdsRemoved() bool {
	for _, i := range m.game.IAPProducts {
		if i.Type == "ads_removal" && m.IsPurchased(i.Key) {
			return true
		}
	}

	return false
}

func (m *Manager) SetLanguage(language language.Tag) language.Tag {
	language = lang.Normalize(language)
	found := false
	for _, l := range m.game.Texts.Languages() {
		if l == language {
			found = true
			break
		}
	}
	if !found {
		language = m.game.Texts.Languages()[0]
	}

	lang.Set(language)
	title := m.game.Texts.Get(language, m.game.System.GameName)
	if title == "" {
		title = "(No Title)"
	}
	ebiten.SetWindowTitle(title)
	return language
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

func (m *Manager) RespondUnlockAchievement(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeUnlockAchievement,
	}
}

func (m *Manager) RespondSaveProgress(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeSaveProgress,
	}
}

func (m *Manager) RespondPurchase(id int, success bool, purchases []byte) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypePurchase,
		Succeeded: success,
		Data:      purchases,
	}
}

func (m *Manager) RespondShowShop(id int, success bool, purchases []byte) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeShowShop,
		Succeeded: success,
		Data:      purchases,
	}
}

func (m *Manager) RespondRestorePurchases(id int, success bool, purchases []byte) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeRestorePurchases,
		Succeeded: success,
		Data:      purchases,
	}
}

func (m *Manager) RespondInterstitialAds(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeInterstitialAds,
	}
}

func (m *Manager) RespondRewardedAds(id int, success bool) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeRewardedAds,
		Succeeded: success,
	}
}

func (m *Manager) RespondOpenLink(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeOpenLink,
	}
}

func (m *Manager) RespondShareImage(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeShareImage,
	}
}

func (m *Manager) RespondChangeLanguage(id int) {
	m.resultCh <- RequestResult{
		ID:   id,
		Type: RequestTypeChangeLanguage,
	}
}

func (m *Manager) RespondGetIAPPrices(id int, success bool, prices []byte) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeIAPPrices,
		Succeeded: success,
		Data:      prices,
	}
}

func (m *Manager) SetPlatformData(key PlatformDataKey, value string) {
	m.setPlatformDataCh <- setPlatformDataArgs{
		key:   key,
		value: value,
	}
}

func (m *Manager) InterstitialAdsLoaded() bool {
	return m.interstitialAdsLoaded
}

func (m *Manager) RewardedAdsLoaded() bool {
	return m.rewardedAdsLoaded
}

func (m *Manager) RespondAsset(id int, success bool, data []byte) {
	m.resultCh <- RequestResult{
		ID:        id,
		Type:      RequestTypeAsset,
		Succeeded: success,
		Data:      data,
	}
}
