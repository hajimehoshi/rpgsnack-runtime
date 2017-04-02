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

// +build !android
// +build !ios
// +build !js

package game

import (
	"encoding/json"
	"log"
	"os"

	datapkg "github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type Requester struct {
	game *Game
}

func (m *Requester) RequestUnlockAchievement(requestID int, achievementID int) {
	log.Printf("request unlock achievement: requestID: %d, achievementID: %d", requestID, achievementID)
	m.game.FinishUnlockAchievement(requestID)
}

func (m *Requester) RequestSaveProgress(requestID int, data []uint8) {
	log.Printf("request save progress: requestID: %d", requestID)
	go func() {
		f, err := os.Create(datapkg.SavePath())
		if err != nil {
			// TODO: Should pass err instead of string?
			m.game.FinishSaveProgress(requestID)
			return
		}
		defer f.Close()
		if _, err := f.Write(data); err != nil {
			m.game.FinishSaveProgress(requestID)
			return
		}
		m.game.FinishSaveProgress(requestID)
	}()
}

func (m *Requester) RequestPurchase(requestID int, productID string) {
	log.Printf("request purchase: requestID: %d, productID: %s", requestID, productID)
	// Add new purchase if unique
	var purchases []string
	if err := json.Unmarshal(datapkg.Purchases(), &purchases); err != nil {
		panic(err)
	}

	isNew := true
	for _, p := range purchases {
		if p == productID {
			isNew = false
			break
		}
	}

	if isNew {
		purchases = append(purchases, productID)
	}

	newPurchases, err := json.Marshal(purchases)
	if err != nil {
		panic(err)
	}

	go func() {
		f, err := os.Create(datapkg.PurchasesPath())
		if err != nil {
			// TODO: Should pass err instead of string?
			m.game.FinishPurchase(requestID, true, newPurchases)
			return
		}
		defer f.Close()
		if _, err := f.Write(newPurchases); err != nil {
			m.game.FinishPurchase(requestID, true, newPurchases)
			return
		}
		m.game.FinishPurchase(requestID, true, newPurchases)
	}()
}

func (m *Requester) RequestRestorePurchases(requestID int) {
	log.Printf("request restore purchase: requestID: %d", requestID)
	m.game.FinishRestorePurchases(requestID, nil)
}

func (m *Requester) RequestInterstitialAds(requestID int) {
	log.Printf("request interstitial ads: requestID: %d", requestID)
	m.game.FinishInterstitialAds(requestID)
}

func (m *Requester) RequestRewardedAds(requestID int) {
	log.Printf("request rewarded ads: requestID: %d", requestID)
	m.game.FinishRewardedAds(requestID, true)
}

func (m *Requester) RequestOpenLink(requestID int, linkType string, data string) {
	log.Printf("request open link: requestID: %d", requestID)
	m.game.FinishOpenLink(requestID)
}

func (m *Requester) RequestShareImage(requestID int, title string, message string, image string) {
	log.Printf("request share image: requestID: %d, title: %s, message: %s, image: %s", requestID, title, message, image)
	m.game.FinishShareImage(requestID)
}
