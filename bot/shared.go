package bot

import (
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/patrickmn/go-cache"
	"time"
)

// mediaCache contains the requests of each user. Will reset in 10 minutes
var mediaCache *cache.Cache
var bot *tgbotapi.BotAPI
var RedditOauth *reddit.Oauth

const RegularMaxUploadSize = 50 * 1000 * 1000 // these must be 1000 not 1024
const PhotoMaxUploadSize = 10 * 1000 * 1000

// NoThumbnailNeededSize is the size which files bigger than it need a thumbnail
// Otherwise the telegram will show them without one
const NoThumbnailNeededSize = 10 * 1000 * 1000

func init() {
	mediaCache = cache.New(5*time.Minute, 10*time.Minute)
}
