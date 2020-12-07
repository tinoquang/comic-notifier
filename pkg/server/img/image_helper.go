package img

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

func downloadOriginImg(name string, imgURL string) (fileName string, err error) {

	fileName = name + filepath.Ext(imgURL)

	err = util.DownloadFile(imgURL, "./"+fileName)
	if err != nil {
		logging.Danger(err)
		return
	}

	return fileName, nil
}

func uploadFileToFirebase(bucket *storage.BucketHandle, file, objectName string) error {

	f, err := os.Open(file)
	if err != nil {
		logging.Danger(err)
		return err
	}

	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	wc := bucket.Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		logging.Danger(err)
		return err
	}
	if err := wc.Close(); err != nil {
		logging.Danger(err)
		return err
	}

	// Set role reader for all users to object to let front-end get file without authorization
	acl := bucket.Object(objectName).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		logging.Danger(err)
		return err
	}
	return nil
}
