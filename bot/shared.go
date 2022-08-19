package bot

import (
	"github.com/HirbodBehnam/RedditDownloaderBot/cache"
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
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

// Markdown is the styling format used in telegram messages
const Markdown = "Markdown"
