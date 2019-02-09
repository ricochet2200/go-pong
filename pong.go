package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"syscall/js"
	"time"

	"github.com/fogleman/gg"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"golang.org/x/image/draw"
)

type Screen struct {
	Width   int
	Height  int
	Balls   []*Ball
	Paddles []*Paddle
	Score   ScoreBoard
}

func NewScreen(id string) *Screen {
	// Init Canvas stuff
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", id)
	width := int(doc.Get("body").Get("clientWidth").Float())
	height := int(doc.Get("body").Get("clientHeight").Float())
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)

	left := &Paddle{Ball{10, 10, 10, 400, 1, 1}}
	right := &Paddle{Ball{int(width) - 10 - 10, 10, 10, 400, 1, 1}}
	paddles := []*Paddle{left, right}

	fmt.Println("on")
	keyDownEvt := js.NewCallback(func(args []js.Value) {
		e := args[0]
		fmt.Println(e)
		fmt.Println(e.Get("which").Int())
		if e.Get("which").Int() == 38 { // up
			right.Move(-20, height)
		} else if e.Get("which").Int() == 40 { // down
			right.Move(20, height)
		} else if e.Get("which").Int() == 87 { // w
			left.Move(-20, height)
		} else if e.Get("which").Int() == 83 { // s
			left.Move(20, height)
		}
	})
	js.Global().Get("document").Call("addEventListener", "keydown", keyDownEvt)

	s := Screen{int(width), int(height), []*Ball{}, paddles, ScoreBoard{0, 0}}
	for i := 0; i < 8; i++ {
		ball := new(Ball)
		ball.Reset(&s)
		s.Balls = append(s.Balls, ball)
		fmt.Println("Add", len(s.Balls))
	}
	return &s
}

func main() {

	screen := NewScreen("mycanvas")

	tick := time.Tick(50 * time.Millisecond)

	for {
		<-tick
		screen.Update()
		screen.Draw()
	}
}

func (screen *Screen) Update() {
	for _, b := range screen.Balls {
		b.Update()
		if b.X+b.Width >= screen.Width {
			screen.Score.Left += 1
			b.Reset(screen)
		} else if b.X <= 0 {
			screen.Score.Right += 1
			b.Reset(screen)
		}
		if b.Y+b.Height >= screen.Height || b.Y <= 0 {
			b.DirectionY *= -1
		}
		for _, p := range screen.Paddles {
			if b.X+b.Width >= p.X && b.X < p.X+p.Width && b.Y+b.Height >= p.Y && b.Y < p.Y+p.Height {
				b.DirectionX *= -1
			}
		}
	}
}

func (screen *Screen) Draw() {
	//start := time.Now()
	background := image.NewRGBA(image.Rect(0, 0, screen.Width, screen.Height))
	for _, s := range screen.Balls {
		i := s.Draw()
		draw.Copy(background, image.Point{s.X, s.Y}, i, i.Bounds(), draw.Src, nil)
	}
	for _, s := range screen.Paddles {
		i := s.Draw()
		draw.Copy(background, image.Point{s.X, s.Y}, i, i.Bounds(), draw.Src, nil)
	}

	c := gg.NewContextForRGBA(background)
	c.SetRGB(0, 0, 0)
	c.DrawStringAnchored(fmt.Sprintf("LEFT: %d", screen.Score.Left), float64(screen.Width/4), 40, 0, 0)
	c.DrawStringAnchored(fmt.Sprintf("RIGHT: %d", screen.Score.Right), float64(screen.Width/4)*3, 40, 0, 0)
	c.Fill()

	cp := make([]byte, len(background.Pix))
	copy(cp, background.Pix)

	y := js.TypedArrayOf(cp)
	js.Global().Call("DrawClamped", screen.Width, screen.Height, y)
	y.Release()
}

type Ball struct {
	X          int
	Y          int
	Width      int
	Height     int
	DirectionX int
	DirectionY int
}

type ScoreBoard struct {
	Left  int
	Right int
}

func (s *Ball) Update() {
	s.X += (10 * s.DirectionX)
	s.Y += (10 * s.DirectionY)
}

func (b *Ball) Reset(screen *Screen) {
	b.X = screen.Width/4 + rand.Intn(screen.Width/2)
	if b.X > screen.Width/2 {
		b.DirectionX = -1
	} else {
		b.DirectionX = 1
	}

	b.Y = screen.Height/4 + rand.Intn(screen.Height/2)
	if b.Y > screen.Height/2 {
		b.DirectionY = -1
	} else {
		b.DirectionY = 1
	}

	b.Width = 20
	b.Height = 20
}

func (s *Ball) Draw() *image.RGBA {

	img := image.NewRGBA(image.Rect(0, 0, s.Width, s.Height))
	gc := draw2dimg.NewGraphicContext(img)

	gc.SetStrokeColor(color.NRGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})
	gc.SetLineWidth(1)
	gc.SetFillColor(color.NRGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})

	// Draw a circle
	draw2dkit.Circle(gc, float64(s.Width/2), float64(s.Height/2), float64(s.Width/2))
	gc.FillStroke()
	return img
}

type Paddle struct {
	Ball
}

func (p *Paddle) Update() {

}

func (p *Paddle) Move(y int, max int) {
	if p.Y+y < 0 {
		p.Y = 0
	} else if p.Y+p.Height+y > max {
		p.Y = max - p.Height
	} else {
		p.Y += y
	}
}

func (p *Paddle) Draw() *image.RGBA {

	img := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	c := gg.NewContextForRGBA(img)
	c.SetRGB(0, 0, 0)
	c.DrawRectangle(0, 0, float64(p.Width), float64(p.Height))
	c.Fill()
	return img
}
