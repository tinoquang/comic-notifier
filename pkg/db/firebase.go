package db

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

// FirebaseDB contains bucket object to communicate with Firebase storage
type FirebaseDB struct {
	bucket *storage.BucketHandle
}

// NewFirebaseConnection create new bucket object to communicate with Firebase storage
func NewFirebaseConnection() *FirebaseDB {

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

	return &FirebaseDB{bucket: bucket}
}

// Get verify comic image is exist in cloud
func (f *FirebaseDB) Get(comicPage, comicName string) error {

	objectName := fmt.Sprintf("%s/%s", comicPage, comicName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	_, err := f.bucket.Object(objectName).Attrs(ctx)
	if err != nil {
		logging.Danger(err)
		return err
	}

	return nil
}

// Upload save file to Firebase storage and make it public
func (f *FirebaseDB) Upload(fileName, objectName string) error {

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
func (f *FirebaseDB) Delete(comicPage, comicName string) error {

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
