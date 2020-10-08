package server

import (
	"github.com/tinoquang/comic-notifier/pkg/conf"
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

// // GetPage (GET /pages/{name})
// func (a *API) GetPage(ctx context.Context, name string) (*model.Page, error) {

// 	return a.store.Page.GetByName(ctx, name)
// }

// // Comics (GET /comics)
// func (a *API) Comics(ctx context.Context) ([]model.Comic, error) {

// 	return a.store.Comic.List(ctx)
// }

// /* ===================== User ============================ */

// // Users (GET /users)
// func (a *API) Users(ctx context.Context) ([]model.User, error) {

// 	return a.store.User.List(ctx)
// }

// // GetUserByPSID (GET /users?psid=)
// func (a *API) GetUserByPSID(ctx context.Context, psid string) (*model.User, error) {
// 	return a.store.User.GetByFBID(ctx, "psid", psid)
// }

// // GetUsersByComicID (GET /comics/{id}/users)
// func (a *API) GetUsersByComicID(ctx context.Context, comicID int) ([]model.User, error) {

// 	return a.store.User.ListByComicID(ctx, comicID)
// }

// /* ===================== Comic ============================ */

// // UpdateComic use when new chapter realease
// func (a *API) UpdateComic(ctx context.Context, comic *model.Comic) (bool, error) {

// 	updated := true
// 	err := getComicInfo(ctx, comic)

// 	if err != nil {
// 		if strings.Contains(err.Error(), "No new chapter") {
// 			updated = false
// 		} else {
// 			util.Danger()
// 		}
// 		return updated, err
// 	}

// 	err = a.store.Comic.Update(ctx, comic)
// 	return updated, err
// }

// // SubscribeComic (POST /users/{id}/comics)
// func (a *API) SubscribeComic(ctx context.Context, field string, id string, comicURL string) (*model.Comic, error) {

// 	parsedURL, err := url.Parse(comicURL)
// 	if err != nil || parsedURL.Host == "" {
// 		return nil, errors.New("Please check your URL")
// 	}

// 	// Check page support, if not send back "Page is not supported"
// 	_, err = a.store.Page.GetByName(ctx, parsedURL.Hostname())
// 	if err != nil {
// 		return nil, errors.New("Sorry, page " + parsedURL.Hostname() + " is not supported yet")
// 	}

// 	// Page URL validated, now check comics already in database
// 	// util.Info("Validated " + page.Name)
// 	comic, err := a.store.Comic.GetByURL(ctx, comicURL)

// 	// If comic is not in database, query it's latest chap,
// 	// add to database, then prepare response with latest chapter
// 	if err != nil {
// 		if strings.Contains(err.Error(), "not found") {

// 			util.Info("Comic is not in DB yet, insert it")
// 			comic = &model.Comic{
// 				Page: parsedURL.Hostname(),
// 				URL:  comicURL,
// 			}
// 			// Get all comic infos includes latest chapter
// 			err = getComicInfo(ctx, comic)
// 			if err != nil {
// 				util.Danger(err)
// 				return nil, errors.New("Please check your URL")
// 			}

// 			// Add new comic to DB
// 			err = a.store.Comic.Create(ctx, comic)
// 			if err != nil {
// 				util.Danger(err)
// 				return nil, errors.New("Please try again later")
// 			}
// 		} else {
// 			util.Danger(err)
// 			return nil, errors.New("Please try again later")
// 		}
// 	}

// 	// Validate users is in user DB or not
// 	// If not, add user to database, return "Subscribed to ..."
// 	// else return "Already subscribed"
// 	user, err := a.store.User.GetByFBID(ctx, field, id)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "not found") {

// 			util.Info("Add new user")

// 			user, err = getUserInfoByID(field, id, a.cfg)
// 			// Check user already exist
// 			if err != nil {
// 				util.Danger(err)
// 				return nil, errors.New("Please try again later")
// 			}
// 			err = a.store.User.Create(ctx, user)

// 			if err != nil {
// 				util.Danger(err)
// 				return nil, errors.New("Please try again later")
// 			}
// 		} else {
// 			util.Danger(err)
// 			return nil, errors.New("Please try again later")
// 		}
// 	}

// 	_, err = a.store.Comic.GetByPSID(ctx, user.PSID, comic.ID)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "not found") {
// 			subscriber := &model.Subscriber{
// 				PSID:    user.PSID,
// 				ComicID: comic.ID,
// 			}

// 			err = a.store.Subscriber.Create(ctx, subscriber)
// 			if err != nil {
// 				util.Danger(err)
// 				return nil, errors.New("Please try again later")
// 			}
// 			return comic, nil
// 		}
// 		util.Danger(err)
// 		return nil, errors.New("Please try again later")
// 	}
// 	return nil, errors.New("Already subscribed")
// }

// // UnsubscribeComic (DELETE /user/{user_id}/comic/{id})
// func (a *API) UnsubscribeComic(ctx context.Context, psid string, comicID int) error {

// 	return a.store.Subscriber.Delete(ctx, psid, comicID)
// }

// // GetUserComic (GET /user/{user_id}/comics/{id})
// func (a *API) GetUserComic(ctx context.Context, psid string, comicID int) (*model.Comic, error) {

// 	return a.store.Comic.GetByPSID(ctx, psid, comicID)
// }

// // GetUserComics (GET /user/{user_id}/comics)
// func (a *API) GetUserComics(ctx context.Context, psid string) ([]model.Comic, error) {

// 	return a.store.Comic.ListByPSID(ctx, psid)
// }
