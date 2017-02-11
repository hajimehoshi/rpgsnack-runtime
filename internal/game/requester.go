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

package game

import (
	"log"
	"strconv"
)

type MockRequester struct {
	game *Game
}

func (m *MockRequester) RequestUnlockAchievement(requestID int, achievementID int) {
	log.Printf("request unlock achievement: requestID: %d, achievementID: %d", requestID, achievementID)
	achievements := strconv.Itoa(achievementID)
	m.game.FinishUnlockAchievement(requestID, achievements, "")
}

func (m *MockRequester) RequestSaveProgress(requestID int, data string) {
	log.Printf("request save progress: requestID: %d", requestID)
	println(data)
	m.game.FinishSaveProgress(requestID, "")
}

func (m *MockRequester) RequestPurchase(requestID int, productID string) {
}

func (m *MockRequester) RequestInterstitialAds(requestID int) {
}

func (m *MockRequester) RequestRewardedAds(requestID int) {
}

func (m *MockRequester) RequestOpenLink(requestID int, linkType string, data string) {
}

func (m *MockRequester) RequestShareImage(requestID int, title string, message string, image string) {
}
