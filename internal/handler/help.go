package handler

import (
	"rps_main/internal/templates"

	"github.com/labstack/echo/v4"
)

func HandleHelp(c echo.Context) error {
	return Render(c, templates.Help())
}
