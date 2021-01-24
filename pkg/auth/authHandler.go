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
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Handler main authenticate handler
type Handler struct {
	store db.Stores
}

// RegisterHandler create new auth route
func RegisterHandler(g *echo.Group, store db.Stores) {

	h := Handler{store: store}

	g.GET("/auth", h.auth)
	g.GET("/status", h.loggedIn, middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(conf.Cfg.JWT.SecretKey),
		Claims:      &jwt.StandardClaims{},
		TokenLookup: "cookie:_session",
	}))
	g.GET("/login", h.login)
	g.GET("/logout", h.logout)

}

func (h *Handler) loggedIn(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (h *Handler) login(c echo.Context) error {

	authURL, _ := url.Parse("https://www.facebook.com/v8.0/dialog/oauth")
	q := authURL.Query()

	q.Add("client_id", conf.Cfg.FBSecret.AppID)
	q.Add("redirect_uri", fmt.Sprintf("%s:%s/auth", conf.Cfg.Host, conf.Cfg.Port))
	q.Add("state", "quangmt2")

	authURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, authURL.String())
}

func (h *Handler) logout(ctx echo.Context) error {

	cookie := &http.Cookie{
		Name:     "_session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	}
	ctx.SetCookie(cookie)
	return ctx.NoContent(http.StatusOK)
}

func (h *Handler) auth(ctx echo.Context) error {

	// Get FB access token
	code := ctx.QueryParam("code")
	state := ctx.QueryParam("state")

	/* Exchange token using given code */
	queries := map[string]string{
		"client_id":     conf.Cfg.FBSecret.AppID,
		"client_secret": conf.Cfg.FBSecret.AppSecret,
		"code":          code,
	}

	if conf.Cfg.Host == "http://localhost" {
		queries["redirect_uri"] = fmt.Sprintf("%s:8080/auth", conf.Cfg.Host)
	} else {
		queries["redirect_uri"] = fmt.Sprintf("%s/auth", conf.Cfg.Host)
	}

	respBody, err := util.MakeGetRequest(conf.Cfg.Webhook.GraphEndpoint+"/oauth/access_token", queries)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	tokenRes := make(map[string]json.RawMessage)
	err = json.Unmarshal(respBody, &tokenRes)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	userAppID, err := h.validateToken(util.ConvertJSONToString(tokenRes["access_token"]))
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	err = h.store.CheckUserExist(ctx.Request().Context(), userAppID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	jwtCookie, err := h.generateJWT(userAppID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	cookie := &http.Cookie{
		Name:     "_session",
		Value:    jwtCookie,
		Expires:  time.Now().AddDate(0, 0, 1),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	ctx.SetCookie(cookie)

	cookie = &http.Cookie{
		Name:    "uaid",
		Value:   userAppID,
		Expires: time.Now().AddDate(0, 1, 0),
	}
	ctx.SetCookie(cookie)

	if conf.Cfg.Host == "http://localhost" {
		return ctx.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s:3000%s", conf.Cfg.Host, state))
	}

	return ctx.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", conf.Cfg.Host, state))
}

func (h *Handler) generateJWT(userAppID string) (string, error) {

	claims := &jwt.StandardClaims{
		Issuer:    conf.Cfg.JWT.Issuer,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().AddDate(0, 1, 0).Unix(),
		Audience:  conf.Cfg.JWT.Audience,
		Id:        userAppID,
	}
	// Create JWT and send back
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(conf.Cfg.JWT.SecretKey))
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
	queries["access_token"] = conf.Cfg.FBSecret.AppToken

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

	if util.ConvertJSONToString(tokenResponse["app_id"]) != conf.Cfg.FBSecret.AppID {
		return "", errors.New("Access token is invalid")
	}

	userAppID = util.ConvertJSONToString(tokenResponse["user_id"])
	return
}
