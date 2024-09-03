package bot

import (
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-faster/errors"
)

// handleGifUpload downloads a gif and then uploads it to Telegram
func (c *Client) handleGifUpload(bot *gotgbot.Bot, gifUrl, title, thumbnailUrl, postUrl string, dimension reddit.Dimension, chatID int64) error {
	// Inform the user we are doing some shit
	stopReportChannel := statusReporter(bot, chatID, gotgbot.ChatActionUploadVideo)
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := c.RedditOauth.DownloadGif(gifUrl)
	if err != nil {
		log.Println("Unable to download GIF", gifUrl, "for post", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t download this GIF.\nHere is the link: "+gifUrl, nil)
		return err
	}
	defer func() { // Cleanup
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	// Upload the gif
	// Check file size
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		_, err = bot.SendMessage(chatID, "The file is too large to upload on Telegram.\nHere is the link: "+gifUrl, nil)
		return err
	}
	// Check thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = c.RedditOauth.DownloadThumbnail(thumbnailUrl)
		if err != nil {
			log.Println("Cannot download GIF thumbnail", thumbnailUrl, ":", err)
		} else {
			defer func() {
				_ = tmpThumbnailFile.Close()
				_ = os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Check dimension
	if dimension.Empty() {
		dimension, err = reddit.GetVideoDimensions(tmpFile.Name())
		if err != nil {
			log.Println("Cannot get dimensions of GIF:", err)
		}
	}
	// Upload it
	animationOpt := &gotgbot.SendAnimationOpts{
		Caption:   addLinkIfNeeded(escapeMarkdown(title), postUrl),
		ParseMode: gotgbot.ParseModeMarkdownV2,
		Width:     dimension.Width,
		Height:    dimension.Height,
	}
	if tmpThumbnailFile != nil {
		animationOpt.Thumbnail = fileReaderFromOsFile(tmpThumbnailFile)
	}
	_, err = bot.SendAnimation(chatID, fileReaderFromOsFile(tmpFile), animationOpt)
	if err != nil {
		log.Println("Unable to upload GIF for post", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t upload this GIF.\nHere is the link: "+gifUrl, nil)
	}
	return err
}

// handleVideoUpload downloads a video and then uploads it to Telegram
func (c *Client) handleVideoUpload(bot *gotgbot.Bot, vidUrl, audioUrl, title, thumbnailUrl, postUrl string, dimension reddit.Dimension, duration, chatID int64) error {
	// Inform the user we are doing some shit
	stopReportChannel := statusReporter(bot, chatID, gotgbot.ChatActionUploadVideo)
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := c.RedditOauth.DownloadVideo(vidUrl, audioUrl)
	if err != nil {
		if errors.Is(err, reddit.FileTooBigError) {
			_, err = bot.SendMessage(chatID, "I couldn’t download this file because it’s too large.\n"+generateVideoUrlsMessage(vidUrl, audioUrl), nil)
		} else {
			log.Println("Unable to download video", vidUrl, "for post", postUrl, ":", err)
			_, err = bot.SendMessage(chatID, "I couldn’t download this video.\n"+generateVideoUrlsMessage(vidUrl, audioUrl), nil)
		}
		return err
	}
	defer func() { // Cleanup
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	// Check file size
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		_, err = bot.SendMessage(chatID, "This file is too large to upload on Telegram.\n"+generateVideoUrlsMessage(vidUrl, audioUrl), nil)
		return err
	}
	// Check thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = c.RedditOauth.DownloadThumbnail(thumbnailUrl)
		if err != nil {
			log.Println("Cannot download video thumbnail", thumbnailUrl, ":", err)
		} else {
			defer func() {
				_ = tmpThumbnailFile.Close()
				_ = os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Check dimension
	if dimension.Empty() {
		dimension, err = reddit.GetVideoDimensions(tmpFile.Name())
		if err != nil {
			log.Println("Cannot get dimensions of video:", err)
		}
	}
	// Upload it
	videoOpt := &gotgbot.SendVideoOpts{
		Duration:          duration,
		Caption:           addLinkIfNeeded(escapeMarkdown(title), postUrl),
		ParseMode:         gotgbot.ParseModeMarkdownV2,
		SupportsStreaming: true,
		Width:             dimension.Width,
		Height:            dimension.Height,
	}
	if tmpThumbnailFile != nil {
		videoOpt.Thumbnail = fileReaderFromOsFile(tmpThumbnailFile)
	}
	_, err = bot.SendVideo(chatID, fileReaderFromOsFile(tmpFile), videoOpt)
	if err != nil {
		log.Println("Unable to upload video for", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t upload this video.\n"+generateVideoUrlsMessage(vidUrl, audioUrl), nil)
	}
	return err
}

// handleVideoUpload downloads a photo and then uploads it to Telegram
func (c *Client) handlePhotoUpload(bot *gotgbot.Bot, photoUrl, title, thumbnailUrl, postUrl string, chatID int64, asPhoto bool) error {
	// Inform the user we are doing some shit
	var stopReportChannel chan struct{}
	if asPhoto {
		stopReportChannel = statusReporter(bot, chatID, gotgbot.ChatActionUploadPhoto)
	} else {
		stopReportChannel = statusReporter(bot, chatID, gotgbot.ChatActionUploadDocument)
	}
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := c.RedditOauth.DownloadPhoto(photoUrl)
	if err != nil {
		log.Println("Unable to download photo", photoUrl, "for post", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t download this image.\nHere is the link: "+photoUrl, nil)
		return err
	}
	defer func() { // Cleanup
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	// Check filesize
	if asPhoto {
		asPhoto = util.CheckFileSize(tmpFile.Name(), PhotoMaxUploadSize) // send photo as file if it is larger than 10MB
	}
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		_, err = bot.SendMessage(chatID, "The file is too large to upload on Telegram.\nHere is the link: "+photoUrl, nil)
		return err
	}
	// Download thumbnail
	var tmpThumbnailFile *os.File = nil
	if !asPhoto && !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		// photos does not support thumbnail...
		tmpThumbnailFile, err = c.RedditOauth.DownloadThumbnail(thumbnailUrl)
		if err != nil {
			log.Println("Cannot download photo thumbnail", thumbnailUrl, ":", err)
		} else {
			defer func() {
				_ = tmpThumbnailFile.Close()
				_ = os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload
	if asPhoto {
		_, err = bot.SendPhoto(chatID, fileReaderFromOsFile(tmpFile), &gotgbot.SendPhotoOpts{
			Caption:   addLinkIfNeeded(escapeMarkdown(title), postUrl),
			ParseMode: gotgbot.ParseModeMarkdownV2,
		})
	} else {
		documentOpt := &gotgbot.SendDocumentOpts{
			Caption:   addLinkIfNeeded(escapeMarkdown(title), postUrl),
			ParseMode: gotgbot.ParseModeMarkdownV2,
		}
		if tmpThumbnailFile != nil {
			documentOpt.Thumbnail = fileReaderFromOsFile(tmpThumbnailFile)
		}
		_, err = bot.SendDocument(chatID, fileReaderFromOsFile(tmpFile), documentOpt)
	}
	if err != nil {
		log.Println("Unable to upload photo for post", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t upload this image.\nHere is the link: "+photoUrl, nil)
	}
	return err
}

// handleAlbumUpload uploads an album to Telegram
func (c *Client) handleAlbumUpload(bot *gotgbot.Bot, album reddit.FetchResultAlbum, chatID int64, asFile bool) error {
	// Report status
	stopReportChannel := statusReporter(bot, chatID, gotgbot.ChatActionUploadPhoto)
	defer close(stopReportChannel)
	// Download each file of album
	var err error
	filePaths := make([]*os.File, 0, len(album.Album))
	defer func() { // cleanup
		for _, f := range filePaths {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()
	fileConfigs := make([]gotgbot.InputMedia, 0, len(album.Album))
	fileLinks := make([]string, 0, len(album.Album))
	for _, media := range album.Album {
		var tmpFile *os.File
		var link string
		var f gotgbot.InputMedia
		switch media.Type {
		case reddit.FetchResultMediaTypePhoto:
			tmpFile, err = c.RedditOauth.DownloadPhoto(media.Link)
			if err == nil {
				if asFile {
					f = gotgbot.InputMediaDocument{Media: fileReaderFromOsFile(tmpFile), Caption: media.Caption}
				} else {
					f = gotgbot.InputMediaPhoto{Media: fileReaderFromOsFile(tmpFile), Caption: media.Caption}
				}
			}
		case reddit.FetchResultMediaTypeGif:
			tmpFile, err = c.RedditOauth.DownloadGif(media.Link)
			if err == nil {
				if asFile {
					f = gotgbot.InputMediaDocument{Media: fileReaderFromOsFile(tmpFile), Caption: media.Caption}
				} else {
					f = gotgbot.InputMediaVideo{Media: fileReaderFromOsFile(tmpFile), Caption: media.Caption}
				}
			}
		case reddit.FetchResultMediaTypeVideo:
			tmpFile, err = c.RedditOauth.DownloadVideo(media.Link, "") // TODO: can i do something about audio URL?
			if err == nil {
				if asFile {
					f = gotgbot.InputMediaDocument{Media: fileReaderFromOsFile(tmpFile), Caption: media.Caption}
				} else {
					f = gotgbot.InputMediaVideo{
						Media:             fileReaderFromOsFile(tmpFile),
						Caption:           media.Caption,
						SupportsStreaming: true,
					}
				}
			}
		}
		if err != nil {
			log.Println("Unable to download album media:", err)
			_, _ = bot.SendMessage(chatID, "I couldn’t download the gallery./nHere is the link: "+link, nil)
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
		_, err = bot.SendMediaGroup(chatID, fileConfigs[i*10:(i+1)*10], nil)
		if err != nil {
			log.Println("Unable to upload gallery:", err)
			_, _ = bot.SendMessage(chatID, generateGalleryFailedMessage(fileLinks[i*10:(i+1)*10]), nil)
		}
	}
	err = nil // needed for last error check
	fileConfigs = fileConfigs[i*10:]
	if len(fileConfigs) == 1 {
		switch f := fileConfigs[0].(type) {
		case gotgbot.InputMediaPhoto:
			_, err = bot.SendPhoto(chatID, f.Media, nil)
		case gotgbot.InputMediaVideo:
			_, err = bot.SendVideo(chatID, f.Media, nil)
		case gotgbot.InputMediaDocument:
			_, err = bot.SendDocument(chatID, f.Media, nil)
		default:
			panic("IMPOSSIBLE")
		}
	} else if len(fileConfigs) > 1 {
		_, err = bot.SendMediaGroup(chatID, fileConfigs, nil)
	}
	if err != nil {
		log.Println("Unable to upload gallery:", err)
		_, err = bot.SendMessage(chatID, generateGalleryFailedMessage(fileLinks[i*10:]), nil)
	}
	return err
}

// handleAudioUpload simply downloads then uploads an audio to Telegram
func (c *Client) handleAudioUpload(bot *gotgbot.Bot, audioURL, title, postUrl string, duration, chatID int64) error {
	// Send status
	stopReportChannel := statusReporter(bot, chatID, gotgbot.ChatActionUploadVoice)
	defer close(stopReportChannel)
	// Create a temp file
	audioFile, err := c.RedditOauth.DownloadAudio(audioURL)
	if err != nil {
		log.Println("Unable to download audio from", audioURL, "for post", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t download the audio.\n"+generateAudioURLMessage(audioURL), nil)
		return err
	}
	defer func() {
		_ = audioFile.Close()
		_ = os.Remove(audioFile.Name())
	}()
	// Simply upload it to telegram
	_, err = bot.SendAudio(chatID, fileReaderFromOsFile(audioFile), &gotgbot.SendAudioOpts{
		Caption:   addLinkIfNeeded(escapeMarkdown(title), postUrl),
		ParseMode: gotgbot.ParseModeMarkdownV2,
		Duration:  duration,
	})
	if err != nil {
		log.Println("Unable to upload audio for post", postUrl, ":", err)
		_, err = bot.SendMessage(chatID, "I couldn’t upload the audio.\n"+generateAudioURLMessage(audioURL), nil)
	}
	return err
}

// statusReporter starts reporting for uploading a thing in telegram
// This function returns a channel which a message must be sent to it when reporting must be stopped
// You can also close the channel to stop the reporter.
func statusReporter(bot *gotgbot.Bot, chatID int64, action string) chan struct{} {
	doneChan := make(chan struct{}, 1)
	go statusReporterGoroutine(bot, chatID, action, doneChan)
	return doneChan
}

// statusReporterGoroutine must be called from another goroutine to report the status of upload
func statusReporterGoroutine(bot *gotgbot.Bot, chatID int64, action string, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second * 5) // we have to send it each 5 seconds
	_, _ = bot.SendChatAction(chatID, action, nil)
	for {
		select {
		case <-ticker.C:
			_, _ = bot.SendChatAction(chatID, action, nil)
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
	sb.WriteString("I couldn’t upload the media.\nHere are the links to the files:")
	for _, media := range medias {
		sb.WriteByte('\n')
		sb.WriteString(media)
	}
	return sb.String()
}
