package model

import (
	"github.com/tinoquang/comic-notifier/pkg/server/img"
)

// Comic model
type Comic struct {
	ID           int    `json:"id"`
	Page         string `json:"page"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	OriginImgURL string `json:"-"`
	CloudImg     string `json:"-"`
	LatestChap   string `json:"latest"`
	ChapURL      string `json:"chap-url"`
}

// UpdateCloudImg fill CloudImg field
func (c *Comic) UpdateCloudImg() error {

	cloudImg, err := img.UploadToFirebase(c.Page, c.Name, c.OriginImgURL)
	if err != nil {
		return err
	}

	c.CloudImg = cloudImg

	return nil
}
