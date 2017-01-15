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
	"bytes"

	eaudio "github.com/hajimehoshi/ebiten/audio"
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

func PlaySE(path string, volume float64) error {
	return theAudio.PlaySE(path, volume)
}

func PlayBGM(path string, volume float64) error {
	return theAudio.PlayBGM(path, volume)
}

func StopBGM() error {
	return theAudio.StopBGM()
}

type audio struct {
	context     *eaudio.Context
	players     map[string]*eaudio.Player
	playing     *eaudio.Player
	playingName string
}

func newAudio() (*audio, error) {
	context, err := eaudio.NewContext(22050)
	if err != nil {
		return nil, err
	}
	return &audio{
		context: context,
		players: map[string]*eaudio.Player{},
	}, nil
}

func (a *audio) Update() error {
	return a.context.Update()
}

func (a *audio) PlaySE(name string, volume float64) error {
	bin := assets.MustAsset("audio/se/" + name + ".wav")
	s, err := wav.Decode(a.context, eaudio.NopCloser(bytes.NewReader(bin)))
	if err != nil {
		return err
	}
	p, err := eaudio.NewPlayer(a.context, s)
	if err != nil {
		return err
	}
	p.SetVolume(volume)
	return p.Play()
}

func (a *audio) PlayBGM(name string, volume float64) error {
	p, ok := a.players[name]
	if !ok {
		bin := assets.MustAsset("audio/bgm/" + name + ".wav")
		s, err := wav.Decode(a.context, eaudio.NopCloser(bytes.NewReader(bin)))
		if err != nil {
			return err
		}
		ss := eaudio.NewInfiniteLoop(s, s.Size())
		player, err := eaudio.NewPlayer(a.context, ss)
		if err != nil {
			return err
		}
		a.players[name] = player
		p = player
	}
	if a.playingName == name {
		a.playing.SetVolume(volume)
		return nil
	}
	if err := p.Rewind(); err != nil {
		return err
	}
	if err := p.Play(); err != nil {
		return err
	}
	p.SetVolume(volume)
	a.playing = p
	a.playingName = name
	return nil
}

func (a *audio) StopBGM() error {
	if a.playing == nil {
		return nil
	}
	if err := a.playing.Pause(); err != nil {
		return err
	}
	a.playing = nil
	a.playingName = ""
	return nil
}
