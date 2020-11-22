package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/api"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
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
	return nil
}

/* ===================== User ============================ */

// Users (GET /user)
func (a *API) Users(ctx echo.Context) error {

	return nil
}

// GetUser (GET /user/{id})
func (a *API) GetUser(ctx echo.Context, id string) error {

	u, err := a.store.User.GetByFBID(ctx.Request().Context(), "psid", id)
	if err != nil {
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
func (a *API) GetUserComics(ctx echo.Context, psid string, params api.GetUserComicsParams) error {

	q, limit, offset := listArgs(params.Q, params.Limit, params.Offset)
	opt := store.NewComicsListOptions(q, limit, offset)

	comicPage := api.ComicPage{}
	comics, err := a.store.Comic.ListByPSID(ctx.Request().Context(), opt, psid)
	if err != nil {
		// Return empty list if not found comic
		if strings.Contains(err.Error(), "not found") {
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
func (a *API) SubscribeComic(ctx echo.Context, id string) error {

	comicURL := ctx.FormValue("comic")
	if comicURL == "" {
		return ctx.NoContent(http.StatusBadRequest)
	}

	c, err := a.store.SubscribeComic(ctx.Request().Context(), "psid", id, comicURL)
	if err != nil {
		if strings.Contains(err.Error(), "check your URL") {
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
func (a *API) UnsubscribeComic(ctx echo.Context, userID string, comicID int) error {

	// Validate if user has subscribed to this comic, if not then this request is invalid
	c, err := a.store.Comic.CheckComicSubscribe(ctx.Request().Context(), userID, comicID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = a.store.Subscriber.Delete(ctx.Request().Context(), userID, comicID)
	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	s, err := a.store.Subscriber.ListByComicID(ctx.Request().Context(), comicID)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logging.Info(s, c)
	// Check if no user subscribe to this comic --> remove this comic from DB
	// if len(s) == 0 {
	// 	img.DeleteImg(string(c.ImgurID))
	// 	a.store.Comic.Delete(ctx.Request().Context(), comicID)
	// }

	return ctx.NoContent(http.StatusOK)
}
