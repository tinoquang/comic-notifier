package server

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/api"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

// API -> server handler for api endpoint
type API struct {
	cfg   *conf.Config
	store *store.Stores
}

// NewAPI return new api interface
func NewAPI(c *conf.Config, s *store.Stores) *API {
	return &API{cfg: c, store: s}
}

// Comics (GET /comics)
func (a *API) Comics(ctx echo.Context) error {

	// _, offset, limit := listArgs(params.Q, params.Limit, params.Offset)
	opt := store.NewComicsListOptions("", 0, 0)
	comicPage := api.ComicPage{}

	comics, err := a.store.Comic.List(ctx.Request().Context(), opt)
	if err != nil {
		if err == store.ErrNotFound {
			return ctx.JSON(http.StatusOK, &comicPage)
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range comics {
		c := comics[i]
		imgURL := c.ImgurLink.Value()
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:         &c.ID,
			Page:       &c.Page,
			Name:       &c.Name,
			Url:        &c.URL,
			LatestChap: &c.LatestChap,
			ImgURL:     &imgURL,
			ChapURL:    &c.ChapURL,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// GetComic (GET /comics/{id})
func (a *API) GetComic(ctx echo.Context, id int) error {

	c, err := a.store.Comic.Get(ctx.Request().Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	imgURL := c.ImgurLink.Value()
	comic := api.Comic{
		Id:         &c.ID,
		Page:       &c.Page,
		Name:       &c.Name,
		Url:        &c.URL,
		LatestChap: &c.LatestChap,
		ImgURL:     &imgURL,
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
		if err == store.ErrNotFound {
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
func (a *API) GetUser(ctx echo.Context, userPSID string) error {

	if !userHasAccess(ctx, userPSID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	u, err := a.store.User.GetByFBID(ctx.Request().Context(), "psid", userPSID)
	if err != nil {
		if err == store.ErrNotFound {
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
func (a *API) GetUserComics(ctx echo.Context, userPSID string, params api.GetUserComicsParams) error {

	if !userHasAccess(ctx, userPSID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	q, limit, offset := listArgs(params.Q, params.Limit, params.Offset)
	opt := store.NewComicsListOptions(q, limit, offset)

	comicPage := api.ComicPage{}
	comics, err := a.store.Comic.ListByPSID(ctx.Request().Context(), opt, userPSID)
	if err != nil {
		// Return empty list if not found comic
		if err == store.ErrNotFound {
			comicPage.Comics = []api.Comic{}
			return ctx.JSON(http.StatusOK, &comicPage)
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusNotFound)
	}

	for i := range comics {
		c := comics[i]
		imgURL := c.ImgurLink.Value()
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:         &c.ID,
			Page:       &c.Page,
			Name:       &c.Name,
			Url:        &c.URL,
			LatestChap: &c.LatestChap,
			ImgURL:     &imgURL,
			ChapURL:    &c.ChapURL,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// SubscribeComic (POST /users/{id}/comics)
func (a *API) SubscribeComic(ctx echo.Context, userPSID string) error {

	if !userHasAccess(ctx, userPSID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	comicURL := ctx.FormValue("comic")
	if comicURL == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}

	c, err := a.store.SubscribeComic(ctx.Request().Context(), "psid", userPSID, comicURL)
	if err != nil {
		if err == store.ErrInvalidURL {
			return ctx.NoContent(http.StatusBadRequest)
		}
		return ctx.NoContent(http.StatusInternalServerError)
	}

	imgURL := c.ImgurLink.Value()
	comic := api.Comic{
		Id:         &c.ID,
		Page:       &c.Page,
		Name:       &c.Name,
		Url:        &c.URL,
		LatestChap: &c.LatestChap,
		ImgURL:     &imgURL,
		ChapURL:    &c.ChapURL,
	}

	return ctx.JSON(http.StatusOK, &comic)
}

// UnsubscribeComic (DELETE /users/{user_id}/comics/{id})
func (a *API) UnsubscribeComic(ctx echo.Context, userPSID string, comicID int) error {

	if !userHasAccess(ctx, userPSID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	// Validate if user has subscribed to this comic, if not then this request is invalid
	c, err := a.store.Comic.CheckComicSubscribe(ctx.Request().Context(), userPSID, comicID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = a.store.Subscriber.Delete(ctx.Request().Context(), userPSID, comicID)
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
		img.DeleteImg(string(c.ImgurID))
		a.store.Comic.Delete(ctx.Request().Context(), comicID)
	}

	return ctx.NoContent(http.StatusOK)
}

func userHasAccess(ctx echo.Context, psid string) bool {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwt.StandardClaims)

	if claims.Id != psid {
		return false
	}

	return true
}
