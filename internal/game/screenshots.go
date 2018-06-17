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

package game

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

type screenshotData struct {
	width  int
	height int
	lang   language.Tag
}

type screenshots struct {
	screenshots       []*screenshotData
	currentScreenshot *screenshotData
	screenshotDir     string
	screenshotCount   int
	origWidth         int
	origHeight        int
	origLang          language.Tag
	finished          bool
}

func newScreenshots(width, height int, langs []language.Tag) *screenshots {
	s := &screenshots{
		screenshotDir: filepath.Join("screenshots", time.Now().Format("20060102_030405")),
		origWidth:     width,
		origHeight:    height,
		origLang:      lang.Get(),
	}
	for _, sc := range []struct {
		width  int
		height int
	}{
		{
			width:  480, // 1242
			height: 720, // 2208
		},
		{
			width:  480, // 2048
			height: 854, // 2732
		},
		{
			width:  480,  // 1125
			height: 1040, // 2436
		},
	} {
		for _, l := range langs {
			s.screenshots = append(s.screenshots,
				&screenshotData{
					width:  sc.width,
					height: sc.height,
					lang:   l,
				})
		}
	}
	return s
}

func (s *screenshots) update(game *Game) {
	if len(s.screenshots) == 0 {
		if !s.finished {
			game.SetScreenSize(s.origWidth, s.origHeight, ebiten.ScreenScale())
			game.sceneManager.SetLanguage(s.origLang)
			s.finished = true
		}
		return
	}

	if sc := s.screenshots[0]; sc != s.currentScreenshot {
		game.SetScreenSize(sc.width, sc.height, ebiten.ScreenScale())
		game.sceneManager.SetLanguage(sc.lang)
		s.screenshotCount = 0
		s.currentScreenshot = sc
	}
	s.screenshotCount++
}

func (s *screenshots) isFinished() bool {
	return s.finished
}

func (s *screenshots) tryDumpScreenshots(screen *ebiten.Image) error {
	// s.screenshotCount >= 2 is necessary to assure that update is done.
	// >= 1 is not enough for some mysterious reason.
	if len(s.screenshots) == 0 || s.screenshotCount < 2 {
		return nil
	}

	sc := s.screenshots[0]

	if err := os.MkdirAll(s.screenshotDir, 0755); err != nil {
		return err
	}

	fn := filepath.Join(s.screenshotDir, fmt.Sprintf("%d-%d-%s.png", sc.width, sc.height, sc.lang))
	fmt.Println(fn)

	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _ := ebiten.NewImage(sc.width, sc.height, ebiten.FilterDefault)
	img.Fill(color.Black)
	img.DrawImage(screen, nil)

	if err := png.Encode(f, img); err != nil {
		return err
	}

	s.screenshots = s.screenshots[1:]
	return nil
}
