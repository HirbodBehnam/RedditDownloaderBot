package helpers

import (
	"encoding/json"
	"io"
	"strings"
)

// GetRedGifsID get id from watch url
func GetRedGifsID(url string) string {
	itemList := strings.Split(url, "/")

	if len(itemList) > 4 {
		return itemList[4]
	}

	return ""
}

type RedGisInfo struct {
	Gif  RedGisGif   `json:"gif"`
	User interface{} `json:"user"`
}

type RedGisGif struct {
	ID           string        `json:"id"`
	CreateDate   int           `json:"createDate"`
	HasAudio     bool          `json:"hasAudio"`
	Width        int           `json:"width"`
	Height       int           `json:"height"`
	Likes        int           `json:"likes"`
	Tags         []string      `json:"tags"`
	Verified     bool          `json:"verified"`
	Views        interface{}   `json:"views"`
	Duration     float64       `json:"duration"`
	Published    bool          `json:"published"`
	Urls         RedGisGifUrls `json:"urls"`
	UserName     string        `json:"userName"`
	Type         int           `json:"type"`
	AvgColor     string        `json:"avgColor"`
	Gallery      interface{}   `json:"gallery"`
	HideHome     bool          `json:"hideHome"`
	HideTrending bool          `json:"hideTrending"`
}

type RedGisGifUrls struct {
	Poster     string `json:"poster"`
	Thumbnail  string `json:"thumbnail"`
	Hd         string `json:"hd"`
	Sd         string `json:"sd"`
	Gif        string `json:"gif"`
	Vthumbnail string `json:"vthumbnail"`
}

// GetRedGifsInfo parse gifs info(json) from request body
func GetRedGifsInfo(body io.Reader) (info RedGisInfo, err error) {
	var ret RedGisInfo
	err = json.NewDecoder(body).Decode(&ret)
	return ret, err
}
