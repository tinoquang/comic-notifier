package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/crawler"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// Server implement main business logic
type Server struct {
	API *API
	Msg *MSG
}

var (
	wg                sync.WaitGroup
	messengerEndpoint string
	pageToken         string
	webhookToken      string
)

// New  create new server
func New(store db.Store, crawler crawler.Crawler) *Server {

	// Get env config
	messengerEndpoint = conf.Cfg.Webhook.GraphEndpoint + "/me/messages"
	webhookToken = conf.Cfg.Webhook.WebhookToken
	pageToken = conf.Cfg.FBSecret.PakeToken

	s := &Server{
		API: NewAPI(store),
		Msg: NewMSG(store, crawler),
	}

	// Start update-comic thread
	go updateComicThread(crawler, store, conf.Cfg.WrkDat.WorkerNum, conf.Cfg.WrkDat.Timeout)
	return s
}

// UpdateThread read comic database and update each comic to each latest chap
func updateComicThread(crwl crawler.Crawler, s db.Store, workerNum, timeout int) {

	// Start update routine, then sleep for a while and re-update
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Get all comics in DB
		comics, err := s.ListComics(ctx)
		cancel() // Call context cancel here to avoid context leak

		if err != nil {
			logging.Info("Get list of comic fails, err", err)
			time.Sleep(time.Duration(timeout) * time.Minute)
			continue
		}

		if len(comics) != 0 {

			logging.Info(fmt.Sprintf("Update %d comic(s) ...", len(comics)))

			// Create workers
			comicPool := make(chan db.Comic, workerNum)
			for i := 0; i < workerNum; i++ {
				go worker(i, s, crwl, &wg, comicPool)
				wg.Add(1)
			}

			// Query successful, for each comic put into job channel for worker to do the update stuffs
			for _, comic := range comics {
				comicPool <- comic
			}
			close(comicPool)

			wg.Wait()
			logging.Info("All comics is updated")
		}

		time.Sleep(time.Duration(timeout) * time.Minute)
	}

	// Never reach here
}

func worker(id int, s db.Store, crwl crawler.Crawler, wg *sync.WaitGroup, comicPool <-chan db.Comic) {

	// Get comic from updateComicThread, which run only when updateComicThread push comic into comicPool
	for comic := range comicPool {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Synchronized firebase img
		err := s.SynchronizedComicImage(&comic)
		if err != nil {
			logging.Danger(err)
		}

		c, err := crwl.GetComicInfo(ctx, comic.Page, comic.Url, comic.ChapUrl)
		if err != nil {

			if err != util.ErrComicUpToDate {
				logging.Danger(err)
			}
			cancel()
			continue
		}
		c.ID = comic.ID
		err = s.UpdateComicChapter(ctx, &c, comic.ImgUrl)
		if err != nil {
			logging.Danger(err)
			cancel()
			continue
		}

		logging.Info("Comic", comic.ID, "-", comic.Name, "new chapter", comic.LatestChap)
		notifyToUsers(ctx, s, &comic)

		cancel() // Call context cancel here to avoid context leak
	}

	wg.Done()
}

func notifyToUsers(ctx context.Context, s db.Store, comic *db.Comic) {

	users, err := s.ListUsersPerComic(ctx, comic.ID)
	if err != nil {
		logging.Danger("Can't send notification for comic %s, err: %s", comic.Name, err.Error())
		return
	}

	for _, user := range users {
		sendMsgTagsReply(user.Psid.String, comic)
	}
}
