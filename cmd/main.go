package main

import (
	"rps_main/internal/game"
	"rps_main/internal/handler"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Static("/", "static")

	e.GET("/", handler.HandleHome)
	e.POST("/play", handler.HandlePlay)

	// Game Server (WebSocket)
	e.GET("/game/:id", func(c echo.Context) error {
		gameId := c.Param("id")
		if gs, ok := game.GetServer(gameId); ok {
			return gs.Connect(c)
		}
		return c.Redirect(301, "/")
	})

	e.Logger.Fatal(e.Start(":9000"))
}
