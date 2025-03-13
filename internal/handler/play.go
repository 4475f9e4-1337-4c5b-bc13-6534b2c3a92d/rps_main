package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"rps_main/internal/models"
	"rps_main/internal/templates"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetNewServerID(settings models.GameSettings) (string, error) {
	url := os.Getenv("GAME_SERVER_URI")
	if url == "" {
		return "", errors.New("GAME_SERVER_URI is missing")
	}
	url = url + "/game"

	settingsJson, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(settingsJson))
	if err != nil {
		return "", errors.New("Failed to create request")
	}
	token, err := CreateJWT("main-ai", os.Getenv("JWT_MM_SECRET"))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Failed to get server ID")
	}

	var newServerResponse struct {
		ServerID string `json:"server_id"`
	}
	if json.Unmarshal(body, &newServerResponse); err != nil {
		return "", errors.New("Failed to unmarshal response")
	}

	return newServerResponse.ServerID, nil
}

func HandlePlay(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		log.Println("No user in context")
		return err
	}

	mode := c.FormValue("mode")
	best := c.FormValue("bestOf")
	bestOf, err := strconv.Atoi(best)
	if err != nil {
		log.Println("Invalid bestOf value")
		return err
	}

	log.Println("Starting game with bestOf:", bestOf, "mode:", mode)
	if mode == "" {
		settings := models.GameSettings{
			GameMode:  "rps",
			Type:      "ai",
			BestOf:    bestOf,
			PlayerOne: user.ID,
			PlayerTwo: "ai",
		}
		id, err := GetNewServerID(settings)
		if err != nil {
			log.Println("Failed to get new server ID")
			return err
		}

		c.SetParamNames("id")
		c.SetParamValues(id)
		return HandleGame(c)
	} else {
		mode = "pvp"
		comp := templates.InQueue(best, mode)
		return Render(c, comp)
	}
}

func HandleGame(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		c.Response().Header().Set("HX-Redirect", "/profile")
		return c.Redirect(http.StatusFound, "/profile")
	}
	comp := templates.Game(id)
	return Render(c, comp)
}
