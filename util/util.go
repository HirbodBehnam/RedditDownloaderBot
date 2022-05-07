package util

import (
	"encoding/base64"
	"github.com/HirbodBehnam/RedditDownloaderBot/config"
	"github.com/google/uuid"
	"log"
	"net/url"
	"os"
	"os/exec"
	"reflect"
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
