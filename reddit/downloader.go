package reddit

import (
	"bytes"
	"github.com/HirbodBehnam/RedditDownloaderBot/config"
	"github.com/HirbodBehnam/RedditDownloaderBot/util"
	"github.com/go-faster/errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// We don't download anything more than this size
const maxDownloadSize = 50 * 1000 * 1000

// FileTooBigError indicates that this file is too big to be uploaded to Telegram
// So we don't download it at first place
var FileTooBigError = errors.New("file too big")

// DownloadPhoto downloads a photo from reddit and returns the saved file in it
func DownloadPhoto(link string) (*os.File, error) {
	// Get the file name
	var fileName string
	{
		u, err := url.Parse(link)
		if err != nil {
			return nil, err
		}
		fileName = u.Path[1:]
	}
	// Generate a temp file
	tmpFile, err := os.CreateTemp("", "*."+fileName)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create temp file")
	}
	// Download the file
	err = downloadToFile(link, tmpFile)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, errors.Wrap(err, "cannot download file")
	}
	// We are good
	return tmpFile, nil
}

// DownloadVideo downloads a video from reddit
// If necessary, it will merge the audio and video with ffmpeg
func DownloadVideo(vidUrl string) (audioUrl string, videoFile *os.File, err error) {
	// Download the video in a temp file
	videoFile, err = os.CreateTemp("", "*.mp4")
	if err != nil {
		err = errors.Wrap(err, "cannot create a temp file for video")
		return
	}
	defer func() {
		// Only delete the video file if error is not nil
		if err != nil {
			videoFile.Close()
			os.Remove(videoFile.Name())
		}
	}()
	err = downloadToFile(vidUrl, videoFile)
	if err != nil {
		err = errors.Wrap(err, "cannot download file")
		return
	}
	// Otherwise, search for an audio file
	audioUrl, hasAudio := HasAudio(vidUrl)
	audFile, err := os.CreateTemp("", "*.mp4")
	if err != nil {
		err = errors.Wrap(err, "cannot create a temp file for audio")
		return
	}
	// We don't need audio file anyway
	defer func() {
		audFile.Close()
		os.Remove(audFile.Name())
	}()
	if downloadToFile(audioUrl, audFile) != nil {
		audioUrl = ""
		hasAudio = false
	}
	// Check ffmpeg; If it doesn't exist, just return the video file
	if !util.DoesFfmpegExists() {
		return audioUrl, videoFile, nil
	}
	// If this file has audio, convert it
	if hasAudio {
		var finalFile *os.File
		// Convert
		finalFile, err = os.CreateTemp("", "*.mp4")
		if err != nil {
			err = errors.Wrap(err, "cannot create a temp file for final video")
			return
		}
		cmd := exec.Command("ffmpeg",
			"-i", videoFile.Name(),
			"-i", audFile.Name(),
			"-c", "copy",
			finalFile.Name(), "-y")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			log.Println("Cannot convert video:", err, "\n", stderr.String())
			finalFile.Close()
			os.Remove(finalFile.Name())
			// We don't return error here
			err = nil
			return audioUrl, videoFile, nil
		}
		// If we have reached here, it means that the conversion was fine
		// So we swap the final file with video file and delete the video file
		videoFile.Close()
		os.Remove(videoFile.Name())
		videoFile = finalFile
	}
	// No we can return the video file
	err = nil // Just be safe
	return audioUrl, videoFile, nil
}

// DownloadGif downloads a gif from reddit
func DownloadGif(link string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "*.mp4")
	if err != nil {
		return nil, err
	}
	err = downloadToFile(link, tmpFile)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, err
	}
	return tmpFile, nil
}

// DownloadThumbnail is basically DownloadPhoto but without the filename
func DownloadThumbnail(link string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		log.Println("Cannot create temp file for thumbnail:", err)
		return nil, err
	}
	// Download to file
	err = downloadToFile(link, tmpFile)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, err
	}
	// We are good
	return tmpFile, nil
}

// DownloadAudio simply downloads an audio file from reddit via direct link
func DownloadAudio(audioUrl string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "*.m4a")
	if err != nil {
		log.Println("Cannot create temp file for audio:", err)
		return nil, err
	}
	// Download to file
	err = downloadToFile(audioUrl, tmpFile)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, err
	}
	// We are good
	return tmpFile, nil
}

// downloadToFile downloads a link to a file
// It also checks where the file is too big to be uploaded to Telegram or not
// If the file is too big, it returns FileTooBigError
func downloadToFile(link string, f *os.File) error {
	resp, err := config.GlobalHttpClient.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("forbidden")
	}
	if resp.ContentLength == -1 {
		return errors.New("unknown length")
	}
	if resp.ContentLength > maxDownloadSize {
		return FileTooBigError
	}
	_, err = io.Copy(f, resp.Body)
	return err
}

// HasAudio checks if a video contains audio
func HasAudio(videoURL string) (audioURL string, hasAudio bool) {
	// Get the audio URL
	audioURL = videoURL[:strings.LastIndex(videoURL, "/")] // base url
	if strings.Contains(videoURL, ".mp4") {                // new reddit api or sth idk
		audioURL += "/DASH_audio.mp4"
	} else { // old format
		audioURL += "/audio"
	}
	// Check if it exists
	resp, err := config.GlobalHttpClient.Head(audioURL)
	return audioURL, err == nil && resp.StatusCode == http.StatusOK
}
