// Copyright 2016 Hajime Hoshi
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

package main

import (
	"bufio"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sort"
)

type runeSort []rune

func (r runeSort) Len() int           { return len(r) }
func (r runeSort) Less(i, j int) bool { return r[i] < r[j] }
func (r runeSort) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func writeInfo(positions map[rune]int) error {
	f, err := os.Create("images/mplus_positions")
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	runes := []rune{}
	for r := range positions {
		runes = append(runes, r)
	}
	sort.Sort(runeSort(runes))
	for _, r := range runes {
		pos := positions[r]
		w.WriteByte(uint8(r))
		w.WriteByte(uint8(r >> 8))
		w.WriteByte(uint8(pos))
		w.WriteByte(uint8(pos >> 8))
	}
	return nil
}

const (
	charWidth  = 12
	charHeight = 16
)

func writeCompactedImage(positions map[rune]int, origImg image.Image) error {
	width := origImg.Bounds().Size().X
	nx := width / charWidth
	palette := color.Palette([]color.Color{
		color.Transparent, color.Opaque,
	})
	dstWidth := 256 * charWidth
	dstHeight := ((len(positions)-1)/256 + 1) * charHeight
	img := image.NewPaletted(image.Rect(0, 0, dstWidth, dstHeight), palette)
	for r, pos := range positions {
		dstX := (pos % 256) * charWidth
		dstY := (pos / 256) * charHeight
		dst := image.Rect(dstX, dstY, dstX+charWidth, dstY+charHeight)
		srcX := int(r) % nx * charWidth
		srcY := int(r) / nx * charHeight
		src := image.Pt(srcX, srcY)
		draw.Draw(img, dst, origImg, src, draw.Src)
	}
	f, err := os.Create("images/mplus.compacted.png")
	if err != nil {
		return err
	}
	defer f.Close()
	e := &png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	if err := e.Encode(f, img); err != nil {
		return err
	}
	return nil
}

func run() error {
	f, err := os.Open("images/mplus.png")
	if err != nil {
		return err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return err
	}
	width, height := img.Bounds().Size().X, img.Bounds().Size().Y
	positions := map[rune]int{}
	count := 0
	charXNum, charYNum := height/charHeight, width/charWidth
	for j := 0; j < charYNum; j++ {
	char:
		for i := 0; i < charXNum; i++ {
			for y := 0; y < charHeight; y++ {
				for x := 0; x < charWidth; x++ {
					c := img.At(i*charWidth+x, j*charHeight+y)
					if _, _, _, a := c.RGBA(); a != 0 {
						positions[rune(j*charXNum+i)] = count
						count++
						continue char
					}
				}
			}
		}
	}
	if err := writeCompactedImage(positions, img); err != nil {
		return err
	}
	if err := writeInfo(positions); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
