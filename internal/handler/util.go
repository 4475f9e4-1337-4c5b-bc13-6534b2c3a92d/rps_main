package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

const (
	accountsUrl = "http://localhost:9010"
	gameDataUrl = "http://localhost:54321"
)

var HttpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func CookieAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") == "" {
			cookie, err := c.Cookie("access_token")
			if err == nil {
				c.Request().Header.Set("Authorization", "Bearer "+cookie.Value)
			}
		}
		return next(c)
	}
}

func EchoJWTMiddleware(secret string) echo.MiddlewareFunc {
 return echojwt.WithConfig(echojwt.Config{
			SigningKey:             []byte(secret),
			ContinueOnIgnoredError: false,
			ErrorHandler: func(c echo.Context, err error) error {
				log.Println("Error JWT Redirect:", err)
				c.Response().Header().Set("HX-Redirect", "/")
				return c.Redirect(http.StatusSeeOther, "/")
			},
		})
}

func Render(ctx echo.Context, component templ.Component) error {
	return component.Render(ctx.Request().Context(), ctx.Response())
}
