package img

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinoquang/comic-notifier/pkg/conf"
)

func TestUploadDeleteImage(t *testing.T) {

	cfg := conf.New("../../../")
	SetEnvVar(cfg)

	assert := assert.New(t)
	imageURL := "https://img.blogtruyen.com/manga/0/139/tokyo one piece halloween 188699.jpg"

	img, err := UploadImagetoImgur("test image", imageURL)
	if err != nil {
		t.Fatal("failed to upload image to imgur")
	}

	assert.NotEmpty(img.ID)
	assert.NotEmpty(img.Link)
	assert.NotEmpty(img.Title)

	err = DeleteImg(img.ID)
	if err != nil {
		t.Fatal("fail to delete image uploaded to imgur")
	}
}
