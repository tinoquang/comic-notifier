package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	wg sync.WaitGroup
)

// UpdateThread read comic database and update each comic to each latest chap
func updateComicThread(s *store.Stores, workerNum, timeout int) {

	// Start update routine, then sleep for a while and re-update
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Get all comics in DB
		opt := store.NewComicsListOptions("", 0, 0)
		comics, err := s.Comic.List(ctx, opt)
		cancel() // Call context cancel here to avoid context leak

		if err != nil {
			logging.Info("Get list of comic fails, sleep sometimes...")
			time.Sleep(time.Duration(timeout) * time.Minute)
			continue
		}

		if len(comics) != 0 {

			logging.Info(fmt.Sprintf("Update %d comic(s) ...", len(comics)))

			// Create workers
			comicPool := make(chan model.Comic, workerNum)
			for i := 0; i < workerNum; i++ {
				go worker(i, s, &wg, comicPool)
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

func worker(id int, s *store.Stores, wg *sync.WaitGroup, comicPool <-chan model.Comic) {

	// Get comic from updateComicThread, which run only when updateComicThread push comic into comicPool
	for comic := range comicPool {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		err := updateComic(ctx, s, &comic)
		if err == nil {
			logging.Info("Comic", comic.ID, "-", comic.Name, "new chapter", comic.LatestChap)
			notifyToUsers(ctx, s, &comic)
		} else if err != util.ErrComicUpToDate {
			logging.Danger(err)
		}

		cancel() // Call context cancel here to avoid context leak
	}

	wg.Done()
}

// UpdateComic use when new chapter realease
func updateComic(ctx context.Context, s *store.Stores, comic *model.Comic) (err error) {

	oldImgURL := comic.OriginImgURL
	// err = crawler.GetComicInfo(ctx, comic)
	if err != nil {
		return
	}

	err = s.Comic.Update(ctx, comic, oldImgURL)
	return
}

func notifyToUsers(ctx context.Context, s *store.Stores, comic *model.Comic) {

	subscribers, err := s.Subscriber.ListByComicID(ctx, comic.ID)
	if err != nil {
		logging.Danger("Can't send notification for comic %s, err: %s", comic.Name, err.Error())
		return
	}

	for _, sub := range subscribers {
		sendMsgTagsReply(sub.PSID, comic)
	}
}
