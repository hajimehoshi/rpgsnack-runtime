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

func RespondUnlockAchievement(id int) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondUnlockAchievement: %v", err)
			}
		}
	}()

	theGame.RespondUnlockAchievement(id)
	return nil
}

func RespondSaveProgress(id int) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondSaveProgress: %v", err)
			}
		}
	}()

	theGame.RespondSaveProgress(id)
	return nil
}

func RespondPurchase(id int, success bool, purchases []uint8) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondPurchase: %v", err)
			}
		}
	}()

	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.RespondPurchase(id, success, p)
	return nil
}

func RespondShowShop(id int, success bool, purchases []uint8) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondShowShop: %v", err)
			}
		}
	}()

	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.RespondShowShop(id, success, p)
	return nil
}

func RespondRestorePurchases(id int, success bool, purchases []uint8) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondRestorePurchases: %v", err)
			}
		}
	}()

	var p []uint8
	if purchases != nil {
		p = make([]uint8, len(purchases))
		copy(p, purchases)
	}
	theGame.RespondRestorePurchases(id, success, p)
	return nil
}

func RespondInterstitialAds(id int) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondInterstitialAds: %v", err)
			}
		}
	}()

	theGame.RespondInterstitialAds(id)
	return nil
}

func RespondRewardedAds(id int, success bool) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondRemoteAds: %v", err)
			}
		}
	}()

	theGame.RespondRewardedAds(id, success)
	return nil
}

func RespondOpenLink(id int) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondOpenLinks: %v", err)
			}
		}
	}()

	theGame.RespondOpenLink(id)
	return nil
}

func RespondShareImage(id int) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondShareImage: %v", err)
			}
		}
	}()

	theGame.RespondShareImage(id)
	return nil
}

func RespondChangeLanguage(id int) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondChangeLanguage: %v", err)
			}
		}
	}()

	theGame.RespondChangeLanguage(id)
	return nil
}

func RespondGetIAPPrices(id int, success bool, prices []uint8) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondGetIAPPrices: %v", err)
			}
		}
	}()

	var p []uint8
	if prices != nil {
		p = make([]uint8, len(prices))
		copy(p, prices)
	}
	theGame.RespondGetIAPPrices(id, success, p)
	return nil
}

func SetPlatformData(key string, value string) (err error) {
	<-startCalled

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

func RespondAsset(id int, success bool, data []uint8) (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at RespondAsset: %v", err)
			}
		}
	}()

	var d []uint8
	if data != nil {
		d = make([]uint8, len(data))
		copy(d, data)
	}
	theGame.RespondAsset(id, success, d)
	return nil
}
