package reddit

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
)

func GetRedgifsVideo(body io.Reader) (videoUrl string, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}
	urlList := []string{}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("property"); name == "og:video" {
			videoUrl, _ = s.Attr("content")
			log.Println("get video url:", videoUrl)
			urlList = append(urlList, videoUrl)
		}
	})

	if len(urlList) == 0 {
		return "", fmt.Errorf("cant find meta")
	}

	videoUrl = urlList[0]

	return videoUrl, nil
}
