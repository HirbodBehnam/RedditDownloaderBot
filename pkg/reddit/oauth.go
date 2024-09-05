package reddit

import (
	"RedditDownloaderBot/pkg/common"
	"RedditDownloaderBot/pkg/util"
	"encoding/json"
	"github.com/go-faster/errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// userAgent of requests
const userAgent = "TelegramBot:Reddit-Downloader-Bot:" + common.Version + " (by /u/HirbodBehnam)"

// postApiPoint is the endpoint format which we should get info about posts
const postApiPoint = "https://api.reddit.com/api/info/?id=t3_"

// commentApiPoint is the endpoint format which we should get info about comments
const commentApiPoint = "https://api.reddit.com/api/info/?id=t1_"

const encodedGrantType = "grant_type=client_credentials&duration=permanent"

// RateLimitErr is returned when we reach the rate limit of Reddit
var RateLimitErr = errors.New("rate limit reached")

// Oauth is a struct which can talk to reddit endpoints
type Oauth struct {
	// The client id of this app
	clientId string
	// The client secret of this app
	clientSecret string
	// The authorization header we should send to each request
	authorizationHeader string
	// When we should make the next request in unix epoch
	rateLimitFreedom int64
	// The HTTP client for Imgur downloads (might use proxy)
	imgurHTTPClient *http.Client
}

// tokenRequestResponse is the result of https://www.reddit.com/api/v1/access_token endpoint
type tokenRequestResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// NewRedditOauth returns a new RedditOauth to be used to get posts from reddit
func NewRedditOauth(clientId, clientSecret string) (*Oauth, error) {
	redditOauth := &Oauth{
		clientId:     clientId,
		clientSecret: clientSecret,
	}
	// Get the token
	nextRefresh, err := redditOauth.createToken()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create initial token")
	}
	// The proxy to download the Imgur media through it. Imgur sometimes
	// blocks some IP addresses like Hetzner for example. It's interesting because
	// even with authorization it does not work. Even accessing through the browser
	// it does not work either. So, someone might use a proxy (like Cloudflare Warp)
	// to bypass this restriction.
	if imgurProxy := os.Getenv("IMGUR_PROXY"); imgurProxy != "" {
		if imgurProxyUrl, _ := url.Parse(imgurProxy); imgurProxyUrl != nil {
			redditOauth.imgurHTTPClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(imgurProxyUrl)}}
		}
	}
	// Refresh the token once in a while
	go redditOauth.tokenRefresh(nextRefresh)
	return redditOauth, nil
}

// tokenRefresh refreshes the
func (o *Oauth) tokenRefresh(nextRefresh time.Duration) {
	for {
		time.Sleep(nextRefresh - time.Minute)
		// Check rate limit
		freedom := atomic.LoadInt64(&o.rateLimitFreedom)
		if time.Now().Unix() < freedom {
			time.Sleep(time.Now().Sub(time.Unix(freedom, 0)))
			nextRefresh = 0 // do not wait in line "time.Sleep(nextRefresh - time.Minute)"
			continue
		}
		// Request the token
		nextRefreshCandidate, err := o.createToken()
		if err != nil {
			log.Printf("cannot re-generate token: %s", err.Error())
			nextRefresh = 2 * time.Minute
		} else {
			nextRefresh = nextRefreshCandidate
		}
	}
}

// createToken creates an RedditOauth.authorizationHeader and returns when will the next token expire
func (o *Oauth) createToken() (time.Duration, error) {
	// Build the request
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(encodedGrantType))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(o.clientId, o.clientSecret)
	// Send the request
	resp, err := common.GlobalHttpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "cannot do the request")
	}
	defer resp.Body.Close()
	// Parse the response
	if resp.StatusCode != http.StatusOK {
		buffer := make([]byte, 100) // 100 chars is ok right?
		n, _ := resp.Body.Read(buffer)
		return 0, errors.Wrapf(err, "status code is not 200. It is %s. Body starts with: %s", resp.Status, string(buffer[:n]))
	}
	var body tokenRequestResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return 0, errors.Wrap(err, "cannot parse response")
	}
	// Set the data
	o.authorizationHeader = "bearer: " + body.AccessToken
	return time.Duration(body.ExpiresIn) * time.Second, nil
}

// GetComment gets the info about a comment from reddit
func (o *Oauth) GetComment(id string) (map[string]interface{}, error) {
	return o.doGetJsonRequest(commentApiPoint + id)
}

// GetPost gets the info about a post from reddit
func (o *Oauth) GetPost(id string) (map[string]interface{}, error) {
	return o.doGetJsonRequest(postApiPoint + id)
}

// FollowRedirect follows a page's redirect and returns the final URL
func (o *Oauth) FollowRedirect(u string) (string, error) {
	resp, err := o.head(u)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return resp.Request.URL.String(), nil
}

func (o *Oauth) doGetJsonRequest(Url string) (map[string]interface{}, error) {
	// Check rate limit
	if time.Now().Unix() < atomic.LoadInt64(&o.rateLimitFreedom) {
		return nil, RateLimitErr
	}
	// Build the request
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create request")
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", o.authorizationHeader)
	// Do the request
	resp, err := common.GlobalHttpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot do the request")
	}
	defer resp.Body.Close()
	// Check the rate limit
	if rateLimit, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining")); err == nil && rateLimit == 0 {
		freedom, _ := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		atomic.StoreInt64(&o.rateLimitFreedom, time.Now().Unix()+int64(freedom))
		return nil, RateLimitErr
	}
	// Read the body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody, err
}

// head will do a head request. Useful to check redirects
func (o *Oauth) head(Url string) (*http.Response, error) {
	// Check rate limit
	if time.Now().Unix() < atomic.LoadInt64(&o.rateLimitFreedom) {
		return nil, RateLimitErr
	}
	// Build the request
	req, err := http.NewRequest("HEAD", Url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create request")
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", o.authorizationHeader)
	return common.GlobalHttpClient.Do(req)
}

// downloadToFile downloads a link to a file
// It also checks where the file is too big to be uploaded to Telegram or not
// If the file is too big, it returns FileTooBigError
func (o *Oauth) downloadToFile(link string, f *os.File) error {
	// Check rate limit
	if time.Now().Unix() < atomic.LoadInt64(&o.rateLimitFreedom) {
		return RateLimitErr
	}
	// Build the request
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return errors.Wrap(err, "cannot create request")
	}
	req.Header.Set("User-Agent", userAgent)
	// Check Imgur and proxy
	var client *http.Client
	if o.imgurHTTPClient != nil && util.IsImgurLink(link) {
		client = o.imgurHTTPClient
	} else {
		client = &common.GlobalHttpClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return errors.New("non 2xx status: " + resp.Status)
	}
	if resp.ContentLength == -1 {
		return errors.New("Unknown length")
	}
	if resp.ContentLength > maxDownloadSize {
		return FileTooBigError
	}
	_, err = io.Copy(f, resp.Body)
	return err
}
