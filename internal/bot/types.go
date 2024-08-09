package bot

import (
	"RedditDownloaderBot/internal/cache"
	"RedditDownloaderBot/pkg/reddit"
)

// Client is the contains the data needed to operate the bot
type Client struct {
	CallbackCache cache.Interface
	RedditOauth   *reddit.Oauth
}

// AllowedUsers is a list of users which can use the bot
// An empty list means that everyone can use the bot
type AllowedUsers []int64

// IsAllowed checks if a user is allowed to use the bot or not
func (a AllowedUsers) IsAllowed(userID int64) bool {
	// Free bot
	if len(a) == 0 {
		return true
	}
	// Loop and search
	for _, id := range a {
		if userID == id {
			return true
		}
	}
	return false
}
