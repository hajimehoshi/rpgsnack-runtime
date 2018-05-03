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

package weather

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

var (
	rainImage *ebiten.Image
	snowImage *ebiten.Image
)

func init() {
	rainImage, _ = ebiten.NewImage(1, 20, ebiten.FilterDefault)
	rainImage.Fill(color.White)
}

func init() {
	snowImage, _ = ebiten.NewImage(3, 3, ebiten.FilterDefault)
	alphas := []byte{
		0x80, 0xff, 0x80,
		0xff, 0xff, 0xff,
		0x80, 0xff, 0x80,
	}
	pix := make([]byte, len(alphas)*4)
	for i, v := range alphas {
		pix[4*i] = v
		pix[4*i+1] = v
		pix[4*i+2] = v
		pix[4*i+3] = v
	}
	snowImage.ReplacePixels(pix)
}

type sprite struct {
	weatherType data.WeatherType
	x           float64
	y           float64
	opacity     int
}

func (s *sprite) update() {
	const (
		screenWidth  = consts.TileXNum * consts.TileSize
		screenHeight = consts.TileYNum * consts.TileSize
	)
	switch s.weatherType {
	case data.WeatherTypeRain:
		s.x -= 2 * math.Sin(math.Pi/16)
		s.y += 2 * math.Cos(math.Pi/16)
		s.opacity -= 6
	case data.WeatherTypeSnow:
		s.x -= 1 * math.Sin(math.Pi/16)
		s.y += 1 * math.Cos(math.Pi/16)
		s.opacity -= 3
	}

	if s.opacity <= 0 {
		s.x = float64(rand.Intn(screenWidth+100) - 100)
		s.y = float64(rand.Intn(screenHeight+200) - 100)
		s.opacity = 160 + rand.Intn(60)
	}
}

func (s *sprite) draw(screen *ebiten.Image) {
	var img *ebiten.Image
	switch s.weatherType {
	case data.WeatherTypeRain:
		img = rainImage
	case data.WeatherTypeSnow:
		img = snowImage
	}

	op := &ebiten.DrawImageOptions{}
	if s.weatherType == data.WeatherTypeRain {
		op.GeoM.Rotate(math.Pi / 16)
	}
	op.GeoM.Translate(s.x, s.y)
	op.ColorM.Scale(1, 1, 1, float64(s.opacity)/255)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

type Weather struct {
	weatherType data.WeatherType
	sprites     []*sprite
}

func New(weatherType data.WeatherType) *Weather {
	const spriteNum = 25

	sprites := make([]*sprite, spriteNum)
	for i := range sprites {
		sprites[i] = &sprite{
			weatherType: weatherType,
		}
	}

	return &Weather{
		weatherType: weatherType,
		sprites:     sprites,
	}
}

func (w *Weather) Update() {
	if w == nil {
		return
	}
	for _, s := range w.sprites {
		s.update()
	}
}

func (w *Weather) Draw(screen *ebiten.Image) {
	if w == nil {
		return
	}
	for _, s := range w.sprites {
		s.draw(screen)
	}
}
