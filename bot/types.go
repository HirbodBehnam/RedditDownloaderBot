package bot

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
