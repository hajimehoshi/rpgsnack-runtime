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
	"fmt"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Requester scene.Requester

func FinishUnlockAchievement(id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishUnlockAchievement: %v", err)
			}
		}
	}()

	theGame.FinishUnlockAchievement(id)
	return nil
}

func FinishSaveProgress(id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishSaveProgress: %v", err)
			}
		}
	}()

	theGame.FinishSaveProgress(id)
	return nil
}

func FinishPurchase(id int, success bool, purchases []uint8) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishPurchase: %v", err)
			}
		}
	}()

	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.FinishPurchase(id, success, p)
	return nil
}

func FinishShowShop(id int, success bool, purchases []uint8) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishShowShop: %v", err)
			}
		}
	}()

	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.FinishShowShop(id, success, p)
	return nil
}

func FinishRestorePurchases(id int, success bool, purchases []uint8) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishRestorePurchases: %v", err)
			}
		}
	}()

	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.FinishRestorePurchases(id, success, p)
	return nil
}

func FinishInterstitialAds(id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishInterstitialAds: %v", err)
			}
		}
	}()

	theGame.FinishInterstitialAds(id)
	return nil
}

func FinishRewardedAds(id int, success bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishRemoteAds: %v", err)
			}
		}
	}()

	theGame.FinishRewardedAds(id, success)
	return nil
}

func FinishOpenLink(id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishOpenLinks: %v", err)
			}
		}
	}()

	theGame.FinishOpenLink(id)
	return nil
}

func FinishShareImage(id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishShareImage: %v", err)
			}
		}
	}()

	theGame.FinishShareImage(id)
	return nil
}

func FinishChangeLanguage(id int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishChangeLanguage: %v", err)
			}
		}
	}()

	theGame.FinishChangeLanguage(id)
	return nil
}

func FinishGetIAPPrices(id int, success bool, prices []uint8) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at FinishGetIAPPrices: %v", err)
			}
		}
	}()

	var p []uint8
	if prices != nil {
		p = make([]uint8, len(prices))
		copy(p, prices)
	}
	theGame.FinishGetIAPPrices(id, success, p)
	return nil
}

func SetPlatformData(key string, value string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at SetPlatformData: %v", err)
			}
		}
	}()

	theGame.SetPlatformData(scene.PlatformDataKey(key), value)
	return nil
}
