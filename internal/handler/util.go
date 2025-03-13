package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"rps_main/internal/models"
	"time"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
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

func CreateJWT(payload, secret string) (string, error) {
	signing := []byte(secret)
	now := time.Now().UTC()
	claims := make(jwt.MapClaims)
	claims["sub"] = payload
	claims["exp"] = now.Add(time.Hour * 24 * 3).Unix()
	claims["iat"] = now.Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signing)
	if err != nil {
		return "", err
	}
	return token, nil
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

func EchoJWTMiddleware(secret string, redirectURL string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:             []byte(secret),
		ContinueOnIgnoredError: false,
		ErrorHandler: func(c echo.Context, err error) error {
			log.Println("Error JWT Redirect:", err)
			c.Response().Header().Set("HX-Redirect", redirectURL)
			return c.Redirect(http.StatusSeeOther, redirectURL)
		},
	})
}

func GetUserFromContext(c echo.Context) (*models.User, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, errors.New("no user in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("no user in context")
	}

	userJSON, ok := claims["user"].(string)
	if !ok {
		return nil, errors.New("no user in context")
	}

	var user models.User
	err := json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return nil, errors.New("no user in context")
	}

	return &user, nil
}

func Render(ctx echo.Context, component templ.Component) error {
	return component.Render(ctx.Request().Context(), ctx.Response())
}
