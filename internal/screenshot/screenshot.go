// Copyright 2018 Hajime Hoshi
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

package screenshot

import (
	"bytes"
	"image/color"
	"image/png"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

type screenshotData struct {
	width  int
	height int
	lang   language.Tag
}

type Screenshot struct {
	screenshots       []*screenshotData
	currentScreenshot *screenshotData
	screenshotCount   int
	origLang          language.Tag
	finished          bool
	pseudoScreen      *ebiten.Image
}

type Size struct {
	Width  int
	Height int
}

func New(sizes []Size, langs []language.Tag) *Screenshot {
	s := &Screenshot{
		origLang: lang.Get(),
	}
	for _, size := range sizes {
		for _, l := range langs {
			s.screenshots = append(s.screenshots,
				&screenshotData{
					width:  size.Width,
					height: size.Height,
					lang:   l,
				})
		}
	}
	return s
}

type SceneManager interface {
	ResetPseudoScreen()
	SetPseudoScreen(screen *ebiten.Image)
	SetLanguage(lang language.Tag) language.Tag
}

func (s *Screenshot) Update(sceneManager SceneManager) {
	if len(s.screenshots) == 0 {
		if !s.finished {
			sceneManager.ResetPseudoScreen()
			sceneManager.SetLanguage(s.origLang)
			s.finished = true
		}
		return
	}

	if sc := s.screenshots[0]; sc != s.currentScreenshot {
		if s.pseudoScreen != nil {
			s.pseudoScreen.Dispose()
			s.pseudoScreen = nil
		}
		s.pseudoScreen, _ = ebiten.NewImage(sc.width, sc.height, ebiten.FilterDefault)
		sceneManager.SetPseudoScreen(s.pseudoScreen)
		sceneManager.SetLanguage(sc.lang)
		s.screenshotCount = 0
		s.currentScreenshot = sc
	}
	s.screenshotCount++
}

func (s *Screenshot) IsFinished() bool {
	return s.finished
}

func (s *Screenshot) TryDump() ([]byte, Size, language.Tag, error) {
	// s.screenshotCount >= 2 is necessary to assure that update is done.
	// >= 1 is not enough for some mysterious reason.
	if len(s.screenshots) == 0 || s.screenshotCount < 2 {
		return nil, Size{}, language.Tag{}, nil
	}

	sc := s.screenshots[0]

	// Make the background black.
	img, _ := ebiten.NewImage(sc.width, sc.height, ebiten.FilterDefault)
	img.Fill(color.Black)
	img.DrawImage(s.pseudoScreen, nil)

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		return nil, Size{}, language.Tag{}, err
	}

	s.screenshots = s.screenshots[1:]
	return buf.Bytes(), Size{sc.width, sc.height}, sc.lang, nil
}
