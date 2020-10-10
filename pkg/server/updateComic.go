package server

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/server/crawler"
	"github.com/tinoquang/comic-notifier/pkg/store"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	wg sync.WaitGroup
)

func worker(store *store.Stores, wg *sync.WaitGroup, comicPool <-chan *model.Comic) {

	for comic := range comicPool {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		util.Info("Comic", comic.ID, "-", comic.Name, "starting update...")

		updated, err := updateComic(ctx, store, comic)

		if err == nil {
			util.Info("Comic", comic.ID, "-", comic.Name, "new chapter", comic.LatestChap)
			notifyToUsers(ctx, store, comic)
		} else {
			if !updated {
				util.Info("Comic", comic.ID, "-", comic.Name, "is up-to-date")
			} else {
				util.Danger(err)
			}
		}
		cancel()
		wg.Done()
	}
}

// UpdateThread read comic database and update each comic to each latest chap
func updateComicThread(store *store.Stores, workerNum, timeout int) {

	util.Info("Start update new chapter routine ...")

	// Create and jobs
	comicPool := make(chan *model.Comic, workerNum)

	for i := 0; i < workerNum; i++ {
		go worker(store, &wg, comicPool)
	}
	// Start infinite loop
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		// Get all comics in DB
		comics, err := store.Comic.List(ctx)
		if err != nil {
			util.Info("Update new chapter fail, err: ", err)
			util.Info("Sleep update routine for 30min, then go check new chapter again..........")
			time.Sleep(time.Duration(timeout) * time.Minute)
			continue
		}

		// Query successful, for each comic put into job channel for worker to do the update stuffs
		for _, comic := range comics {
			comicPool <- &comic
			wg.Add(1)
		}

		// Wait util all comics is updated, then sleep 30min and start checking again
		wg.Wait()
		util.Info("All comics is up-to-date")

		cancel()
		time.Sleep(time.Duration(timeout) * time.Minute)
	}

}

// UpdateComic use when new chapter realease
func updateComic(ctx context.Context, store *store.Stores, comic *model.Comic) (bool, error) {

	updated := true
	err := crawler.GetComicInfo(ctx, comic)

	if err != nil {
		if strings.Contains(err.Error(), "No new chapter") {
			updated = false
		} else {
			util.Danger()
		}
		return updated, err
	}

	err = store.Comic.Update(ctx, comic)
	return updated, err
}

func notifyToUsers(ctx context.Context, store *store.Stores, comic *model.Comic) {

	subscribers, err := store.Subscriber.ListByComicID(ctx, comic.ID)
	if err != nil {
		util.Danger(err)
		return
	}

	for _, s := range subscribers {
		util.Info("Notify ", comic.Name, " to user ID", s.PSID)
		sendMsgTagsReply(s.PSID, comic)
	}
}
