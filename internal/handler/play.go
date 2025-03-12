package handler

import (
	"log"
	"rps_main/internal/templates"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type GameServer struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func HandlePlay(c echo.Context) error {
	bestOf, err := strconv.Atoi(c.FormValue("bestOf"))
	if err != nil {
		log.Println("Invalid bestOf value")
		return err
	}
	mode := c.FormValue("mode")

	log.Println("Starting game with bestOf:", bestOf, "mode:", mode)
	if mode == "" {
		mode = "ai"
		comp := templates.Game("1")
		return Render(c, comp)
	} else {
		mode = "pvp"
		comp := templates.InQueue("1")
		return Render(c, comp)
	}

	// Create GameServer
	//id := game.NewServer(bestOf)
	// Send Game Template
	//comp := templates.Game(id)
	//return Render(c, comp)
}
