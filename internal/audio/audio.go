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

type byteStream struct {
	*bytes.Reader
}

func (b *byteStream) Close() error { return nil }

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

type audio struct {
	context *eaudio.Context
}

func newAudio() (*audio, error) {
	context, err := eaudio.NewContext(22050)
	if err != nil {
		return nil, err
	}
	return &audio{
		context: context,
	}, nil
}

func (a *audio) Update() error {
	return a.context.Update()
}

func (a *audio) PlaySE(path string, volume float64) error {
	bin := assets.MustAsset("audio/se/" + path + ".wav")
	s, err := wav.Decode(a.context, &byteStream{bytes.NewReader(bin)})
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
