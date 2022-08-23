package cache

import "RedditDownloaderBot/pkg/reddit"

// CallbackDataCached is the data we store associated with an ID which is CallbackButtonData.ID
// We store this type in mediaCache
type CallbackDataCached struct {
	// The list of links which the one in CallbackButtonData.LinkKey is used
	Links map[int]string
	// Title of the post
	Title string
	// Thumbnail link of the post
	ThumbnailLink string
	// The Links[AudioIndex] contains the audio of a video
	// If there is no audio, this must be -1
	AudioIndex int
	// The duration of video
	Duration int
	// What media is this
	Type reddit.FetchResultMediaType
}
