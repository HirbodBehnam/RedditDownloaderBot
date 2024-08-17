package bot

import "RedditDownloaderBot/pkg/reddit"

const RegularMaxUploadSize = 50 * 1000 * 1000 // these must be 1000 not 1024
const PhotoMaxUploadSize = 10 * 1000 * 1000

// NoThumbnailNeededSize is the size which files bigger than it need a thumbnail
// Otherwise the telegram will show them without one
const NoThumbnailNeededSize = 10 * 1000 * 1000

// The maximum dimensions which a thumbnail can have.
// From the Telegram docs this must be 320x320 but based on my tests,
// the dimensions does not matter if the file size is less than 200k.
var maxThumbnailDimensions = reddit.Dimension{
	Width:  1080,
	Height: 1080,
}
