package msg

import (
	"context"
	"sync"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	wg sync.WaitGroup
)

func worker(svr ServerInterface, wg *sync.WaitGroup, comicPool <-chan model.Comic) {

	for comic := range comicPool {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		util.Info("Comic ID", comic.ID, ": ", comic.Name, ": starting update...")

		updated, err := svr.UpdateComic(ctx, &comic)

		if err == nil {
			util.Info("Comic ID", comic.ID, ": ", comic.Name, " new chapter ", comic.LatestChap)
			notifyToUsers(ctx, svr, &comic)
		} else {
			if !updated {
				util.Info("Comic ID", comic.ID, ": ", comic.Name, "has no update")
			} else {
				util.Danger(err)
			}
		}
		cancel()
		wg.Done()
	}
}

// UpdateThread read comic database and update each comic to each latest chap
func updateThread(svr ServerInterface, workerNum, timeout int) {

	util.Info("Start update new chapter routine ...")

	// Create and jobs
	comicPool := make(chan model.Comic, workerNum)

	for i := 0; i < workerNum; i++ {
		go worker(svr, &wg, comicPool)
	}
	// Start infinite loop
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		// Get all comics in DB
		comics, err := svr.Comics(ctx)
		if err != nil {
			util.Info("Update new chapter fail, err: ", err)
			util.Info("Sleep update routine for 30min, then go check new chapter again..........")
			time.Sleep(time.Duration(timeout) * time.Minute)
			continue
		}

		// Query successful, for each comic put into job channel for worker to do the update stuffs
		for _, comic := range comics {
			comicPool <- comic
			wg.Add(1)
		}

		// Wait util all comics is updated, then sleep 30min and start checking again
		wg.Wait()
		util.Info("All comics is up-to-date")

		cancel()
		time.Sleep(time.Duration(timeout) * time.Minute)
	}

}

func notifyToUsers(cxt context.Context, svr ServerInterface, comic *model.Comic) {

	users, err := svr.GetUsersByComicID(cxt, comic.ID)
	if err != nil {
		util.Danger(err)
		return
	}

	for _, u := range users {
		util.Info("Notify ", comic.Name, " to user ", u.Name)
		sendMsgTagsReply(u.PSID, comic)
	}
}
