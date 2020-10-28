package main

import (
	"github.com/nunnatsa/walkinHat/hat"
	"github.com/nunnatsa/walkinHat/storage"
	"os"

	"html/template"
	"log"
	"net/http"
)

func main() {
	ch := make(chan *hat.Pixel)
	_ = hat.NewHat(ch)
	st := storage.NewStorage(ch)

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
