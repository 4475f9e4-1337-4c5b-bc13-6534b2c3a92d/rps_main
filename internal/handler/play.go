package handler

import (
	"rps_main/internal/game"
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
		return err
	}
	c.Logger().Print(bestOf)

	// Create GameServer
	id := game.NewServer(bestOf)

	// Send Game Template
	comp := templates.Game(id)
	return Render(c, comp)
}
