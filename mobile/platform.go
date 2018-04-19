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

package mobile

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Requester scene.Requester

func FinishUnlockAchievement(id int) {
	theGame.FinishUnlockAchievement(id)
}

func FinishSaveProgress(id int) {
	theGame.FinishSaveProgress(id)
}

func FinishPurchase(id int, success bool, purchases []uint8) {
	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.FinishPurchase(id, success, p)
}

func FinishShowShop(id int, success bool, purchases []uint8) {
	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.FinishShowShop(id, success, p)
}

func FinishRestorePurchases(id int, success bool, purchases []uint8) {
	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.FinishRestorePurchases(id, success, p)
}

func FinishInterstitialAds(id int) {
	theGame.FinishInterstitialAds(id)
}

func FinishRewardedAds(id int, success bool) {
	theGame.FinishRewardedAds(id, success)
}

func FinishOpenLink(id int) {
	theGame.FinishOpenLink(id)
}

func FinishShareImage(id int) {
	theGame.FinishShareImage(id)
}

func FinishChangeLanguage(id int) {
	theGame.FinishChangeLanguage(id)
}

func FinishGetIAPPrices(id int, success bool, prices []uint8) {
	var p []uint8
	if prices != nil {
		p = make([]uint8, len(prices))
		copy(p, prices)
	}
	theGame.FinishGetIAPPrices(id, success, p)
}

func SetPlatformData(key string, value string) {
	theGame.SetPlatformData(scene.PlatformDataKey(key), value)
}
