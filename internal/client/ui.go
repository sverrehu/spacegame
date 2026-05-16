package client

import (
	"fmt"
	"log"
	"maps"
	"math"
	"slices"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

var playerNameFont text.Face
var messagesFont text.Face
var scoresFont text.Face
var statusFont text.Face
var resurrectFont text.Face
var helpFont text.Face

var messages GameMessages
var chatMessages GameMessages

var lastThrust = model.ThrustNone
var lastTurn = model.TurnNone
var lastPhaser = false
var lastBomb = false
var helpWanted = false

const width, height = 1500, 850

func startUI() {
	waitForWorldReady()
	initSounds()
	defer func() { teardownSounds() }()
	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("SHH Space Game").
		WithSize(width, height).
		WithContinuousRender(false))

	messages.Clear()
	chatMessages.Clear()

	var canvas *ggcanvas.Canvas
	var animToken *gogpu.AnimationToken
	var frame int

	fontSource := utils.LoadFontSource()
	defer func() { _ = fontSource.Close() }()
	playerNameFont = fontSource.Face(16)
	messagesFont = fontSource.Face(24)
	scoresFont = fontSource.Face(20)
	statusFont = fontSource.Face(20)
	resurrectFont = fontSource.Face(34)
	helpFont = fontSource.Face(26)

	app.OnDraw(func(dc *gogpu.Context) {
		if frame == 0 {
			animToken = app.StartAnimation()
		}
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}
		if canvas == nil {
			provider := app.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Fatalf("Failed to create canvas: %v", err)
			}
		}
		cw, ch := canvas.Size()
		if cw != w || ch != h {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("Resize error: %v", err)
			}
			cw, ch = w, h
		}
		err := canvas.Draw(func(cc *gg.Context) {
			renderFrame(cc)
		})
		if err != nil {
			log.Printf("Draw error: %v", err)
		}
		err = canvas.Render(dc.RenderTarget())
		if err != nil {
			log.Printf("Frame %d: Render error: %v", frame, err)
		}
		app.RequestRedraw()
		frame++
	})

	app.OnClose(func() {
		if animToken != nil {
			animToken.Stop()
		}
		gg.CloseAccelerator()
	})

	app.EventSource().OnKeyPress(onKeyPress())
	app.EventSource().OnKeyRelease(onKeyRelease())
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func onKeyPress() func(key gpucontext.Key, mods gpucontext.Modifiers) {
	return func(key gpucontext.Key, mods gpucontext.Modifiers) {
		if key == gpucontext.KeyLeft {
			server.sendTurnMessage(model.TurnLeft)
			lastTurn = model.TurnLeft
		} else if key == gpucontext.KeyRight {
			server.sendTurnMessage(model.TurnRight)
			lastTurn = model.TurnRight
		} else if key == gpucontext.KeyUp {
			server.sendThrustMessage(model.ThrustForward)
			lastThrust = model.ThrustForward
		} else if key == gpucontext.KeyDown {
			server.sendThrustMessage(model.ThrustBack)
			lastThrust = model.ThrustBack
		} else if !lastPhaser && (key == gpucontext.KeySpace || key == gpucontext.KeyLeftControl || key == gpucontext.KeyRightControl) {
			server.sendFirePhaserMessage()
			lastPhaser = true
		} else if !lastBomb && (key == gpucontext.KeyLeftAlt || key == gpucontext.KeyRightAlt || key == gpucontext.KeyLeftShift || key == gpucontext.KeyRightShift) {
			server.sendFireBombMessage()
			lastBomb = true
		} else if key == gpucontext.KeyR {
			if !myShip().IsAlive {
				server.sendResurrectMessage()
			}
		} else if key == gpucontext.KeyF1 {
			helpWanted = true
		}
	}
}

func onKeyRelease() func(key gpucontext.Key, mods gpucontext.Modifiers) {
	return func(key gpucontext.Key, mods gpucontext.Modifiers) {
		if key == gpucontext.KeyLeft && lastTurn == model.TurnLeft {
			server.sendTurnMessage(model.TurnNone)
			lastTurn = model.TurnNone
		} else if key == gpucontext.KeyRight && lastTurn == model.TurnRight {
			server.sendTurnMessage(model.TurnNone)
			lastTurn = model.TurnNone
		} else if key == gpucontext.KeyUp && lastThrust == model.ThrustForward {
			server.sendThrustMessage(model.ThrustNone)
			lastThrust = model.ThrustNone
		} else if key == gpucontext.KeyDown && lastThrust == model.ThrustBack {
			server.sendThrustMessage(model.ThrustNone)
			lastThrust = model.ThrustNone
		} else if key == gpucontext.KeySpace || key == gpucontext.KeyLeftControl || key == gpucontext.KeyRightControl {
			lastPhaser = false
		} else if key == gpucontext.KeyLeftAlt || key == gpucontext.KeyRightAlt || key == gpucontext.KeyLeftShift || key == gpucontext.KeyRightShift {
			lastBomb = false
		} else if key == gpucontext.KeyF1 {
			helpWanted = false
		}
	}
}

func renderFrame(cc *gg.Context) {
	cc.ClearWithColor(gg.RGBA2(0, 0, 0, 1))
	// world relative objects
	drawBounds(cc)
	for _, star := range world.Stars {
		drawStar(cc, star)
	}
	for _, ship := range world.Ships {
		drawShip(cc, ship)
	}
	for _, phaser := range world.Phasers {
		drawPhaser(cc, phaser)
	}
	for _, bomb := range world.Bombs {
		drawBomb(cc, bomb)
	}
	for _, bombPack := range world.BombPacks {
		drawBombPack(cc, bombPack)
	}
	for _, explosion := range world.Explosions {
		drawExplosion(cc, explosion)
	}
	// screen relative objects
	drawRadar(cc)
	drawStatus(cc)
	drawScores(cc)
	drawAllMessages(cc)
	drawResurrectMessage(cc)
	drawHelp(cc)
}

func drawBounds(cc *gg.Context) {
	cc.SetRGB(0.5, 0.5, 0.5)
	cc.MoveTo(rX(0), rY(0))
	cc.LineTo(rX(world.Width-1), rY(0))
	cc.LineTo(rX(world.Width-1), rY(world.Height-1))
	cc.LineTo(rX(0), rY(world.Height-1))
	cc.LineTo(rX(0), rY(0))
	_ = cc.Stroke()
}

func drawStar(cc *gg.Context, star *model.Star) {
	cc.SetRGB(star.Color.R, star.Color.G, star.Color.B)
	cc.DrawPoint(rX(star.Position.X), rY(star.Position.Y), 1.1)
	_ = cc.Fill()
}

func drawShip(cc *gg.Context, ship *model.Ship) {
	if !ship.IsAlive {
		return
	}
	worldShape := ship.GetWorldRelativeShape()
	cc.SetRGB(ship.Color.R, ship.Color.G, ship.Color.B)
	cc.MoveTo(rX(worldShape[0].X), rY(worldShape[0].Y))
	for _, point := range worldShape[1:] {
		cc.LineTo(rX(point.X), rY(point.Y))
	}
	cc.ClosePath()
	_ = cc.Fill()
	if len(ship.Name) > 0 {
		cc.SetRGB(0.39, 0.39, 1)
		cc.SetFont(playerNameFont)
		cc.DrawStringAnchored(ship.Name, rX(ship.Position.X), rY(ship.Position.Y-17), 0.5, 1)
	}
}

func drawPhaser(cc *gg.Context, phaser *model.Phaser) {
	cc.SetRGB(phaser.Color.R, phaser.Color.G, phaser.Color.B)
	cc.DrawPoint(rX(phaser.Position.X), rY(phaser.Position.Y), 1.3)
	_ = cc.Fill()
}

func drawBomb(cc *gg.Context, bomb *model.Bomb) {
	cc.SetRGB(1, 1, 0)
	diameter := 7.0
	r := (diameter - 2.0) / 2.0
	if bomb.Flip {
		cc.DrawLine(rX(bomb.Position.X)-r-2, rY(bomb.Position.Y), rX(bomb.Position.X)+r+2, rY(bomb.Position.Y))
		cc.DrawLine(rX(bomb.Position.X), rY(bomb.Position.Y)-r-2, rX(bomb.Position.X), rY(bomb.Position.Y)+r+2)
	} else {
		cc.DrawLine(rX(bomb.Position.X)-r-1, rY(bomb.Position.Y)-r-1, rX(bomb.Position.X)+r+1, rY(bomb.Position.Y)+r+1)
		cc.DrawLine(rX(bomb.Position.X)-r-1, rY(bomb.Position.Y)+r+1, rX(bomb.Position.X)+r+1, rY(bomb.Position.Y)-r-1)
	}
	_ = cc.Stroke()
	cc.SetRGB(bomb.Color.R, bomb.Color.G, bomb.Color.B)
	cc.DrawCircle(rX(bomb.Position.X), rY(bomb.Position.Y), r)
	_ = cc.Fill()
}

func drawBombPack(cc *gg.Context, bombPack *model.BombPack) {
	diameter := 7.0
	d2 := diameter - 2.0
	r := d2 / 2.0
	cc.SetRGB(bombPack.Color.R, bombPack.Color.G, bombPack.Color.B)
	cc.DrawCircle(rX(bombPack.Position.X-d2), rY(bombPack.Position.Y-d2), r)
	cc.DrawCircle(rX(bombPack.Position.X), rY(bombPack.Position.Y-d2), r)
	cc.DrawCircle(rX(bombPack.Position.X-r), rY(bombPack.Position.Y), r)
	_ = cc.Fill()
}

func drawExplosion(cc *gg.Context, explosion *model.Explosion) {
	cc.SetRGB(0.8, 0.8, 1.0)
	cc.DrawCircle(rX(explosion.Position.X), rY(explosion.Position.Y), explosion.OuterRadius)
	_ = cc.Fill()
	if explosion.InnerRadius > 0 {
		cc.SetRGB(0, 0, 0)
		cc.DrawCircle(rX(explosion.Position.X), rY(explosion.Position.Y), explosion.InnerRadius)
		_ = cc.Fill()
	}
}

func drawAllMessages(cc *gg.Context) {
	drawMessages(cc, &messages, 1.0, 0.78, 0.0, 0)
	drawMessages(cc, &chatMessages, 1.0, 1.0, 0.0, height/2)
}

func drawMessages(cc *gg.Context, messages *GameMessages, r, g, b float64, startY float64) {
	lineHeight := messagesFont.Metrics().Ascent + messagesFont.Metrics().Descent
	y := startY + lineHeight
	x := 10.0
	cc.SetRGB(r, g, b)
	cc.SetFont(messagesFont)
	for _, msg := range messages.GetMessages() {
		cc.DrawString(msg.Text, x, y)
		y += lineHeight
	}
	_ = cc.Fill()
}

func drawRadar(cc *gg.Context) {
	radarWidth := width / 7.0
	radarHeight := (radarWidth * world.Height) / world.Width
	radarX := width - radarWidth - 10
	radarY := height - radarHeight - 10
	cc.SetRGB(0, 0, 0)
	cc.DrawRectangle(radarX, radarY, radarWidth, radarHeight)
	_ = cc.Fill()
	cc.SetRGB(0, 1, 0)
	cc.DrawRectangle(radarX, radarY, radarWidth, radarHeight)
	_ = cc.Stroke()
	cc.Push() // to get rid of the clipping later
	cc.ClipRect(radarX, radarY, radarWidth, radarHeight)
	/* show a frame indicating the area seen in the window */
	cc.SetRGB(0.5, 0.5, 0.5)
	cc.DrawRectangle(radarX+((myShip().Position.X-width/2)*radarWidth)/world.Width,
		radarY+((myShip().Position.Y-height/2)*radarHeight)/world.Height,
		(width*radarWidth)/world.Width,
		(height*radarHeight)/world.Height)
	_ = cc.Stroke()
	cc.Pop() // get rid of the clipping
	// place dots for each ship
	cc.SetRGB(0, 1, 0)
	for _, ship := range world.Ships {
		if !ship.IsAlive {
			continue
		}
		x := 1 + (ship.Position.X*(radarWidth-1))/world.Width
		y := 1 + (ship.Position.Y*(radarHeight-1))/world.Height
		cc.DrawPoint(radarX+x, radarY+y, 1.3)
	}
	_ = cc.Stroke()
	// place dots for each bomb pack
	cc.SetRGB(1, 0, 0)
	for _, bombPack := range world.BombPacks {
		x := 1 + (bombPack.Position.X*(radarWidth-1))/world.Width
		y := 1 + (bombPack.Position.Y*(radarHeight-1))/world.Height
		cc.DrawPoint(radarX+x, radarY+y, 1.3)
	}
	_ = cc.Stroke()
}

func drawStatus(cc *gg.Context) {
	ascent := statusFont.Metrics().Ascent
	lineHeight := ascent + messagesFont.Metrics().Descent
	labelWidth := statusFont.Advance("Damage: ") // longest status indicator
	meterWidth := 2 * labelWidth
	x := 10.0
	x2 := x + labelWidth
	y := height - 3*lineHeight
	cc.SetRGB(0, 1, 0)
	cc.SetFont(scoresFont)
	cc.DrawString("Bombs:", x, y)
	cc.DrawString(fmt.Sprintf("%d", myShip().BombsLeft), x2, y)
	cc.DrawString("Heat:", x, y+lineHeight)
	heatPctWidth := meterWidth * math.Min(float64(myShip().PhaserHeat), 100) / 100.0
	safeHeatPctWidth := meterWidth * math.Min(float64(myShip().PhaserHeat), 75) / 100.0
	if myShip().PhaserHeat > 75 {
		cc.SetRGB(1, 0, 0)
		cc.DrawRectangle(x2, y+lineHeight-ascent, heatPctWidth, ascent)
		_ = cc.Fill()
		cc.SetRGB(0, 1, 0)
	}
	cc.DrawRectangle(x2, y+lineHeight-ascent, safeHeatPctWidth, ascent)
	_ = cc.Fill()
	cc.DrawRectangle(x2, y+lineHeight-ascent, meterWidth, ascent)
	_ = cc.Stroke()
	cc.DrawString("Damage:", x, y+2*lineHeight)
	damagePctWidth := meterWidth * math.Min(float64(myShip().Damage), 100) / 100.0
	cc.SetRGB(1, 0, 0)
	cc.DrawRectangle(x2, y+2*lineHeight-ascent, damagePctWidth, ascent)
	_ = cc.Fill()
	cc.SetRGB(0, 1, 0)
	cc.DrawRectangle(x2, y+2*lineHeight-ascent, meterWidth, ascent)
	_ = cc.Stroke()
}

func drawScores(cc *gg.Context) {
	scores := toScores(slices.Collect(maps.Values(world.Ships)))
	lines := make([]string, 0, len(scores))
	maxWidth := 0.0
	for _, score := range scores {
		s := fmt.Sprintf("%5.2f (%3d/%3d) %-15.15s", score.Ratio, score.Score, score.AntiScore, score.Name)
		w := scoresFont.Advance(s)
		if w > maxWidth {
			maxWidth = w
		}
		lines = append(lines, s)
	}
	lineHeight := scoresFont.Metrics().Ascent + messagesFont.Metrics().Descent
	x := width - maxWidth
	y := lineHeight
	cc.SetRGB(1, 1, 0)
	cc.SetFont(scoresFont)
	for _, line := range lines {
		cc.DrawString(line, x, y)
		y += lineHeight
	}
	_ = cc.Fill()
}

func drawResurrectMessage(cc *gg.Context) {
	if myShip().IsAlive {
		return
	}
	cc.SetRGB(1, 1, 1)
	cc.SetFont(resurrectFont)
	cc.DrawStringAnchored("You are dead. Press R to respawn.", width/2, height/2, 0.5, 0.5)
	_ = cc.Fill()
}

func drawHelp(cc *gg.Context) {
	if !helpWanted {
		return
	}
	lines := [...]string{
		"         SHH Space Game",
		"",
		"       by Sverre H Huseby",
		"",
		"",
		"Arrows            Turn and Thrust",
		"Ctrl or Space     Fire Phaser",
		"Alt or Shift      Fire Targeting Bomb",
		//"A                 Toggle Audio",
		"R                 Respawn",
		//"T                 Enter a Chat Line",
	}
	lineHeight := helpFont.Metrics().Ascent + helpFont.Metrics().Descent
	maxWidth := 0.0
	for _, line := range lines {
		w := helpFont.Advance(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	x := (width - maxWidth) / 2.0
	y := (height - float64(len(lines))*lineHeight) / 2.0
	cc.SetFont(helpFont)
	cc.SetRGB(1, 1, 1)
	for _, line := range lines {
		cc.DrawString(line, x, y)
		y += lineHeight
	}
	_ = cc.Stroke()
}

func rX(x float64) float64 {
	return x - myShip().Position.X + width/2
}

func rY(y float64) float64 {
	return y - myShip().Position.Y + height/2
}
