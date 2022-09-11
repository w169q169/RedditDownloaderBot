package reddit

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
)

func GetRedgifsVideo(body io.Reader) (videoUrl string, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("property"); name == "og:video" {

			videoUrl, _ = s.Attr("content")
		}
	})

	if videoUrl == "" {
		return "", fmt.Errorf("cant find meta")
	}

	return videoUrl, nil
}
