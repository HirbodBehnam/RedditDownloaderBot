package reddit

import (
	"RedditDownloaderBot/pkg/common"
	"bytes"
	"encoding/json"
	"github.com/go-faster/errors"
	"io"
	"log"
	"net/http"
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
	nextRefresh, err := redditOauth.createToken()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create initial token")
	}
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
	// Parse the response
	if resp.StatusCode != http.StatusOK {
		buffer := new(bytes.Buffer)
		io.CopyN(buffer, resp.Body, 100) // 100 chars is ok right?
		resp.Body.Close()
		return 0, errors.Wrapf(err, "status code is not 200. It is %s. Body starts with: %s", resp.Status, buffer.String())
	}
	var body tokenRequestResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
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
	// Check the rate limit
	if rateLimit, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining")); err == nil && rateLimit == 0 {
		freedom, _ := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		atomic.StoreInt64(&o.rateLimitFreedom, time.Now().Unix()+int64(freedom))
		resp.Body.Close()
		return nil, RateLimitErr
	}
	// Read the body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	resp.Body.Close()
	return responseBody, err
}
