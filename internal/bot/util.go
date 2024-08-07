package bot

import (
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"os"
	"strings"
)

// If this value is false, we will not add the link of the post to each message caption.
var disableIncludeLinkInCaption = util.ParseEnvironmentVariableBool("DISABLE_LINK_IN_CAPTION")

// The characters which needs to be escaped based on
// https://core.telegram.org/bots/api#formatting-options
var markdownEscaper = strings.NewReplacer("_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!")

// createPhotoInlineKeyboard creates inline keyboards to get the quality info of a photo
// Each row represents a quality and each row has two columns: Send as photo or send as file
// The id must match the ID in the mediaCache
func createPhotoInlineKeyboard(id string, medias reddit.FetchResultMedia) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(medias.Medias))
	for i, media := range medias.Medias {
		column := make([]tgbotapi.InlineKeyboardButton, 2)
		// One button to download as photo
		info := CallbackButtonData{
			ID:      id,
			LinkKey: i,
			Mode:    CallbackButtonDataModePhoto,
		}
		column[0] = tgbotapi.NewInlineKeyboardButtonData("Photo "+media.Quality, info.String())
		// One button to download as file
		info.Mode = CallbackButtonDataModeFile
		column[1] = tgbotapi.NewInlineKeyboardButtonData("File "+media.Quality, info.String())
		// Add to rows
		rows[i] = column
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// createGifInlineKeyboard creates an inline keyboard for downloading gifs based on given reddit.FetchResultMedia
func createGifInlineKeyboard(id string, medias reddit.FetchResultMedia) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(medias.Medias))
	for i, media := range medias.Medias {
		// One button to download as gif only
		// They don't support the file format
		info := CallbackButtonData{
			ID:      id,
			LinkKey: i,
		}
		// Add to rows
		rows[i] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Gif "+media.Quality, info.String())}
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// createVideoInlineKeyboard creates an inline keyboard for downloading gifs based on given reddit.FetchResultMedia
func createVideoInlineKeyboard(id string, medias reddit.FetchResultMedia) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(medias.Medias))
	for i, media := range medias.Medias {
		// One button to download as gif only
		// They don't support the file format
		info := CallbackButtonData{
			ID:      id,
			LinkKey: i,
		}
		// Add to rows
		rows[i] = []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(media.Quality, info.String())}
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// Adds the link of the post to a text if needed (the INCLUDE_LINK is set)
func addLinkIfNeeded(text, link string) string {
	if disableIncludeLinkInCaption {
		return text
	}
	return text + "\n\n" + "[🔗 Link](" + link + ")"
}

// escapeMarkdown will escape the characters which are not ok in markdown
func escapeMarkdown(text string) string {
	return markdownEscaper.Replace(text)
}

// telegramUploadOsFile is wrapper for os.File in order to make it uploadable in Telegram
type telegramUploadOsFile struct {
	*os.File
}

func (f telegramUploadOsFile) NeedsUpload() bool {
	return true
}

func (f telegramUploadOsFile) UploadData() (string, io.Reader, error) {
	// Move the file pointer to beginning
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return "", nil, err
	}
	// Note: I can use io.NopCloser in order to make the bot not close the file
	s, err := f.Stat()
	if err != nil {
		return "", nil, err
	}
	return s.Name(), f, nil
}

func (f telegramUploadOsFile) SendData() string {
	panic("telegramUploadOsFile must be uploaded")
}
