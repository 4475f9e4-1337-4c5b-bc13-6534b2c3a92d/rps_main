package main

import (
	"log"
	"os"
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

	e.GET("/profile", handler.HandleProfile, handler.CookieAuthMiddleware, handler.EchoJWTMiddleware(secret, "/"))
	e.GET("/profile/menu", handler.HandleProfileMenu)
	e.GET("/profile/:id/history", handler.HandleProfileHistory)
	e.GET("/profile/:id/stats", handler.HandleProfileStats)

	auth := e.Group("/auth")
	auth.POST("/login", handler.ForwardAuthRequest)
	auth.POST("/register", handler.ForwardAuthRequest)

	e.POST("/play", handler.HandlePlay, handler.CookieAuthMiddleware, handler.EchoJWTMiddleware(secret, "/profile"))
	e.GET("/game/:id", handler.HandleGame, handler.CookieAuthMiddleware, handler.EchoJWTMiddleware(secret, "/profile"))

	e.Logger.Fatal(e.Start(":9000"))
}
