package util

import (
	"encoding/base64"
	"encoding/json"
	"github.com/HirbodBehnam/RedditDownloaderBot/config"
	"github.com/google/uuid"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"unsafe"
)

// IsUrl checks if a string is an url
// From https://stackoverflow.com/a/55551215/4213397
func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

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
func GetRedGifsInfo(body io.ReadCloser) (info *RedGisInfo, err error) {
	ret := &RedGisInfo{}

	err = json.NewDecoder(body).Decode(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// FollowRedirect follows a page's redirect and returns the final URL
func FollowRedirect(u string) (string, error) {
	resp, err := config.GlobalHttpClient.Head(u)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return resp.Request.URL.String(), nil
}

// DoesFfmpegExists returns true if ffmpeg is found
func DoesFfmpegExists() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// CheckFileSize checks the size of file before sending it to telegram
func CheckFileSize(f string, allowed int64) bool {
	fi, err := os.Stat(f)
	if err != nil {
		log.Println("Cannot get file size:", err.Error())
		return false
	}
	return fi.Size() <= allowed
}

// UUIDToBase64 uses the not standard base64 encoding to encode an uuid.UUID as string
// So instead of 36 chars we have 24
func UUIDToBase64(id uuid.UUID) string {
	return base64.StdEncoding.EncodeToString(id[:])
}

// ByteToString converts a byte slice to string
func ByteToString(b []byte) string {
	// From strings.Builder.String()
	return *(*string)(unsafe.Pointer(&b))
}

// StringToByte converts a string to byte array without copy
// Please be really fucking careful with this function
// DO NOT APPEND ANYTHING TO UNDERLYING SLICE AND DO NOT CHANGE IT
// Use this function to call function like unmarshal text manually
func StringToByte(s string) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)), len(s))
}

// ToJsonString converts an object to json string
func ToJsonString(object any) string {
	data, _ := json.Marshal(object)
	return ByteToString(data)
}
