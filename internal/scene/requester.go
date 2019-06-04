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

package scene

type Requester interface {
	RequestUnlockAchievement(requestID int, achievementID int)
	RequestSaveProgress(requestID int, data []byte)
	RequestSavePermanent(requestID int, data []byte)
	RequestPurchase(requestID int, productID string)
	RequestShowShop(requestID int, data string)
	RequestRestorePurchases(requestID int)
	RequestInterstitialAds(requestID int, forceAds bool)
	RequestRewardedAds(requestID int, forceAds bool)
	RequestOpenURL(requestID int, url string)
	RequestOpenApp(requestID int, appName string, appData string)
	RequestOpenNews(requestID int, appID string)
	RequestOpenReview(requestID int)
	RequestAdsInitialize()
	RequestShareImage(requestID int, title string, message string, image []byte)
	RequestTerminateGame()
	RequestChangeLanguage(requestID int, lang string)
	RequestReview()
	RequestSendAnalytics(eventName string, value string)
	RequestVibration(vibrationType string)
	RequestAsset(requestID int, key string)
	RequestMarkNewsRead(newsID int64)
}

type RequestType int

const (
	RequestTypeUnlockAchievement RequestType = iota
	RequestTypeSaveProgress
	RequestTypeSavePermanent
	RequestTypePurchase
	RequestTypeShowShop
	RequestTypeRestorePurchases
	RequestTypeInterstitialAds
	RequestTypeRewardedAds
	RequestTypeOpenURL
	RequestTypeOpenApp
	RequestTypeOpenNews
	RequestTypeOpenReview
	RequestTypeShareImage
	RequestTypeChangeLanguage
	RequestTypeAsset
)

type RequestResult struct {
	ID        int
	Type      RequestType
	Succeeded bool
	Data      []byte
}

type requesterImpl struct {
	Requester

	// lastSaveData holds the last save data so as not to be GCed
	// before processing on the mobile side finishes.
	lastSaveData      []byte
	lastSavePermanent []byte
	lastShareImage    []byte
}

func (r *requesterImpl) RequestSaveProgress(requestID int, data []byte) {
	r.lastSaveData = data
	r.Requester.RequestSaveProgress(requestID, data)
}

func (r *requesterImpl) RequestSavePermanent(requestID int, data []byte) {
	// Passing data causes "cgo argument has Go pointer to Go pointer" error. We don't know why.
	bs := make([]byte, len(data))
	copy(bs, data)
	r.lastSavePermanent = bs
	r.Requester.RequestSavePermanent(requestID, bs)
}

func (r *requesterImpl) RequestShareImage(requestID int, title string, message string, image []byte) {
	r.lastShareImage = image
	r.Requester.RequestShareImage(requestID, title, message, image)
}
