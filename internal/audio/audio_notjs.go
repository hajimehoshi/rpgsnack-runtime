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

// +build !js

package audio

import (
	"fmt"

	eaudio "github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/audio/wav"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
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

func PlayBGM(name string, volume float64) {
	theAudio.PlayBGM(name, volume)
}

func PlayingBGMName() string {
	return theAudio.playingBGMName
}

func PlayingBGMVolume() float64 {
	return theAudio.playingBGMVolume
}

func StopBGM() {
	theAudio.StopBGM()
}

type audio struct {
	context          *eaudio.Context
	players          map[string]*eaudio.Player
	sePlayers        map[*eaudio.Player]struct{}
	playing          *eaudio.Player
	playingBGMName   string
	playingBGMVolume float64
	err              error
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
	}, nil
}

func (a *audio) Update() error {
	if a.err != nil {
		return a.err
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
	StopBGM()
	for p := range a.sePlayers {
		if err := p.Pause(); err != nil {
			a.err = err
			return
		}
	}
	a.sePlayers = map[*eaudio.Player]struct{}{}
}

func (a *audio) getPlayer(path string, loop bool) (*eaudio.Player, error) {
	mp3Path := path + ".mp3"
	wavPath := path + ".wav"
	if assets.Exists(mp3Path) {
		bin := assets.GetResource(mp3Path)
		s, err := mp3.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
		if err != nil {
			return nil, fmt.Errorf("audio: decode error: %s, %v", mp3Path, err)
		}
		if loop {
			return eaudio.NewPlayer(a.context, eaudio.NewInfiniteLoop(s, s.Size()))
		}
		return eaudio.NewPlayer(a.context, s)
	}

	if assets.Exists(wavPath) {
		bin := assets.GetResource(wavPath)
		s, err := wav.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
		if err != nil {
			return nil, fmt.Errorf("audio: decode error: %s, %v", wavPath, err)
		}
		if loop {
			return eaudio.NewPlayer(a.context, eaudio.NewInfiniteLoop(s, s.Size()))
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
	p.SetVolume(volume)
	p.Play()
	a.sePlayers[p] = struct{}{}
}

func (a *audio) PlayBGM(name string, volume float64) {
	if a.err != nil {
		return
	}
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
		a.playing.SetVolume(volume)
		return
	}
	if a.playing != nil {
		a.playing.Pause()
	}
	if err := p.Rewind(); err != nil {
		a.err = err
		return
	}
	p.SetVolume(volume)
	p.Play()
	a.playing = p
	a.playingBGMName = name
	a.playingBGMVolume = volume
}

func (a *audio) StopBGM() {
	if a.err != nil {
		return
	}
	if a.playing == nil {
		return
	}
	if err := a.playing.Pause(); err != nil {
		a.err = err
		return
	}
	a.playing = nil
	a.playingBGMName = ""
	a.playingBGMVolume = 0
}
