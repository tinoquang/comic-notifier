package server

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/api"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// API -> server handler for api endpoint
type API struct {
	store *store.Stores
}

// NewAPI return new api interface
func NewAPI(s *store.Stores) *API {
	return &API{store: s}
}

// Comics (GET /comics)
func (a *API) Comics(ctx echo.Context) error {

	// _, offset, limit := listArgs(params.Q, params.Limit, params.Offset)
	opt := store.NewComicsListOptions("", 0, 0)
	comicPage := api.ComicPage{}

	comics, err := a.store.Comic.List(ctx.Request().Context(), opt)
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.JSON(http.StatusOK, &comicPage)
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range comics {
		c := comics[i]
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:         &c.ID,
			Page:       &c.Page,
			Name:       &c.Name,
			Url:        &c.URL,
			LatestChap: &c.LatestChap,
			ImgURL:     &c.CloudImg,
			ChapURL:    &c.ChapURL,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// GetComic (GET /comics/{id})
func (a *API) GetComic(ctx echo.Context, id int) error {

	c, err := a.store.Comic.Get(ctx.Request().Context(), id)
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	comic := api.Comic{
		Id:         &c.ID,
		Page:       &c.Page,
		Name:       &c.Name,
		Url:        &c.URL,
		LatestChap: &c.LatestChap,
		ImgURL:     &c.CloudImg,
		ChapURL:    &c.ChapURL,
	}
	return ctx.JSON(http.StatusOK, &comic)
}

/* ===================== User ============================ */

// Users (GET /user)
func (a *API) Users(ctx echo.Context) error {

	userPage := []api.User{}

	users, err := a.store.User.List(ctx.Request().Context())
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.JSON(http.StatusOK, &userPage)
		}

		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range users {
		u := users[i]

		userPage = append(userPage, api.User{
			Psid:       &u.PSID,
			Appid:      &u.AppID,
			Name:       &u.Name,
			ProfilePic: &u.ProfilePic,
			Comics:     nil,
		})

	}
	return ctx.JSON(http.StatusOK, &userPage)
}

// GetUser (GET /user/{id})
func (a *API) GetUser(ctx echo.Context, userAppID string) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	u, err := a.store.User.GetByFBID(ctx.Request().Context(), "appID", userAppID)
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	user := api.User{
		Psid:       &u.PSID,
		Appid:      &u.AppID,
		Name:       &u.Name,
		ProfilePic: &u.ProfilePic,
		Comics:     nil,
	}

	return ctx.JSON(http.StatusOK, &user)
}

// GetUserComics (GET users/{id}/comics)
func (a *API) GetUserComics(ctx echo.Context, userAppID string, params api.GetUserComicsParams) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	u, err := a.store.User.GetByFBID(ctx.Request().Context(), "appID", userAppID)
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	q, limit, offset := listArgs(params.Q, params.Limit, params.Offset)
	opt := store.NewComicsListOptions(q, limit, offset)

	comicPage := api.ComicPage{}
	comics, err := a.store.Comic.ListByPSID(ctx.Request().Context(), opt, u.PSID)
	if err != nil {
		// Return empty list if not found comic
		if err == util.ErrNotFound {
			comicPage.Comics = []api.Comic{}
			return ctx.JSON(http.StatusOK, &comicPage)
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusNotFound)
	}

	for i := range comics {
		c := comics[i]
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:         &c.ID,
			Page:       &c.Page,
			Name:       &c.Name,
			Url:        &c.URL,
			LatestChap: &c.LatestChap,
			ImgURL:     &c.CloudImg,
			ChapURL:    &c.ChapURL,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// SubscribeComic (POST /users/{id}/comics)
func (a *API) SubscribeComic(ctx echo.Context, userAppID string) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	u, err := a.store.User.GetByFBID(ctx.Request().Context(), "appID", userAppID)
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	comicURL := ctx.FormValue("comic")
	if comicURL == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}

	c, err := a.store.SubscribeComic(ctx.Request().Context(), u.PSID, comicURL)
	if err != nil {
		if err == util.ErrInvalidURL {
			return ctx.NoContent(http.StatusBadRequest)
		}
		return ctx.NoContent(http.StatusInternalServerError)
	}

	comic := api.Comic{
		Id:         &c.ID,
		Page:       &c.Page,
		Name:       &c.Name,
		Url:        &c.URL,
		LatestChap: &c.LatestChap,
		ImgURL:     &c.CloudImg,
		ChapURL:    &c.ChapURL,
	}

	return ctx.JSON(http.StatusOK, &comic)
}

// UnsubscribeComic (DELETE /users/{user_id}/comics/{id})
func (a *API) UnsubscribeComic(ctx echo.Context, userAppID string, comicID int) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	u, err := a.store.User.GetByFBID(ctx.Request().Context(), "appID", userAppID)
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Validate if user has subscribed to this comic, if not then this request is invalid
	c, err := a.store.Comic.CheckComicSubscribe(ctx.Request().Context(), u.PSID, comicID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = a.store.Subscriber.Delete(ctx.Request().Context(), u.PSID, comicID)
	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	s, err := a.store.Subscriber.ListByComicID(ctx.Request().Context(), comicID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Check if no user subscribe to this comic --> remove this comic from DB
	if len(s) == 0 {
		a.store.Comic.Delete(ctx.Request().Context(), c)
	}

	return ctx.NoContent(http.StatusOK)
}

func userHasAccess(ctx echo.Context, appID string) bool {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwt.StandardClaims)

	if claims.Id != appID {
		return false
	}

	return true
}
