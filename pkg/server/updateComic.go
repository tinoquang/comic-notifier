package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

var updateDone chan struct{}

func initUpdateService(updateLock *sync.Mutex, crwl infoCrawler, s db.Store, workerNum, timeout int) {

	updateDone = make(chan struct{})
	go updateComicService(updateLock, crwl, s, workerNum, timeout)
}

// UpdateThread read comic database and update each comic to each latest chap
func updateComicService(updateLock *sync.Mutex, crwl infoCrawler, s db.Store, workerNum, timeout int) {

	var wg sync.WaitGroup

	// Start update routine, then sleep for a while and re-update
	for {

		// Need to verify NotifyService is not running
		updateLock.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Get all comics in DB
		comics, err := s.ListComics(ctx)
		cancel() // Call context cancel here to avoid context leak

		if err != nil {
			logging.Danger("Get list of comic fails, err", err)
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

		updateLock.Unlock()
		updateDone <- struct{}{}
		time.Sleep(time.Duration(timeout) * time.Minute)
		// time.Sleep(15 * time.Second)
	}

	// Never reach here
}

func worker(id int, s db.Store, crwl infoCrawler, wg *sync.WaitGroup, comicPool <-chan db.Comic) {

	// Get comic from updateComicThread, which run only when updateComicThread push comic into comicPool
	for oldComic := range comicPool {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		// Synchronized firebase img
		err := s.SyncComicImage(&oldComic)
		if err != nil {
			logging.Danger(err)
		}

		c, err := crwl.GetComicInfo(ctx, oldComic.Url, true)
		if err != nil {
			logging.Danger(err)
			cancel()
			continue
		}

		if c.Page != "hocvientruyentranh.net" {
			if c.LastUpdate.Sub(oldComic.LastUpdate) < 0 { // Avoid update old chapter
				cancel()
				continue
			}
		}

		if c.ChapUrl == oldComic.ChapUrl {
			cancel()
			continue
		}

		c.ID = oldComic.ID
		err = s.UpdateNewChapter(ctx, &c, oldComic.ImgUrl)
		if err != nil {
			logging.Danger(err)
			cancel()
			continue
		}

		logging.Info("Comic", c.ID, "-", c.Name, "new chapter", c.LatestChap)
		addNewNotification(ctx, s, c)

		cancel() // Call context cancel here to avoid context leak
	}

	wg.Done()
}

func addNewNotification(ctx context.Context, s db.Store, comic db.Comic) {

	users, err := s.ListUsersPerComic(ctx, comic.ID)
	if err != nil {
		logging.Danger("Can't send notification for comic %s, err: %s", comic.Name, err.Error())
		return
	}

	for _, user := range users {
		newNotification <- notification{
			userID: user.Psid.String,
			comic:  comic,
			retry:  0,
		}
	}
}
