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

import (
	"fmt"
	"io/ioutil"

	eaudio "github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/audio/vorbis"
	"github.com/hajimehoshi/ebiten/audio/wav"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/interpolation"
)

var theAudio = &audio{}

func init() {
	a, err := newAudio()
	if err != nil {
		panic(err)
	}
	theAudio = a
}

func Update() error {
	return theAudio.Update()
}

func Stop() {
	theAudio.Stop()
}

func PlaySE(name string, volume float64) {
	theAudio.PlaySE(name, volume)
}

func PlayBGM(name string, volume float64, fadeTimeInFrames int) {
	theAudio.PlayBGM(name, volume, fadeTimeInFrames)
}

func PlayingBGMName() string {
	return theAudio.playingBGMName
}

func PlayingBGMVolume() float64 {
	return theAudio.bgmVolume.Dst()
}

func StopBGM(fadeTimeInFrames int) {
	theAudio.StopBGM(fadeTimeInFrames)
}

func ResumeBGM() {
	theAudio.ResumeBGM()
}

func PauseBGM() {
	theAudio.PauseBGM()
}

type audio struct {
	context        *eaudio.Context
	players        map[string]*eaudio.Player
	sePlayers      map[*eaudio.Player]struct{}
	playing        *eaudio.Player
	playingBGMName string

	wavCache map[string][]byte

	bgmVolume interpolation.I

	toStopBGM bool
	paused    bool

	err error
}

func newAudio() (*audio, error) {
	context, err := eaudio.NewContext(44100)
	if err != nil {
		return nil, err
	}
	return &audio{
		context:   context,
		players:   map[string]*eaudio.Player{},
		sePlayers: map[*eaudio.Player]struct{}{},
		wavCache:  map[string][]byte{},
	}, nil
}

func (a *audio) Update() error {
	if a.err != nil {
		return a.err
	}

	if !a.paused {
		a.bgmVolume.Update()
	}
	if a.playing != nil {
		a.playing.SetVolume(a.bgmVolume.Current() * volumeBias)
	}
	if a.toStopBGM && !a.bgmVolume.IsChanging() {
		a.playing.Pause()
		if err := a.playing.Rewind(); err != nil {
			return err
		}
		a.playing = nil
		a.playingBGMName = ""
		a.toStopBGM = false
	}

	closed := []*eaudio.Player{}
	for p := range a.sePlayers {
		if !p.IsPlaying() {
			closed = append(closed, p)
		}
	}
	for _, p := range closed {
		delete(a.sePlayers, p)
	}
	return nil
}

func (a *audio) Stop() {
	if a.err != nil {
		return
	}
	StopBGM(0)
	for p := range a.sePlayers {
		p.Pause()
		// These SEs are no longer used, so positions should not be cared.
	}
	a.sePlayers = map[*eaudio.Player]struct{}{}
}

func (a *audio) getPlayer(path string, loop bool) (*eaudio.Player, error) {
	mp3Path := path + ".mp3"
	oggPath := path + ".ogg"
	wavPath := path + ".wav"
	if assets.Exists(mp3Path) {
		bin := assets.GetResource(mp3Path)
		s, err := mp3.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
		if err != nil {
			return nil, fmt.Errorf("audio: decode error: %s, %v", mp3Path, err)
		}
		if loop {
			return eaudio.NewPlayer(a.context, eaudio.NewInfiniteLoop(s, s.Length()))
		}
		return eaudio.NewPlayer(a.context, s)
	}

	if assets.Exists(oggPath) {
		bin := assets.GetResource(oggPath)
		s, err := vorbis.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
		if err != nil {
			return nil, fmt.Errorf("audio: decode error: %s, %v", oggPath, err)
		}
		if loop {
			return eaudio.NewPlayer(a.context, eaudio.NewInfiniteLoop(s, s.Length()))
		}
		return eaudio.NewPlayer(a.context, s)
	}

	if assets.Exists(wavPath) {
		if _, ok := a.wavCache[wavPath]; !ok {
			bin := assets.GetResource(wavPath)
			s, err := wav.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
			if err != nil {
				return nil, fmt.Errorf("audio: decode error: %s, %v", wavPath, err)
			}
			bs, err := ioutil.ReadAll(s)
			if err != nil {
				return nil, err
			}
			a.wavCache[wavPath] = bs
		}
		bs := a.wavCache[wavPath]
		s := eaudio.BytesReadSeekCloser(bs)
		if loop {
			return eaudio.NewPlayer(a.context, eaudio.NewInfiniteLoop(s, int64(len(bs))))
		}
		return eaudio.NewPlayer(a.context, s)
	}

	return nil, fmt.Errorf("audio: %s not found", path)
}

func (a *audio) PlaySE(name string, volume float64) {
	if a.err != nil {
		return
	}
	p, err := a.getPlayer("audio/se/"+name, false)
	if err != nil {
		a.err = err
		return
	}
	p.SetVolume(volume * volumeBias)
	p.Play()
	a.sePlayers[p] = struct{}{}
}

func (a *audio) PlayBGM(name string, volume float64, fadeTimeInFrames int) {
	if a.err != nil {
		return
	}

	a.toStopBGM = false
	a.paused = false
	a.bgmVolume.Set(volume, fadeTimeInFrames)

	p, ok := a.players[name]
	if !ok {
		player, err := a.getPlayer("audio/bgm/"+name, true)
		if err != nil {
			a.err = err
			return
		}
		a.players[name] = player
		p = player
	}
	if a.playingBGMName == name {
		p.SetVolume(a.bgmVolume.Current() * volumeBias)
		p.Play()
		return
	}
	if a.playing != nil {
		a.playing.Pause()
		if err := a.playing.Rewind(); err != nil {
			a.err = err
			return
		}
	}

	p.Play()
	p.SetVolume(a.bgmVolume.Current() * volumeBias)
	a.playing = p
	a.playingBGMName = name
}

func (a *audio) ResumeBGM() {
	if a.err != nil {
		return
	}
	if a.playing == nil {
		return
	}
	a.PlayBGM(a.playingBGMName, a.bgmVolume.Current(), 0)
}

func (a *audio) PauseBGM() {
	if a.err != nil {
		return
	}
	if a.playing == nil {
		return
	}
	a.playing.Pause()
	a.paused = true
}

func (a *audio) StopBGM(fadeTimeInFrames int) {
	if a.err != nil {
		return
	}
	if a.playing == nil {
		return
	}
	a.toStopBGM = true
	if a.paused {
		fadeTimeInFrames = 0
		a.paused = false
	}
	a.bgmVolume.Set(0, fadeTimeInFrames)
}
