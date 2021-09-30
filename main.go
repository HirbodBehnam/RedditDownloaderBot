package main

import (
	"github.com/HirbodBehnam/RedditDownloaderBot/bot"
	"github.com/HirbodBehnam/RedditDownloaderBot/config"
	reddit2 "github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	"github.com/HirbodBehnam/RedditDownloaderBot/util"
	"log"
	"os"
)

func main() {
	var err error
	if len(os.Args) < 4 {
		log.Fatal("Please pass the bot token, reddit client app and reddit client secret as arguments.")
	}
	log.Println("Reddit Downloader Bot v" + config.Version)
	if !util.DoesFfmpegExists() {
		log.Println("WARNING: ffmpeg is not installed on your system")
	}
	bot.RedditOauth, err = reddit2.NewRedditOauth(os.Args[2], os.Args[3])
	if err != nil {
		log.Fatal("Cannot initialize the reddit oauth:", err.Error())
	}
	bot.RunBot(os.Args[1])
}
