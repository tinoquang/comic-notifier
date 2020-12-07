package img

import (
	"fmt"
	"testing"

	"github.com/tinoquang/comic-notifier/pkg/conf"
)

func TestUploadDeleteImage(t *testing.T) {

	conf.Init()
	err := InitFirebaseBucket()
	if err != nil {
		t.Fatal("failed to create firebase bucket object:", err)
	}
	// assert := assert.New(t)
	imgURLs := []string{
		"https://img.blogtruyen.com/manga/0/139/tokyo one piece halloween 188699.jpg",
		"https://img.blogtruyen.com/manga/8/8981/giphy (2).gif",
		"https://cdn2.beeng.net/mangas/2020/07/26/05/a-tu-la-tay-du-ngoai-truyen.jpg",
	}

	for i, url := range imgURLs {
		_, err := UploadToFirebase("beeng", fmt.Sprintf("test_%d", i), url)
		if err != nil {
			t.Fatal("failed to upload image to imgur:", err)
		}

		if err != nil {
			t.Fatal("fail to delete image uploaded to imgur:", err)
		}
	}

}

func TestUpdateImageSuccess(t *testing.T) {

}

func TestUpdateSameImage(t *testing.T) {

}
