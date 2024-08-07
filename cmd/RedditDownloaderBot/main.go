package main

import (
	"RedditDownloaderBot/internal/bot"
	"RedditDownloaderBot/internal/cache"
	"RedditDownloaderBot/pkg/common"
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	"github.com/go-faster/errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	errors.DisableTrace()
	var err error
	log.Println("Reddit Downloader Bot v" + common.Version)
	if !util.DoesFfmpegExists() {
		log.Println("Warning: FFmpeg is not installed on your computer.")
	}
	// Load the variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	botToken := os.Getenv("BOT_TOKEN")
	if clientID == "" || clientSecret == "" || botToken == "" {
		log.Fatalln("Please set CLIENT_ID, CLIENT_SECRET, and BOT_TOKEN according to the Readme file on GitHub.")
	}
	botClient := bot.Client{}
	// Start up database
	if redisAddress, redisPort := os.Getenv("REDIS_ADDRESS"), os.Getenv("REDIS_PORT"); redisAddress != "" && redisPort != "" {
		// Parse ttl
		ttl, _ := time.ParseDuration(os.Getenv("REDIS_TTL"))
		if ttl <= 0 {
			ttl = 5 * time.Minute
		}
		botClient.CallbackCache, err = cache.NewRedisCache(redisAddress+":"+redisPort, os.Getenv("REDIS_PASSWORD"), ttl)
		if err != nil {
			log.Fatalln("Cannot connect to Redis:", err)
		}
	} else { // Simple in cache memory
		botClient.CallbackCache = cache.NewMemoryCache(5*time.Minute, 10*time.Minute)
	}
	defer botClient.CallbackCache.Close()
	// Start the reddit oauth
	botClient.RedditOauth, err = reddit.NewRedditOauth(clientID, clientSecret)
	if err != nil {
		log.Fatalln("Cannot initialize the Reddit OAuth:", err.Error())
	}
	botClient.RunBot(botToken, getAllowedUsers())
}

// getAllowedUsers gets the list of users which are allowed to use the bot
func getAllowedUsers() (allowedIDs []int64) {
	usersString := strings.Split(os.Getenv("ALLOWED_USERS"), ",")
	allowedIDs = make([]int64, 0, len(usersString))
	for _, idString := range usersString {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err == nil {
			allowedIDs = append(allowedIDs, id)
		}
	}
	return
}
