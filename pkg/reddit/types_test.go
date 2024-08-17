package reddit

import (
	"github.com/stretchr/testify/assert"
	"math"
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

func TestFetchedThumbnails_SelectThumbnail(t *testing.T) {
	tests := []struct {
		Name         string
		Input        FetchedThumbnails
		Dim          Dimension
		ExpectedLink string
	}{
		{
			Name:  "NullList",
			Input: nil,
			Dim: Dimension{
				Width:  math.MaxInt64,
				Height: math.MaxInt64,
			},
			ExpectedLink: "",
		},
		{
			Name:  "ZeroList",
			Input: FetchedThumbnails{},
			Dim: Dimension{
				Width:  math.MaxInt64,
				Height: math.MaxInt64,
			},
			ExpectedLink: "",
		},
		{
			Name: "NoLimits",
			Input: FetchedThumbnails{
				{
					Link: "1",
					Dim: Dimension{
						Width:  10,
						Height: 10,
					},
				},
				{
					Link: "2",
					Dim: Dimension{
						Width:  100,
						Height: 100,
					},
				},
				{
					Link: "3",
					Dim: Dimension{
						Width:  1000,
						Height: 1000,
					},
				},
				{
					Link: "4",
					Dim: Dimension{
						Width:  10000,
						Height: 10000,
					},
				},
			},
			Dim: Dimension{
				Width:  math.MaxInt64,
				Height: math.MaxInt64,
			},
			ExpectedLink: "4",
		},
		{
			Name: "EqualWidthHeight",
			Input: FetchedThumbnails{
				{
					Link: "1",
					Dim: Dimension{
						Width:  10,
						Height: 10,
					},
				},
				{
					Link: "2",
					Dim: Dimension{
						Width:  100,
						Height: 100,
					},
				},
				{
					Link: "3",
					Dim: Dimension{
						Width:  1000,
						Height: 1000,
					},
				},
				{
					Link: "4",
					Dim: Dimension{
						Width:  10000,
						Height: 10000,
					},
				},
			},
			Dim: Dimension{
				Width:  100,
				Height: 100,
			},
			ExpectedLink: "2",
		},
		{
			Name: "LessWidth",
			Input: FetchedThumbnails{
				{
					Link: "1",
					Dim: Dimension{
						Width:  10,
						Height: 10,
					},
				},
				{
					Link: "2",
					Dim: Dimension{
						Width:  100,
						Height: 100,
					},
				},
				{
					Link: "3",
					Dim: Dimension{
						Width:  1000,
						Height: 1000,
					},
				},
				{
					Link: "4",
					Dim: Dimension{
						Width:  10000,
						Height: 10000,
					},
				},
			},
			Dim: Dimension{
				Width:  500,
				Height: 5000,
			},
			ExpectedLink: "2",
		},
		{
			Name: "LessHeight",
			Input: FetchedThumbnails{
				{
					Link: "1",
					Dim: Dimension{
						Width:  10,
						Height: 10,
					},
				},
				{
					Link: "2",
					Dim: Dimension{
						Width:  100,
						Height: 100,
					},
				},
				{
					Link: "3",
					Dim: Dimension{
						Width:  1000,
						Height: 1000,
					},
				},
				{
					Link: "4",
					Dim: Dimension{
						Width:  10000,
						Height: 10000,
					},
				},
			},
			Dim: Dimension{
				Width:  5000,
				Height: 500,
			},
			ExpectedLink: "2",
		},
		{
			Name: "LessWidthHeight",
			Input: FetchedThumbnails{
				{
					Link: "1",
					Dim: Dimension{
						Width:  10,
						Height: 10,
					},
				},
				{
					Link: "2",
					Dim: Dimension{
						Width:  100,
						Height: 100,
					},
				},
				{
					Link: "3",
					Dim: Dimension{
						Width:  1000,
						Height: 1000,
					},
				},
				{
					Link: "4",
					Dim: Dimension{
						Width:  10000,
						Height: 10000,
					},
				},
			},
			Dim: Dimension{
				Width:  50,
				Height: 50,
			},
			ExpectedLink: "1",
		},
		{
			Name: "VerySmallWidthHeight",
			Input: FetchedThumbnails{
				{
					Link: "1",
					Dim: Dimension{
						Width:  10,
						Height: 10,
					},
				},
				{
					Link: "2",
					Dim: Dimension{
						Width:  100,
						Height: 100,
					},
				},
				{
					Link: "3",
					Dim: Dimension{
						Width:  1000,
						Height: 1000,
					},
				},
				{
					Link: "4",
					Dim: Dimension{
						Width:  10000,
						Height: 10000,
					},
				},
			},
			Dim: Dimension{
				Width:  1,
				Height: 1,
			},
			ExpectedLink: "1",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			link := test.Input.SelectThumbnail(test.Dim)
			assert.Equal(t, test.ExpectedLink, link)
		})
	}
}
