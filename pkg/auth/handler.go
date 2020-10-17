package auth

import (
	"encoding/json"
	"net/http"
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

	g.POST("", h.login)
}

func (h *Handler) login(c echo.Context) error {

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
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

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
		util.Danger()
		return err
	}

	return nil
}
