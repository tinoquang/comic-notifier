package server

import (
	"net/http"

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

	comicPage := api.ComicPage{}

	comics, err := a.store.Comic.List(ctx.Request().Context())
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range comics {
		c := comics[i]
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:      &c.ID,
			Page:    &c.Page,
			Name:    &c.Name,
			Url:     &c.URL,
			Latest:  &c.LatestChap,
			ChapUrl: &c.ChapURL,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// GetComic (GET /comics/{id})
func (a *API) GetComic(ctx echo.Context, id int) error {
	return nil
}

/* ===================== User ============================ */

// GetUserComics (GET users/{id}/comics)
func (a *API) GetUserComics(ctx echo.Context, id string) error {

	comicPage := api.ComicPage{}
	comics, err := a.store.Comic.ListByPSID(ctx.Request().Context(), id)
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range comics {
		c := comics[i]
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:      &c.ID,
			Page:    &c.Page,
			Name:    &c.Name,
			Url:     &c.URL,
			Latest:  &c.LatestChap,
			ChapUrl: &c.ChapURL,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}
