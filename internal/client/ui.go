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

type UI struct {
	client         ClientInterface
	playerNameFont text.Face
	messagesFont   text.Face
	scoresFont     text.Face
	statusFont     text.Face
	resurrectFont  text.Face
	helpFont       text.Face

	lastThrust model.Thrust
	lastTurn   model.Turn
	lastPhaser bool
	lastBomb   bool
	helpWanted bool
}

const width, height = 1500, 850

func NewUI() *UI {
	return &UI{
		lastThrust: model.ThrustNone,
		lastTurn:   model.TurnNone,
	}
}

func (ui *UI) startUI(cl ClientInterface) {
	ui.client = cl
	initSounds()
	defer func() { teardownSounds() }()
	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("SHH Space Game").
		WithSize(width, height).
		WithContinuousRender(false))

	ui.client.InfoMessages().Clear()
	ui.client.ChatMessages().Clear()

	var canvas *ggcanvas.Canvas
	var animToken *gogpu.AnimationToken
	var frame int

	fontSource := utils.LoadFontSource()
	defer func() { _ = fontSource.Close() }()
	ui.playerNameFont = fontSource.Face(16)
	ui.messagesFont = fontSource.Face(24)
	ui.scoresFont = fontSource.Face(20)
	ui.statusFont = fontSource.Face(20)
	ui.resurrectFont = fontSource.Face(34)
	ui.helpFont = fontSource.Face(26)

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
			ui.renderFrame(cc)
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

	app.EventSource().OnKeyPress(ui.onKeyPress())
	app.EventSource().OnKeyRelease(ui.onKeyRelease())
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func (ui *UI) onKeyPress() func(key gpucontext.Key, mods gpucontext.Modifiers) {
	return func(key gpucontext.Key, mods gpucontext.Modifiers) {
		if key == gpucontext.KeyLeft {
			ui.client.TurnLeft()
			ui.lastTurn = model.TurnLeft
		} else if key == gpucontext.KeyRight {
			ui.client.TurnRight()
			ui.lastTurn = model.TurnRight
		} else if key == gpucontext.KeyUp {
			ui.client.ThrustForward()
			ui.lastThrust = model.ThrustForward
		} else if key == gpucontext.KeyDown {
			ui.client.ThrustBack()
			ui.lastThrust = model.ThrustBack
		} else if !ui.lastPhaser && (key == gpucontext.KeySpace || key == gpucontext.KeyLeftControl || key == gpucontext.KeyRightControl) {
			ui.client.FirePhaser()
			ui.lastPhaser = true
		} else if !ui.lastBomb && (key == gpucontext.KeyLeftAlt || key == gpucontext.KeyRightAlt || key == gpucontext.KeyLeftShift || key == gpucontext.KeyRightShift) {
			ui.client.FireBomb()
			ui.lastBomb = true
		} else if key == gpucontext.KeyR {
			if !ui.client.MyShip().IsAlive {
				ui.client.Resurrect()
			}
		} else if key == gpucontext.KeyF1 {
			ui.helpWanted = true
		}
	}
}

func (ui *UI) onKeyRelease() func(key gpucontext.Key, mods gpucontext.Modifiers) {
	return func(key gpucontext.Key, mods gpucontext.Modifiers) {
		if key == gpucontext.KeyLeft && ui.lastTurn == model.TurnLeft {
			ui.client.TurnNone()
			ui.lastTurn = model.TurnNone
		} else if key == gpucontext.KeyRight && ui.lastTurn == model.TurnRight {
			ui.client.TurnNone()
			ui.lastTurn = model.TurnNone
		} else if key == gpucontext.KeyUp && ui.lastThrust == model.ThrustForward {
			ui.client.ThrustNone()
			ui.lastThrust = model.ThrustNone
		} else if key == gpucontext.KeyDown && ui.lastThrust == model.ThrustBack {
			ui.client.ThrustNone()
			ui.lastThrust = model.ThrustNone
		} else if key == gpucontext.KeySpace || key == gpucontext.KeyLeftControl || key == gpucontext.KeyRightControl {
			ui.lastPhaser = false
		} else if key == gpucontext.KeyLeftAlt || key == gpucontext.KeyRightAlt || key == gpucontext.KeyLeftShift || key == gpucontext.KeyRightShift {
			ui.lastBomb = false
		} else if key == gpucontext.KeyF1 {
			ui.helpWanted = false
		}
	}
}

func (ui *UI) renderFrame(cc *gg.Context) {
	world := ui.client.World()
	myShip := ui.client.MyShip()
	cc.ClearWithColor(gg.RGBA2(0, 0, 0, 1))
	// world relative objects
	ui.drawBounds(cc, world, myShip)
	for _, star := range world.Stars {
		ui.drawStar(cc, star, myShip)
	}
	for _, ship := range world.Ships {
		ui.drawShip(cc, ship, myShip)
	}
	for _, phaser := range world.Phasers {
		ui.drawPhaser(cc, phaser, myShip)
	}
	for _, bomb := range world.Bombs {
		ui.drawBomb(cc, bomb, myShip)
	}
	for _, bombPack := range world.BombPacks {
		ui.drawBombPack(cc, bombPack, myShip)
	}
	for _, explosion := range world.Explosions {
		ui.drawExplosion(cc, explosion, myShip)
	}
	// screen relative objects
	ui.drawRadar(cc, world, myShip)
	ui.drawStatus(cc, myShip)
	ui.drawScores(cc, world)
	ui.drawAllMessages(cc)
	ui.drawResurrectMessage(cc, myShip)
	ui.drawHelp(cc)
}

func (ui *UI) drawBounds(cc *gg.Context, world *model.World, myShip *model.Ship) {
	cc.SetRGB(0.5, 0.5, 0.5)
	cc.MoveTo(ui.rX(0, myShip), ui.rY(0, myShip))
	cc.LineTo(ui.rX(world.Width-1, myShip), ui.rY(0, myShip))
	cc.LineTo(ui.rX(world.Width-1, myShip), ui.rY(world.Height-1, myShip))
	cc.LineTo(ui.rX(0, myShip), ui.rY(world.Height-1, myShip))
	cc.LineTo(ui.rX(0, myShip), ui.rY(0, myShip))
	_ = cc.Stroke()
}

func (ui *UI) drawStar(cc *gg.Context, star *model.Star, myShip *model.Ship) {
	cc.SetRGB(star.Color.R, star.Color.G, star.Color.B)
	cc.DrawPoint(ui.rX(star.Position.X, myShip), ui.rY(star.Position.Y, myShip), 1.1)
	_ = cc.Fill()
}

func (ui *UI) drawShip(cc *gg.Context, ship *model.Ship, myShip *model.Ship) {
	if !ship.IsAlive {
		return
	}
	worldShape := ship.WorldRelativeShape()
	cc.SetRGB(ship.Color.R, ship.Color.G, ship.Color.B)
	cc.MoveTo(ui.rX(worldShape[0].X, myShip), ui.rY(worldShape[0].Y, myShip))
	for _, point := range worldShape[1:] {
		cc.LineTo(ui.rX(point.X, myShip), ui.rY(point.Y, myShip))
	}
	cc.ClosePath()
	_ = cc.Fill()
	if len(ship.Name) > 0 {
		cc.SetRGB(0.39, 0.39, 1)
		cc.SetFont(ui.playerNameFont)
		cc.DrawStringAnchored(ship.Name, ui.rX(ship.Position.X, myShip), ui.rY(ship.Position.Y-17, myShip), 0.5, 1)
	}
}

func (ui *UI) drawPhaser(cc *gg.Context, phaser *model.Phaser, myShip *model.Ship) {
	cc.SetRGB(phaser.Color.R, phaser.Color.G, phaser.Color.B)
	cc.DrawPoint(ui.rX(phaser.Position.X, myShip), ui.rY(phaser.Position.Y, myShip), 1.3)
	_ = cc.Fill()
}

func (ui *UI) drawBomb(cc *gg.Context, bomb *model.Bomb, myShip *model.Ship) {
	cc.SetRGB(1, 1, 0)
	diameter := 7.0
	r := (diameter - 2.0) / 2.0
	sx := ui.rX(bomb.Position.X, myShip)
	sy := ui.rY(bomb.Position.Y, myShip)
	if bomb.Flip {
		cc.DrawLine(sx-r-2, sy, sx+r+2, sy)
		cc.DrawLine(sx, sy-r-2, sx, sy+r+2)
	} else {
		cc.DrawLine(sx-r-1, sy-r-1, sx+r+1, sy+r+1)
		cc.DrawLine(sx-r-1, sy+r+1, sx+r+1, sy-r-1)
	}
	_ = cc.Stroke()
	cc.SetRGB(bomb.Color.R, bomb.Color.G, bomb.Color.B)
	cc.DrawCircle(sx, sy, r)
	_ = cc.Fill()
}

func (ui *UI) drawBombPack(cc *gg.Context, bombPack *model.BombPack, myShip *model.Ship) {
	diameter := 7.0
	d2 := diameter - 2.0
	r := d2 / 2.0
	cc.SetRGB(bombPack.Color.R, bombPack.Color.G, bombPack.Color.B)
	cc.DrawCircle(ui.rX(bombPack.Position.X-d2, myShip), ui.rY(bombPack.Position.Y-d2, myShip), r)
	cc.DrawCircle(ui.rX(bombPack.Position.X, myShip), ui.rY(bombPack.Position.Y-d2, myShip), r)
	cc.DrawCircle(ui.rX(bombPack.Position.X-r, myShip), ui.rY(bombPack.Position.Y, myShip), r)
	_ = cc.Fill()
}

func (ui *UI) drawExplosion(cc *gg.Context, explosion *model.Explosion, myShip *model.Ship) {
	cc.SetRGB(0.8, 0.8, 1.0)
	cc.DrawCircle(ui.rX(explosion.Position.X, myShip), ui.rY(explosion.Position.Y, myShip), explosion.OuterRadius)
	_ = cc.Fill()
	if explosion.InnerRadius > 0 {
		cc.SetRGB(0, 0, 0)
		cc.DrawCircle(ui.rX(explosion.Position.X, myShip), ui.rY(explosion.Position.Y, myShip), explosion.InnerRadius)
		_ = cc.Fill()
	}
}

func (ui *UI) drawAllMessages(cc *gg.Context) {
	ui.drawMessages(cc, ui.client.InfoMessages(), 1.0, 0.78, 0.0, 0)
	ui.drawMessages(cc, ui.client.ChatMessages(), 1.0, 1.0, 0.0, height/2)
}

func (ui *UI) drawMessages(cc *gg.Context, messages *GameMessages, r, g, b float64, startY float64) {
	lineHeight := ui.messagesFont.Metrics().Ascent + ui.messagesFont.Metrics().Descent
	y := startY + lineHeight
	x := 10.0
	cc.SetRGB(r, g, b)
	cc.SetFont(ui.messagesFont)
	for _, msg := range messages.GetMessages() {
		cc.DrawString(msg.Text, x, y)
		y += lineHeight
	}
	_ = cc.Fill()
}

func (ui *UI) drawRadar(cc *gg.Context, world *model.World, myShip *model.Ship) {
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
	cc.DrawRectangle(radarX+((myShip.Position.X-width/2)*radarWidth)/world.Width,
		radarY+((myShip.Position.Y-height/2)*radarHeight)/world.Height,
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

func (ui *UI) drawStatus(cc *gg.Context, myShip *model.Ship) {
	ascent := ui.statusFont.Metrics().Ascent
	lineHeight := ascent + ui.messagesFont.Metrics().Descent
	labelWidth := ui.statusFont.Advance("Damage: ") // longest status indicator
	meterWidth := 2 * labelWidth
	x := 10.0
	x2 := x + labelWidth
	y := height - 3*lineHeight
	cc.SetRGB(0, 1, 0)
	cc.SetFont(ui.scoresFont)
	cc.DrawString("Bombs:", x, y)
	cc.DrawString(fmt.Sprintf("%d", myShip.BombsLeft), x2, y)
	cc.DrawString("Heat:", x, y+lineHeight)
	heatPctWidth := meterWidth * math.Min(float64(myShip.PhaserHeat), 100) / 100.0
	safeHeatPctWidth := meterWidth * math.Min(float64(myShip.PhaserHeat), 75) / 100.0
	if myShip.PhaserHeat > 75 {
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
	damagePctWidth := meterWidth * math.Min(float64(myShip.Damage), 100) / 100.0
	cc.SetRGB(1, 0, 0)
	cc.DrawRectangle(x2, y+2*lineHeight-ascent, damagePctWidth, ascent)
	_ = cc.Fill()
	cc.SetRGB(0, 1, 0)
	cc.DrawRectangle(x2, y+2*lineHeight-ascent, meterWidth, ascent)
	_ = cc.Stroke()
}

func (ui *UI) drawScores(cc *gg.Context, world *model.World) {
	scores := toScores(slices.Collect(maps.Values(world.Ships)))
	lines := make([]string, 0, len(scores))
	maxWidth := 0.0
	for _, score := range scores {
		s := fmt.Sprintf("%5.2f (%3d/%3d) %-15.15s", score.Ratio, score.Score, score.AntiScore, score.Name)
		w := ui.scoresFont.Advance(s)
		if w > maxWidth {
			maxWidth = w
		}
		lines = append(lines, s)
	}
	lineHeight := ui.scoresFont.Metrics().Ascent + ui.scoresFont.Metrics().Descent
	x := width - maxWidth
	y := lineHeight
	cc.SetRGB(1, 1, 0)
	cc.SetFont(ui.scoresFont)
	for _, line := range lines {
		cc.DrawString(line, x, y)
		y += lineHeight
	}
	_ = cc.Fill()
}

func (ui *UI) drawResurrectMessage(cc *gg.Context, myShip *model.Ship) {
	if myShip.IsAlive {
		return
	}
	cc.SetRGB(1, 1, 1)
	cc.SetFont(ui.resurrectFont)
	cc.DrawStringAnchored("You are dead. Press R to respawn.", width/2, height/2, 0.5, 0.5)
	_ = cc.Fill()
}

func (ui *UI) drawHelp(cc *gg.Context) {
	if !ui.helpWanted {
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
	lineHeight := ui.helpFont.Metrics().Ascent + ui.helpFont.Metrics().Descent
	maxWidth := 0.0
	for _, line := range lines {
		w := ui.helpFont.Advance(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	x := (width - maxWidth) / 2.0
	y := (height - float64(len(lines))*lineHeight) / 2.0
	cc.SetFont(ui.helpFont)
	cc.SetRGB(1, 1, 1)
	for _, line := range lines {
		cc.DrawString(line, x, y)
		y += lineHeight
	}
	_ = cc.Stroke()
}

func (ui *UI) rX(x float64, myShip *model.Ship) float64 {
	return x - myShip.Position.X + width/2
}

func (ui *UI) rY(y float64, myShip *model.Ship) float64 {
	return y - myShip.Position.Y + height/2
}
