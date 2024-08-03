package bot

import (
	"RedditDownloaderBot/internal/cache"
	"RedditDownloaderBot/pkg/reddit"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallbackCache is used to cache the data of the callback queries
var CallbackCache cache.Interface
var bot *tgbotapi.BotAPI
var RedditOauth *reddit.Oauth

const RegularMaxUploadSize = 50 * 1000 * 1000 // these must be 1000 not 1024
const PhotoMaxUploadSize = 10 * 1000 * 1000

// NoThumbnailNeededSize is the size which files bigger than it need a thumbnail
// Otherwise the telegram will show them without one
const NoThumbnailNeededSize = 10 * 1000 * 1000

// MarkdownV2 is the styling format used in telegram messages
const MarkdownV2 = "MarkdownV2"
