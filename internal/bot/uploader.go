package bot

import (
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-faster/errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleGifUpload downloads a gif and then uploads it to Telegram
func handleGifUpload(gifUrl, title, thumbnailUrl, postUrl string, chatID int64) {
	// Inform the user we are doing some shit
	stopReportChannel := statusReporter(chatID, "upload_video")
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := reddit.DownloadGif(gifUrl)
	if err != nil {
		log.Println("Cannot download file", gifUrl, ":", err)
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t download this file.\nHere is the link: "+gifUrl))
		return
	}
	defer func() { // Cleanup
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	// Upload the gif
	// Check file size
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		bot.Send(tgbotapi.NewMessage(chatID, "This file is too large to upload on Telegram.\nHere is the link: "+gifUrl))
		return
	}
	// Check thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = reddit.DownloadThumbnail(thumbnailUrl)
		if err == nil {
			defer func() {
				tmpThumbnailFile.Close()
				os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload it
	msg := tgbotapi.NewAnimation(chatID, telegramUploadOsFile{tmpFile})
	msg.Caption = addLinkIfNeeded(escapeMarkdown(title), postUrl)
	msg.ParseMode = MarkdownV2
	if tmpThumbnailFile != nil {
		msg.Thumb = telegramUploadOsFile{tmpThumbnailFile}
	}
	_, err = bot.Send(msg)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t upload this file.\nHere is the link: "+gifUrl))
		log.Println("Cannot upload file:", err)
		return
	}
}

// handleVideoUpload downloads a video and then uploads it to Telegram
func handleVideoUpload(vidUrl, audioUrl, title, thumbnailUrl, postUrl string, duration int, chatID int64) {
	// Inform the user we are doing some shit
	stopReportChannel := statusReporter(chatID, "upload_video")
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := reddit.DownloadVideo(vidUrl, audioUrl)
	if err != nil {
		if errors.Is(err, reddit.FileTooBigError) {
			bot.Send(tgbotapi.NewMessage(chatID, "I can't download the file because it’s too large.\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		} else {
			log.Println("Cannot download file", vidUrl, ":", err)
			bot.Send(tgbotapi.NewMessage(chatID, "I can't download the file.\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		}
		return
	}
	defer func() { // Cleanup
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	// Check file size
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		bot.Send(tgbotapi.NewMessage(chatID, "This file is too large to upload on Telegram.\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		return
	}
	// Check thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = reddit.DownloadThumbnail(thumbnailUrl)
		if err == nil {
			defer func() {
				tmpThumbnailFile.Close()
				os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload it
	msg := tgbotapi.NewVideo(chatID, telegramUploadOsFile{tmpFile})
	msg.Caption = addLinkIfNeeded(escapeMarkdown(title), postUrl)
	msg.Duration = duration
	msg.SupportsStreaming = true
	msg.ParseMode = MarkdownV2
	if tmpThumbnailFile != nil {
		msg.Thumb = telegramUploadOsFile{tmpThumbnailFile}
	}
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Cannot upload file:", err)
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t upload this file.\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		return
	}
}

// handleVideoUpload downloads a photo and then uploads it to Telegram
func handlePhotoUpload(photoUrl, title, thumbnailUrl, postUrl string, chatID int64, asPhoto bool) {
	// Inform the user we are doing some shit
	var stopReportChannel chan struct{}
	if asPhoto {
		stopReportChannel = statusReporter(chatID, "upload_photo")
	} else {
		stopReportChannel = statusReporter(chatID, "upload_document")
	}
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := reddit.DownloadPhoto(photoUrl)
	if err != nil {
		log.Println("Cannot download file", photoUrl, ":", err)
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t download this file.\nHere is the link: "+photoUrl))
		return
	}
	defer func() { // Cleanup
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	// Check filesize
	if asPhoto {
		asPhoto = util.CheckFileSize(tmpFile.Name(), PhotoMaxUploadSize) // send photo as file if it is larger than 10MB
	}
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		bot.Send(tgbotapi.NewMessage(chatID, "This file is too large to upload on Telegram.\nHere is the link: "+photoUrl))
		return
	}
	// Download thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = reddit.DownloadThumbnail(thumbnailUrl)
		if err == nil {
			defer func() {
				tmpThumbnailFile.Close()
				os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload
	var msg tgbotapi.Chattable
	if asPhoto {
		photo := tgbotapi.NewPhoto(chatID, telegramUploadOsFile{tmpFile})
		photo.Caption = addLinkIfNeeded(escapeMarkdown(title), postUrl)
		photo.ParseMode = MarkdownV2
		if tmpThumbnailFile != nil {
			photo.Thumb = telegramUploadOsFile{tmpThumbnailFile}
		}
		msg = photo
	} else {
		photo := tgbotapi.NewDocument(chatID, telegramUploadOsFile{tmpFile})
		photo.Caption = addLinkIfNeeded(escapeMarkdown(title), postUrl)
		photo.ParseMode = MarkdownV2
		if tmpThumbnailFile != nil {
			photo.Thumb = telegramUploadOsFile{tmpThumbnailFile}
		}
		msg = photo
	}
	_, err = bot.Send(msg)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t upload this file.\nHere is the link: "+photoUrl))
		log.Println("Cannot upload file:", err)
		return
	}
}

// handleAlbumUpload uploads an album to Telegram
func handleAlbumUpload(album reddit.FetchResultAlbum, chatID int64, asFile bool) {
	// Report status
	stopReportChannel := statusReporter(chatID, "upload_photo")
	defer close(stopReportChannel)
	// Download each file of album
	var err error
	filePaths := make([]*os.File, 0, len(album.Album))
	defer func() { // cleanup
		for _, f := range filePaths {
			f.Close()
			os.Remove(f.Name())
		}
	}()
	fileConfigs := make([]interface{}, 0, len(album.Album))
	fileLinks := make([]string, 0, len(album.Album))
	for _, media := range album.Album {
		var tmpFile *os.File
		var link string
		var f interface{}
		switch media.Type {
		case reddit.FetchResultMediaTypePhoto:
			tmpFile, err = reddit.DownloadPhoto(media.Link)
			if err == nil {
				if asFile {
					uploadFile := tgbotapi.NewInputMediaDocument(telegramUploadOsFile{tmpFile})
					uploadFile.Caption = media.Caption
					f = uploadFile
				} else {
					uploadFile := tgbotapi.NewInputMediaPhoto(telegramUploadOsFile{tmpFile})
					uploadFile.Caption = media.Caption
					f = uploadFile
				}
			}
		case reddit.FetchResultMediaTypeGif:
			tmpFile, err = reddit.DownloadGif(media.Link)
			if err == nil {
				if asFile {
					uploadFile := tgbotapi.NewInputMediaDocument(telegramUploadOsFile{tmpFile})
					uploadFile.Caption = media.Caption
					f = uploadFile
				} else {
					uploadFile := tgbotapi.NewInputMediaVideo(telegramUploadOsFile{tmpFile})
					uploadFile.Caption = media.Caption
					f = uploadFile
				}
			}
		case reddit.FetchResultMediaTypeVideo:
			tmpFile, err = reddit.DownloadVideo(media.Link, "") // TODO: can i do something about audio URL?
			if err == nil {
				if asFile {
					uploadFile := tgbotapi.NewInputMediaDocument(telegramUploadOsFile{tmpFile})
					uploadFile.Caption = media.Caption
					f = uploadFile
				} else {
					uploadFile := tgbotapi.NewInputMediaVideo(telegramUploadOsFile{tmpFile})
					uploadFile.Caption = media.Caption
					uploadFile.SupportsStreaming = true
					f = uploadFile
				}
			}
		}
		if err != nil {
			log.Println("cannot download media of gallery:", err)
			bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t download the gallery media.\nHere is the link: "+link))
			continue
		}
		fileConfigs = append(fileConfigs, f)
		link = media.Link
		fileLinks = append(fileLinks, media.Link)
		filePaths = append(filePaths, tmpFile)
	}
	// Now upload 10 of them at once
	i := 0
	for ; i < len(fileConfigs)/10; i++ {
		_, err = bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, fileConfigs[i*10:(i+1)*10]))
		if err != nil {
			log.Println("Cannot upload gallery:", err)
			bot.Send(tgbotapi.NewMessage(chatID, generateGalleryFailedMessage(fileLinks[i*10:(i+1)*10])))
		}
	}
	err = nil // needed for last error check
	fileConfigs = fileConfigs[i*10:]
	if len(fileConfigs) == 1 {
		switch f := fileConfigs[0].(type) {
		case tgbotapi.InputMediaPhoto:
			_, err = bot.Send(tgbotapi.NewPhoto(chatID, f.Media))
		case tgbotapi.InputMediaVideo:
			_, err = bot.Send(tgbotapi.NewVideo(chatID, f.Media))
		case tgbotapi.InputMediaDocument:
			_, err = bot.Send(tgbotapi.NewDocument(chatID, f.Media))
		}
	} else if len(fileConfigs) > 1 {
		_, err = bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, fileConfigs))
	}
	if err != nil {
		log.Println("cannot upload gallery:", err)
		bot.Send(tgbotapi.NewMessage(chatID, generateGalleryFailedMessage(fileLinks[i*10:])))
	}
}

// handleAudioUpload simply downloads then uploads an audio to Telegram
func handleAudioUpload(audioURL, title, postUrl string, duration int, chatID int64) {
	// Send status
	stopReportChannel := statusReporter(chatID, "upload_voice")
	defer close(stopReportChannel)
	// Create a temp file
	audioFile, err := reddit.DownloadAudio(audioURL)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t download the audio file.\nHere is the link: "+generateAudioURLMessage(audioURL)))
		return
	}
	defer func() {
		audioFile.Close()
		os.Remove(audioFile.Name())
	}()
	// Simply upload it to telegram
	msg := tgbotapi.NewAudio(chatID, telegramUploadOsFile{audioFile})
	msg.Caption = addLinkIfNeeded(escapeMarkdown(title), postUrl)
	msg.ParseMode = MarkdownV2
	msg.Duration = duration
	_, err = bot.Send(msg)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "I couldn’t upload the audio file.\nHere is the link: "+generateAudioURLMessage(audioURL)))
		return
	}
}

// statusReporter starts reporting for uploading a thing in telegram
// This function returns a channel which a message must be sent to it when reporting must be stopped
// You can also close the channel to stop the reporter
func statusReporter(chatID int64, action string) chan struct{} {
	doneChan := make(chan struct{}, 1)
	go statusReporterGoroutine(chatID, action, doneChan)
	return doneChan
}

// statusReporterGoroutine must be called from another goroutine to report the status of upload
func statusReporterGoroutine(chatID int64, action string, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second * 5) // we have to send it each 5 seconds
	actionObject := tgbotapi.NewChatAction(chatID, action)
	bot.Send(actionObject)
	for {
		select {
		case <-ticker.C:
			bot.Send(actionObject)
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// generateVideoUrlsMessage generates a text message which it can be used to give the user
// the requested video and audio URL
func generateVideoUrlsMessage(videoUrl, audioUrl string) string {
	var sb strings.Builder
	sb.Grow(150)
	sb.WriteString("Here is the link to the video file: ")
	sb.WriteString(videoUrl)
	if audioUrl != "" {
		sb.WriteString("\n")
		sb.WriteString(generateAudioURLMessage(audioUrl))
	}
	return sb.String()
}

// generateAudioURLMessage generates a text to send to user when downloading an audio fails
func generateAudioURLMessage(audioURL string) string {
	return "Here is the link to the audio file: " + audioURL
}

// generateGalleryFailedMessage generates an error message to send to user when uploading gallery goes wrong
// The medias is the array of links to medias in the message which was meant to be uploaded
func generateGalleryFailedMessage(medias []string) string {
	var sb strings.Builder
	sb.Grow(len(medias) * 120) // each link length I guess
	sb.WriteString("I couldn’t upload the gallery media files.\nHere are the links:")
	for _, media := range medias {
		sb.WriteByte('\n')
		sb.WriteString(media)
	}
	return sb.String()
}
