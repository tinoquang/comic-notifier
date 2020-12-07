package img

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/pkg/errors"
	"github.com/tinoquang/comic-notifier/pkg/conf"
	"github.com/tinoquang/comic-notifier/pkg/logging"
)

var (
	bucket      *storage.BucketHandle
	ErrUpToDate = errors.Errorf("Image is up-to-date")
)

// InitFirebaseBucket create firebase bucket
func InitFirebaseBucket() (err error) {

	config := &firebase.Config{
		StorageBucket: conf.Cfg.FirebaseBucket.Name,
	}

	app, err := firebase.NewApp(context.Background(), config, conf.Cfg.FirebaseBucket.Option)
	if err != nil {
		return
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		return
	}

	bucket, err = client.DefaultBucket()
	if err != nil {
		return
	}

	return nil
}

// UploadToFirebase add image to Imgur gallery and return link to new image
func UploadToFirebase(prefix, name, imgURL string) (cloudImg string, err error) {

	// download img first
	fileName, err := downloadOriginImg(name, imgURL)
	if err != nil {
		return "", err
	}

	// upload image to firebase
	err = uploadFileToFirebase(bucket, "./"+fileName, fmt.Sprintf("%s/%s", prefix, name)) // ex: prefix = beeng.net, fileName = tay-du.jpg
	if err != nil {
		return "", err
	}

	// delete img whether upload success or not, to save disk
	err = os.Remove("./" + fileName)
	return fmt.Sprintf("%s/%s/%s", conf.Cfg.FirebaseBucket.URL, prefix, fileName), err
}

// DeleteFirebaseImg delete img in imgur
func DeleteFirebaseImg(prefix, name string) error {

	// ext := filepath.Ext(cloudImg)
	objectName := fmt.Sprintf("%s/%s", prefix, name) //, ext)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	err := bucket.Object(objectName).Delete(ctx)
	if err != nil {
		logging.Danger(err)
		return err
	}

	return nil
}
