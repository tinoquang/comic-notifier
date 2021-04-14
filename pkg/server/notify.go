package server

import (
	"sync"
	"time"

	"github.com/tinoquang/comic-notifier/pkg/conf"
	db "github.com/tinoquang/comic-notifier/pkg/db/sqlc"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

type notification struct {
	userID string
	comic  db.Comic
	retry  int // number attempts to send notification to user
}

var (
	newNotification    chan notification
	failedNotification chan notification
)

func initNotifyService(updateLock *sync.Mutex) {

	newNotification = make(chan notification, conf.Cfg.WrkDat.NotifyWorkerNum)
	failedNotification = make(chan notification, conf.Cfg.WrkDat.NotifyWorkerNum)

	go notifyService(updateLock)
}

func notifyService(updateLock *sync.Mutex) {

	var wg sync.WaitGroup
	for {

		// Need to verify updateService is not running, to avoid missing notification
		updateLock.Lock()
		notificationPool := make(chan notification, conf.Cfg.WrkDat.NotifyWorkerNum)

		// Resend all failNotification first
	failed:
		for {
			select {
			case n := <-failedNotification:
				notificationPool <- n
			default:
				break failed
			}
		}

		// Send all newNotification
	new:
		for {
			select {
			case n := <-newNotification:
				notificationPool <- n
			default:
				break new
			}
		}

		for i := 0; i < conf.Cfg.WrkDat.NotifyWorkerNum; i++ {
			go notifyWorker(i, &wg, notificationPool)
			wg.Add(1)
		}

		close(notificationPool)

		wg.Wait()
		updateLock.Unlock()

		// logging.Info("Notifyservice wait")
		select {
		case <-time.After(20 * time.Minute):
			// logging.Info("Notifywait timeout")
		case <-updateDone:
			// logging.Info("Received done signal from updateService")
		}
	}

}

func notifyWorker(id int, wg *sync.WaitGroup, notify <-chan notification) {

	for n := range notify {
		err := sendMsgTagsReply(n.userID, &n.comic)
		if err != nil {
			n.retry++

			// Retry sending notification 5 times before consider this is an error
			if n.retry < 5 {
				failedNotification <- n
			} else {
				logging.Danger("Can't send notify for comic", n.comic.Name, "to user", n.userID, "err", err)

			}
		}
		// logging.Info("Notify", id, " success for comic", n.comic.Name, "to user", n.userID, "retry time =", n.retry)
	}

	wg.Done()
}
