package reddit

import (
	"os"
	"testing"
)

func TestGetRedgifsVideo(t *testing.T) {
	t.Run("normal", func(t *testing.T) {

		f, err := os.Open("./test_data/redit_video.html")
		if err != nil {
			t.Errorf("open file failed:%v", err)
			return
		}

		defer f.Close()

		videoUrl, err := GetRedgifsVideo(f)
		if err != nil {
			t.Errorf("get video url failed:%v", err)
			return
		}

		expectURL := "https://thumbs4.redgifs.com/AdventurousSquareMarlin-mobile.mp4?expires=1662901200&signature=4fde43aece844be3c7612500d1eb3d50587457a1a8d4e054089ab916d70a9b3d&for=103.116.72.24"

		if videoUrl != expectURL {
			t.Errorf("want:%v, get :%v", expectURL, videoUrl)
			return
		}

	})
}
