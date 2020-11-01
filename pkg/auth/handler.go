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

	g.GET("/login", h.login)
	g.GET("/auth", h.auth)
}

func (h *Handler) login(c echo.Context) error {

	authURL, _ := url.Parse("https://www.facebook.com/v8.0/dialog/oauth")
	q := authURL.Query()

	q.Add("client_id", h.cfg.FBSecret.AppID)
	q.Add("redirect_uri", fmt.Sprintf("%s:%s/auth", h.cfg.Host, h.cfg.Port))
	q.Add("state", "quangmt2")

	authURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusMovedPermanently, authURL.String())
}

func (h *Handler) auth(c echo.Context) error {

	// Get FB access token
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if state != "quangmt2" {
		return c.NoContent(http.StatusBadRequest)
	}

	/* Exchange token using given code */
	queries := map[string]string{
		"client_id":     h.cfg.FBSecret.AppID,
		"redirect_uri":  fmt.Sprintf("%s:%s/auth", h.cfg.Host, h.cfg.Port),
		"client_secret": h.cfg.FBSecret.AppSecret,
		"code":          code,
	}

	respBody, err := util.MakeGetRequest(h.cfg.Webhook.GraphEndpoint+"/oauth/access_token", queries)
	if err != nil {
		util.Danger(err)
		return c.NoContent(http.StatusBadRequest)
	}

	tokenRes := make(map[string]json.RawMessage)
	err = json.Unmarshal(respBody, &tokenRes)
	if err != nil {
		util.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	userAppID, err := h.validateToken(util.ConvertJSONToString(tokenRes["access_token"]))
	if err != nil {
		util.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	jwtCookie, err := h.generateJWT(userAppID)
	if err != nil {
		util.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	cookie := &http.Cookie{
		Name:     "_session",
		Value:    jwtCookie,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	return c.Redirect(http.StatusPermanentRedirect, "/")
}

func (h *Handler) generateJWT(userAppID string) (string, error) {

	// Create JWT and send back
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = userAppID
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(h.cfg.JWT))
	if err != nil {
		return "", err
	}

	return t, nil

}

func (h *Handler) validateToken(token string) (userAppID string, err error) {

	userAppID = ""
	tokenResponse := map[string]json.RawMessage{}
	queries := make(map[string]string)
	queries["input_token"] = token
	queries["access_token"] = h.cfg.FBSecret.AppToken

	respBody, err := util.MakeGetRequest("https://graph.facebook.com/debug_token", queries)
	if err != nil {
		util.Danger(err)
		return
	}

	err = json.Unmarshal(respBody, &tokenResponse)
	err = json.Unmarshal(tokenResponse["data"], &tokenResponse)
	if err != nil {
		util.Danger()
		return
	}

	if util.ConvertJSONToString(tokenResponse["app_id"]) != h.cfg.FBSecret.AppID {
		return "", errors.New("Access token is invalid")
	}

	userAppID = util.ConvertJSONToString(tokenResponse["user_id"])
	return
}
