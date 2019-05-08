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

package audio

var volumeBias float64 = 0.8
var seVolumeBias float64 = 1.0
var bgmVolumeBias float64 = 1.0

func SetSEVolume(v float64) {
	seVolumeBias = v
}

func SetBGMVolume(v float64) {
	bgmVolumeBias = v
}

func SEVolume() float64 {
	return seVolumeBias
}

func BGMVolume() float64 {
	return bgmVolumeBias
}

func setMasterVolume(v float64) {
	volumeBias = v
}

func ToggleMute() {
	if volumeBias == 0 {
		setMasterVolume(0.8)
		return
	}
	setMasterVolume(0)
}
