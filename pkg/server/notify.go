package server

import (
	"time"

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

func initNotifyService() {

	newNotification = make(chan notification, 100)
	failedNotification = make(chan notification, 100)

	go notifyService()
}

func notifyService() {

	for {
		notificationPool := make(chan notification, 100)

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

		for i := 0; i < 100; i++ {
			go notifyWorker(i, notificationPool)
		}

		close(notificationPool)
		logging.Info("Notify service sleep")
		time.Sleep(time.Duration(5) * time.Minute)
	}

}

func notifyWorker(id int, notify <-chan notification) {

	for n := range notify {
		err := sendMsgTagsReply(n.userID, &n.comic)
		if err != nil {
			logging.Danger("Can't send notify for comic", n.comic.Name, "to user", n.userID, "retry time =", n.retry)
			n.retry++

			// After 4 times retry, consider this
			if n.retry < 5 {
				failedNotification <- n
			}
		}
		logging.Danger("Notify success for comic", n.comic.Name, "to user", n.userID, "retry time =", n.retry)
	}
}
