package cache

import "RedditDownloaderBot/pkg/reddit"

// CallbackDataCached is the data we store associated with an ID which is CallbackButtonData.ID
// We store this type in mediaCache
type CallbackDataCached struct {
	// The link of the post itself
	PostLink string
	// The list of links which the one in CallbackButtonData.LinkKey is used
	Links map[int]Media
	// Title of the post
	Title string
	// Thumbnail link of the post. Note that this is the preferred link which will be used.
	// Not all the links
	ThumbnailLink string
	// The description for the post. Also known as selftext
	Description string
	// The Links[AudioIndex] contains the audio of a video
	// If there is no audio, this must be -1
	AudioIndex int
	// The duration of video
	Duration int64
	// What media is this
	Type reddit.FetchResultMediaType
}

// Media holds the information for a media in reddit
type Media struct {
	Link string
	// Width of the thing. Can be zero
	Width int64
	// Height of the thing. Can be zero
	Height int64
}

// CallbackAlbumCached is the album data which the user will request to be either
// received in media or file format.
type CallbackAlbumCached struct {
	// The link of the post itself
	PostLink string
	// The album data
	Album reddit.FetchResultAlbum
}
