package main

import (
	"log"
	"os"
	"rps_main/internal/game"
	"rps_main/internal/handler"

	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	e := echo.New()
	e.HideBanner = true
	e.Static("/", "static")
	e.GET("/", handler.HandleHome, handler.CookieAuthMiddleware, echojwt.WithConfig(echojwt.Config{
		SigningKey:             []byte(secret),
		ContinueOnIgnoredError: true,
		SuccessHandler: func(c echo.Context) {
			log.Println("JWT On Home Auth")
			c.Response().Header().Set("HX-Redirect", "/profile")
		},
		ErrorHandler: func(c echo.Context, err error) error {
			log.Println("No JWT On Home no Auth", err)
			return nil
		},
	}))

	e.GET("/login", handler.HandleLogin)
	e.GET("/register", handler.HandleRegister)
	e.GET("/logout", handler.HandleLogout)

	e.GET("/profile", handler.HandleProfile, handler.CookieAuthMiddleware, handler.EchoJWTMiddleware(secret))
	e.GET("/profile/:id/history", handler.HandleProfileHistory)
	e.GET("/profile/:id/stats", handler.HandleProfileStats)

	// Forward requests to microservices

	auth := e.Group("/auth")
	auth.POST("/login", handler.ForwardAuthRequest)
	auth.POST("/register", handler.ForwardAuthRequest)

	// mmUrl, err := url.Parse("http://localhost:9100")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// mmTargets := []*middleware.ProxyTarget{{URL: mmUrl}}
	// e.GET("/mm/cancel", handler.CancelMatchmaking)
	// mm := e.Group("/mm")
	// mm.POST("/join", handler.ForwardMatchmakingRequest)
	// mm.POST("/leave", handler.ForwardMatchmakingRequest)
	// mm.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(mmTargets)))

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
