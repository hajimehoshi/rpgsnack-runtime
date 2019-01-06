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

// +build !js

package game

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	datapkg "github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type Requester struct {
	game *Game
}

func (m *Requester) RequestUnlockAchievement(requestID int, achievementID int) {
	log.Printf("request unlock achievement: requestID: %d, achievementID: %d", requestID, achievementID)
	m.game.RespondUnlockAchievement(requestID)
}

func (m *Requester) RequestSaveProgress(requestID int, data []uint8) {
	log.Printf("request save progress: requestID: %d", requestID)
	go func() {
		defer m.game.RespondSaveProgress(requestID)

		f, err := os.Create(datapkg.SavePath())
		if err != nil {
			// TODO: Should pass err instead of string?
			panic(err)
		}
		defer f.Close()

		if _, err := f.Write(data); err != nil {
			panic(err)
			return
		}
	}()
}

func (m *Requester) RequestPurchase(requestID int, productID string) {
	log.Printf("request purchase: requestID: %d, productID: %s", requestID, productID)
	go func() {
		result := ([]uint8)("[]")
		defer m.game.RespondPurchase(requestID, true, result)

		var purchases []string
		b, err := ioutil.ReadFile(datapkg.PurchasesPath())
		if err != nil && !os.IsNotExist(err) {
			return
		}
		if b != nil {
			result = b
			if err := json.Unmarshal(b, &purchases); err != nil {
				return
			}
			for _, p := range purchases {
				if p == productID {
					return
				}
			}
		}

		purchases = append(purchases, productID)
		b, err = json.Marshal(purchases)
		if err != nil {
			panic(err)
		}

		result = b
		if err := ioutil.WriteFile(datapkg.PurchasesPath(), b, 0666); err != nil {
			panic(err)
		}
	}()
}

func (m *Requester) RequestShowShop(requestID int, data string) {
	log.Printf("request to ShowShop data:%s", data)
	//TODO Mock purchase selection
	m.game.RespondShowShop(requestID, true, []byte("[\"bronze_support\"]"))
}

func (m *Requester) RequestRestorePurchases(requestID int) {
	log.Printf("request restore purchase: requestID: %d", requestID)
	m.game.RespondRestorePurchases(requestID, true, nil)
}

func (m *Requester) RequestInterstitialAds(requestID int) {
	log.Printf("request interstitial ads: requestID: %d", requestID)
	go func() {
		time.Sleep(time.Second)
		m.game.RespondInterstitialAds(requestID)
	}()
}

func (m *Requester) RequestRewardedAds(requestID int) {
	log.Printf("request rewarded ads: requestID: %d", requestID)
	go func() {
		time.Sleep(time.Second)
		m.game.RespondRewardedAds(requestID, true)
	}()
}

func (m *Requester) RequestOpenLink(requestID int, linkType string, data string) {
	log.Printf("request open link: requestID: %d %s %s", requestID, linkType, data)
	m.game.RespondOpenLink(requestID)
}

func (m *Requester) RequestShareImage(requestID int, title string, message string, image []byte) {
	log.Printf("request share image: requestID: %d, title: %s, message: %s", requestID, title, message)
	m.game.RespondShareImage(requestID)
}

func (m *Requester) RequestChangeLanguage(requestID int, lang string) {
	log.Printf("request change language: requestID: %d, lang: %s", requestID, lang)
	go func() {
		defer m.game.RespondChangeLanguage(requestID)
		f, err := os.Create(datapkg.LanguagePath())
		if err != nil {
			// TODO: Should pass err instead of string?
			panic(err)
		}
		defer f.Close()
		j, err := json.Marshal(lang)
		if err != nil {
			panic(err)
		}
		if _, err := f.Write(j); err != nil {
			panic(err)
		}
	}()
}

func (m *Requester) RequestGetIAPPrices(requestID int) {
	log.Printf("request IAP prices: requestID: %d", requestID)
	m.game.RespondGetIAPPrices(requestID, true, []byte("{\"ads_removal\": \"$0.99\"}"))
}

func (m *Requester) RequestTerminateGame() {
	log.Printf("request terminate game")
}

func (m *Requester) RequestReview() {
	log.Printf("request review")
}

func (m *Requester) RequestSendAnalytics(eventName string, value string) {
	log.Printf("request to send an analytics event: %s value: %s", eventName, value)
}

func (m *Requester) RequestAsset(requestID int, key string) {
	// TODO: Implement this
	log.Printf("request asset %s", key)
	go func() {
		m.game.RespondAsset(requestID, true, []byte{})
	}()
}
