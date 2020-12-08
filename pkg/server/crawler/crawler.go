package crawler

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/model"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

var (
	crawlerMap map[string]Crawler
)

// Crawler interface
type Crawler interface {
	crawl(ctx context.Context, comic *model.Comic, detector detectSpoiler) (err error)
}
type detectSpoiler interface {
	detect(chapURL string, attr1, attr2 string) error
}

// New init crawlerMap contain page which is supported
func New() {
	crawlerMap = make(map[string]Crawler)
	crawlerMap["beeng.net"] = beeng{}
	crawlerMap["blogtruyen.vn"] = blogtruyen{}
	// crawlerMap["truyendep.com"] = mangaK{}
	crawlerMap["truyentranhtuan.com"] = truyentranhtuan{}
	crawlerMap["truyentranh.net"] = truyentranhnet{}
}

// GetComicInfo return link of latest chapter of a page
func GetComicInfo(ctx context.Context, comic *model.Comic) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			logging.Danger()
		}
		return
	}()

	if _, ok := crawlerMap[comic.Page]; !ok {
		return util.ErrPageNotSupported
	}

	return crawlComic(ctx, comic, crawlerMap[comic.Page])
}

func crawlComic(ctx context.Context, comic *model.Comic, crawler Crawler) error {
	return crawler.crawl(ctx, comic, spoiler{})
}
