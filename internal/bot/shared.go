package bot

import "RedditDownloaderBot/pkg/reddit"

const regularMaxUploadSize = 50 * 1000 * 1000 // these must be 1000 not 1024
const photoMaxUploadSize = 10 * 1000 * 1000

// noThumbnailNeededSize is the size which files bigger than it need a thumbnail
// Otherwise the telegram will show them without one
const noThumbnailNeededSize = 10 * 1000 * 1000

// maxTextSize is the maximum text size which can be sent in the bot as a message
const maxTextSize = 4096

// The maximum dimensions which a thumbnail can have.
// From the Telegram docs this must be 320x320 but based on my tests,
// the dimensions does not matter if the file size is less than 200k.
var maxThumbnailDimensions = reddit.Dimension{
	Width:  1080,
	Height: 1080,
}
