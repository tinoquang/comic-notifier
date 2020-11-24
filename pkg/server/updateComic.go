package server

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/crawler"
	"github.com/tinoquang/comic-notifier/pkg/server/img"
	"github.com/tinoquang/comic-notifier/pkg/store"
)

var (
	wg sync.WaitGroup
)

func worker(s *store.Stores, wg *sync.WaitGroup, comicPool <-chan model.Comic) {

	for comic := range comicPool {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		// logging.Info("Comic", comic.ID, "-", comic.Name, "starting update...")

		err := updateComic(ctx, s, &comic)

		if err == nil {
			logging.Info("Comic", comic.ID, "-", comic.Name, "new chapter", comic.LatestChap)
			notifyToUsers(ctx, s, &comic)
		} else {
			if strings.Contains(err.Error(), "No new chapter") {
				// logging.Info("Comic", comic.ID, "-", comic.Name, "is up-to-date")
			} else {
				logging.Danger(err)
			}
		}
		cancel()
		wg.Done()
	}
}

// UpdateThread read comic database and update each comic to each latest chap
func updateComicThread(s *store.Stores, workerNum, timeout int) {

	// Create and jobs
	comicPool := make(chan model.Comic, workerNum)

	for i := 0; i < workerNum; i++ {
		go worker(s, &wg, comicPool)
	}
	// Start infinite loop
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		// Get all comics in DB
		opt := store.NewComicsListOptions("", 0, 0)
		comics, err := s.Comic.List(ctx, opt)
		if err != nil {
			logging.Info("Get list of comic fails, sleep sometimes...")
			time.Sleep(time.Duration(timeout) * time.Minute)
			continue
		}

		logging.Info(fmt.Sprintf("Update %d comic(s) ...", len(comics)))

		// Query successful, for each comic put into job channel for worker to do the update stuffs
		for _, comic := range comics {
			comicPool <- comic
			wg.Add(1)
		}

		// Wait util all comics is updated, then sleep 30min and start checking again
		wg.Wait()
		logging.Info("All comics is up-to-date")

		cancel()
		time.Sleep(time.Duration(timeout) * time.Minute)
	}

}

// UpdateComic use when new chapter realease
func updateComic(ctx context.Context, s *store.Stores, comic *model.Comic) (err error) {

	err = crawler.GetComicInfo(ctx, comic)
	if err != nil {
		return
	}

	img.UpdateImage(string(comic.ImgurID), comic)
	err = s.Comic.Update(ctx, comic)
	return
}

func notifyToUsers(ctx context.Context, s *store.Stores, comic *model.Comic) {

	subscribers, err := s.Subscriber.ListByComicID(ctx, comic.ID)
	if err != nil {
		logging.Danger(err)
		return
	}

	for _, sub := range subscribers {
		logging.Info("Notify ", comic.Name, " to user ID", sub.PSID)
		sendMsgTagsReply(sub.PSID, comic)
	}
}
