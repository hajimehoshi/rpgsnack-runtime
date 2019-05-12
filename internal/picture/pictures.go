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

package picture

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/interpolation"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/tint"
)

type tintingImageCacheKey struct {
	picture *picture

	// colorM represents color matrix value.
	// ebiten.ColorM cannot be used as a key directly so far.
	// See https://github.com/hajimehoshi/ebiten/issues/866
	colorM [20]float64
}

// tintingImageCache is an image cache with color matrix information to reduce graphics operations.
var tintingImageCache = map[tintingImageCacheKey]*ebiten.Image{}

type Pictures struct {
	pictures []*picture
}

func (p *Pictures) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("pictures")
	e.BeginArray()
	for _, pic := range p.pictures {
		e.EncodeInterface(pic)
	}
	e.EndArray()

	e.EndMap()
	return e.Flush()
}

func (p *Pictures) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)

	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "pictures":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				p.pictures = make([]*picture, n)
				for i := 0; i < n; i++ {
					if !d.SkipCodeIfNil() {
						p.pictures[i] = &picture{}
						d.DecodeInterface(p.pictures[i])
					}
				}
			}
		}
	}

	if err := d.Error(); err != nil {
		return fmt.Errorf("pictures: Pictures.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (p *Pictures) ensurePictures(id int) {
	if len(p.pictures) < id+1 {
		p.pictures = append(p.pictures, make([]*picture, id+1-len(p.pictures))...)
	}
}

func (p *Pictures) TouchingPictureID(x, y int) int {
	tx := float64(x)
	ty := float64(y)
	for i := range p.pictures {
		id := len(p.pictures) - 1 - i
		pic := p.pictures[id]
		if pic == nil || pic.image == nil || !pic.touchable {
			continue
		}
		sx, sy := pic.image.Size()
		var m ebiten.GeoM
		m.Translate(-pic.x.Current(), -pic.y.Current())
		m.Rotate(-pic.angle.Current())
		m.Scale(1.0/pic.scaleX.Current(), 1.0/pic.scaleY.Current())
		m.Translate(math.Floor(((1+pic.originX)*float64(sx))/2), math.Floor(((1+pic.originY)*float64(sy))/2))

		nx, ny := m.Apply(tx, ty)
		if 0 <= nx && nx <= float64(sx) && 0 <= ny && ny <= float64(sy) {
			return id
		}
	}
	return 0
}

func (p *Pictures) MoveTo(id int, x, y int, count int) {
	p.ensurePictures(id)
	if p.pictures[id] == nil {
		return
	}
	p.pictures[id].moveTo(x, y, count)
}

func (p *Pictures) Scale(id int, scaleX, scaleY float64, count int) {
	p.ensurePictures(id)
	if p.pictures[id] == nil {
		return
	}
	p.pictures[id].scale(scaleX, scaleY, count)
}

func (p *Pictures) Rotate(id int, angle float64, count int) {
	p.ensurePictures(id)
	if p.pictures[id] == nil {
		return
	}
	p.pictures[id].rotate(angle, count)
}

func (p *Pictures) Fade(id int, opacity float64, count int) {
	p.ensurePictures(id)
	if p.pictures[id] == nil {
		return
	}
	p.pictures[id].fade(opacity, count)
}

func (p *Pictures) Tint(id int, red, green, blue, gray float64, count int) {
	p.ensurePictures(id)
	if p.pictures[id] == nil {
		return
	}
	p.pictures[id].setTint(red, green, blue, gray, count)
}

func (p *Pictures) ChangeImage(id int, imageName string) {
	p.ensurePictures(id)
	if p.pictures[id] == nil {
		return
	}
	p.pictures[id].changeImage(imageName)
}

func (p *Pictures) Update() {
	for _, pic := range p.pictures {
		if pic == nil {
			continue
		}
		pic.update()
	}
}

func (p *Pictures) Draw(screen *ebiten.Image, offsetX, offsetY int, priority data.PicturePriorityType) {
	for _, pic := range p.pictures {
		if pic == nil {
			continue
		}
		if pic.priority == priority {
			pic.draw(screen, offsetX, offsetY)
		}
	}
}

func (p *Pictures) Add(id int, name string, x, y int, scaleX, scaleY, angle, opacity float64, originX, originY float64, blendType data.ShowPictureBlendType, priority data.PicturePriorityType, touchable bool) {
	p.ensurePictures(id)
	var image *ebiten.Image
	if name != "" {
		image = assets.GetLocalizedImage("pictures/" + name)
	}
	p.pictures[id] = &picture{
		imageName: name,
		image:     image,
		x:         interpolation.New(float64(x)),
		y:         interpolation.New(float64(y)),
		scaleX:    interpolation.New(scaleX),
		scaleY:    interpolation.New(scaleY),
		angle:     interpolation.New(angle),
		opacity:   interpolation.New(opacity),
		originX:   originX,
		originY:   originY,
		blendType: blendType,
		priority:  priority,
		touchable: touchable,
	}
}

func (p *Pictures) Remove(id int) {
	p.ensurePictures(id)
	p.pictures[id] = nil
}

type picture struct {
	imageName string
	image     *ebiten.Image
	x         *interpolation.I
	y         *interpolation.I
	scaleX    *interpolation.I
	scaleY    *interpolation.I
	angle     *interpolation.I
	opacity   *interpolation.I
	tint      tint.Tint
	originX   float64
	originY   float64
	blendType data.ShowPictureBlendType
	priority  data.PicturePriorityType
	touchable bool
}

func (p *picture) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("imageName")
	e.EncodeString(p.imageName)

	e.EncodeString("x")
	e.EncodeInterface(p.x)

	e.EncodeString("y")
	e.EncodeInterface(p.y)

	e.EncodeString("scaleX")
	e.EncodeInterface(p.scaleX)

	e.EncodeString("scaleY")
	e.EncodeInterface(p.scaleY)

	e.EncodeString("angle")
	e.EncodeInterface(p.angle)

	e.EncodeString("opacity")
	e.EncodeInterface(p.opacity)

	e.EncodeString("tint")
	e.EncodeInterface(&p.tint)

	e.EncodeString("originX")
	e.EncodeFloat64(p.originX)

	e.EncodeString("originY")
	e.EncodeFloat64(p.originY)

	e.EncodeString("blendType")
	e.EncodeString(string(p.blendType))

	e.EncodeString("priority")
	e.EncodeString(string(p.priority))

	e.EncodeString("touchable")
	e.EncodeBool(p.touchable)

	e.EndMap()
	return e.Flush()
}

func (p *picture) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)

	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "imageName":
			p.imageName = d.DecodeString()
			if p.imageName != "" {
				p.image = assets.GetLocalizedImage("pictures/" + p.imageName)
			}
		case "x":
			p.x = &interpolation.I{}
			d.DecodeInterface(p.x)
		case "y":
			p.y = &interpolation.I{}
			d.DecodeInterface(p.y)
		case "scaleX":
			p.scaleX = &interpolation.I{}
			d.DecodeInterface(p.scaleX)
		case "scaleY":
			p.scaleY = &interpolation.I{}
			d.DecodeInterface(p.scaleY)
		case "angle":
			p.angle = &interpolation.I{}
			d.DecodeInterface(p.angle)
		case "opacity":
			p.opacity = &interpolation.I{}
			d.DecodeInterface(p.opacity)
		case "tint":
			d.DecodeInterface(&p.tint)
		case "originX":
			p.originX = d.DecodeFloat64()
		case "originY":
			p.originY = d.DecodeFloat64()
		case "blendType":
			p.blendType = data.ShowPictureBlendType(d.DecodeString())
		case "priority":
			p.priority = data.PicturePriorityType(d.DecodeString())
		case "touchable":
			p.touchable = d.DecodeBool()
		}
	}

	// TODO Implement Decoder
	if p.priority == "" {
		p.priority = data.PicturePriorityOverlay
	}

	if err := d.Error(); err != nil {
		return fmt.Errorf("pictures: picture.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (p *picture) moveTo(x, y int, count int) {
	p.x.Set(float64(x), count)
	p.y.Set(float64(y), count)
}

func (p *picture) scale(scaleX, scaleY float64, count int) {
	p.scaleX.Set(scaleX, count)
	p.scaleY.Set(scaleY, count)
}

func (p *picture) rotate(angle float64, count int) {
	p.angle.Set(angle, count)
}

func (p *picture) fade(opacity float64, count int) {
	p.opacity.Set(opacity, count)
}

func (p *picture) setTint(red, green, blue, gray float64, count int) {
	p.tint.Set(red, green, blue, gray, count)
}

func (p *picture) changeImage(imageName string) {
	p.imageName = imageName
	if imageName == "" {
		p.image = nil
	} else {
		p.image = assets.GetLocalizedImage("pictures/" + p.imageName)
	}
	for k := range tintingImageCache {
		if k.picture == p {
			delete(tintingImageCache, k)
		}
	}
}

func (p *picture) update() {
	p.x.Update()
	p.y.Update()
	p.scaleX.Update()
	p.scaleY.Update()
	p.angle.Update()
	p.opacity.Update()
	p.tint.Update()
}

func (p *picture) draw(screen *ebiten.Image, offsetX, offsetY int) {
	if p.image == nil {
		return
	}

	sx, sy := p.image.Size()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(math.Floor(((-1-p.originX)*float64(sx))/2), math.Floor(((-1-p.originY)*float64(sy))/2))
	op.GeoM.Scale(p.scaleX.Current(), p.scaleY.Current())
	op.GeoM.Rotate(p.angle.Current())
	op.GeoM.Translate(p.x.Current(), p.y.Current())
	op.GeoM.Translate(float64(offsetX), float64(offsetY))

	p.tint.Apply(&op.ColorM)
	if p.opacity.Current() < 1 {
		op.ColorM.Scale(1, 1, 1, p.opacity.Current())
	}

	img := p.image
	if !isDiagonal(op.ColorM) && !p.tint.IsChanging() {
		img = p.getCachedImage(op.ColorM)
		op.ColorM = ebiten.ColorM{}
	}

	switch p.blendType {
	case data.ShowPictureBlendTypeNormal:
		// Use default
	case data.ShowPictureBlendTypeAdd:
		op.CompositeMode = ebiten.CompositeModeLighter
	}

	screen.DrawImage(img, op)
}

func (p *picture) getCachedImage(cm ebiten.ColorM) *ebiten.Image {
	if p.image == nil {
		return nil
	}

	k := tintingImageCacheKey{
		picture: p,
		colorM:  colorMToFloats(cm),
	}
	if img, ok := tintingImageCache[k]; ok {
		return img
	}

	img := applyColorM(p.image, cm)
	tintingImageCache[k] = img
	// TODO: Now there is no restriction on the size of tintingImageCache. Adjust this if needed.
	return img
}

func isDiagonal(cm ebiten.ColorM) bool {
	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			if i == j {
				continue
			}
			if cm.Element(i, j) != 0 {
				return false
			}
		}
	}
	return true
}

func applyColorM(img *ebiten.Image, cm ebiten.ColorM) *ebiten.Image {
	w, h := img.Size()
	newImg, _ := ebiten.NewImage(w, h, ebiten.FilterDefault)
	op := &ebiten.DrawImageOptions{}
	op.ColorM = cm
	newImg.DrawImage(img, op)
	return newImg
}

func colorMToFloats(cm ebiten.ColorM) (es [20]float64) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			es[i*5+j] = cm.Element(i, j)
		}
	}
	return
}
