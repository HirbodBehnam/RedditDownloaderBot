package bot

import (
	"encoding/json"
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	"github.com/HirbodBehnam/RedditDownloaderBot/util"
)

// CallbackButtonData is the data which is sent to us after user clicks on an inline button
// It might look dumb, but I use json to store in button query data
type CallbackButtonData struct {
	// The id to search the cache for it.
	// ID is the base64 of the uuid.UUID instead of hex
	ID string `json:"u"`
	// In the map of the data, CallbackDataCached.Links is the one we are looking for
	// We use this key to get the url of the media we are looking for
	LinkKey int `json:"l"`
	// Mode might be used for some medias to apply an option
	Mode CallbackButtonDataMode `json:"m,omitempty"`
}

// CallbackButtonDataMode specifies some options of callback data if needed
type CallbackButtonDataMode uint8

const (
	// CallbackButtonDataModePhoto means that we should use photo instead of file to send it to Telegram
	CallbackButtonDataModePhoto CallbackButtonDataMode = iota
	// CallbackButtonDataModeFile means that we should use file instead of photo to send it to Telegram
	CallbackButtonDataModeFile
)

// String returns the json format of CallbackButtonData
func (c *CallbackButtonData) String() string {
	b, _ := json.Marshal(c)
	return util.ByteToString(b)
}

// CallbackDataCached is the data we store associated with an ID which is CallbackButtonData.ID
// We store this type in mediaCache
type CallbackDataCached struct {
	// The list of links which the one in CallbackButtonData.LinkKey is used
	Links map[int]string
	// Title of the post
	Title string
	// Thumbnail link of the post
	ThumbnailLink string
	// What media is this
	Type reddit.FetchResultMediaType
}
