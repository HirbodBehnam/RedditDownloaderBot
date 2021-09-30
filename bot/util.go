package bot

import (
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
