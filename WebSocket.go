package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var pool Pool

// This will let us know if the client goes away so we can remove it from the pool
func readLoop(c *Client) {
	for {
		if _, _, err := c.Conn.NextReader(); err != nil {
			log.Println("readLoop", err)
			if err := c.Conn.Close(); err != nil {
				log.Print(err)
			}
			pool.Unregister <- c
			break
		}
	}
}

type Client struct {
	ID   string // IP address and port for the registrant
	Conn *websocket.Conn
	//	Pool *Pool
}

//type Message struct {
//	Type int    `json:"type"`
//	Body string `json:"body"`
//}

// Pool of client registrations
type Pool struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

func (p *Pool) Init() {
	p.Clients = make(map[*Client]bool)
	p.Register = make(chan *Client)
	p.Unregister = make(chan *Client)
	p.Broadcast = make(chan []byte)
}

// NewPool creates the pool for client registrations
//func NewPool() *Pool {
//	return &Pool{
//		Clients:    make(map[*Client]bool),
//		Register:   make(chan *Client),
//		Unregister: make(chan *Client),
//		Broadcast:  make(chan []byte),
//	}
//}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			go readLoop(client)
			log.Println("Size of Connection Pool: ", len(pool.Clients), client.ID, " added.")
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			log.Println("Size of Connection Pool: ", len(pool.Clients), client.ID, " dropped off.")
			break
		case message := <-pool.Broadcast:
			//			fmt.Println("Sending message to all clients in Pool")
			for client := range pool.Clients {
				if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Broadcast update error - %s\n", err)
					delete(pool.Clients, client)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return conn, nil
}
