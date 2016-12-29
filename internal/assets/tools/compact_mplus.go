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
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func writeInfo(lines []bool) error {
	f, err := os.Create("images/mplus.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		v := 0
		if line {
			v = 1
		}
		fmt.Fprintf(f, "%d,", v)
	}
	return nil
}

const charHeight = 16

func writeCompactedImage(lines []bool, origImg image.Image) error {
	count := 0
	for _, line := range lines {
		if !line {
			continue
		}
		count++
	}
	width := origImg.Bounds().Size().X
	img := image.NewRGBA(image.Rect(0, 0, width, count*charHeight))
	n := 0
	for i, line := range lines {
		if !line {
			continue
		}
		dst := image.Rect(0, n*charHeight, width, (n+1)*charHeight)
		src := image.Pt(0, i*charHeight)
		draw.Draw(img, dst, origImg, src, draw.Src)
		n++
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
	_, height := img.Bounds().Size().X, img.Bounds().Size().Y
	lines := make([]bool, height/charHeight)
	for j := 0; j < height; j++ {
		for i := 0; i < height; i++ {
			c := img.At(i, j)
			_, _, _, a := c.RGBA()
			if a == 0 {
				continue
			}
			lines[j/charHeight] = true
			j += charHeight - j%charHeight
		}
	}
	if err := writeCompactedImage(lines, img); err != nil {
		return err
	}
	if err := writeInfo(lines); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
