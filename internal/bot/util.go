package bot

import (
	"RedditDownloaderBot/internal/cache"
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	"github.com/PaulSonOfLars/gotgbot/v2"
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
func createPhotoInlineKeyboard(id string, medias reddit.FetchResultMedia) gotgbot.InlineKeyboardMarkup {
	rows := make([][]gotgbot.InlineKeyboardButton, len(medias.Medias))
	for i, media := range medias.Medias {
		column := make([]gotgbot.InlineKeyboardButton, 2)
		// One button to download as photo
		info := CallbackButtonData{
			ID:      id,
			LinkKey: i,
			Mode:    CallbackButtonDataModePhoto,
		}
		column[0] = gotgbot.InlineKeyboardButton{
			Text:         "Photo " + media.Quality,
			CallbackData: info.String(),
		}
		// One button to download as file
		info.Mode = CallbackButtonDataModeFile
		column[1] = gotgbot.InlineKeyboardButton{
			Text:         "File " + media.Quality,
			CallbackData: info.String(),
		}
		// Add to rows
		rows[i] = column
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// createGifInlineKeyboard creates an inline keyboard for downloading gifs based on given reddit.FetchResultMedia
func createGifInlineKeyboard(id string, medias reddit.FetchResultMedia) gotgbot.InlineKeyboardMarkup {
	rows := make([][]gotgbot.InlineKeyboardButton, len(medias.Medias))
	for i, media := range medias.Medias {
		// One button to download as gif only
		// They don't support the file format
		info := CallbackButtonData{
			ID:      id,
			LinkKey: i,
		}
		// Add to rows
		rows[i] = []gotgbot.InlineKeyboardButton{{
			Text:         "GIF " + media.Quality,
			CallbackData: info.String(),
		}}
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// createVideoInlineKeyboard creates an inline keyboard for downloading gifs based on given reddit.FetchResultMedia
func createVideoInlineKeyboard(id string, medias reddit.FetchResultMedia) gotgbot.InlineKeyboardMarkup {
	rows := make([][]gotgbot.InlineKeyboardButton, len(medias.Medias))
	for i, media := range medias.Medias {
		// One button to download as gif only
		// They don't support the file format
		info := CallbackButtonData{
			ID:      id,
			LinkKey: i,
		}
		// Add to rows
		rows[i] = []gotgbot.InlineKeyboardButton{{
			Text:         media.Quality,
			CallbackData: info.String(),
		}}
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
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

// Create a gotgbot.FileReader from a os.File
func fileReaderFromOsFile(file *os.File) *gotgbot.FileReader {
	// Move the file pointer to beginning
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil
	}
	s, err := file.Stat()
	if err != nil {
		return nil
	}
	return &gotgbot.FileReader{Name: s.Name(), Data: file}
}

// getLinkMapOfFetchResultMediaEntries will get a map which holds the index of each entry plus its
// dimension
func getLinkMapOfFetchResultMediaEntries(entries reddit.FetchResultMediaEntries) map[int]cache.Media {
	result := make(map[int]cache.Media, len(entries))
	for i, media := range entries {
		result[i] = cache.Media{
			Link:   media.Link,
			Width:  media.Dim.Width,
			Height: media.Dim.Height,
		}
	}
	return result
}

// Sends a post description to the bot if it exists.
// Replies to a message. Can optionally use markdown if the description
// fits in a message.
func sendPostDescription(bot *gotgbot.Bot, description string, sentMessage *gotgbot.Message, useMarkdown bool) error {
	var err error = nil
	if description != "" { // if the description is empty don't do anything
		if len(description) > maxTextSize {
			_, err = bot.SendDocument(sentMessage.Chat.Id, &gotgbot.FileReader{
				Name: "description.txt",
				Data: strings.NewReader(description),
			}, &gotgbot.SendDocumentOpts{ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: sentMessage.MessageId,
			}})
		} else {
			opts := new(gotgbot.SendMessageOpts)
			if useMarkdown {
				opts.ParseMode = gotgbot.ParseModeMarkdownV2
			}
			_, err = sentMessage.Reply(bot, description, opts)
		}
	}
	return err
}
