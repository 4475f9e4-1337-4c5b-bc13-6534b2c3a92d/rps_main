package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"rps_main/internal/game"
	"rps_main/internal/handler"

	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Static("/", "static")
	e.GET("/", handler.HandleHome)

	e.GET("/login", handler.HandleLogin)
	e.GET("/register", handler.HandleRegister)

	secret := os.Getenv("JWT_SECRET")
	e.GET("/profile", handler.HandleProfile, echojwt.WithConfig(echojwt.Config{
		SigningKey:             []byte(secret),
		ContinueOnIgnoredError: true,
		ErrorHandler: func(c echo.Context, err error) error {
			c.Redirect(http.StatusFound, "/")
			return nil
		},
	}))

	authUrl, err := url.Parse("http://localhost:9010")
	if err != nil {
		log.Fatal(err)
	}
	authTargets := []*middleware.ProxyTarget{
		{URL: authUrl},
	}

	mmUrl, err := url.Parse("http://localhost:9100")
	if err != nil {
		log.Fatal(err)
	}
	mmTargets := []*middleware.ProxyTarget{
		{URL: mmUrl},
	}

	auth := e.Group("/auth")
	auth.POST("/login", handler.ForwardLogin)
	auth.POST("/register", handler.ForwardRegister)
	auth.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(authTargets)))

	mm := e.Group("/mm")
	mm.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(mmTargets)))

	e.POST("/play", handler.HandlePlay)
	e.GET("/game/:id", func(c echo.Context) error {
		gameId := c.Param("id")
		if gs, ok := game.GetServer(gameId); ok {
			return gs.Connect(c)
		}
		return c.Redirect(301, "/")
	})

	e.Logger.Fatal(e.Start(":9000"))
}
