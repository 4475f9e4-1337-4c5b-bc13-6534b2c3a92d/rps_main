package handler

import (
	"log"

	"github.com/labstack/echo/v4"
)

func ForwardRegister(c echo.Context) error {
	log.Println("ForwardRegister")
	return nil
}

func ForwardLogin(c echo.Context) error {
	log.Println("ForwardLogin")
	return nil
}
