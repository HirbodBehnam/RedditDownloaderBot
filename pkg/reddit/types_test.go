package reddit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchResultMedia_HasAudio(t *testing.T) {
	assertion := assert.New(t)
	tests := []struct {
		input         FetchResultMedia
		expectedIndex int
		expectedHas   bool
	}{{
		input: FetchResultMedia{
			Medias: nil,
			Type:   FetchResultMediaTypePhoto,
		},
		expectedIndex: -1,
		expectedHas:   false,
	}, {
		input: FetchResultMedia{
			Medias: nil,
			Type:   FetchResultMediaTypeGif,
		},
		expectedIndex: -1,
		expectedHas:   false,
	}, {
		input: FetchResultMedia{
			Medias: nil,
			Type:   FetchResultMediaTypeVideo,
		},
		expectedIndex: -1,
		expectedHas:   false,
	}, {
		input: FetchResultMedia{
			Medias: FetchResultMediaEntries([]FetchResultMediaEntry{
				{
					Quality: "shahkar",
				},
			}),
			Type: FetchResultMediaTypeVideo,
		},
		expectedIndex: -1,
		expectedHas:   false,
	}, {
		input: FetchResultMedia{
			Medias: FetchResultMediaEntries([]FetchResultMediaEntry{
				{
					Quality: "shahkar",
				},
				{
					Quality: DownloadAudioQuality,
				},
			}),
			Type: FetchResultMediaTypeVideo,
		},
		expectedIndex: 1,
		expectedHas:   true,
	}, {
		input: FetchResultMedia{
			Medias: FetchResultMediaEntries([]FetchResultMediaEntry{
				{
					Quality: DownloadAudioQuality,
				},
				{
					Quality: "shahkar",
				},
			}),
			Type: FetchResultMediaTypeVideo,
		},
		expectedIndex: -1,
		expectedHas:   false, // FetchResultMedia.HasAudio only checks the last element in array
	}, {
		input: FetchResultMedia{
			Medias: FetchResultMediaEntries([]FetchResultMediaEntry{
				{
					Quality: DownloadAudioQuality,
				},
			}),
			Type: FetchResultMediaTypeVideo,
		},
		expectedIndex: 0,
		expectedHas:   true,
	}}

	for i, test := range tests {
		index, has := test.input.HasAudio()
		assertion.Equalf(test.expectedHas, has, "unexpected has on test number %d", i+1)
		assertion.Equalf(test.expectedIndex, index, "unexpected index on test number %d", i+1)
	}
}
