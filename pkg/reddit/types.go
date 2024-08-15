package reddit

// DownloadAudioQuality is the string to send to user when they want to download audio of a video
const DownloadAudioQuality = "Audio"

// FetchError is an error type which might be returned from StartFetch function
type FetchError struct {
	// NormalError is the real error message which has caused the error. This can be printed to stdout for
	// debugging, but it may contain sensitive data. So it should not be sent to users.
	// There is a special case which this string is empty. That means that there was a problem with user's
	// request, and we should not log anything.
	NormalError string
	// BotError on the other hand, must be sent to user. It should never be empty.
	BotError string
}

// Error returns the normal error which might contain sensitive information
func (e FetchError) Error() string {
	return e.NormalError
}

// FetchResultText is a result of StartFetch which represents a reddit text
type FetchResultText struct {
	// Title of the post
	Title string
	// The text
	Text string
}

// FetchResultComment is a result of StartFetch which represents a reddit comment text
type FetchResultComment struct {
	// The text of comment
	Text string
}

// FetchResultMediaEntry contains the quality and the link to a media in reddit
type FetchResultMediaEntry struct {
	// Link is the link to get this media
	Link string
	// The quality of this media
	Quality string
}

// FetchResultMediaEntries is a list of FetchResultMediaEntry
type FetchResultMediaEntries []FetchResultMediaEntry

// ToLinkMap creates a map of int -> string which represents the index of each entry with the link of it
func (e FetchResultMediaEntries) ToLinkMap() map[int]string {
	result := make(map[int]string, len(e))
	for i, media := range e {
		result[i] = media.Link
	}
	return result
}

// FetchResultMediaType says either is media is photo, gif or video
type FetchResultMediaType byte

const (
	FetchResultMediaTypePhoto FetchResultMediaType = iota
	FetchResultMediaTypeGif
	FetchResultMediaTypeVideo
)

// FetchResultMedia is the result of the
type FetchResultMedia struct {
	// Medias is the list of all available media in different qualities
	Medias FetchResultMediaEntries
	// This is the link to the thumbnail of this media
	// Might be empty
	ThumbnailLink string
	// Title is the title of the post
	Title string
	// Duration of the video. This entry does not matter on other types
	Duration int64
	// Types says what kind of media is this
	Type FetchResultMediaType
}

// HasAudio checks if a video does have audio
// It returns false if FetchResultMedia.Type is not FetchResultMediaTypeVideo
// index will be -1 if it doesn't have audio
func (f FetchResultMedia) HasAudio() (index int, has bool) {
	if f.Type != FetchResultMediaTypeVideo {
		return -1, false
	}
	// Edge case
	if len(f.Medias) == 0 {
		return -1, false
	}
	// We just need to check the final element in order to see if it does have audio or not
	has = f.Medias[len(f.Medias)-1].Quality == DownloadAudioQuality
	index = len(f.Medias) - 1
	if !has {
		index = -1
	}
	return
}

// FetchResultAlbumEntry is one media of album
type FetchResultAlbumEntry struct {
	// Link is the link to get this media
	Link string
	// The caption of this media
	Caption string
	// Types says what kind of media is this
	Type FetchResultMediaType
}

// FetchResultAlbum is a result of reddit album
type FetchResultAlbum struct {
	Album []FetchResultAlbumEntry
}

// Dimension of a media
type Dimension struct {
	Width  int64
	Height int64
}
