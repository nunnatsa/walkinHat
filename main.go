package main

import (
	"math/rand"
	"os"

	"github.com/gorilla/websocket"
	"github.com/nathany/bobblehat/sense/screen/color"

	"html/template"
	"log"
	"net/http"
)

const (
	rmask = 0xF800
	gmask = 0x07E0
	bmask = 0x001F
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getColor() color.Color {
	return color.New(uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)))
}

func main() {
	ch := make(chan *Pixel)
	_ = NewHat(ch)
	st := NewStorage(ch)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	http.Handle("/color", st)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Panic(err)
	}
}

func init() {
	_, err := os.Stat("static/index.html")
	if os.IsNotExist(err) {
		t, err := template.ParseFiles("static/index.gohtml")
		if err != nil {
			log.Panic(err)
		}
		f, err := os.Create("static/index.html")
		if err != nil {
			log.Panic(err)
		}

		hostname, err := os.Hostname()
		if err != nil {
			log.Panic(err)
		}
		err = t.Execute(f, hostname)
		if err != nil {
			log.Panic(err)
		}
	}
}
