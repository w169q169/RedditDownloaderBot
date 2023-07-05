package reddit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDownloadVideo(t *testing.T) {
	t.Run("imgur", func(t *testing.T) {
		vidUrl := "https://i.imgur.com/1G9MyZA.mp4"

		audioUrl, videoFile, err := DownloadVideo(vidUrl)
		assert.NoError(t, err)
		assert.NotNil(t, videoFile)
		_ = audioUrl
	})
}
