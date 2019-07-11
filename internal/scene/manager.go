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
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/screenshot"
)

var TierTypes = [...]string{"tier1_donation", "tier2_donation", "tier3_donation", "tier4_donation"}

type Scene interface {
	Update(manager *Manager) error
	Draw(screen *ebiten.Image)
	Resize(width, height int)
}

type setPlatformDataArgs struct {
	key   PlatformDataKey
	value string
}

type Manager struct {
	width                 int
	height                int
	requester             Requester
	current               Scene
	next                  Scene
	fadingInCountMax      int
	fadingInCount         int
	fadingOutCountMax     int
	fadingOutCount        int
	lastRequestID         int
	resultCh              chan RequestResult
	results               map[int]*RequestResult
	setPlatformDataCh     chan setPlatformDataArgs
	game                  *data.Game
	progress              []byte
	permanent             *Permanent
	purchases             []string
	interstitialAdsLoaded bool
	rewardedAdsLoaded     bool
	credits               *data.Credits
	prices                map[string]string
	popupNewsID           int64

	blackImage *ebiten.Image
	turbo      bool

	screenshot    *screenshot.Screenshot
	screenshotDir string
	pseudoScreen  *ebiten.Image
	frozenScreen  *ebiten.Image
	origWidth     int
	origHeight    int

	needsSharingScreenshot bool

	// offscreen is for scaling.
	offscreen *ebiten.Image
}

type PlatformDataKey string

const (
	PlatformDataKeyInterstitialAdsLoaded PlatformDataKey = "interstitial_ads_loaded"
	PlatformDataKeyRewardedAdsLoaded     PlatformDataKey = "rewarded_ads_loaded"
	PlatformDataKeyBackButton            PlatformDataKey = "backbutton"
	PlatformDataKeyCredits               PlatformDataKey = "credits"
	PlatformDataKeyPrices                PlatformDataKey = "prices"
	PlatformDataKeyPopupNewsID           PlatformDataKey = "popup_news_id"
)

func NewManager(width, height int, requester Requester, game *data.Game, progress []byte, permanent []byte, purchases []string, fadingInCount int) *Manager {
	p := &Permanent{}
	if len(permanent) > 0 {
		if err := msgpack.Unmarshal(permanent, p); err != nil {
			panic(fmt.Sprintf("scene: msgpack encoding error: %v", err))
		}
	}

	m := &Manager{
		width:             width,
		height:            height,
		requester:         &requesterImpl{Requester: requester},
		resultCh:          make(chan RequestResult, 1),
		results:           map[int]*RequestResult{},
		setPlatformDataCh: make(chan setPlatformDataArgs, 1),
		game:              game,
		progress:          progress,
		permanent:         p,
		purchases:         purchases,
		fadingInCount:     fadingInCount,
		fadingInCountMax:  fadingInCount,
	}
	m.blackImage, _ = ebiten.NewImage(16, 16, ebiten.FilterDefault)
	m.blackImage.Fill(color.Black)

	audio.SetBGMVolume(float64(m.BGMVolume()) / 100.0)
	audio.SetSEVolume(float64(m.SEVolume()) / 100.0)

	m.credits = &data.Credits{
		Sections: []data.CreditsSection{
			{
				Header: "AUTHOR",
				Body:   []string{"???"},
			},
			{
				Header:      "PLATINUM SPONSOR",
				HeaderColor: "#bbaaee",
				Body:        nil,
			},
			{
				Header:      "GOLD SPONSOR",
				HeaderColor: "#ffd700",
				Body:        nil,
			},
			{
				Header:      "SILVER SPONSOR",
				HeaderColor: "#c0c0c0",
				Body:        nil,
			},
			{
				Header:      "BRONZE SPONSOR",
				HeaderColor: "#cd7f32",
				Body:        nil,
			},
		},
	}

	return m
}

func (m *Manager) InitScene(scene Scene) {
	if m.current != nil {
		panic("scene: the current scene must not be nil")
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
		m.current.Resize(width, height)
	}
}

func (m *Manager) BottomOffset() int {
	if consts.HasExtraBottomGrid(m.height) {
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

func (m *Manager) SetPseudoScreen(screen *ebiten.Image) {
	if m.origWidth == 0 || m.origHeight == 0 {
		m.origWidth = m.width
		m.origHeight = m.height
	}
	if m.frozenScreen == nil {
		m.frozenScreen, _ = ebiten.NewImage(m.width, m.height, ebiten.FilterDefault)
		m.Draw(m.frozenScreen)
	}

	m.pseudoScreen = screen
	m.SetScreenSize(m.pseudoScreen.Size())
}

func (m *Manager) ResetPseudoScreen() {
	m.pseudoScreen = nil
	m.frozenScreen.Dispose()
	m.frozenScreen = nil
	m.SetScreenSize(m.origWidth, m.origHeight)
	m.origWidth = 0
	m.origHeight = 0
}

func (m *Manager) Update() error {
	backPressed := false
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
			backPressed = true
		case PlatformDataKeyCredits:
			var credits *data.Credits
			if err := json.Unmarshal([]byte(a.value), &credits); err != nil {
				return err
			}
			m.credits = credits
		case PlatformDataKeyPrices:
			var prices map[string]string
			if err := json.Unmarshal([]byte(a.value), &prices); err != nil {
				return err
			}
			m.prices = prices
		case PlatformDataKeyPopupNewsID:
			i, err := strconv.ParseInt(a.value, 10, 64)
			if err != nil {
				return err
			}
			m.popupNewsID = i
		default:
			log.Printf("platform data key not implemented: %s", a.key)
		}
	default:
	}

	if input.IsMuteButtonTriggered() {
		audio.ToggleMute()
	}

	if m.screenshot == nil {
		if input.IsScreenshotButtonTriggered() {
			sizes := []screenshot.Size{
				{
					Width:  480, // 1242
					Height: 720, // 2208
				},
				{
					Width:  480, // 2048
					Height: 854, // 2732
				},
				{
					Width:  480,  // 1125
					Height: 1040, // 2436
				},
			}
			m.screenshot = screenshot.New(sizes, m.game.Texts.Languages())
			m.screenshotDir = filepath.Join("screenshots", time.Now().Format("20060102_030405"))
		} else if m.needsSharingScreenshot {
			m.screenshot = screenshot.New([]screenshot.Size{
				{
					Width:  480,  // 1125
					Height: 1040, // 2436
				},
			}, []language.Tag{lang.Get()})
		}
	}
	if m.screenshot != nil {
		m.screenshot.Update(m)
		if m.screenshot.IsFinished() {
			m.screenshot = nil
		}
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
		if backPressed {
			input.PressBackButton()
			backPressed = false
		}
		if m.next != nil {
			if m.fadingOutCount == 0 {
				m.current = m.next
				m.next = nil
			}
		} else {
			if err := m.current.Update(m); err != nil {
				return err
			}
		}
		if 0 < m.fadingOutCount {
			m.fadingOutCount--
		} else if 0 < m.fadingInCount {
			m.fadingInCount--
		}
	}
	return nil
}

func (m *Manager) ShareScreenshot() {
	m.needsSharingScreenshot = true
}

func (m *Manager) widthScale() float64 {
	ow, _ := m.Size()
	return float64(m.width) / float64(ow)
}

func (m *Manager) Draw(screen *ebiten.Image) error {
	if m.pseudoScreen != nil {
		screen.DrawImage(m.frozenScreen, nil)
		m.pseudoScreen.Clear()
		screen = m.pseudoScreen
	}

	m.drawWithScale(screen)

	if m.screenshot != nil {
		img, size, l, err := m.screenshot.TryDump()
		if err != nil {
			return err
		}
		if img != nil {
			if m.needsSharingScreenshot {
				subject := m.game.Texts.Get(lang.Get(), m.game.System.GameName)
				body := m.game.Texts.Get(lang.Get(), m.game.System.ScreenshotMessage)
				m.Requester().RequestShareImage(0, subject, body, img)
				m.needsSharingScreenshot = false
			} else {
				if err := os.MkdirAll(m.screenshotDir, 0755); err != nil {
					return err
				}
				fn := filepath.Join(m.screenshotDir, fmt.Sprintf("%d-%d-%s.png", size.Width, size.Height, l))
				fmt.Println(fn)
				ioutil.WriteFile(fn, img, 0666)
			}
		}
	}
	return nil
}

func (m *Manager) drawWithScale(screen *ebiten.Image) {
	if m.widthScale() == 1 {
		m.drawImpl(screen)
		return
	}

	if m.offscreen != nil {
		w, h := m.offscreen.Size()
		if cw, ch := m.Size(); w != cw || h != ch {
			m.offscreen.Dispose()
			m.offscreen, _ = ebiten.NewImage(cw, ch, ebiten.FilterDefault)
		}
	} else {
		w, h := m.Size()
		m.offscreen, _ = ebiten.NewImage(w, h, ebiten.FilterDefault)
	}

	m.offscreen.Clear()
	m.drawImpl(m.offscreen)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(m.widthScale(), 1)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(m.offscreen, op)
}

func (m *Manager) drawImpl(screen *ebiten.Image) {
	m.current.Draw(screen)
	if 0 < m.fadingInCount || 0 < m.fadingOutCount {
		alpha := 0.0
		if 0 < m.fadingOutCount {
			alpha = 1 - float64(m.fadingOutCount)/float64(m.fadingOutCountMax)
		} else {
			alpha = float64(m.fadingInCount) / float64(m.fadingInCountMax)
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
	return len(m.progress) > 0
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

func (m *Manager) ShopData(name data.ShopType, tabs []bool) []byte {
	p := &data.ShopPopup{}
	for i, t := range tabs {
		if !t {
			continue
		}
		shop := m.Game().GetShop(name, i)
		if shop == nil || len(shop.Products) == 0 {
			continue
		}
		p.Tabs = append(p.Tabs, &data.ShopPopupTab{
			Name:     m.game.Texts.Get(lang.Get(), shop.TabName),
			Products: m.getShopProducts(shop.Products),
		})
	}

	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *Manager) DynamicShopData(products []int) []byte {
	p := &data.ShopPopup{}
	p.Tabs = append(p.Tabs, &data.ShopPopupTab{
		Name:     "",
		Products: m.getShopProducts(products),
	})

	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *Manager) getShopProducts(products []int) []*data.ShopProduct {
	data := []*data.ShopProduct{}
	shopProducts := m.Game().GetShopProducts(products)
	for _, shopProduct := range shopProducts {
		if m.IsAvailable(shopProduct.ID) {
			shopProduct.Unlocked = m.IsUnlocked(shopProduct.ID)
			data = append(data, shopProduct)
		}
	}

	return data
}

func (m *Manager) SponsorTier() int {
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

func (m *Manager) GoTo(next Scene) {
	m.GoToWithFading(next, 0, 0)
}

func (m *Manager) GoToWithFading(next Scene, fadingOutCount, fadingInCount int) {
	if 0 < m.fadingInCount || 0 < m.fadingOutCount {
		// TODO: Should panic here?
		return
	}
	m.next = next
	m.fadingInCount = fadingInCount
	m.fadingInCountMax = fadingInCount
	m.fadingOutCount = fadingOutCount
	m.fadingOutCountMax = fadingOutCount
}

func (m *Manager) GenerateRequestID() int {
	m.lastRequestID++
	return m.lastRequestID
}

func (m *Manager) Credits() *data.Credits {
	return m.credits
}

func (m *Manager) Price(key string) string {
	if m.prices == nil {
		return ""
	}
	return m.prices[key]
}

func (m *Manager) ReceiveResultIfExists(id int) *RequestResult {
	if r, ok := m.results[id]; ok {
		delete(m.results, id)
		return r
	}
	return nil
}

func (m *Manager) RequestRewardedAds(requestID int, forceAds bool) {
	m.Requester().RequestRewardedAds(requestID, forceAds)
}

func (m *Manager) RequestSaveVolume(requestID int, seVolume int, bgmVolume int) {
	m.permanent.SEMute = 100 - seVolume
	m.permanent.BGMMute = 100 - bgmVolume
	bytes, err := msgpack.Marshal(m.permanent)
	if err != nil {
		panic(fmt.Sprintf("scene: msgpack encoding error: %v", err))
	}
	m.Requester().RequestSavePermanent(requestID, bytes)
}

func (m *Manager) RequestSavePermanentVariable(requestID int, permanentVariableID int, value int64) {
	if len(m.permanent.Variables) < permanentVariableID+1 {
		zeros := make([]int64, permanentVariableID+1-len(m.permanent.Variables))
		m.permanent.Variables = append(m.permanent.Variables, zeros...)
	}
	m.permanent.Variables[permanentVariableID] = value

	bytes, err := msgpack.Marshal(m.permanent)
	if err != nil {
		panic(fmt.Sprintf("scene: msgpack encoding error: %v", err))
	}
	m.Requester().RequestSavePermanent(requestID, bytes)
}

func (m *Manager) RequestSaveVibrationEnabled(requestID int, vibrationEnabled bool) {
	m.permanent.VibrationDisabled = !vibrationEnabled
	bytes, err := msgpack.Marshal(m.permanent)
	if err != nil {
		panic(fmt.Sprintf("scene: msgpack encoding error: %v", err))
	}
	m.Requester().RequestSavePermanent(requestID, bytes)
}

func (m *Manager) PermanentVariableValue(id int) int64 {
	if len(m.permanent.Variables) < id+1 {
		zeros := make([]int64, id+1-len(m.permanent.Variables))
		m.permanent.Variables = append(m.permanent.Variables, zeros...)
	}
	return m.permanent.Variables[id]
}

func (m *Manager) VibrationEnabled() bool {
	return !m.permanent.VibrationDisabled
}

func (m *Manager) BGMVolume() int {
	return 100 - m.permanent.BGMMute
}

func (m *Manager) SEVolume() int {
	return 100 - m.permanent.SEMute
}

func (m *Manager) RequestSavePermanentMinigame(requestID int, minigameID, score int, lastActiveAt int64) {
	if len(m.permanent.Minigames) < minigameID+1 {
		zeros := make([]*MinigameData, minigameID+1-len(m.permanent.Minigames))
		m.permanent.Minigames = append(m.permanent.Minigames, zeros...)
	}
	m.permanent.Minigames[minigameID] = &MinigameData{Score: score, LastActiveAt: lastActiveAt}

	bytes, err := msgpack.Marshal(m.permanent)
	if err != nil {
		panic(fmt.Sprintf("scene: msgpack encoding error: %v", err))
	}
	m.Requester().RequestSavePermanent(requestID, bytes)
}

func (m *Manager) PermanentMinigame(id int) *MinigameData {
	if len(m.permanent.Minigames) < id+1 {
		return nil
	}
	return m.permanent.Minigames[id]
}

func (m *Manager) RespondUnlockAchievement(id int) {
	go func() {
		m.resultCh <- RequestResult{
			ID:   id,
			Type: RequestTypeUnlockAchievement,
		}
	}()
}

func (m *Manager) RespondSaveProgress(id int) {
	go func() {
		m.resultCh <- RequestResult{
			ID:   id,
			Type: RequestTypeSaveProgress,
		}
	}()
}

func (m *Manager) RespondSavePermanent(id int) {
	go func() {
		m.resultCh <- RequestResult{
			ID:   id,
			Type: RequestTypeSavePermanent,
		}
	}()
}

func (m *Manager) RespondPurchase(id int, success bool, purchases []byte) {
	go func() {
		m.resultCh <- RequestResult{
			ID:        id,
			Type:      RequestTypePurchase,
			Succeeded: success,
			Data:      purchases,
		}
	}()
}

func (m *Manager) RespondShowShop(id int, success bool, purchases []byte) {
	go func() {
		m.resultCh <- RequestResult{
			ID:        id,
			Type:      RequestTypeShowShop,
			Succeeded: success,
			Data:      purchases,
		}
	}()
}

func (m *Manager) RespondRestorePurchases(id int, success bool, purchases []byte) {
	go func() {
		m.resultCh <- RequestResult{
			ID:        id,
			Type:      RequestTypeRestorePurchases,
			Succeeded: success,
			Data:      purchases,
		}
	}()
}

func (m *Manager) RespondInterstitialAds(id int, success bool) {
	go func() {
		m.resultCh <- RequestResult{
			ID:        id,
			Type:      RequestTypeInterstitialAds,
			Succeeded: success,
		}
	}()
}

func (m *Manager) RespondRewardedAds(id int, success bool) {
	go func() {
		m.resultCh <- RequestResult{
			ID:        id,
			Type:      RequestTypeRewardedAds,
			Succeeded: success,
		}
	}()
}

func (m *Manager) RespondOpenLink(id int) {
	go func() {
		m.resultCh <- RequestResult{
			ID:   id,
			Type: RequestTypeOpenLink,
		}
	}()
}

func (m *Manager) RespondShareImage(id int) {
	go func() {
		m.resultCh <- RequestResult{
			ID:   id,
			Type: RequestTypeShareImage,
		}
	}()
}

func (m *Manager) RespondChangeLanguage(id int) {
	go func() {
		m.resultCh <- RequestResult{
			ID:   id,
			Type: RequestTypeChangeLanguage,
		}
	}()
}

func (m *Manager) SetPlatformData(key PlatformDataKey, value string) {
	go func() {
		m.setPlatformDataCh <- setPlatformDataArgs{
			key:   key,
			value: value,
		}
	}()
}

func (m *Manager) InterstitialAdsLoaded() bool {
	return m.interstitialAdsLoaded
}

func (m *Manager) RewardedAdsLoaded() bool {
	return m.rewardedAdsLoaded
}

func (m *Manager) RespondAsset(id int, success bool, data []byte) {
	go func() {
		m.resultCh <- RequestResult{
			ID:        id,
			Type:      RequestTypeAsset,
			Succeeded: success,
			Data:      data,
		}
	}()
}
