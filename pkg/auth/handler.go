package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Handler main authenticate handler
type Handler struct {
	cfg *conf.Config
}

// RegisterHandler create new auth route
func RegisterHandler(g *echo.Group, cfg *conf.Config) {

	h := Handler{cfg: cfg}

	g.GET("/logged_in", h.loggedIn)
	g.POST("/login", h.login)
}

func (h *Handler) loggedIn(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) login(c echo.Context) error {

	authURL, _ := url.Parse("https://www.facebook.com/v7.0/dialog/oauth")
	q := authURL.Query()

	q.Add("client_id", h.cfg.FBSecret.AppID)
	q.Add("redirect_uri", fmt.Sprintf("%s:%s/auth", h.cfg.Host, h.cfg.Port))
	q.Add("state", "quangmt2")

	authURL.RawQuery = q.Encode()

	return c.NoContent(http.StatusOK)
}

func (h *Handler) generateJWT(c echo.Context) error {

	name := c.FormValue("name")
	userID := c.FormValue("userID")
	appToken := c.FormValue("app-token")
	err := h.validateToken(appToken, userID)
	if err != nil {
		util.Danger(err)
		return echo.ErrUnauthorized
	}

	// Create JWT and send back
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = name
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(h.cfg.JWT))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

func (h *Handler) validateToken(token, userID string) error {

	tokenResponse := map[string]json.RawMessage{}
	queries := make(map[string]string)
	queries["input_token"] = token
	queries["access_token"] = h.cfg.FBSecret.AppToken

	respBody, err := util.MakeGetRequest("https://graph.facebook.com/debug_token", queries)
	if err != nil {
		util.Danger()
		return err
	}

	err = json.Unmarshal(respBody, &tokenResponse)
	err = json.Unmarshal(tokenResponse["data"], &tokenResponse)
	if err != nil {
		util.Danger()
		return err
	}

	if util.ConvertJSONToString(tokenResponse["app_id"]) != h.cfg.FBSecret.AppID ||
		util.ConvertJSONToString(tokenResponse["user_id"]) != userID {
		return errors.New("User info is invalid")
	}

	return nil
}
