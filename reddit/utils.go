package reddit

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
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

	for _, videoUrlItem := range urlList {
		videoSize, err := GetVideoFileSize(videoUrlItem)
		log.Printf("video url:%v size:%v err:%v \n", videoUrlItem, videoSize, err)
	}

	videoUrl = urlList[0]

	return videoUrl, nil
}

func GetVideoFileSize(videoUrl string) (ret int, err error) {
	resp, err := http.Head(videoUrl)
	if err != nil {
		return 0, errors.WithMessagef(err, "run head url failed")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("get wrong http status:%v", resp.StatusCode)
	}

	// the Header "Content-Length" will let us know
	// the total file size to download
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	if err != nil {
		return 0, errors.WithMessagef(err, "get content-length failed")
	}

	return size, nil
}
