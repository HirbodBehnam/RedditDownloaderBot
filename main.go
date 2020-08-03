package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
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
	"path"
	"strconv"
	"strings"
	"time"
)

var UserMedia *cache.Cache
var bot *tgbotapi.BotAPI

const VERSION = "1.4.1"
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36"

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
		UserMedia.Delete(data[2:])
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

// Handles gallery posts like this: https://www.reddit.com/r/needforspeed/comments/i1p817/heres_some_of_my_favorite_screenshots_i_took
func HandelGallery(files map[string]interface{}, id int64) {
	// loop and download all files
	fileConfigs := make([]interface{}, 0)
	for _, imageRoot := range files {
		// extract the url
		image := imageRoot.(map[string]interface{})
		if image["status"].(string) != "valid" { // i have not encountered anything else except valid so far
			continue
		}
		link := image["s"].(map[string]interface{})["u"].(string)
		// for some reasons, i have to remove all "amp;" from the url in order to make this work
		link = strings.ReplaceAll(link, "amp;", "")
		// now download the file
		fileConfigs = append(fileConfigs, tgbotapi.NewInputMediaPhoto(link)) // TODO: this is a bad idea. I have to wait for multiple uploads in the bot api and fix this. Read more: https://github.com/go-telegram-bot-api/telegram-bot-api/pull/356
	}
	// upload all of them to telegram
	msg := tgbotapi.NewMediaGroup(id, fileConfigs)
	_, err := bot.Send(msg)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot upload files: "+err.Error()))
		log.Println("Cannot upload file:", err)
		return
	}
}

// download and send the gif
func HandleGifFinal(gifUrl, title string, id int64) {
	firstMessage, _ := bot.Send(tgbotapi.NewMessage(id, "Downloading GIF..."))
	defer bot.Send(tgbotapi.NewDeleteMessage(id, firstMessage.MessageID))
	// generate a temp file
	tmpFile, err := ioutil.TempFile("", "*.mp4")
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
	audioUrl := vidUrl[:strings.LastIndex(vidUrl, "/")] // base url
	if strings.Contains(vidUrl, ".mp4") {               // new reddit api or sth idk
		audioUrl += "/DASH_audio.mp4"
	} else { // old format
		audioUrl += "/audio"
	}
	hasAudio := DownloadFile(audioUrl, audFile) == nil
	// merge audio and video if needed
	toUpload := vidFile.Name()
	_, _ = bot.Send(tgbotapi.NewDeleteMessage(id, infoMessage.MessageID))
	if hasAudio {
		// check ffmpeg first
		if !CheckFfmpegExists() {
			log.Println("ffmpeg not found!")
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot convert video: ffmpeg is not installed on server;\nHere is the link to video: "+vidUrl+
				"\nAnd here is the audio: "+audioUrl))
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
				msg.Text += "\nHere is also the link to audio file: " + audioUrl
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
			log.Printf("Recovering from panic in StartFetch error is: %v, The url was:%v\n", r, postUrl)
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot get data. (panic)"))
		}
	}()
	var postId string
	// get the id
	{
		var u *url.URL = nil
		// check all lines for links. In new reddit update, sharing via Telegram adds the post title at its first
		lines := strings.Split(postUrl, "\n")
		for _, line := range lines {
			u, _ = url.Parse(line)
			if u != nil && (u.Host == "www.reddit.com" || u.Host == "reddit.com") {
				postUrl = line
				break
			}
			u = nil // this is for last loop. If u is nil after that final loop, it means that there is no reddit url in text
		}
		if u == nil {
			_, _ = bot.Send(tgbotapi.NewMessage(id, "Cannot parse reddit the url. Does your text contain a reddit url?"))
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
	// get the title
	title := root["title"].(string)
	title = html.UnescapeString(title)
	// check cross post
	if _, crossPost := root["crosspost_parent_list"]; crossPost {
		c := root["crosspost_parent_list"].([]interface{})
		if len(c) != 0 {
			root = c[0].(map[string]interface{})
		}
	}
	// check it
	msg := tgbotapi.NewMessage(id, "")
	msg.ReplyToMessageID = msgId
	if hint, exists := root["post_hint"]; exists {
		switch hint.(string) {
		case "image": // image or gif
			msg.Text = "Please select the quality"
			if root["url"].(string)[len(root["url"].(string))-3:] == "gif" {
				// check imgur gifs
				if strings.HasPrefix(root["url"].(string), "https://i.imgur.com") { // Example: https://www.reddit.com/r/dankmemes/comments/gag117/you_daughter_of_a_bitch_im_in/
					gifDownloadUrl := root["url"].(string)
					lastSlash := strings.LastIndex(gifDownloadUrl, "/")
					gifDownloadUrl = gifDownloadUrl[:lastSlash+1] + "download" + gifDownloadUrl[lastSlash:]
					HandleGifFinal(gifDownloadUrl, title, id)
					return
				}
				msg.ReplyMarkup = GenerateInlineKeyboardPhoto(root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["variants"].(map[string]interface{})["mp4"].(map[string]interface{}), title, true) // this is normal reddit gif
			} else {
				msg.ReplyMarkup = GenerateInlineKeyboardPhoto(root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{}), title, false)
			}
		case "link": // link
			u := root["url"].(string)
			if u[len(u)-4:] == "gifv" && strings.HasPrefix(u, "https://i.imgur.com") { // imgur gif
				HandleGifFinal(u[:len(u)-4]+"mp4", title, id)
				return
			}
			msg.Text = html.UnescapeString(title + "\n" + u) // a normal link
		case "hosted:video": // v.reddit
			msg.Text = "Please select the quality"
			vid := root["media"].(map[string]interface{})["reddit_video"].(map[string]interface{})
			keyboard := GenerateInlineKeyboardVideo(vid["fallback_url"].(string), title)
			if keyboard.InlineKeyboard != nil {
				msg.ReplyMarkup = keyboard
			} else { // just dl and send the main video
				HandleVideoFinal(vid["fallback_url"].(string), title, id)
				return
			}
		case "rich:video": // files hosted other than reddit; This bot currently supports gfycat.com
			if urlObject, domainExists := root["domain"]; domainExists {
				switch urlObject.(string) {
				case "gfycat.com": // just act like gif
					msg.Text = "Please select the quality"
					images := root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})
					if _, hasVariants := images["variants"]; hasVariants {
						if mp4, hasMp4 := images["variants"].(map[string]interface{})["mp4"]; hasMp4 {
							msg.ReplyMarkup = GenerateInlineKeyboardPhoto(mp4.(map[string]interface{}), title, true)
							break
						}
					}
					// check reddit_video_preview
					if vid, hasVid := root["preview"].(map[string]interface{})["reddit_video_preview"]; hasVid {
						if u, hasUrl := vid.(map[string]interface{})["fallback_url"]; hasUrl {
							msg.ReplyMarkup = GenerateInlineKeyboardVideo(u.(string), title)
							break
						}
					}
					msg.Text = "Cannot get the video. Here is the direct link to gfycat:\n" + root["url"].(string)
				case "streamable.com": // Example: https://streamable.com/u2jzoo
					// download the source at first
					source, err := DownloadString(root["url"].(string))
					if err != nil {
						msg.Text = "Cannot get the source code of " + root["url"].(string)
						break
					}
					// get the meta tag og:video
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(source)))
					if err != nil {
						msg.Text = "Cannot parse the source code of " + root["url"].(string)
						break
					}
					doc.Find("meta").Each(func(i int, s *goquery.Selection) {
						if name, _ := s.Attr("property"); name == "og:video" {
							videoUrl, _ := s.Attr("content")
							HandleVideoFinal(videoUrl, title, id)
						}
					})
					return
				default:
					msg.Text = "This bot does not support downloading from " + urlObject.(string) + "\nThe url field in json is " + root["url"].(string)
				}
			} else {
				msg.Text = "The type of this post is rich:video but it does not contains `domain`"
			}
		default:
			msg.Text = "This post type is not supported: " + hint.(string)
		}
	} else { // text or gallery
		if data, ok := root["media_metadata"]; ok { // gallery
			HandelGallery(data.(map[string]interface{}), id)
			return
		}
		// text
		msg.Text = html.UnescapeString(title + "\n" + root["selftext"].(string)) // just make sure that the markdown is ok
		msg.Text = strings.ReplaceAll(msg.Text, "&#x200B;", "")                  // https://www.reddit.com/r/OutOfTheLoop/comments/9abjhm/what_does_x200b_mean/
		msg.ParseMode = "markdown"
	}
	_, _ = bot.Send(msg)
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

func GenerateInlineKeyboardVideo(vidUrl, title string) tgbotapi.InlineKeyboardMarkup {
	m := make(map[string]string) // I store this in cache
	var keyboard [][]tgbotapi.InlineKeyboardButton
	// at first generate a guid for cache
	id := guuid.New().String()
	// get max res
	u, _ := url.Parse(vidUrl)
	u.RawQuery = ""
	res := path.Base(u.Path)[strings.LastIndex(path.Base(u.Path), "_")+1:] // the max res of video
	base := u.String()[:strings.LastIndex(u.String(), "/")]                // base url is this: https://v.redd.it/3lelz0i6crx41
	newFormat := strings.Contains(res, ".mp4")                             // this is new reddit format. The filenames are like DASH_480.mp4
	if newFormat {
		res = res[:strings.Index(res, ".")] // remove format to get the max quality
	}
	// list all of the qualities
	startAdd := false
	for k, v := range QUALITY {
		if v == res || startAdd {
			startAdd = true
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v+"p", "0"+strconv.Itoa(k)+id))) // 0 makes the data compatible with phototype
			m[strconv.Itoa(k)] = base + "/DASH_" + v
			if newFormat {
				m[strconv.Itoa(k)] += ".mp4"
			}
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
	req.Header.Set("User-Agent", UserAgent)
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
	req.Header.Set("User-Agent", UserAgent)
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
