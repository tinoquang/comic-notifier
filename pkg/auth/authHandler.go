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
	"github.com/labstack/echo/v4/middleware"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Handler main authenticate handler
type Handler struct {
	cfg   *conf.Config
	store *store.Stores
}

// RegisterHandler create new auth route
func RegisterHandler(g *echo.Group, cfg *conf.Config, store *store.Stores) {

	h := Handler{cfg: cfg, store: store}

	g.GET("/auth", h.auth)
	g.GET("/status", h.loggedIn, middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(cfg.JWT.SecretKey),
		Claims:      &jwt.StandardClaims{},
		TokenLookup: "cookie:_session",
	}))
	g.GET("/login", h.login)
	g.GET("/logout", h.logout)

}

func (h *Handler) loggedIn(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) login(c echo.Context) error {

	authURL, _ := url.Parse("https://www.facebook.com/v8.0/dialog/oauth")
	q := authURL.Query()

	q.Add("client_id", h.cfg.FBSecret.AppID)
	q.Add("redirect_uri", fmt.Sprintf("%s:%s/auth", h.cfg.Host, h.cfg.Port))
	q.Add("state", "quangmt2")

	authURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, authURL.String())
}

func (h *Handler) logout(c echo.Context) error {

	cookie := &http.Cookie{
		Name:     "_session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	return c.NoContent(http.StatusOK)
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
		"client_secret": h.cfg.FBSecret.AppSecret,
		"code":          code,
	}

	if h.cfg.Host == "http://localhost" {
		queries["redirect_uri"] = fmt.Sprintf("%s:8080/auth", h.cfg.Host)
	} else {
		queries["redirect_uri"] = fmt.Sprintf("%s/auth", h.cfg.Host)
	}

	respBody, err := util.MakeGetRequest(h.cfg.Webhook.GraphEndpoint+"/oauth/access_token", queries)
	if err != nil {
		logging.Danger(err)
		return c.NoContent(http.StatusBadRequest)
	}

	tokenRes := make(map[string]json.RawMessage)
	err = json.Unmarshal(respBody, &tokenRes)
	if err != nil {
		logging.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	userAppID, err := h.validateToken(util.ConvertJSONToString(tokenRes["access_token"]))
	if err != nil {
		logging.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	user, err := util.GetUserInfoFromFB(h.cfg, "appid", userAppID)
	if err != nil {
		logging.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	jwtCookie, err := h.generateJWT(user.AppID)
	if err != nil {
		logging.Danger(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// Save user info to DB to get later
	// err = h.store.User.Create(c.Request().Context(), user)
	// if err != nil {
	// 	logging.Danger(err)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }

	cookie := &http.Cookie{
		Name:     "_session",
		Value:    jwtCookie,
		Expires:  time.Now().AddDate(0, 0, 1),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	cookie = &http.Cookie{
		Name:    "upid",
		Value:   user.AppID,
		Expires: time.Now().AddDate(0, 1, 0),
	}
	c.SetCookie(cookie)

	if h.cfg.Host == "http://localhost" {
		return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s:3000", h.cfg.Host))
	}

	return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s", h.cfg.Host))
}

func (h *Handler) generateJWT(userAppID string) (string, error) {

	claims := &jwt.StandardClaims{
		Issuer:    h.cfg.JWT.Issuer,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().AddDate(0, 1, 0).Unix(),
		Audience:  h.cfg.JWT.Audience,
		Id:        userAppID,
	}
	// Create JWT and send back
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(h.cfg.JWT.SecretKey))
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
		logging.Danger(err)
		return
	}

	err = json.Unmarshal(respBody, &tokenResponse)
	err = json.Unmarshal(tokenResponse["data"], &tokenResponse)
	if err != nil {
		logging.Danger()
		return
	}

	if util.ConvertJSONToString(tokenResponse["app_id"]) != h.cfg.FBSecret.AppID {
		return "", errors.New("Access token is invalid")
	}

	userAppID = util.ConvertJSONToString(tokenResponse["user_id"])
	return
}
