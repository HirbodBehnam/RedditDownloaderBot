package bot

import (
	"RedditDownloaderBot/internal/cache"
	"RedditDownloaderBot/pkg/common"
	"RedditDownloaderBot/pkg/reddit"
	"RedditDownloaderBot/pkg/util"
	"encoding/json"
	"errors"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// RunBot runs the bot with the specified token
func (c *Client) RunBot(token string, allowedUsers AllowedUsers) {
	// Setup the bot
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		log.Fatal("Cannot initialize the bot: ", err.Error())
	}
	log.Println("Bot authorized on account.", bot.Username)
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(_ *gotgbot.Bot, _ *ext.Context, err error) ext.DispatcherAction {
			log.Println("An error occurred while handling update: ", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)
	// Add handlers
	dispatcher.AddHandler(handlers.NewCallback(func(_ *gotgbot.CallbackQuery) bool {
		return true
	}, c.handleCallback))
	dispatcher.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return allowedUsers.IsAllowed(msg.From.Id)
	}, c.handleMessage))
	// Wait for updates
	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 60,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 60,
			},
		},
	})
	if err != nil {
		panic("Failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started . . .\n", bot.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

func (c *Client) handleMessage(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Only text messages are allowed
	if ctx.Message.Text == "" {
		_, err := ctx.EffectiveChat.SendMessage(bot, "Please send a Reddit post.", nil)
		return err
	}
	// Check if the message is command. I don't use command handler because I'll lose
	// the userID control.
	switch ctx.Message.Text {
	case "/start":
		_, err := ctx.EffectiveChat.SendMessage(bot, "Hey!\n\nJust send me a post or comment, and I’ll download it for you.", nil)
		return err
	case "/about":
		_, err := ctx.EffectiveChat.SendMessage(bot, "Reddit Downloader Bot v"+common.Version+"\nBy Hirbod Behnam\nSource: https://github.com/HirbodBehnam/RedditDownloaderBot", nil)
		return err
	case "/help":
		_, err := ctx.EffectiveChat.SendMessage(bot, "You can send me Reddit posts or comments. If it’s text only, I’ll send a text message. If it’s an image or video, I’ll upload and send the content along with the title and link.", nil)
		return err
	default:
		return c.fetchPostDetailsAndSend(bot, ctx)
	}
}

// fetchPostDetailsAndSend gets the basic info about the post being sent to us
func (c *Client) fetchPostDetailsAndSend(bot *gotgbot.Bot, ctx *ext.Context) error {
	result, realPostUrl, fetchErr := c.RedditOauth.StartFetch(ctx.Message.Text)
	if fetchErr != nil {
		if fetchErr.NormalError != "" {
			log.Println("Cannot fetch the post", ctx.Message.Text, ":", fetchErr.NormalError)
		}
		_, err := ctx.EffectiveMessage.Reply(bot, fetchErr.BotError, nil)
		return err
	}
	// Check the result type
	toSendText := ""
	toSendOpt := &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
	}
	switch data := result.(type) {
	case reddit.FetchResultText:
		toSendText = addLinkIfNeeded(data.Title+"\n"+data.Text, realPostUrl)
	case reddit.FetchResultComment:
		toSendText = addLinkIfNeeded(data.Text, realPostUrl)
	case reddit.FetchResultMedia:
		if len(data.Medias) == 0 {
			toSendText = "No media found."
			break
		}
		// If there is one media quality, download it
		// Also allow the user to choose between photo or document in image
		if len(data.Medias) == 1 && data.Type != reddit.FetchResultMediaTypePhoto {
			switch data.Type {
			case reddit.FetchResultMediaTypeGif:
				return handleGifUpload(bot, data.Medias[0].Link, data.Title, data.ThumbnailLink, realPostUrl, ctx.EffectiveChat.Id)
			case reddit.FetchResultMediaTypeVideo:
				// If the video does have an audio, ask user if they want the audio
				if _, hasAudio := data.HasAudio(); !hasAudio {
					// Otherwise, just download the video
					return handleVideoUpload(bot, data.Medias[0].Link, "", data.Title, data.ThumbnailLink, realPostUrl, data.Duration, ctx.EffectiveChat.Id)
				}
			default:
				panic("Shash")
			}
		}
		// Allow the user to select quality
		toSendText = "Please select the quality."
		idString := util.UUIDToBase64(uuid.New())
		audioIndex, _ := data.HasAudio()
		switch data.Type {
		case reddit.FetchResultMediaTypePhoto:
			toSendOpt.ReplyMarkup = createPhotoInlineKeyboard(idString, data)
		case reddit.FetchResultMediaTypeGif:
			toSendOpt.ReplyMarkup = createGifInlineKeyboard(idString, data)
		case reddit.FetchResultMediaTypeVideo:
			toSendOpt.ReplyMarkup = createVideoInlineKeyboard(idString, data)
		}
		// Insert the id in cache
		err := c.CallbackCache.SetMediaCache(idString, cache.CallbackDataCached{
			PostLink:      realPostUrl,
			Links:         data.Medias.ToLinkMap(),
			Title:         data.Title,
			ThumbnailLink: data.ThumbnailLink,
			Type:          data.Type,
			Duration:      data.Duration,
			AudioIndex:    audioIndex,
		})
		if err != nil {
			log.Println("Cannot set the media cache in database:", err)
		}
	case reddit.FetchResultAlbum:
		idString := util.UUIDToBase64(uuid.New())
		err := c.CallbackCache.SetAlbumCache(idString, data)
		if err != nil {
			log.Println("Cannot set the album cache in database:", err)
		}
		toSendText = "Download album as media or file?"
		toSendOpt.ReplyMarkup = gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
				gotgbot.InlineKeyboardButton{
					Text: "Media",
					CallbackData: CallbackButtonData{
						ID:   idString,
						Mode: CallbackButtonDataModePhoto,
					}.String(),
				},
				gotgbot.InlineKeyboardButton{
					Text: "File",
					CallbackData: CallbackButtonData{
						ID:   idString,
						Mode: CallbackButtonDataModeFile,
					}.String(),
				},
			}},
		}
	default:
		log.Printf("unknown type: %T\n", result)
		toSendText = "Unknown type (Please report this on the main GitHub project.)"
	}
	_, err := ctx.EffectiveMessage.Reply(bot, toSendText, toSendOpt)
	if err != nil {
		toSendOpt.ParseMode = gotgbot.ParseModeNone // fall back and don't format message
		_, err = ctx.EffectiveMessage.Reply(bot, toSendText, toSendOpt)
	}
	return err
}

// handleCallback handles the callback query of selecting a quality for any media type
func (c *Client) handleCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Don't crash!
	defer func() {
		if r := recover(); r != nil {
			_, _ = ctx.EffectiveChat.SendMessage(bot, "Cannot get data. (panic)", nil)
			log.Println("Recovering from panic:", r)
		}
	}()
	// Delete the message
	_, _ = bot.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.GetMessageId(), nil)
	// Parse the data
	var data CallbackButtonData
	err := json.Unmarshal([]byte(ctx.CallbackQuery.Data), &data)
	if err != nil {
		_, err = ctx.EffectiveChat.SendMessage(bot, "Broken callback data", nil)
		return err
	}
	// Get the cache from database
	cachedData, err := c.CallbackCache.GetAndDeleteMediaCache(data.ID)
	if errors.Is(err, cache.NotFoundErr) {
		// Check albums
		var album reddit.FetchResultAlbum
		album, err = c.CallbackCache.GetAndDeleteAlbumCache(data.ID)
		if err == nil {
			return handleAlbumUpload(bot, album, ctx.EffectiveChat.Id, data.Mode == CallbackButtonDataModeFile)
		} else if errors.Is(err, cache.NotFoundErr) {
			// It does not exist...
			_, err = ctx.EffectiveChat.SendMessage(bot, "Please resend the link.", nil)
			return err
		}
		// Fall to report internal error
	}
	// Check other errors
	if err != nil {
		log.Println("Cannot get Callback ID from database:", err)
		_, err = ctx.EffectiveChat.SendMessage(bot, "Internal error", nil)
		return err
	}
	// Check the link
	link, exists := cachedData.Links[data.LinkKey]
	if !exists {
		_, err = ctx.EffectiveChat.SendMessage(bot, "Please resend the link.", nil)
		return err
	}
	// Check the media type
	switch cachedData.Type {
	case reddit.FetchResultMediaTypeGif:
		return handleGifUpload(bot, link, cachedData.Title, cachedData.ThumbnailLink, cachedData.PostLink, ctx.EffectiveChat.Id)
	case reddit.FetchResultMediaTypePhoto:
		return handlePhotoUpload(bot, link, cachedData.Title, cachedData.ThumbnailLink, cachedData.PostLink, ctx.EffectiveChat.Id, data.Mode == CallbackButtonDataModePhoto)
	case reddit.FetchResultMediaTypeVideo:
		if data.LinkKey == cachedData.AudioIndex {
			return handleAudioUpload(bot, link, cachedData.Title, cachedData.PostLink, cachedData.Duration, ctx.EffectiveChat.Id)
		} else {
			audioURL := cachedData.Links[cachedData.AudioIndex]
			return handleVideoUpload(bot, link, audioURL, cachedData.Title, cachedData.ThumbnailLink, cachedData.PostLink, cachedData.Duration, ctx.EffectiveChat.Id)
		}
	}
	// What
	panic("Unknown media type: " + strconv.Itoa(int(cachedData.Type)))
}
