package main

import (
	"bytes"
	"encoding/json"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	guuid "github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var UserMedia *cache.Cache
var bot *tgbotapi.BotAPI

const VERSION = "1.0.0"

var QUALITY = []string{"1080", "720", "480", "360", "240", "96"}

func main() {
	var err error
	if len(os.Args) < 2 {
		log.Fatal("Please pass the bot token as argument.")
	}
	// load bot
	bot, err = tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		log.Fatal("Cannot initialize the bot:", err.Error())
	}
	log.Println("Reddit Downloader Bot v" + VERSION)
	if !CheckFfmpegExists() {
		log.Println("WARNING: ffmpeg is not installed on your system")
	}
	log.Println("Bot authorized on account", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Cannot get updates channel:", err.Error())
	}

	UserMedia = cache.New(5*time.Minute, 10*time.Minute)
	// fetch updates
	for update := range updates {
		if update.CallbackQuery != nil {
			go HandleCallback(update.CallbackQuery.Data, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			continue
		}
		if update.Message == nil { // ignore any non-Message
			continue
		}
		// check if the message is command
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Hello and welcome!\nJust send me the link of the post to download it for you."))
			case "about":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Reddit Downloader Bot v"+VERSION+"\nBy Hirbod Behnam\nSource: https://github.com/HirbodBehnam/RedditDownloaderBot"))
			case "help":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Just send me the link of the reddit post. If it's text, I will send the text of the post. If it's a photo or video, I will send the it with the title as caption."))
			default:
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry this command is not recognized; Try /help"))
			}
			continue
		}
		// only text massages are allowed
		if update.Message.Text == "" {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please send the link to bot"))
			continue
		}
		go StartFetch(update.Message.Text, update.Message.Chat.ID, update.Message.MessageID)
	}
}

// this method runs when the user chooses one of the resolutions
func HandleCallback(data string, id int64, msgId int) {
	// at first get the url from cache
	// the first char is requested type (media or file)
	_, _ = bot.DeleteMessage(tgbotapi.NewDeleteMessage(id, msgId))
	if d, exists := UserMedia.Get(data[2:]); exists {
		m := d.(map[string]string)
		if m["type"] == "0" { // photo
			HandlePhotoFinal(m[data[1:2]], m["title"], id, data[:1] == "0")
		} else if m["type"] == "1" { // video
			HandleVideoFinal(m[data[1:2]], m["title"], id)
		} else { // gif; type = 2
			HandleGifFinal(m[data[1:2]], m["title"], id)
		}
		UserMedia.Delete(data[1:])
	} else {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Please resend the link to bot"))
	}
}

// download and send the photo
func HandlePhotoFinal(photoUrl, title string, id int64, asPhoto bool) {
	// get the file name
	var fileName string
	{
		u, err := url.Parse(photoUrl)
		if err != nil {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse url: "+err.Error()))
			return
		}
		fileName = u.Path[1:]
	}
	// generate a temp file
	tmpFile, err := ioutil.TempFile("", "*."+fileName)
	if err != nil {
		log.Println("Cannot create temp file:", err)
		_, _ = bot.Send(tgbotapi.NewMessage(id, "internal error"))
		return
	}
	defer os.Remove(tmpFile.Name()) // clean up
	// download the file
	err = DownloadFile(photoUrl, tmpFile)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot download file: "+err.Error()))
		return
	}
	// send the file to telegram
	if asPhoto {
		msg := tgbotapi.NewPhotoUpload(id, tmpFile.Name())
		msg.Caption = title
		_, err = bot.Send(msg)
	} else {
		msg := tgbotapi.NewDocumentUpload(id, tmpFile.Name())
		msg.Caption = title
		_, err = bot.Send(msg)
	}
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot upload file: "+err.Error()))
		log.Println("Cannot upload file:", err)
		return
	}
}

// download and send the gif
func HandleGifFinal(gifUrl, title string, id int64) {
	firstMessage, _ := bot.Send(tgbotapi.NewMessage(id, "Downloading GIF..."))
	defer bot.Send(tgbotapi.NewDeleteMessage(id, firstMessage.MessageID))
	var fileName string
	{
		u, err := url.Parse(gifUrl)
		if err != nil {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse url: "+err.Error()))
			return
		}
		fileName = u.Path[1:]
	}
	// generate a temp file
	tmpFile, err := ioutil.TempFile("", "*."+fileName+".mp4")
	if err != nil {
		log.Println("Cannot create temp file:", err)
		_, _ = bot.Send(tgbotapi.NewMessage(id, "internal error"))
		return
	}
	defer os.Remove(tmpFile.Name()) // clean up
	// download the file
	err = DownloadFile(gifUrl, tmpFile)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot download file: "+err.Error()))
		return
	}
	// upload it
	msg := tgbotapi.NewAnimationUpload(id, tmpFile.Name())
	msg.Caption = title
	_, err = bot.Send(msg)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot upload file: "+err.Error()))
		log.Println("Cannot upload file:", err)
		return
	}
}

func HandleVideoFinal(vidUrl, title string, id int64) {
	infoMessage, _ := bot.Send(tgbotapi.NewMessage(id, "Downloading Video..."))
	// maybe add filename later?
	vidFile, err := ioutil.TempFile("", "*.mp4")
	if err != nil {
		log.Println("Cannot create temp file:", err)
		_, _ = bot.Send(tgbotapi.NewMessage(id, "internal error"))
		return
	}
	defer os.Remove(vidFile.Name())
	hasAudio := true
	audFile, err := ioutil.TempFile("", "*.mp4")
	if err != nil {
		log.Println("Cannot create temp file:", err)
		_, _ = bot.Send(tgbotapi.NewMessage(id, "internal error"))
		return
	}
	defer os.Remove(audFile.Name())
	// download the video
	err = DownloadFile(vidUrl, vidFile)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot download file: "+err.Error()))
		return
	}
	// download the audio if available
	err = DownloadFile(vidUrl[:strings.LastIndex(vidUrl, "/")]+"/audio", audFile)
	if err != nil {
		log.Println(err)
		hasAudio = false
	}
	// merge audio and video if needed
	toUpload := vidFile.Name()
	_, _ = bot.Send(tgbotapi.NewDeleteMessage(id, infoMessage.MessageID))
	if hasAudio {
		// check ffmpeg first
		if !CheckFfmpegExists() {
			log.Println("ffmpeg not found!")
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot convert video: ffmpeg is not installed on server"))
			return
		}
		// convert
		infoMessage, _ = bot.Send(tgbotapi.NewMessage(id, "Converting video..."))
		finalFile, err := ioutil.TempFile("", "*.mp4")
		if err != nil {
			log.Println("Cannot create temp file:", err)
			_, _ = bot.Send(tgbotapi.NewMessage(id, "internal error"))
			return
		}
		defer os.Remove(finalFile.Name())
		cmd := exec.Command("ffmpeg", "-i", vidFile.Name(), "-i", audFile.Name(), "-c", "copy", finalFile.Name(), "-y")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			log.Println("Cannot convert video:", err)
			log.Println(string(stderr.Bytes()))
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot convert video"))
			return
		}
		_, _ = bot.Send(tgbotapi.NewDeleteMessage(id, infoMessage.MessageID))
		toUpload = finalFile.Name()
	}
	// before upload, check the file size
	{
		fi, err := os.Stat(toUpload)
		if err != nil {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot read file on server"))
			log.Println("Cannot read file for stats:", err)
			return
		}
		if fi.Size() > 50*1000*1000 { // for some reasons, this is not 50 * 1024 * 1024
			msg := tgbotapi.NewMessage(id, "This file is too big to upload it on telegram!\nHere is the link to video: "+vidUrl)
			if hasAudio {
				msg.Text += "\nHere is also the link to audio file: " + vidUrl[:strings.LastIndex(vidUrl, "/")] + "/audio"
			}
			_, _ = bot.Send(msg)
			return
		}
	}
	// upload the file
	infoMessage, _ = bot.Send(tgbotapi.NewMessage(id, "Uploading video..."))
	msg := tgbotapi.NewVideoUpload(id, toUpload)
	msg.Caption = title
	_, err = bot.Send(msg)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot upload file: "+err.Error()))
		log.Println("Cannot upload file:", err)
		return
	}
	_, _ = bot.Send(tgbotapi.NewDeleteMessage(id, infoMessage.MessageID))
}

// this method starts when the user sends the link
func StartFetch(postUrl string, id int64, msgId int) {
	// dont crash the whole thing
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovering from panic in printAllOperations error is: %v \n", r)
		}
	}()
	var postId string
	// get the id
	{
		u, err := url.Parse(postUrl)
		if err != nil {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse the url. Is the thing you send a url?"))
			return
		}
		split := strings.Split(u.Path, "/")
		if len(split) < 5 {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "This url looks too small"))
			return
		}
		postId = split[4]
	}
	// now download the json
	rawJson, err := DownloadString("https://api.reddit.com/api/info/?id=t3_" + postId)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot download page: "+err.Error()))
		return
	}
	// parse the json
	var root map[string]interface{}
	err = json.Unmarshal(rawJson, &root)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse the page as json:"+err.Error()))
		return
	}
	rawJson = nil // gc stuff
	// get post type
	// to do so, I check data->children[0]->data->post_hint
	{
		data, exists := root["data"]
		if !exists {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse the page data: cannot find node `data`"))
			return
		}
		children, exists := data.(map[string]interface{})["children"]
		if !exists {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse the page data: cannot find node `data->children`"))
			return
		}
		data = children.([]interface{})[0]
		data, exists = data.(map[string]interface{})["data"]
		if !exists {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse the page data: cannot find node `data->children[0]->data`"))
			return
		}
		root = data.(map[string]interface{})
	}
	// check it
	msg := tgbotapi.NewMessage(id, "")
	msg.ReplyToMessageID = msgId
	if hint, exists := root["post_hint"]; exists {
		switch hint.(string) {
		case "image": // image or gif
			msg.Text = "Please select the quality"
			if root["url"].(string)[len(root["url"].(string))-3:] == "gif" {
				msg.ReplyMarkup = GenerateInlineKeyboardPhoto(root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["variants"].(map[string]interface{})["mp4"].(map[string]interface{}), root["title"].(string), true)
			} else {
				msg.ReplyMarkup = GenerateInlineKeyboardPhoto(root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{}), root["title"].(string), false)
			}
			_, _ = bot.Send(msg)
		case "link": // link
			u := root["url"].(string)
			if u[len(u)-4:] == "gifv" && strings.HasPrefix(u, "https://i.imgur.com") { // imgur gif
				HandleGifFinal(u[:len(u)-4]+"mp4", root["title"].(string), id)
				return
			}
			msg.Text = html.UnescapeString(root["title"].(string) + "\n" + u) // a normal link
			_, _ = bot.Send(msg)
		case "hosted:video": // v.reddit
			msg.Text = "Please select the quality"
			vid := root["media"].(map[string]interface{})["reddit_video"].(map[string]interface{})
			msg.ReplyMarkup = GenerateInlineKeyboardVideo(vid["fallback_url"].(string), root["url"].(string), root["title"].(string))
			_, _ = bot.Send(msg)
		default:
			msg.Text = "This post type is not supported: " + hint.(string)
			_, _ = bot.Send(msg)
		}
	} else { // text
		msg.Text = html.UnescapeString(root["title"].(string) + "\n" + root["selftext"].(string)) // just make sure that the markdown is ok
		msg.Text = strings.ReplaceAll(msg.Text, "&#x200B;", "")                                   // https://www.reddit.com/r/OutOfTheLoop/comments/9abjhm/what_does_x200b_mean/
		msg.ParseMode = "markdown"
		_, _ = bot.Send(msg)
	}
}

// generates an inline keyboard for user to choose the quality of media and stores it in cache db
func GenerateInlineKeyboardPhoto(data map[string]interface{}, title string, isGif bool) tgbotapi.InlineKeyboardMarkup {
	var mediaType string
	if isGif {
		mediaType = "Gif"
	} else {
		mediaType = "Picture"
	}
	m := make(map[string]string) // I store this in cache
	var keyboard [][]tgbotapi.InlineKeyboardButton
	// at first generate a guid for cache
	id := guuid.New().String()
	// at first include source image
	{
		tKeyboard := make([]tgbotapi.InlineKeyboardButton, 2) // two button in raw: as media or as file
		u, w, h := ExtractLinkAndRes(data["source"])
		tKeyboard[0] = tgbotapi.NewInlineKeyboardButtonData(mediaType+" "+w+"×"+h, "00"+id)
		tKeyboard[1] = tgbotapi.NewInlineKeyboardButtonData("File "+w+"×"+h, "10"+id)
		m["0"] = u
		if isGif {
			keyboard = append(keyboard, tKeyboard[:1]) // file type is not supported for gifs
		} else {
			keyboard = append(keyboard, tKeyboard)
		}
	}
	// now get all other thumbs
	for k, v := range data["resolutions"].([]interface{}) {
		tKeyboard := make([]tgbotapi.InlineKeyboardButton, 2) // two button in raw: as media or as file
		u, w, h := ExtractLinkAndRes(v)
		tKeyboard[0] = tgbotapi.NewInlineKeyboardButtonData(mediaType+" "+w+"×"+h, "0"+strconv.Itoa(k+1)+id)
		tKeyboard[1] = tgbotapi.NewInlineKeyboardButtonData("File "+w+"×"+h, "1"+strconv.Itoa(k+1)+id)
		m[strconv.Itoa(k+1)] = u
		if isGif {
			keyboard = append(keyboard, tKeyboard[:1]) // file type is not supported for gifs
		} else {
			keyboard = append(keyboard, tKeyboard)
		}
	}
	if isGif {
		m["type"] = "2" // gif type
	} else {
		m["type"] = "0" // photo type
	}
	m["title"] = title
	UserMedia.Set(id, m, cache.DefaultExpiration)
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

func GenerateInlineKeyboardVideo(vidUrl, base string, title string) tgbotapi.InlineKeyboardMarkup {
	m := make(map[string]string) // I store this in cache
	var keyboard [][]tgbotapi.InlineKeyboardButton
	// at first generate a guid for cache
	id := guuid.New().String()
	// get max res
	res := vidUrl[len(vidUrl)-20 : len(vidUrl)-16]
	if res[0] == '_' {
		res = res[1:]
	}
	// list all of the qualities
	startAdd := false
	for k, v := range QUALITY {
		if v == res || startAdd {
			startAdd = true
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v+"p", "0"+strconv.Itoa(k)+id))) // 0 makes the data compatible with phototype
			m[strconv.Itoa(k)] = base + "/DASH_" + v
		}
	}
	m["type"] = "1" // video
	m["title"] = title
	UserMedia.Set(id, m, cache.DefaultExpiration)
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// extracts the data from "source":{ "url":"https://preview.redd.it/utx00pfe4cp41.jpg?auto=webp&amp;s=de4ff82478b12df6369b8d7eeca3894f094e87e1", "width":624, "height":960 } stuff
// first return values are url, width, height
func ExtractLinkAndRes(data interface{}) (string, string, string) {
	kv := data.(map[string]interface{})
	return html.UnescapeString(kv["url"].(string)), strconv.Itoa(int(kv["width"].(float64))), strconv.Itoa(int(kv["height"].(float64)))
}

// downloads a URL's data as string
// the user agent must change
func DownloadString(Url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return nil, err
	}
	// mimic chrome
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return nil, errors.New("forbidden")
	}
	return ioutil.ReadAll(resp.Body)
}

// downloads a web page to file
func DownloadFile(Url string, file *os.File) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return err
	}
	// mimic chrome
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return errors.New("forbidden")
	}
	_, err = io.Copy(file, resp.Body)
	return err
}

// returns true if ffmpeg is found
func CheckFfmpegExists() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}
