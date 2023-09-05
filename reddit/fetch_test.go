package reddit

import (
	"testing"

	"github.com/HirbodBehnam/RedditDownloaderBot/util"
	"github.com/stretchr/testify/assert"
)

func TestGetPostId(t *testing.T) {
	tests := []struct {
		TestName          string
		Url               string
		NeedsInternet     bool
		ExpectedID        string
		ExpectedIsComment bool
		ExpectedError     string // is empty if no error must be thrown
	}{
		{
			TestName:          "Normal Post",
			Url:               "https://www.reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},

		{TestName: "Normal Post",
			Url:               "https://www.reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Normal Comment",
			Url:               "https://www.reddit.com/r/gaming/comments/vdrdxu/comment/icm3y72/?utm_source=share&utm_medium=web2x&context=3",
			NeedsInternet:     false,
			ExpectedID:        "icm3y72",
			ExpectedIsComment: true,
			ExpectedError:     "",
		},
		{
			TestName:          "redd.it Link",
			Url:               "https://redd.it/kmi4d3",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "v.redd.it Link",
			Url:               "https://v.redd.it/rhs0ixoyc7j91",
			NeedsInternet:     true,
			ExpectedID:        "wul62b",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Post With Other Lines",
			Url:               "Prop Hunt Was Fun\nhttps://www.reddit.com/r/Unexpected/comments/wul62b/prop_hunt_was_fun/\nhttps://google.com",
			NeedsInternet:     false,
			ExpectedID:        "wul62b",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Invalid Url",
			Url:               "",
			NeedsInternet:     false,
			ExpectedID:        "",
			ExpectedIsComment: false,
			ExpectedError:     "Cannot parse reddit the url. Does your text contain a reddit url?",
		},
		{
			TestName:          "Short Url",
			Url:               "https://www.reddit.com/r/Unexpected/comments",
			NeedsInternet:     false,
			ExpectedID:        "",
			ExpectedIsComment: false,
			ExpectedError:     "Cannot parse reddit the url. Does your text contain a reddit url?",
		},
		{
			TestName:          "Short Reddit Url",
			Url:               "https://www.reddit.com/wul62b",
			NeedsInternet:     false,
			ExpectedID:        "wul62b",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Old Post",
			Url:               "https://old.reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Normal Post 2",
			Url:               "https://reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Normal Post No Transport",
			Url:               "reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "New Shared Links",
			Url:               "https://reddit.com/r/UkraineWarVideoReport/s/AKk56RlMN6",
			NeedsInternet:     true,
			ExpectedID:        "15ma9tp",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			// Check internet if needed
			if test.NeedsInternet {
				if _, err := util.FollowRedirect(test.Url); err != nil {
					t.Skip("cannot connect to internet:", err)
					return
				}
			}
			// Get the id
			id, isComment, err := getPostID(test.Url)
			if err != nil {
				assert.Equal(t, test.ExpectedError, err.BotError)
			}
			assert.Equal(t, test.ExpectedIsComment, isComment)
			assert.Equal(t, test.ExpectedID, id)
		})
	}
}
