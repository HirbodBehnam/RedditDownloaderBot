package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsImgurLink(t *testing.T) {
	tests := []struct {
		TestName string
		Link     string
		Expected bool
	}{
		{
			TestName: "ImgurLink1",
			Link:     "https://i.imgur.com/7ZYm2NC.mp4",
			Expected: true,
		},
		{
			TestName: "ImgurLink2",
			Link:     "https://imgur.com/wfwfw",
			Expected: true,
		},
		{
			TestName: "ImgurLink3",
			Link:     "https://shash.imgur.com/wfwfw.jpeg?shash=yes",
			Expected: true,
		},
		{
			TestName: "ImgurUppercase",
			Link:     "https://i.imGur.cOm/7ZYm2NC.mp4",
			Expected: true,
		},
		{
			TestName: "RedditLink",
			Link:     "https://external-preview.redd.it/eHhsa3JrdDl4YmlkMYG42k61zUHLZWYmXgKxVFtbkqT2ytev2qoJoAjMPjdm.png?format=pjpg&auto=webp&s=892d3a60ccd4d1a602637f0ffb974645fe1cea09",
			Expected: false,
		},
		{
			TestName: "BrokenLink",
			Link:     ":(",
			Expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			assert.Equal(t, test.Expected, IsImgurLink(test.Link))
		})
	}
}
