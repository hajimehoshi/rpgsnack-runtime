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

// +build android ios

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

func FinishPurchase(id int, success bool) {
	theGame.FinishPurchase(id, success)
}

func FinishRestorePurchases(id int, purchases []uint8) {
	theGame.FinishRestorePurchases(id, purchases)
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

func SetPlatformData(key string, value int) {
	theGame.SetPlatformData(scene.PlatformDataKey(key), int(value))
}
