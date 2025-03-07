package handler

import (
	"log"
	"rps_main/internal/templates"

	"github.com/labstack/echo/v4"
)

func HandleHome(c echo.Context) error {
	comp := templates.Layout(templates.Home(), "RPS")
	return Render(c, comp)
}

func HandleLogin(c echo.Context) error {
	comp := templates.LoginForm()
	return Render(c, comp)
}

func HandleRegister(c echo.Context) error {
	comp := templates.RegisterForm()
	return Render(c, comp)
}

func HandleProfile(c echo.Context) error {
	log.Println("HandleProfile")
	comp := templates.Layout(templates.Home(), "RPS - Profile")
	return Render(c, comp)
}
