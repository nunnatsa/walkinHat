package hat

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/nathany/bobblehat/sense/screen"
	"github.com/nathany/bobblehat/sense/screen/color"
	"github.com/nathany/bobblehat/sense/stick"
)

const (
	dataTemplate = `{"color": "%s", "x": %d, "y": %d}`

	rmask = 0xF800
	gmask = 0x07E0
	bmask = 0x001F
)

type Pixel struct {
	x, y int
	c    color.Color
}

func (px Pixel) MarshalJSON() ([]byte, error) {
	return []byte(px.String()), nil
}

func (px Pixel) String() string {
	return fmt.Sprintf(dataTemplate, colorToHTML(px.c), px.x, px.y)
}

func colorToHTML(c color.Color) string {
	r := (uint16(c) & rmask) >> 8
	g := (uint16(c) & gmask) >> 3
	b := (uint16(c) & bmask) << 3

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

type Hat struct {
	ch    chan<- *Pixel
	input *stick.Device
	px    Pixel
}

func NewHat(ch chan<- *Pixel) *Hat {
	h := &Hat{
		ch: ch,
	}

	go h.do()
	return h
}

func (h *Hat) init() {
	var err error
	h.input, err = stick.Open("/dev/input/event0")
	if err != nil {
		log.Panic(err)
	}

	rand.Seed(time.Now().UnixNano())

	h.px = Pixel{rand.Intn(8), rand.Intn(8), getColor()}
	h.drawPixel()
}

func (h Hat) drawPixel() {
	fb := screen.NewFrameBuffer()
	h.ch <- &h.px
	fb.SetPixel(h.px.x, h.px.y, h.px.c)
	screen.Draw(fb)
}

func (h *Hat) do() {
	h.init()
	// Set up a signals channel (stop the loop using Ctrl-C)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	for {
		select {
		case <-signals:
			screen.Clear()
			fmt.Println("")
			os.Exit(0)
		case event := <-h.input.Events:
			changed := false
			switch event.Code {
			case stick.Enter:
				changed = true
				h.px.c = getColor()
			case stick.Up:
				if h.px.y > 0 {
					changed = true
					h.px.y--
				}
			case stick.Down:
				if h.px.y < 7 {
					changed = true
					h.px.y++
				}
			case stick.Left:
				if h.px.x > 0 {
					changed = true
					h.px.x--
				}
			case stick.Right:
				if h.px.x < 7 {
					changed = true
					h.px.x++
				}
			}
			if changed {
				h.drawPixel()
			}
		}
	}
}

func getColor() color.Color {
	return color.New(uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)))
}
