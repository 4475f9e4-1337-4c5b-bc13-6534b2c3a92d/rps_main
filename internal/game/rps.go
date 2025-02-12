package game

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

const (
	StatePlaying = iota
	StateRoundOver
	StateGameOver
)

const (
	TypeAI  = "ai"
	TypePvP = "pvp"
)

type move string

const (
	ChoiceRock     move = "rock"
	ChoicePaper    move = "paper"
	ChoiceScissors move = "scissors"
)

const RoundTime = 30 * time.Second

type HTMXMessage struct {
	Headers interface{} `json:"HEADERS"`
	Choice  string      `json:"choice"`
}

type ClientMove struct {
	PlayerId string `json:"playerId"`
	Choice   move   `json:"choice"`
}

type Client struct {
	id   string
	gs   *GameServer
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.gs.unregister <- c
	}()

	for {
		// parse message json
		htmxMsg := &HTMXMessage{}
		err := c.conn.ReadJSON(htmxMsg)
		if err != nil {
			fmt.Println("decode err", err)
			break
		}

		// do something with message
		fmt.Println("htmxMsg", htmxMsg.Choice)
		// call server setChoice
		//c.gs.setChoice(c.id, move(htmxMsg.Text))
		//fmt.Println("htmxMsg", htmxMsg)
		//c.gs.broadcast <- message
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("write err", err)
			break
		}
	}
}

type Score struct {
	PlayerOne int
	PlayerTwo int
	Draw      int
}

type Game struct {
	Id              string
	GameMode        string
	Type            string
	State           int
	Round           int
	Timer           time.Duration
	PlayerOneChoice move
	PlayerTwoChoice move
	Score           Score
}

type GameServer struct {
	id         string
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	ticker     *time.Ticker
	game       *Game
	mutex      sync.Mutex
}

type serverManager struct {
	servers map[string]*GameServer
	mutex   sync.Mutex
}

var manager = serverManager{
	servers: make(map[string]*GameServer),
}

func GetServer(id string) (*GameServer, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	gs, ok := manager.servers[id]
	return gs, ok
}

func NewServer(bestOf int) string {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	id := uuid.NewString()
	manager.servers[id] = &GameServer{
		id:         id,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		ticker:     time.NewTicker(time.Second),
		game: &Game{
			Id:    id,
			Type:  TypeAI,
			State: StatePlaying,
			Round: 1,
			Timer: RoundTime,
			Score: Score{
				PlayerOne: 0,
				PlayerTwo: 0,
				Draw:      0,
			},
		},
	}
	fmt.Println("new server", id)
	manager.servers[id].ticker.Stop()
	go manager.servers[id].run()
	return id
}

func (gs *GameServer) run() {
	fmt.Println("running server", gs.id)
	for {
		select {
		case client := <-gs.register:
			gs.clients[client] = true
			fmt.Println("client registered", client.id)
		case client := <-gs.unregister:
			if _, ok := gs.clients[client]; ok {
				close(client.send)
				delete(gs.clients, client)
				client.conn.Close()
			}

		// Send message to all clients
		case msg := <-gs.broadcast:
			for client := range gs.clients {
				client.send <- msg
			}

		case <-gs.ticker.C:
			if gs.game.State == StatePlaying {
				gs.mutex.Lock()
				gs.game.Timer -= time.Second
				gs.mutex.Unlock()
				if gs.game.Timer <= 0 {
					gs.endRound()
				} else {
					//gs.broadcast <- temaplates.GameTimer(gs.game.Timer)
				}
			}
		}
	}
}

func (gs *GameServer) startRound() {
	gs.mutex.Lock()
	gs.game.State = StatePlaying
	gs.game.PlayerOneChoice = ""
	gs.game.PlayerTwoChoice = ""
	gs.game.Round++
	gs.game.Timer = RoundTime
	gs.ticker.Reset(time.Second)
	gs.mutex.Unlock()
}

func (gs *GameServer) setChoice(playerId string, choice move) {
	fmt.Println("setChoice", playerId, choice)
	gs.mutex.Lock()
	if playerId == "playerOne" {
		gs.game.PlayerOneChoice = choice
	} else if playerId == "playerTwo" {
		gs.game.PlayerTwoChoice = choice
	}
	gs.mutex.Unlock()
	gs.checkRoundOver()
}

func (gs *GameServer) checkRoundOver() {
	if gs.game.PlayerOneChoice != "" && gs.game.PlayerTwoChoice != "" {
		gs.endRound()
	}
}

func (gs *GameServer) endRound() {
	gs.mutex.Lock()
	gs.game.State = StateRoundOver
	gs.mutex.Unlock()
}

func (gs *GameServer) endGame() {
	gs.mutex.Lock()
	gs.game.State = StateGameOver
	gs.mutex.Unlock()
}

func (gs *GameServer) Connect(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("ws err", err)
		return err
	}

	client := &Client{
		id:   uuid.NewString(),
		gs:   gs,
		conn: conn,
		send: make(chan []byte),
	}
	gs.register <- client
	fmt.Println("client pre", client.id)

	go client.readPump()
	go client.writePump()
	fmt.Println("client after", client.id)

	return nil
}
