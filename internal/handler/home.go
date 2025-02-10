package handler

import (
	"rps_main/internal/templates"

	"github.com/labstack/echo/v4"
)

func HandleHome(c echo.Context) error {
	comp := templates.Layout(templates.Home(), "RPS")
	return Render(c, comp)
}
