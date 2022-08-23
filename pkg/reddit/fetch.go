package reddit

import (
	"RedditDownloaderBot/pkg/common"
	"RedditDownloaderBot/pkg/reddit/helpers"
	"RedditDownloaderBot/pkg/util"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html"
	"log"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// If this variable is true, it means that we don't allow nsfw posts to be downloaded
var denyNsfw = util.ParseEnvironmentVariableBool("DENY_NSFW")

// This error is returned if NSFW posts are disabled via denyNsfw and a nsfw post is requested
var nsfwNotAllowedErr = &FetchError{
	NormalError: "",
	BotError:    "NSFW posts are disabled.",
}

// qualities is the possible qualities of videos in reddit
var qualities = [...]string{"1080", "720", "480", "360", "240", "96"}

var giphyCommentRegex = regexp.MustCompile("!\\[gif]\\(giphy\\|(\\w+)(?:\\|downsized)?\\)")

// StartFetch gets the post info from url
// The fetchResult can be one of the following types:
// FetchResultText
// FetchResultComment
// FetchResultMedia
// FetchResultAlbum
func (o *Oauth) StartFetch(postUrl string) (fetchResult interface{}, fetchError *FetchError) {
	// Don't crash the whole application
	defer func() {
		if r := recover(); r != nil {
			fetchError = &FetchError{
				NormalError: fmt.Sprintf("recovering from panic in StartFetch error is: %v, The url was: %v", r, postUrl),
				BotError:    "Cannot get data. (panic)\nMaybe deleted post or invalid url?",
			}
		}
	}()
	// Get the post ID
	postId, isComment, fetchError := getPostID(postUrl)
	if fetchError != nil {
		return
	}
	if isComment {
		root, err := o.GetComment(postId)
		if err != nil {
			return nil, &FetchError{
				NormalError: "cannot download comment: " + err.Error(),
				BotError:    "Cannot download comment",
			}
		}
		// Check gif comments
		text := root["data"].(map[string]interface{})["children"].([]interface{})[0].(map[string]interface{})["data"].(map[string]interface{})["body"].(string)
		if matches := giphyCommentRegex.FindStringSubmatch(text); len(matches) == 2 {
			return FetchResultMedia{
				Medias: []FetchResultMediaEntry{{
					Link:    fmt.Sprintf("https://i.giphy.com/media/%s/giphy.gif", matches[1]),
					Quality: "giphy",
				}},
				Type:  FetchResultMediaTypeGif,
				Title: strings.ReplaceAll(text, matches[0], ""),
			}, nil
		}
		// Normal comment
		return FetchResultComment{text}, nil
	}
	// Now download the json
	root, err := o.GetPost(postId)
	if err != nil {
		fetchError = &FetchError{
			NormalError: "cannot get the post data: " + err.Error(),
			BotError:    "Cannot get the post data",
		}
		return
	}
	// Get post type
	// To do so, I check data->children[0]->data->post_hint
	{
		data, exists := root["data"]
		if !exists {
			fetchError = &FetchError{
				NormalError: "cannot parse the page data: cannot find node `data`",
				BotError:    "Cannot parse the page data: cannot find node `data`",
			}
			return
		}
		children, exists := data.(map[string]interface{})["children"]
		if !exists {
			fetchError = &FetchError{
				NormalError: "cannot parse the page data: cannot find node `data->children`",
				BotError:    "Cannot parse the page data: cannot find node `data->children`",
			}
			return
		}
		data = children.([]interface{})[0]
		data, exists = data.(map[string]interface{})["data"]
		if !exists {
			fetchError = &FetchError{
				NormalError: "cannot parse the page data: cannot find node `data->children[0]->data`",
				BotError:    "Cannot parse the page data: cannot find node `data->children[0]->data`",
			}
			return
		}
		root = data.(map[string]interface{})
	}
	// Check if the post is nsfw and bot forbids them
	nsfw, _ := root["over_18"].(bool)
	if denyNsfw && nsfw {
		return nil, nsfwNotAllowedErr
	}
	// Get the title
	title := root["title"].(string)
	title = html.UnescapeString(title)
	// Check thumbnail; This must be done before checking cross posts
	thumbnailUrl := ""
	if t, ok := root["thumbnail"]; ok {
		thumbnailUrl = t.(string)
		// Check the url; Sometimes, the value of this is default
		if !util.IsUrl(thumbnailUrl) {
			thumbnailUrl = ""
		}
	}
	// Check cross post
	if _, crossPost := root["crosspost_parent_list"]; crossPost {
		c := root["crosspost_parent_list"].([]interface{})
		if len(c) != 0 {
			root = c[0].(map[string]interface{})
		}
	}
	// Check it
	if hint, exists := root["post_hint"]; exists {
		switch hint.(string) {
		case "image": // image or gif
			result := FetchResultMedia{
				ThumbnailLink: thumbnailUrl,
				Title:         title,
			}
			if root["url"].(string)[len(root["url"].(string))-3:] == "gif" {
				result.Type = FetchResultMediaTypeGif
				// Check imgur gifs
				if strings.HasPrefix(root["url"].(string), "https://i.imgur.com") { // Example: https://www.reddit.com/r/dankmemes/comments/gag117/you_daughter_of_a_bitch_im_in/
					gifDownloadUrl := root["url"].(string)
					lastSlash := strings.LastIndex(gifDownloadUrl, "/")
					gifDownloadUrl = gifDownloadUrl[:lastSlash+1] + "download" + gifDownloadUrl[lastSlash:]
					result.Medias = []FetchResultMediaEntry{{
						Link:    gifDownloadUrl,
						Quality: "imgur", // It doesn't matter
					}}
				} else {
					result.Medias = extractPhotoGifQualities(root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["variants"].(map[string]interface{})["mp4"].(map[string]interface{}))
				}
			} else {
				result.Type = FetchResultMediaTypePhoto
				// Send the original file as well if it's on reddit
				if link, ok := root["url"].(string); ok && strings.HasPrefix(link, "https://i.redd.it/") {
					result.Medias = []FetchResultMediaEntry{
						{
							link,
							"Original",
						},
					}
				}
				result.Medias = append(result.Medias, extractPhotoGifQualities(root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{}))...)
			}
			return result, nil
		case "link": // link
			u := root["url"].(string)
			if u[len(u)-4:] == "gifv" && strings.HasPrefix(u, "https://i.imgur.com") { // imgur gif
				return FetchResultMedia{
					Medias: []FetchResultMediaEntry{{
						Link:    u[:len(u)-4] + "mp4",
						Quality: "imgur", // It doesn't matter
					}},
					ThumbnailLink: thumbnailUrl,
					Title:         title,
					Type:          FetchResultMediaTypeGif,
				}, nil
			}
			return FetchResultText{
				Title: title,
				Text:  html.UnescapeString(title + "\n" + u),
			}, nil
		case "hosted:video": // v.reddit
			redditVideo := root["media"].(map[string]interface{})["reddit_video"].(map[string]interface{})
			duration, _ := redditVideo["duration"].(float64) // Do not panic if duration does not exist. Just let the Telegram handle it
			vid := redditVideo["fallback_url"].(string)
			return FetchResultMedia{
				Medias:        extractVideoQualities(vid),
				ThumbnailLink: thumbnailUrl,
				Title:         title,
				Duration:      int(duration),
				Type:          FetchResultMediaTypeVideo,
			}, nil
		case "rich:video": // files hosted other than reddit; This bot currently supports gfycat.com
			if urlObject, domainExists := root["domain"]; domainExists {
				switch urlObject.(string) {
				case "gfycat.com": // just act like gif
					images := root["preview"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})
					if _, hasVariants := images["variants"]; hasVariants {
						if mp4, hasMp4 := images["variants"].(map[string]interface{})["mp4"]; hasMp4 {
							return FetchResultMedia{
								Medias:        extractPhotoGifQualities(mp4.(map[string]interface{})),
								ThumbnailLink: thumbnailUrl,
								Title:         title,
								Type:          FetchResultMediaTypeGif,
							}, nil
						}
					}
					// Check reddit_video_preview
					if vid, hasVid := root["preview"].(map[string]interface{})["reddit_video_preview"]; hasVid {
						if u, hasUrl := vid.(map[string]interface{})["fallback_url"]; hasUrl {
							return FetchResultMedia{
								Medias:        extractVideoQualities(u.(string)),
								ThumbnailLink: thumbnailUrl,
								Title:         title,
								Type:          FetchResultMediaTypeVideo,
							}, nil
						}
					}
					return nil, &FetchError{
						NormalError: "cannot get the media from gfycat. The main url was " + postUrl,
						BotError:    "Cannot get the video. Here is the direct link to gfycat:\n" + root["url"].(string),
					}
				case "streamable.com": // example: https://streamable.com/u2jzoo
					// Download the source at first
					source, err := common.GlobalHttpClient.Get(root["url"].(string))
					if err != nil {
						return nil, &FetchError{
							NormalError: "cannot get the source code of " + root["url"].(string) + ": " + err.Error(),
							BotError:    "Cannot get the source code of " + root["url"].(string),
						}
					}
					defer source.Body.Close()
					// Get the meta tag og:video
					doc, err := goquery.NewDocumentFromReader(source.Body)
					if err != nil {
						return nil, &FetchError{
							NormalError: "cannot get the parse code of " + root["url"].(string) + ": " + err.Error(),
							BotError:    "Cannot get the parse code of " + root["url"].(string),
						}
					}
					result := FetchResultMedia{
						Medias: []FetchResultMediaEntry{{
							Link:    "",
							Quality: "streamable",
						}},
						ThumbnailLink: thumbnailUrl,
						Title:         title,
						Type:          FetchResultMediaTypeVideo,
					}
					doc.Find("meta").Each(func(i int, s *goquery.Selection) {
						if name, _ := s.Attr("property"); name == "og:video" {
							result.Medias[0].Link, _ = s.Attr("content")
						}
					})
					return result, nil
				case "redgifs.com":
					// get redgifs info from api
					redgifsid := helpers.GetRedGifsID(root["url"].(string))
					if redgifsid == "" {
						return nil, &FetchError{
							NormalError: "cannot get redgifs id  from " + root["url"].(string) + ": " + err.Error(),
							BotError:    "Cannot get redgifs id  from  " + root["url"].(string),
						}
					}

					// api for redgifs is in https://i.redgifs.com/docs/index.html
					infoUrl := fmt.Sprintf("https://api.redgifs.com/v2/gifs/%s", redgifsid)

					source, err := common.GlobalHttpClient.Get(infoUrl)
					if err != nil {
						return nil, &FetchError{
							NormalError: "cannot get redgifs info " + infoUrl + ": " + err.Error(),
							BotError:    "Cannot get redgifs info " + infoUrl,
						}
					}
					defer source.Body.Close()
					// get video urls
					doc, err := helpers.GetRedGifsInfo(source.Body)
					if err != nil {
						return nil, &FetchError{
							NormalError: "cannot get the parse redgifs info from " + infoUrl + ": " + err.Error(),
							BotError:    "Cannot get the parse redgifs info from " + infoUrl,
						}
					}
					result := FetchResultMedia{
						Medias: []FetchResultMediaEntry{
							{
								Quality: "hd",
								Link:    doc.Gif.Urls.Hd,
							},
							{
								Quality: "sd",
								Link:    doc.Gif.Urls.Sd,
							},
						},
						ThumbnailLink: doc.Gif.Urls.Thumbnail,
						Title:         title,
						Type:          FetchResultMediaTypeVideo,
					}

					if doc.Gif.Urls.Gif != "" {
						result.Medias = append(result.Medias, FetchResultMediaEntry{
							Quality: "gif",
							Link:    doc.Gif.Urls.Gif,
						})
					}
					return result, nil
				default:
					return nil, &FetchError{
						NormalError: "",
						BotError:    "This bot does not support downloading from " + urlObject.(string) + "\nThe url field in json is " + root["url"].(string),
					}
				}
			} else {
				return nil, &FetchError{
					NormalError: "",
					BotError:    "The type of this post is rich:video but it does not contains `domain`",
				}
			}
		default:
			return nil, &FetchError{
				NormalError: "",
				BotError:    "This post type is not supported: " + hint.(string),
			}
		}
	} else { // text or gallery
		if gData, ok := root["gallery_data"]; ok { // gallery
			if data, ok := root["media_metadata"]; ok {
				return getGalleryData(data.(map[string]interface{}), gData.(map[string]interface{})["items"].([]interface{})), nil
			}
		}
		// Text
		return FetchResultText{
			Title: title,
			Text:  strings.ReplaceAll(html.UnescapeString(root["selftext"].(string)), "&#x200B;", ""),
		}, nil
	}
}

func getPostID(postUrl string) (postID string, isComment bool, err *FetchError) {
	var u *url.URL = nil
	// Check all lines for links. In new reddit update, sharing via Telegram adds the post title at its first
	lines := strings.Split(postUrl, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
			line = "https://" + line
		}
		u, _ = url.Parse(line)
		if u == nil {
			continue
		}
		if u.Host == "redd.it" {
			if len(u.Path) <= 1 {
				continue
			}
			p := u.Path[1:] // remove the first /
			if strings.Contains(p, "/") {
				continue
			}
			// redd.it links are never comments
			return p, false, nil
		}
		if u.Host == "v.redd.it" {
			followedUrl, err := util.FollowRedirect(line)
			if err != nil {
				continue
			}
			u, _ = url.Parse(followedUrl)
		}
		if u.Host == "www.reddit.com" || u.Host == "reddit.com" || u.Host == "old.reddit.com" {
			postUrl = line
			break
		}
		u = nil // this is for last loop. If u is nil after that final loop, it means that there is no reddit url in text
	}
	if u == nil {
		err = &FetchError{
			NormalError: "",
			BotError:    "Cannot parse reddit the url. Does your text contain a reddit url?",
		}
		return
	}
	split := strings.Split(u.Path, "/")
	if len(split) == 2 { // www.reddit.com/x
		return split[1], false, nil
	}
	if len(split) < 5 {
		err = &FetchError{
			NormalError: "",
			BotError:    "Cannot parse reddit the url. Does your text contain a reddit url?",
		}
		return
	}
	if len(split) >= 7 && split[6] != "" {
		return split[6], true, nil
	}
	return split[4], false, nil
}

// getGalleryData extracts the gallery data from gallery json
func getGalleryData(files map[string]interface{}, galleryDataItems []interface{}) FetchResultAlbum {
	album := make([]FetchResultAlbumEntry, 0, len(galleryDataItems))
	for _, data := range galleryDataItems {
		galleryRoot := files[data.(map[string]interface{})["media_id"].(string)]
		// Extract the url
		image := galleryRoot.(map[string]interface{})
		if image["status"].(string) != "valid" { // I have not encountered anything else except valid so far
			continue
		}
		dataType := image["e"].(string)
		// Check the type
		switch dataType {
		case "Image":
			link := image["s"].(map[string]interface{})["u"].(string)
			// For some reasons, I have to remove all "amp;" from the url in order to make this work
			link = strings.ReplaceAll(link, "amp;", "")
			// Get the caption
			var caption string
			if c, ok := data.(map[string]interface{})["caption"]; ok {
				caption = c.(string)
			}
			if c, ok := data.(map[string]interface{})["outbound_url"]; ok {
				caption += "\n" + c.(string)
			}
			// Append to the album
			album = append(album, FetchResultAlbumEntry{
				Link:    link,
				Caption: caption,
				Type:    FetchResultMediaTypePhoto,
			})
		case "AnimatedImage":
			link := image["s"].(map[string]interface{})["mp4"].(string)
			link = strings.ReplaceAll(link, "amp;", "")
			// Get the caption
			var caption string
			if c, ok := data.(map[string]interface{})["caption"]; ok {
				caption = c.(string)
			}
			if c, ok := data.(map[string]interface{})["outbound_url"]; ok {
				caption += "\n" + c.(string)
			}
			// Append to the album
			album = append(album, FetchResultAlbumEntry{
				Link:    link,
				Caption: caption,
				Type:    FetchResultMediaTypeGif,
			})
		case "RedditVideo":
			id := image["id"].(string)
			w := image["x"].(float64)
			h := image["y"].(float64)
			// Get the quality
			res := "96"
			if w >= 1920 && h >= 1080 { // is this the best way?
				res = "1080"
			} else if w >= 1280 && h >= 720 {
				res = "720"
			} else if w >= 854 && h >= 480 {
				res = "480"
			} else if w >= 640 && h >= 360 {
				res = "360"
			} else if w >= 426 && h >= 240 {
				res = "240"
			}
			link := "https://v.redd.it/" + id + "/DASH_" + res + ".mp4"
			// Get the caption
			var caption string
			if c, ok := data.(map[string]interface{})["caption"]; ok {
				caption = c.(string)
			}
			// Append to the album
			album = append(album, FetchResultAlbumEntry{
				Link:    link,
				Caption: caption,
				Type:    FetchResultMediaTypeVideo,
			})
		default:
			log.Println("Unknown type in send gallery:", dataType)
		}
	}
	return FetchResultAlbum{album}
}

// extractPhotoGifQualities creates an array of FetchResultMediaEntry which are the qualities
// of the photo or gif and their links
func extractPhotoGifQualities(data map[string]interface{}) []FetchResultMediaEntry {
	resolutions := data["resolutions"].([]interface{})
	result := make([]FetchResultMediaEntry, 0, 1+len(resolutions))
	// Include source image at last to keep the increasing quality
	// Just a note for myself: This can be different from the one in resolutions
	{
		u, w, h := extractLinkAndRes(data["source"])
		result = append(result, FetchResultMediaEntry{
			Link:    u,
			Quality: w + "×" + h,
		})
	}
	// Now get all other thumbs
	for i := len(resolutions) - 1; i >= 0; i-- {
		u, w, h := extractLinkAndRes(resolutions[i])
		if i == len(resolutions)-1 { // In first case, the sizes can be same. Example: https://www.reddit.com/r/dankmemes/comments/vqphiy/more_than_bargain_for/
			if w+"×"+h == result[0].Quality {
				continue
			}
		}
		result = append(result, FetchResultMediaEntry{
			Link:    u,
			Quality: w + "×" + h,
		})
	}
	return result
}

// extractVideoQualities gets all possible qualities from a main video url
// It uses the quality in video URL to determine all lower qualities which are available
func extractVideoQualities(vidUrl string) []FetchResultMediaEntry {
	result := make([]FetchResultMediaEntry, 0, len(qualities)+1) // +1 for audio
	// Get max res
	u, _ := url.Parse(vidUrl)
	u.RawQuery = ""
	res := path.Base(u.Path)[strings.LastIndex(path.Base(u.Path), "_")+1:] // the max res of video
	base := u.String()[:strings.LastIndex(u.String(), "/")]                // base url is this: https://v.redd.it/3lelz0i6crx41
	newFormat := strings.Contains(res, ".mp4")                             // this is new reddit format. The filenames are like DASH_480.mp4
	if newFormat {
		res = res[:strings.Index(res, ".")] // remove format to get the max quality
	}
	// List all the qualities
	startAdd := false
	for _, quality := range qualities {
		if quality == res || startAdd {
			startAdd = true
			link := base + "/DASH_" + quality
			if newFormat {
				link += ".mp4"
			}
			result = append(result, FetchResultMediaEntry{
				Link:    link,
				Quality: quality + "p",
			})
		}
	}
	// If the result is still empty, just append the main url to it
	if len(result) == 0 {
		result = append(result, FetchResultMediaEntry{
			Link:    vidUrl,
			Quality: "source",
		})
	}
	// Check for audio
	if audioLink, hasAudio := HasAudio(result[0].Link); hasAudio {
		result = append(result, FetchResultMediaEntry{
			Link:    audioLink,
			Quality: DownloadAudioQuality,
		})
	}
	return result
}

// extractLinkAndRes extracts the data from "source":{ "url":"https://preview.redd.it/utx00pfe4cp41.jpg?auto=webp&amp;s=de4ff82478b12df6369b8d7eeca3894f094e87e1", "width":624, "height":960 } stuff
// First return values are url, width, height
func extractLinkAndRes(data interface{}) (u string, width string, height string) {
	kv := data.(map[string]interface{})
	return html.UnescapeString(kv["url"].(string)), strconv.Itoa(int(kv["width"].(float64))), strconv.Itoa(int(kv["height"].(float64)))
}
