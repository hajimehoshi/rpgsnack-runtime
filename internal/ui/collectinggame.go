// Copyright 2019 The RPGSnack Authors
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

package ui

import (
	"image"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

type token struct {
	x            int
	y            int
	spawnX       int
	spawnY       int
	targetX      int
	targetY      int
	flyTimer     int
	shouldDelete bool
}

const (
	flyTime          = 10
	actorPosX        = 60
	actorPosY        = 36
	actorWidth       = 24
	actorHeight      = 32
	actorFrameCount  = 4
	actorCollectTime = 6
	inputHitRadius   = 16
	tokenFallSpeed   = 3
	maxTokenCount    = 300
)

func randomValue(min, max int) int {
	return min + rand.Intn(max-min)
}

func newToken(spawnX, spawnY, targetX, targetY int, animate bool) *token {
	y := 0
	if !animate {
		y = spawnY
	}
	return &token{
		x:       spawnX,
		y:       y,
		spawnX:  spawnX,
		spawnY:  spawnY,
		targetX: targetX,
		targetY: targetY,
	}
}

func (c *token) UpdateAsChild(x, y int) {
	// Flying
	if c.flyTimer > 0 {
		c.flyTimer--
		c.x = c.spawnX + (c.targetX-c.spawnX)*(flyTime-c.flyTimer)/flyTime
		c.y = c.spawnY + (c.targetY-c.spawnY)*(flyTime-c.flyTimer)/flyTime

		if c.flyTimer == 0 {
			c.shouldDelete = true
		}
		return
	}

	// Falling
	if c.y < c.spawnY {
		c.y += tokenFallSpeed
		return
	}

	// Touch
	if input.Pressed() {
		ix, iy := input.Position()
		dx := c.x + x - ix/consts.TileScale
		dy := c.y + y - iy/consts.TileScale
		if (dx*dx + dy*dy) < inputHitRadius*inputHitRadius {
			audio.PlaySE("system/minigametouch", 1.0)
			c.flyTimer = flyTime
		}
	}
}

func (f *token) DrawAsChild(screen *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(f.x+x), float64(f.y+y))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	screen.DrawImage(assets.GetImage("system/minigame/token.png"), op)
}

const (
	actorFrameTime       = 15
	tokenSpawnTime       = 180
	tokenBoostSpawnTime  = 5
	boostTime            = 300
	maxOnlineSpawnCount  = 15
	maxOfflineSpawnCount = 50
	offlineSpawnInterval = 60 // sec
)

type collectingGame struct {
	minigameID      int
	tokens          []*token
	actorTimer      int
	collectTimer    int
	boostTimer      int
	tokenSpawnTimer int
}

func newCollectingGame() *collectingGame {
	return &collectingGame{
		tokens:          make([]*token, 0),
		tokenSpawnTimer: tokenSpawnTime - 60,
	}
}

type Minigame interface {
	ID() int
	Score() int
	ReqScore() int
	AddScore(score int)
	Active() bool
	LastActiveAt() int64
	MarkLastActive()
	Success() bool
}

func (c *collectingGame) collect(minigame Minigame) {
	c.collectTimer = actorCollectTime
	minigame.AddScore(1)
}

func (c *collectingGame) ActivateBoostMode() {
	c.boostTimer = boostTime
}

func (c *collectingGame) boosting() bool {
	return c.boostTimer > 0
}

func (c *collectingGame) spawnToken(animate bool) *token {
	if animate {
		audio.PlaySE("system/minigamespawn", 1.0)
	}
	return newToken(randomValue(10, 130), randomValue(62, 100), actorPosX+actorWidth/2, actorPosY+actorHeight/2, animate)
}

func (c *collectingGame) CanGetReward() bool {
	return len(c.tokens) < maxTokenCount && !c.boosting()
}

func (c *collectingGame) UpdateAsChild(minigame Minigame, x, y int) {
	if !minigame.Active() {
		return
	}

	if c.minigameID != minigame.ID() {
		c.tokens = []*token{}
	}
	c.minigameID = minigame.ID()

	tokens := []*token{}
	c.tokenSpawnTimer += 1

	// Auto drops of token
	if c.boostTimer > 0 {
		c.boostTimer -= 1
		if c.tokenSpawnTimer >= tokenBoostSpawnTime {
			tokens = append(tokens, c.spawnToken(true))
			c.tokenSpawnTimer = 0
		}
	} else {
		if c.tokenSpawnTimer >= tokenSpawnTime && len(c.tokens) < maxOnlineSpawnCount {
			tokens = append(tokens, c.spawnToken(true))
			c.tokenSpawnTimer = 0
		}
	}

	// Populate offline spawns
	offlineSpawnCount := 0
	if minigame.LastActiveAt() > 0 {
		offlineSpawnCount = int(time.Now().Unix()-minigame.LastActiveAt()) / offlineSpawnInterval
	}
	for i := 0; i < offlineSpawnCount; i++ {
		if len(tokens) >= maxOfflineSpawnCount {
			break
		}
		tokens = append(tokens, c.spawnToken(false))
	}
	minigame.MarkLastActive()

	for _, token := range c.tokens {
		if token.shouldDelete {
			c.collect(minigame)
		} else {
			token.UpdateAsChild(x, y)
			tokens = append(tokens, token)
		}
	}
	c.tokens = tokens

	if c.collectTimer > 0 {
		c.collectTimer -= 1
	}

	c.actorTimer = (c.actorTimer + 1) % (actorFrameTime * actorFrameCount)
}

func (c *collectingGame) DrawAsChild(screen *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x+actorPosX), float64(y+actorPosY))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)

	frame := c.actorTimer / actorFrameTime
	var actorImage *ebiten.Image
	if c.boostTimer > 0 {
		actorImage = assets.GetImage("system/minigame/actorBoost.png")
	} else {
		actorImage = assets.GetImage("system/minigame/actor.png")
	}
	if c.collectTimer > 0 {
		screen.DrawImage(assets.GetImage("system/minigame/actorCollect.png"), op)
	} else {
		screen.DrawImage(actorImage.SubImage(image.Rect(actorWidth*frame, 0, actorWidth*(frame+1), actorHeight)).(*ebiten.Image), op)
	}

	for _, token := range c.tokens {
		token.DrawAsChild(screen, x, y)
	}
}
