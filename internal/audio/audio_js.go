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

// +build js

package audio

import (
	"errors"
	"fmt"

	"github.com/gopherjs/gopherjs/js"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
)

type audio struct {
	context   *js.Object
	bgmSource *js.Object
	bgmGain   *js.Object
	bgmName   string
	dataCache map[string][]byte
}

var theCurrentAudio = &audio{
	dataCache: map[string][]byte{},
}

func init() {
	class := js.Global.Get("AudioContext")
	if class == js.Undefined {
		class = js.Global.Get("webkitAudioContext")
	}
	if class == js.Undefined {
		panic("audio: AudioContext is not available")
	}
	theCurrentAudio.context = class.New()

	var f *js.Object
	f = js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		node := theCurrentAudio.context.Call("createBufferSource")
		node.Call("start", 0)

		js.Global.Get("document").Call("removeEventListener", "mouseup", f)
		js.Global.Get("document").Call("removeEventListener", "touchend", f)
		return nil
	})
	js.Global.Get("document").Call("addEventListener", "mouseup", f)
	js.Global.Get("document").Call("addEventListener", "touchend", f)
}

func seekNextFrame(buf []byte) ([]byte, bool) {
	// TODO: Need to skip tags explicitly? (hajimehoshi/go-mp3#9)

	if len(buf) < 1 {
		return nil, false
	}
	buf = buf[1:]

	for {
		if buf[0] == 0xff && buf[1]&0xfe == 0xfe {
			break
		}
		buf = buf[1:]
		if len(buf) < 2 {
			return nil, false
		}
	}
	return buf, true
}

var errTryAgain = errors.New("try again")

func (a *audio) decode(data []byte) (*js.Object, error) {
	ch := make(chan error)
	var buf *js.Object
	a.context.Call("decodeAudioData", js.NewArrayBuffer(data), func(buffer *js.Object) {
		buf = buffer
		close(ch)
	}, func(err *js.Object) {
		if err != nil {
			ch <- fmt.Errorf("audio: decodeAudioData failed: %v", err)
		} else {
			// On Safari, error value might be null and it is needed to retry decoding
			// from the next frame.
			ch <- errTryAgain
		}
		close(ch)
	})
	if err := <-ch; err != nil {
		return nil, err
	}
	return buf, nil
}

func (a *audio) createSource(group string, name string, volume float64, loop bool) (*js.Object, error) {
	data, ok := a.dataCache[name]
	if !ok {
		mp3Path := fmt.Sprintf("audio/%s/%s.mp3", group, name)
		wavPath := fmt.Sprintf("audio/%s/%s.wav", group, name)

		switch {
		case assets.Exists(mp3Path):
			data = assets.GetResource(mp3Path)
		case assets.Exists(wavPath):
			data = assets.GetResource(wavPath)
		default:
			return nil, fmt.Errorf("audio: invalid audio format: %s", name)
		}
		a.dataCache[name] = data
	}
	n := a.context.Call("createBufferSource")
	n.Set("loop", loop)

	var buffer *js.Object
	for {
		var err error
		buffer, err = a.decode(data)
		if err == errTryAgain {
			d, ok := seekNextFrame(data)
			if !ok {
				return nil, fmt.Errorf("audio: Decode failed: invalid format?")
			}
			data = d
			continue
		}
		if err != nil {
			return nil, err
		}
		break
	}

	n.Set("buffer", buffer)
	return n, nil
}

func Update() error {
	return nil
}

func Stop() error {
	StopBGM()
	return nil
}

func PlaySE(name string, volume float64) error {
	n, err := theCurrentAudio.createSource("se", name, volume, false)
	if err != nil {
		return err
	}

	g := theCurrentAudio.context.Call("createGain")
	g.Get("gain").Set("value", volume)
	n.Call("connect", g)
	g.Call("connect", theCurrentAudio.context.Get("destination"))
	n.Call("start", 0)
	return nil
}

func PlayBGM(name string, volume float64) error {
	StopBGM()

	n, err := theCurrentAudio.createSource("bgm", name, volume, true)
	if err != nil {
		return err
	}

	g := theCurrentAudio.context.Call("createGain")
	g.Get("gain").Set("value", volume)
	n.Call("connect", g)
	g.Call("connect", theCurrentAudio.context.Get("destination"))
	n.Call("start", 0)

	theCurrentAudio.bgmSource = n
	theCurrentAudio.bgmGain = g
	theCurrentAudio.bgmName = name
	return nil
}

func PlayingBGMName() string {
	return theCurrentAudio.bgmName
}

func PlayingBGMVolume() float64 {
	if theCurrentAudio.bgmSource == nil {
		return 0
	}
	return theCurrentAudio.bgmGain.Get("gain").Get("volume").Float()
}

func StopBGM() error {
	if theCurrentAudio.bgmSource == nil {
		return nil
	}
	theCurrentAudio.bgmSource.Call("stop", 0)
	theCurrentAudio.bgmSource = nil
	theCurrentAudio.bgmName = ""
	return nil
}
