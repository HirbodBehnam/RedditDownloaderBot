package util

import (
	"RedditDownloaderBot/pkg/common"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"unsafe"
)

// IsUrl checks if a string is an url
// From https://stackoverflow.com/a/55551215/4213397
func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// FollowRedirect follows a page's redirect and returns the final URL
func FollowRedirect(u string) (string, error) {
	resp, err := common.GlobalHttpClient.Head(u)
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

// ToJsonString converts an object to json string
func ToJsonString(object any) string {
	data, _ := json.Marshal(object)
	return ByteToString(data)
}

// ParseEnvironmentVariableBool parses an environment variable which must represent a bool.
// It returns false if the variable data is malformed or non-existent
func ParseEnvironmentVariableBool(name string) bool {
	result, _ := strconv.ParseBool(os.Getenv(name))
	return result
}
