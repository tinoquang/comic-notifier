package img

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/model"
)

func TestUploadDeleteImage(t *testing.T) {

	cfg := conf.New("../../../")
	SetEnvVar(cfg)

	assert := assert.New(t)
	imageURLs := []string{
		"https://img.blogtruyen.com/manga/0/139/tokyo one piece halloween 188699.jpg",
		"https://img.blogtruyen.com/manga/8/8981/giphy (2).gif",
		"https://cdn2.beeng.net/mangas/2020/07/26/05/a-tu-la-tay-du-ngoai-truyen.jpg",
	}

	for _, url := range imageURLs {
		img, err := UploadImagetoImgur("test image", url)
		if err != nil {
			t.Fatal("failed to upload image to imgur:", err)
		}

		assert.NotEmpty(img.ID)
		assert.NotEmpty(img.Link)
		assert.NotEmpty(img.Title)
		assert.NotEmpty(img.Description)

		err = DeleteImg(img.ID)
		if err != nil {
			t.Fatal("fail to delete image uploaded to imgur:", err)
		}
	}

}

func TestUpdateImageSuccess(t *testing.T) {

	cfg := conf.New("../../../")
	SetEnvVar(cfg)

	assert := assert.New(t)
	imageURLs := []string{
		"https://img.blogtruyen.com/manga/0/139/tokyo one piece halloween 188699.jpg",
		"https://cdn2.beeng.net/mangas/2020/07/26/05/a-tu-la-tay-du-ngoai-truyen.jpg",
	}

	img, err := UploadImagetoImgur("test image", imageURLs[0])
	if err != nil {
		t.Fatal("failed to upload image to imgur:", err)
	}

	assert.NotEmpty(img.ID)
	assert.NotEmpty(img.Link)
	assert.NotEmpty(img.Title)
	assert.NotEmpty(img.Description)

	c := &model.Comic{
		ImageURL: imageURLs[1],
	}

	err = UpdateImage(img.ID, c)
	if err != nil {
		t.Fatal("failed to upload image to imgur:", err)
	}

	assert.NotEmpty(c.ImgurID)
	assert.NotEmpty(c.ImgurLink)

	img, err = GetImageFromImgur(c.ImgurID.Value())
	if err != nil {
		t.Fatal("failed to get updated image:", err)
	}

	assert.NotEmpty(img.ID)
	assert.NotEmpty(img.Link)
	assert.NotEmpty(img.Title)
	assert.NotEmpty(img.Description)

	err = DeleteImg(c.ImgurID.Value())
	if err != nil {
		t.Fatal("fail to delete image uploaded to imgur", err)
	}
}

func TestUpdateSameImage(t *testing.T) {

	cfg := conf.New("../../../")
	SetEnvVar(cfg)

	assert := assert.New(t)
	imageURLs := "https://img.blogtruyen.com/manga/0/139/tokyo one piece halloween 188699.jpg"

	img, err := UploadImagetoImgur("test image", imageURLs)
	if err != nil {
		t.Fatal("failed to upload image to imgur:", err)
	}

	assert.NotEmpty(img.ID)
	assert.NotEmpty(img.Link)
	assert.NotEmpty(img.Title)
	assert.NotEmpty(img.Description)

	c := &model.Comic{
		ImageURL: imageURLs,
	}

	err = UpdateImage(img.ID, c)
	if err == nil || err != ErrUpToDate {
		t.Error("fail update same image success")
	}

	err = DeleteImg(img.ID)
	if err != nil {
		t.Fatal("fail to delete image uploaded to imgur", err)
	}
}
