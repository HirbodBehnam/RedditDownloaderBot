package reddit

import (
	"encoding/json"
	"errors"
	"github.com/HirbodBehnam/RedditDownloaderBot/config"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// userAgent of requests
const userAgent = "TelegramBot:Reddit-Downloader-Bot:" + config.Version + " (by /u/HirbodBehnam)"

// postApiPoint is the endpoint format which we should get info about posts
const postApiPoint = "https://api.reddit.com/api/info/?id=t3_"

// commentApiPoint is the endpoint format which we should get info about comments
const commentApiPoint = "https://api.reddit.com/api/info/?id=t1_"

const encodedGrantType = "grant_type=client_credentials&duration=permanent"

var RateLimitError = errors.New("rate limit reached")

// Oauth is a struct which can talk to reddit endpoints
type Oauth struct {
	// When we should make the next request
	rateLimit time.Time
	// The client id of this app
	clientId string
	// The client secret of this app
	clientSecret string
	// The authorization header we should send to each request
	authorizationHeader string
	// The refresh token to refresh the authorizationHeader
	refreshToken string
	// The mutex for rate limit
	rateLimitMutex sync.RWMutex
}

// tokenRequestResponse is the result of https://www.reddit.com/api/v1/access_token endpoint
type tokenRequestResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// NewRedditOauth returns a new RedditOauth to be used to get posts from reddit
func NewRedditOauth(clientId, clientSecret string) (*Oauth, error) {
	redditOauth := &Oauth{
		clientId:     clientId,
		clientSecret: clientSecret,
	}
	err, nextRefresh := redditOauth.createToken()
	if err != nil {
		return nil, err
	}
	go redditOauth.tokenRefresh(nextRefresh)
	return redditOauth, nil
}

// tokenRefresh refreshes the
func (o *Oauth) tokenRefresh(nextRefresh time.Duration) {
	for {
		time.Sleep(nextRefresh - time.Minute)
		// Check rate limit
		o.rateLimitMutex.RLock()
		freedom := o.rateLimit
		o.rateLimitMutex.RUnlock()
		if time.Now().Before(freedom) {
			time.Sleep(time.Now().Sub(freedom))
			nextRefresh = 0 // do not wait in line "time.Sleep(nextRefresh - time.Minute)"
			continue
		}
		// Request the token
		err, nextRefreshCandidate := o.refreshTokenFunction()
		if err != nil {
			log.Printf("cannot refresh token: %s", err.Error())
			nextRefresh = 2 * time.Minute
		} else {
			nextRefresh = nextRefreshCandidate
		}
	}
}

// createToken creates an RedditOauth.authorizationHeader and returns when will the next token expire
func (o *Oauth) createToken() (error, time.Duration) {
	// Build the request
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(encodedGrantType))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(o.clientId, o.clientSecret)
	// Send the request
	resp, err := config.GlobalHttpClient.Do(req)
	if err != nil {
		return err, 0
	}
	// Parse the response
	var body tokenRequestResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	_ = resp.Body.Close()
	if err != nil {
		return err, 0
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("status code is not 200. It is " + resp.Status), 0
	}
	// Set the data
	o.authorizationHeader = "bearer: " + body.AccessToken
	o.refreshToken = body.RefreshToken
	return nil, time.Duration(body.ExpiresIn) * time.Second
}

// refreshTokenFunction refreshes the token from reddit servers
func (o *Oauth) refreshTokenFunction() (error, time.Duration) {
	// Build the request
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", o.refreshToken)
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(o.clientId, o.clientSecret)
	// Send the request
	resp, err := config.GlobalHttpClient.Do(req)
	if err != nil {
		return err, 0
	}
	// Parse the response
	var body tokenRequestResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	_ = resp.Body.Close()
	if err != nil {
		return err, 0
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("status code is not 200. It is " + resp.Status), 0
	}
	return nil, time.Duration(body.ExpiresIn) * time.Second
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
	o.rateLimitMutex.RLock()
	if time.Now().Before(o.rateLimit) {
		o.rateLimitMutex.RUnlock()
		return nil, RateLimitError
	}
	o.rateLimitMutex.RUnlock()
	// Build the request
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", o.authorizationHeader)
	// Do the request
	resp, err := config.GlobalHttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// Check the rate limit
	if rateLimit, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining")); err == nil && rateLimit == 0 {
		freedom, _ := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		o.rateLimitMutex.Lock()
		o.rateLimit = time.Now().Add(time.Duration(freedom) * time.Second)
		o.rateLimitMutex.Unlock()
	}
	// Read the body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	_ = resp.Body.Close()
	return responseBody, err
}
