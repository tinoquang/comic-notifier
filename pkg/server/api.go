package server

import (
	"database/sql"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/tinoquang/comic-notifier/pkg/api"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// API -> server handler for api endpoint
type API struct {
	store db.Stores
}

// NewAPI return new api interface
func NewAPI(s db.Stores) *API {
	return &API{store: s}
}

// Comics (GET /comics)
func (a *API) Comics(ctx echo.Context, params api.ComicsParams) error {

	comicPage := api.ComicPage{}
	comics, err := a.store.ListComics(ctx.Request().Context())

	if err != nil {
		if err == util.ErrNotFound {
			return ctx.JSON(http.StatusOK, &comicPage)
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range comics {
		c := comics[i]
		id := int(c.ID)
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:         &id,
			Page:       &c.Page,
			Name:       &c.Name,
			Url:        &c.Url,
			LatestChap: &c.LatestChap,
			ImgURL:     &c.CloudImgUrl,
			ChapURL:    &c.ChapUrl,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// GetComic (GET /comics/{id})
func (a *API) GetComic(ctx echo.Context, id int) error {

	c, err := a.store.GetComic(ctx.Request().Context(), int32(id))
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	comic := api.Comic{
		Id:         &id,
		Page:       &c.Page,
		Name:       &c.Name,
		Url:        &c.Url,
		LatestChap: &c.LatestChap,
		ImgURL:     &c.CloudImgUrl,
		ChapURL:    &c.ChapUrl,
	}
	return ctx.JSON(http.StatusOK, &comic)
}

/* ===================== User ============================ */

// Users (GET /user)
func (a *API) Users(ctx echo.Context) error {

	userPage := []api.User{}
	users, err := a.store.ListUsers(ctx.Request().Context())
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.JSON(http.StatusOK, &userPage)
		}

		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for i := range users {
		u := users[i]
		responseUser := createResponseUser(u)
		userPage = append(userPage, responseUser)

	}
	return ctx.JSON(http.StatusOK, &userPage)
}

// GetUser (GET /user/{id})
func (a *API) GetUser(ctx echo.Context, userAppID string) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	u, err := a.store.GetUserByAppID(ctx.Request().Context(), sql.NullString{String: userAppID, Valid: true})
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "404 - Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	user := createResponseUser(u)
	return ctx.JSON(http.StatusOK, &user)
}

// GetUserComics (GET users/{id}/comics)
func (a *API) GetUserComics(ctx echo.Context, userAppID string, params api.GetUserComicsParams) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	user, err := a.store.GetUserByAppID(ctx.Request().Context(), sql.NullString{String: userAppID, Valid: true})
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	comicPage := api.ComicPage{}

	comics, err := a.store.ListComicsByAppID(ctx.Request().Context(), user.Appid.String)
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
		comicID := int(c.ID)
		comicPage.Comics = append(comicPage.Comics, api.Comic{
			Id:         &comicID,
			Page:       &c.Page,
			Name:       &c.Name,
			Url:        &c.Url,
			LatestChap: &c.LatestChap,
			ImgURL:     &c.CloudImgUrl,
			ChapURL:    &c.ChapUrl,
		})
	}
	return ctx.JSON(http.StatusOK, &comicPage)
}

// SubscribeComic (POST /users/{id}/comics)
// func (a *API) SubscribeComic(ctx echo.Context, userAppID string) error {

// 	if !userHasAccess(ctx, userAppID) {
// 		return ctx.NoContent(http.StatusForbidden)
// 	}

// 	user, err := a.store.GetUserByAppID(ctx.Request().Context(), sql.NullString{String: userAppID, Valid: true})
// 	if err != nil {
// 		if err == util.ErrNotFound {
// 			return ctx.String(http.StatusNotFound, "Not found")
// 		}
// 		logging.Danger(err)
// 		return ctx.NoContent(http.StatusInternalServerError)
// 	}

// 	comicURL := ctx.FormValue("comic")
// 	if comicURL == "" {
// 		return ctx.NoContent(http.StatusBadRequest)
// 	}

// 	c, err := a.store.SubscribeComic(ctx.Request().Context(), u.PSID, comicURL)
// 	if err != nil {
// 		if err == util.ErrInvalidURL {
// 			return ctx.NoContent(http.StatusBadRequest)
// 		}
// 		return ctx.NoContent(http.StatusInternalServerError)
// 	}

// 	comicID := int(c.ID)
// 	comic := api.Comic{
// 		Id:         &comicID,
// 		Page:       &c.Page,
// 		Name:       &c.Name,
// 		Url:        &c.Url,
// 		LatestChap: &c.LatestChap,
// 		ImgURL:     &c.CloudImgUrl,
// 		ChapURL:    &c.ChapUrl,
// 	}

// 	return ctx.JSON(http.StatusOK, &comic)
// }

// UnsubscribeComic (DELETE /users/{user_id}/comics/{id})
func (a *API) UnsubscribeComic(ctx echo.Context, userAppID string, comicID int) error {

	if !userHasAccess(ctx, userAppID) {
		return ctx.NoContent(http.StatusForbidden)
	}

	_, err := a.store.GetUserByAppID(ctx.Request().Context(), sql.NullString{String: userAppID, Valid: true})
	if err != nil {
		if err == util.ErrNotFound {
			return ctx.String(http.StatusNotFound, "Not found")
		}
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Validate if user has subscribed to this comic, if not then this request is invalid
	_, err = a.store.GetSubscriberByAppIDAndComicID(ctx.Request().Context(), db.GetSubscriberByAppIDAndComicIDParams{
		UserAppid: userAppID,
		ComicID:   int32(comicID),
	})
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = a.store.DeleteSubscriberByAppID(ctx.Request().Context(), db.DeleteSubscriberByAppIDParams{
		UserAppid: userAppID,
		ComicID:   int32(comicID),
	})
	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	s, err := a.store.ListSubscriberByComicID(ctx.Request().Context(), int32(comicID))
	if err != nil {
		logging.Danger(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// Check if no user subscribe to this comic --> remove this comic from DB
	if len(s) == 0 {
		a.store.DeleteComic(ctx.Request().Context(), int32(comicID))
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

func createResponseUser(u db.User) (responseUser api.User) {

	if u.Psid.Valid {
		responseUser.Psid = &u.Psid.String
	} else {
		responseUser.Psid = nil
	}

	if u.Appid.Valid {
		responseUser.Appid = &u.Appid.String
	} else {
		responseUser.Appid = nil
	}

	if u.ProfilePic.Valid {
		responseUser.ProfilePic = &u.ProfilePic.String
	} else {
		responseUser.ProfilePic = nil
	}

	responseUser.Name = &u.Name
	responseUser.Comics = nil
	return
}
