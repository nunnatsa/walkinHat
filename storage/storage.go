package storage

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"

	"github.com/nunnatsa/walkingHat/hat"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// thread safe id provider. Produces an int64 id from a thread safe counter
type IDProvider struct {
	lock    *sync.Mutex
	counter int64
}

func (p *IDProvider) getNextID() int64 {
	p.lock.Lock()
	p.counter++
	newID := p.counter
	p.lock.Unlock()
	return newID
}

var idp = &IDProvider{lock: &sync.Mutex{}}

// Storage stores a Pixel. It updates it from an incoming channel, and distributes the change to its listeners.
type Storage struct {
	data     []byte
	clients  map[int64]chan []byte
	dataLock *sync.RWMutex
}

// register new listener. Return the the client id and the current Pixel
func (st Storage) Register(ch chan []byte) (int64, []byte) {
	id := idp.getNextID()
	st.clients[id] = ch
	log.Println("register new client", id)

	st.dataLock.RLock()
	data := st.data
	st.dataLock.RUnlock()
	return id, data
}

// remove a listener
func (st *Storage) Deregister(id int64) {
	close(st.clients[id])
	delete(st.clients, id)
	log.Println("client de-registered", id)
}

// get updates from the Hat, and distribute it to the listener
func (st *Storage) do(ch <-chan *hat.Pixel) {
	for data := range ch {
		st.dataLock.Lock()
		st.data = []byte(data.String())
		newData := st.data
		st.dataLock.Unlock()
		for _, client := range st.clients {
			client <- newData
		}
	}
}

func NewStorage(ch <-chan *hat.Pixel) *Storage {
	st := &Storage{clients: map[int64]chan []byte{}, dataLock: &sync.RWMutex{}}

	go st.do(ch)

	return st
}

// This request handler opens a websocket to the browser, and then notify the current pixel to the browser each time it
// changed.
// When the websocket connection is established, the function register itself as a new storage listener. As result, it
// receives the current pixel and sends it to the websocket. Then it listen to changes in the pixel and each time it
// changed, the function updates the websocket
//
func (st Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		ch := make(chan []byte)

		id, px := st.Register(ch)
		defer st.Deregister(id)

		if err := conn.WriteMessage(websocket.TextMessage, px); err != nil {
			log.Println(err)
			return
		}

		for px = range ch {
			if err := conn.WriteMessage(websocket.TextMessage, px); err != nil {
				log.Println(err)
				return
			}

		}
	}
}
