package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"../trace"
)

type room struct {
	forward chan []byte
	join chan *client
	leave chan *client
	clients map[*client]bool
	tracer trace.Tracer
}

func (r *room) run() {
	for {
		select {
			case client := <- r.join:
				r.clients[client] = true
				r.tracer.Trace("New client joined")
			case client := <-r.leave:
				r.clients[client] = false
				r.tracer.Trace("Client left")
			case msg := <- r.forward:
				r.tracer.Trace("got message")
				for client := range r.clients {
					r.tracer.Trace("sent msg to client")
					client.send <- msg
				}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
	}
	client := &client{
		socket: socket,
		send: make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() {r.leave <- client}()
	go client.write()
	client.read()
}

func newRoom() *room {
	return &room {
		forward: make(chan []byte),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
		tracer: trace.Off(),
	}
}