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

// +build js

package game

import (
	"encoding/base64"
	"log"

	"github.com/gopherjs/gopherjs/js"
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
	js.Global.Get("localStorage").Call("setItem", "progress", base64.StdEncoding.EncodeToString(data))
	m.game.RespondSaveProgress(requestID)
}

func (m *Requester) RequestPurchase(requestID int, productID string) {
	log.Printf("request purchase: requestID: %d, productID: %s", requestID, productID)
	m.game.RespondPurchase(requestID, true, nil)
}

func (m *Requester) RequestShowShop(requestID int, data string) {
	log.Printf("request to ShowShop")
	//TODO Mock purchase selection
	m.game.RespondShowShop(requestID, true, []byte("[\"bronze_support\"]"))
}

func (m *Requester) RequestRestorePurchases(requestID int) {
	log.Printf("request restore purchase: requestID: %d", requestID)
	m.game.RespondRestorePurchases(requestID, true, nil)
}

func (m *Requester) RequestInterstitialAds(requestID int) {
	log.Printf("request interstitial ads: requestID: %d", requestID)
	m.game.RespondInterstitialAds(requestID)
}

func (m *Requester) RequestRewardedAds(requestID int) {
	log.Printf("request rewarded ads: requestID: %d", requestID)
	m.game.RespondRewardedAds(requestID, true)
}

func (m *Requester) RequestOpenLink(requestID int, linkType string, data string) {
	log.Printf("request open link: requestID: %d %s %s", requestID, linkType, data)
	m.game.RespondOpenLink(requestID)
}

func (m *Requester) RequestShareImage(requestID int, title string, message string, image string) {
	log.Printf("request share image: requestID: %d, title: %s, message: %s, image: %s", requestID, title, message, image)
	m.game.RespondShareImage(requestID)
}

func (m *Requester) RequestTerminateGame() {
	log.Printf("request terminate game")
}

func (m *Requester) RequestChangeLanguage(requestID int, lang string) {
	log.Printf("request change language: requestID: %d, lang: %s", requestID, lang)
	m.game.RespondChangeLanguage(requestID)
}

func (m *Requester) RequestGetIAPPrices(requestID int) {
	log.Printf("request IAP prices: requestID: %d", requestID)
	m.game.RespondGetIAPPrices(requestID, true, []byte("{\"ads_removal\": \"$0.99\"}"))
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
	m.game.RespondAsset(requestID, true, []byte{})
}
