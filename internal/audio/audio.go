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

func Stop() error {
	return theAudio.Stop()
}

func PlaySE(path string, volume float64) error {
	return theAudio.PlaySE(path, volume)
}

func PlayBGM(path string, volume float64) error {
	return theAudio.PlayBGM(path, volume)
}

func PlayingBGMName() string {
	return theAudio.playingBGMName
}

func PlayingBGMVolume() float64 {
	return theAudio.playingBGMVolume
}

func StopBGM() error {
	return theAudio.StopBGM()
}

type audio struct {
	context          *eaudio.Context
	players          map[string]*eaudio.Player
	sePlayers        map[*eaudio.Player]struct{}
	playing          *eaudio.Player
	playingBGMName   string
	playingBGMVolume float64
}

func newAudio() (*audio, error) {
	context, err := eaudio.NewContext(22050)
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
	if err := a.context.Update(); err != nil {
		return err
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

func (a *audio) Stop() error {
	if err := StopBGM(); err != nil {
		return err
	}
	for p := range a.sePlayers {
		if err := p.Pause(); err != nil {
			return err
		}
	}
	a.sePlayers = map[*eaudio.Player]struct{}{}
	return nil
}

func (a *audio) PlaySE(name string, volume float64) error {
	bin := assets.GetResource("audio/se/" + name + ".wav")
	s, err := wav.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
	if err != nil {
		return err
	}
	p, err := eaudio.NewPlayer(a.context, s)
	if err != nil {
		return err
	}
	p.SetVolume(volume)
	p.Play()
	a.sePlayers[p] = struct{}{}
	return nil
}

func (a *audio) PlayBGM(name string, volume float64) error {
	p, ok := a.players[name]
	if !ok {
		mp3Path := "audio/bgm/" + name + ".mp3"
		wavPath := "audio/bgm/" + name + ".wav"
		var ss eaudio.ReadSeekCloser
		if assets.Exists(mp3Path) {
			bin := assets.GetResource(mp3Path)
			s, err := mp3.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
			if err != nil {
				return fmt.Errorf("audio: decode error: %s, %v", mp3Path, err)
			}
			ss = eaudio.NewInfiniteLoop(s, s.Size())
		} else if assets.Exists(wavPath) {
			bin := assets.GetResource(wavPath)
			s, err := wav.Decode(a.context, eaudio.BytesReadSeekCloser(bin))
			if err != nil {
				return fmt.Errorf("audio: decode error: %s, %v", wavPath, err)
			}
			ss = eaudio.NewInfiniteLoop(s, s.Size())
		} else {
			return fmt.Errorf("audio: %s not found", name)
		}
		player, err := eaudio.NewPlayer(a.context, ss)
		if err != nil {
			return err
		}
		a.players[name] = player
		p = player
	}
	if a.playingBGMName == name {
		a.playing.SetVolume(volume)
		return nil
	} else if a.playing != nil {
		a.playing.Pause()
	}
	if err := p.Rewind(); err != nil {
		return err
	}
	p.Play()
	p.SetVolume(volume)
	a.playing = p
	a.playingBGMName = name
	a.playingBGMVolume = volume
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
	a.playingBGMName = ""
	a.playingBGMVolume = 0
	return nil
}
