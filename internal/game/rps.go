package game

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"rps_main/internal/templates"
	"strconv"
	"strings"
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

const RoundTime = time.Second * 30
const RoundEndTime = time.Second * 3

type HTMXMessage struct {
	Headers  interface{} `json:"HEADERS"`
	Choice   string      `json:"choice"`
	PlayerId string      `json:"playerId"`
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
			fmt.Println("close", err)
			break
		}

		// set move
		c.gs.setChoice(c.id, move(htmxMsg.Choice))
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		//fmt.Println("write:\n", string(msg))
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
	BestOf          int
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
	tickerDone chan bool
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
		tickerDone: make(chan bool),
		game: &Game{
			Id:       id,
			GameMode: "rps",
			Type:     TypeAI,
			BestOf:   bestOf,
			State:    StatePlaying,
			Round:    0,
			Timer:    RoundTime,
			Score: Score{
				PlayerOne: 0,
				PlayerTwo: 0,
				Draw:      0,
			},
		},
	}
	fmt.Println("New server", id[:strings.Index(id, "-")], "nServers:", len(manager.servers))
	go manager.servers[id].run()
	return id
}

func (gs *GameServer) Connect(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("ws err", err)
		return err
	}
	client := &Client{
		gs:   gs,
		conn: conn,
		send: make(chan []byte),
	}
	gs.register <- client
	go client.readPump()
	go client.writePump()
	return nil
}

func (gs *GameServer) cleanup() {
	fmt.Println("cleaning up server", gs.id)
	gs.tickerDone <- true
	close(gs.broadcast)
	close(gs.register)
	close(gs.unregister)
	manager.mutex.Lock()
	delete(manager.servers, gs.id)
	defer manager.mutex.Unlock()
}

func (gs *GameServer) run() {
	//fmt.Println("running server", gs.id)
	gs.tick()
	for {
		select {
		case client := <-gs.register:
			gs.clients[client] = true
			client.id = "playerOne"
			fmt.Println("client registered", client.id)
			if len(gs.clients) == 2 || gs.game.Type == TypeAI {
				gs.startRound()
			}
		case client := <-gs.unregister:
			if _, ok := gs.clients[client]; ok {
				fmt.Println("client unregistered", client.id)
				close(client.send)
				delete(gs.clients, client)
				client.conn.Close()
			}
			if len(gs.clients) == 0 {
				gs.cleanup()
				return
			}

		// Send message to all clients
		case msg := <-gs.broadcast:
			for client := range gs.clients {
				client.send <- msg
			}
		}
	}
}

func (gs *GameServer) tick() {
	if gs.ticker == nil {
		gs.ticker = time.NewTicker(time.Second)
	}
	go func() {
		for {
			select {
			case <-gs.tickerDone:
				gs.ticker.Stop()
				return

			case <-gs.ticker.C:
				if gs.game.State == StatePlaying {
					gs.mutex.Lock()
					gs.game.Timer -= time.Second
					gs.mutex.Unlock()
					if gs.game.Timer <= 0 {
						gs.endRound()
					} else {
						gs.broadcast <- []byte(renderScoreboard(gs.game))
					}
				} else if gs.game.State == StateRoundOver {
					gs.mutex.Lock()
					gs.game.Timer -= time.Second
					gs.mutex.Unlock()
					if gs.game.Timer <= 0 {
						// Check End Game
						nRound := (gs.game.BestOf / 2) + 1
						if gs.game.Score.PlayerOne == nRound || gs.game.Score.PlayerTwo == nRound {
							winner := getWinner(gs.game.PlayerOneChoice, gs.game.PlayerTwoChoice)
							gs.endGame(winner)
							return
						} else {
							gs.startRound()
							gs.broadcast <- []byte(renderSelection(gs.game))
						}
					} else {
						gs.broadcast <- []byte(renderScoreboard(gs.game))
					}
				}
			}
		}
	}()
}

func (gs *GameServer) startRound() {
	gs.mutex.Lock()
	gs.game.State = StatePlaying
	gs.game.PlayerOneChoice = ""
	gs.game.PlayerTwoChoice = ""
	gs.game.Round++
	gs.game.Timer = RoundTime
	gs.mutex.Unlock()
	gs.ticker.Reset(time.Second)
}

func makeAIChoice() move {
	choices := []move{ChoiceRock, ChoicePaper, ChoiceScissors}
	return choices[rand.Intn(len(choices))]
}

func (gs *GameServer) setChoice(playerId string, choice move) {
	if gs.game.State != StatePlaying {
		return
	}

	fmt.Println("setChoice", playerId, choice)
	gs.mutex.Lock()
	if playerId == "playerOne" {
		gs.game.PlayerOneChoice = choice
	} else if playerId == "playerTwo" {
		gs.game.PlayerTwoChoice = choice
	}
	if gs.game.Type == TypeAI {
		gs.game.PlayerTwoChoice = makeAIChoice()
	}

	gs.mutex.Unlock()

	// Check round is over
	if gs.game.PlayerOneChoice != "" && gs.game.PlayerTwoChoice != "" {
		gs.endRound()
	}
}

func getWinner(moveOne, moveTwo move) int {
	if moveOne == moveTwo {
		return 0
	}
	switch moveOne {
	case "rock":
		if moveTwo == "scissors" {
			return 1
		}
	case "paper":
		if moveTwo == "rock" {
			return 1
		}
	case "scissors":
		if moveTwo == "paper" {
			return 1
		}
	}
	return 2
}

func (gs *GameServer) endRound() {
	gs.mutex.Lock()
	gs.game.State = StateRoundOver
	winner := getWinner(gs.game.PlayerOneChoice, gs.game.PlayerTwoChoice)
	//fmt.Println("endRound winner:", winner, gs.game.PlayerOneChoice, gs.game.PlayerTwoChoice)
	// Score the game
	if winner == 0 {
		gs.game.Score.Draw++
	} else if winner == 1 {
		gs.game.Score.PlayerOne++
	} else if winner == 2 {
		gs.game.Score.PlayerTwo++
	}
	gs.mutex.Unlock()

	// Render round result
	gs.game.Timer = RoundEndTime
	gs.broadcast <- []byte(renderEndRound(gs.game, winner))
}

func (gs *GameServer) endGame(winner int) {
	fmt.Println("endGame")
	gs.mutex.Lock()
	gs.game.State = StateGameOver
	gs.mutex.Unlock()

	// Render game result
	gs.broadcast <- []byte(renderEndGame(gs.game, winner))
}

func renderScoreboard(g *Game) string {
	buf := new(bytes.Buffer)
	var playerOneScore string = strconv.Itoa(g.Score.PlayerOne)
	var playerTwoScore string = strconv.Itoa(g.Score.PlayerTwo)
	var timer string = g.Timer.String()
	if g.Timer <= 0 {
		timer = "_"
	}
	templates.Scoreboard(playerOneScore, playerTwoScore, timer).Render(context.Background(), buf)
	return buf.String()
}

func renderSelection(g *Game) string {
	buf := new(bytes.Buffer)
	templates.SelectionScreen().Render(context.Background(), buf)
	html := buf.String() + renderScoreboard(g)
	return html
}

func renderEndRound(g *Game, winner int) string {
	buf := new(bytes.Buffer)
	winnerName := ""
	if winner == 0 {
		winnerName = "draw"
	} else if winner == 1 {
		winnerName = "player1"
	} else if winner == 2 {
		winnerName = "player2"
	}
	templates.ResultScreen(string(g.PlayerOneChoice), string(g.PlayerTwoChoice), winnerName).Render(context.Background(), buf)
	html := buf.String() + renderScoreboard(g)
	return html
}

func renderEndGame(g *Game, winner int) string {
	buf := new(bytes.Buffer)
	winnerName := ""
	if winner == 1 {
		winnerName = "player1"
	} else if winner == 2 {
		winnerName = "player2"
	}
	templates.EndScreen(winnerName).Render(context.Background(), buf)
	html := buf.String() + renderScoreboard(g)
	return html
}
