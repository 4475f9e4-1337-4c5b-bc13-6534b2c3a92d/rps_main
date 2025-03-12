package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"rps_main/internal/templates"

	"github.com/labstack/echo/v4"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type AuthError struct {
	Error string `json:"error"`
}

func ForwardAuthRequest(c echo.Context) error {
	path := accountsUrl + c.Request().URL.Path
	log.Println("Forwarding request to:", path)
	req, err := http.NewRequest(http.MethodPost, path, c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create request"})
	}
	req.Header = c.Request().Header
	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "failed to forward request"})
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var authResponse AuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
			log.Println(err)
		}
		cookie := new(http.Cookie)
		cookie.Name = "access_token"
		cookie.Value = authResponse.AccessToken
		cookie.Path = "/"
		c.SetCookie(cookie)
		c.Response().Header().Set("HX-Redirect", "/profile")
		return c.NoContent(http.StatusSeeOther)

	case http.StatusBadRequest:
		var authError AuthError
		if err := json.NewDecoder(resp.Body).Decode(&authError); err != nil {
			log.Println(err)
		}
		log.Println("Error Response:", authError.Error)
		return Render(c, templates.FormErrors([]string{authError.Error}))
	default:
		log.Println("Unexpected status code:", resp.StatusCode)
		return Render(c, templates.FormErrors([]string{"Uknown Error"}))
	}
}
