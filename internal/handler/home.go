package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"rps_main/internal/models"
	"rps_main/internal/templates"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func HandleHome(c echo.Context) error {
	_, err := getUserFromContext(c)
	if err != nil {
		log.Println("render home")
		comp := templates.Layout(templates.Home(), "RPS", false)
		return Render(c, comp)
	}

	log.Println("Home Redir to Profile")
	c.Response().Header().Set("HX-Redirect", "/profile")
	return c.Redirect(http.StatusSeeOther, "/profile")
}

func HandleRegister(c echo.Context) error {
	comp := templates.RegisterForm()
	return Render(c, comp)
}

func HandleLogin(c echo.Context) error {
	comp := templates.LoginForm()
	return Render(c, comp)
}

func HandleLogout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "access_token"
	cookie.Value = ""
	cookie.Expires = time.Now()
	c.SetCookie(cookie)
	c.Response().Header().Set("HX-Redirect", "/")
	return c.Redirect(http.StatusSeeOther, "/")
}

func renderNoProfile(c echo.Context) error {
	comp := templates.Layout(templates.Profile(nil), "RPS - Profile", false)
	return Render(c, comp)
}

func getUserFromContext(c echo.Context) (*models.User, error) {
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

func getUserGames(c echo.Context, id string) ([]models.GameData, error) {
	path := gameDataUrl + "/api/v1/game/puu-id/" + id
	req, err := http.NewRequest(http.MethodGet, path, c.Request().Body)
	if err != nil {
		return nil, errors.New("failed to create request")
	}
	req.Header.Add("Accept", "application/json")
	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to forward request")
	}
	defer resp.Body.Close()

	var games []models.GameData
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		log.Println(err)
		return nil, errors.New("failed to decode response")
	}

	return games, nil
}

func HandleProfile(c echo.Context) error {
	user, err := getUserFromContext(c)
	if err != nil {
		return renderNoProfile(c)
	}

	// Logged in user
	comp := templates.Layout(templates.Profile(user), "RPS - Profile", user != nil)
	return Render(c, comp)
}

func HandleProfileHistory(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		log.Println("No user id")
		return c.HTML(http.StatusNoContent, "")
	}

	games, err := getUserGames(c, id)
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusNoContent, "")
	}

	// Reverse the order of the games, recent games first
	for i, j := 0, len(games)-1; i < j; i, j = i+1, j-1 {
		games[i], games[j] = games[j], games[i]
	}

	return Render(c, templates.MatchHistory(games, id))
}

func HandleProfileStats(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		log.Println("No user id")
		return c.HTML(http.StatusNoContent, "")
	}

	games, err := getUserGames(c, id)
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusNoContent, "")
	}

	losses := 0
	wins := 0
	for _, game := range games {
		if game.Result.Winner == id {
			wins++
		} else {
			losses++
		}
	}

	stats := models.UserStats{
		Rating: 0,
		Games:  len(games),
		Wins:   wins,
		Losses: losses,
	}
	return Render(c, templates.UserCard(stats))
}
