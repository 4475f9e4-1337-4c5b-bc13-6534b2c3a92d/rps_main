package handler

import (
	"rps_main/internal/game"
	"rps_main/internal/templates"

	"github.com/labstack/echo/v4"
)

func HandlePlay(c echo.Context) error {
	c.Logger().Print(c.FormValue("bestOf"))
	return Render(c, templates.Help())
	g := game.New()
	comp := templates.Layout(templates.PlayAI(g), "RPS - Play")
	return Render(c, comp)
}
