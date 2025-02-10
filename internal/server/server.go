package server

import (
	"rps_main/internal/handler"

	"github.com/labstack/echo/v4"
)

type server struct {
	port string
	echo *echo.Echo
}

func NewServer(port string) (*server, error) {
	return &server{
		port: port,
	}, nil
}

func (s *server) Start() error {
	s.echo = echo.New()
	s.echo.HideBanner = true
	s.echo.Static("/", "static")
	s.echo.GET("/", handler.HandleHome)
	s.echo.POST("/play", handler.HandlePlay)

	s.echo.Logger.Fatal(s.echo.Start(s.port))
	return nil
}
