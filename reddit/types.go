package reddit

// FetchError is an error type which might be returned from StartFetch function
type FetchError struct {
	NormalError string
	BotError    string
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
	result := make(map[int]string)
	for i, media := range e {
		result[i] = media.Link
	}
	return result
}

// FetchResultMedia is the result of the
type FetchResultMedia struct {
	// Medias is the list of all available media in different qualities
	Medias FetchResultMediaEntries
	// This is the link to the thumbnail of this media
	// Might be empty
	ThumbnailLink string
	// Title is the title of the post
	Title string
	// Types says what kind of media is this
	Type FetchResultMediaType
}

// FetchResultMediaType says either is media is photo, gif or video
type FetchResultMediaType byte

const (
	FetchResultMediaTypePhoto FetchResultMediaType = iota
	FetchResultMediaTypeGif
	FetchResultMediaTypeVideo
)

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
