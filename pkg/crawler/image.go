package crawler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
	"github.com/tinoquang/comic-notifier/pkg/util"
)

// New return new DB connection

// firebaseConnection contains bucket object to communicate with Firebase storage
type firebaseConnection struct {
	bucket *storage.BucketHandle
}

// NewFirebaseConnection create new bucket object to communicate with Firebase storage
func newFirebaseConnection() *firebaseConnection {

	var bucket *storage.BucketHandle

	config := &firebase.Config{
		StorageBucket: conf.Cfg.FirebaseBucket.Name,
	}

	app, err := firebase.NewApp(context.Background(), config, conf.Cfg.FirebaseBucket.Option)
	if err != nil {
		panic(err)
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		panic(err)
	}

	bucket, err = client.DefaultBucket()
	if err != nil {
		panic(err)
	}

	return &firebaseConnection{bucket: bucket}
}

// GetImg verify comic image is exist in cloud
func (f *firebaseConnection) GetImg(comicPage, comicName string) error {

	objectName := fmt.Sprintf("%s/%s", comicPage, comicName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	_, err := f.bucket.Object(objectName).Attrs(ctx)
	if err != nil {
		return err
	}

	return nil
}

// UploadImg include download image to local then upload it to firebase cloud
func (f *firebaseConnection) UploadImg(comicPage, comicName, imgURL string) (err error) {

	// Image will be uploaded to folder: page/name.ext in Firebas storage, so we need to pass comicPage and comicName

	// download img first, path will be ./name.ext
	fileName := comicName + filepath.Ext(imgURL)
	err = util.DownloadFile(imgURL, "./"+fileName)
	if err != nil {
		logging.Danger(err)
		return
	}

	// upload image to firebase
	err = f.upload("./"+fileName, fmt.Sprintf("%s/%s", comicPage, comicName)) // ex: prefix = beeng.net, fileName = tay-du.jpg
	if err != nil {
		return err
	}

	// delete img whether upload success or not, since we don't need it anyway
	err = os.Remove("./" + fileName)
	return err
}

// Upload save file to Firebase storage and make it public
func (f *firebaseConnection) upload(fileName, objectName string) error {

	file, err := os.Open(fileName)
	if err != nil {
		logging.Danger(err)
		return err
	}

	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	wc := f.bucket.Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(wc, file); err != nil {
		logging.Danger(err)
		return err
	}
	if err := wc.Close(); err != nil {
		logging.Danger(err)
		return err
	}

	// Set role reader for all users to object to let front-end get file without authorization
	acl := f.bucket.Object(objectName).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		logging.Danger(err)
		return err
	}
	return nil
}

// Delete remove img in cloud
func (f *firebaseConnection) DeleteImg(comicPage, comicName string) error {

	objectName := fmt.Sprintf("%s/%s", comicPage, comicName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	err := f.bucket.Object(objectName).Delete(ctx)
	if err != nil {
		logging.Danger(err)
		return err
	}

	return nil
}
