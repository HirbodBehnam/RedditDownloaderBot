package oauth

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

// UserAgent of requests
const UserAgent = "TelegramBot:Reddit-Downloader-Bot:" + config.Version + " (by /u/HirbodBehnam)"

// PostApiPoint is the endpoint format which we should get info about posts
const PostApiPoint = "https://api.reddit.com/api/info/?id=t3_"

// CommentApiPoint is the endpoint format which we should get info about comments
const CommentApiPoint = "https://api.reddit.com/api/info/?id=t1_"

const encodedGrantType = "grant_type=client_credentials&duration=permanent"

var RateLimitError = errors.New("rate limit reached")

// RedditOauth is an struct which can talk to reddit endpoints
type RedditOauth struct {
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
func NewRedditOauth(clientId, clientSecret string) (*RedditOauth, error) {
	redditOauth := &RedditOauth{
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
func (r *RedditOauth) tokenRefresh(nextRefresh time.Duration) {
	for {
		time.Sleep(nextRefresh - time.Minute)
		err, nextRefreshCandidate := r.refreshTokenFunction()
		if err != nil {
			log.Printf("cannot refresh token: %s", err.Error())
			nextRefresh = time.Minute
		} else {
			nextRefresh = nextRefreshCandidate
		}
	}
}

// createToken creates an RedditOauth.authorizationHeader and returns when will the next token expire
func (r *RedditOauth) createToken() (error, time.Duration) {
	// Build the request
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(encodedGrantType))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(r.clientId, r.clientSecret)
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
	r.authorizationHeader = "bearer: " + body.AccessToken
	r.refreshToken = body.RefreshToken
	return nil, time.Duration(body.ExpiresIn) * time.Second
}

// refreshTokenFunction refreshes the token from reddit servers
func (r *RedditOauth) refreshTokenFunction() (error, time.Duration) {
	// Build the request
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", r.refreshToken)
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(r.clientId, r.clientSecret)
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
func (r *RedditOauth) GetComment(id string) (map[string]interface{}, error) {
	return r.doGetJsonRequest(CommentApiPoint + id)
}

// GetPost gets the info about a post from reddit
func (r *RedditOauth) GetPost(id string) (map[string]interface{}, error) {
	return r.doGetJsonRequest(PostApiPoint + id)
}

func (r *RedditOauth) doGetJsonRequest(Url string) (map[string]interface{}, error) {
	// Check rate limit
	r.rateLimitMutex.RLock()
	if time.Now().Before(r.rateLimit) {
		r.rateLimitMutex.RUnlock()
		return nil, RateLimitError
	}
	r.rateLimitMutex.RUnlock()
	// Build the request
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Authorization", r.authorizationHeader)
	// Do the request
	resp, err := config.GlobalHttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// Check the rate limit
	if rateLimit, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining")); err == nil && rateLimit == 0 {
		freedom, _ := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		r.rateLimitMutex.Lock()
		r.rateLimit = time.Now().Add(time.Duration(freedom) * time.Second)
		r.rateLimitMutex.Unlock()
	}
	// Read the body
	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	_ = resp.Body.Close()
	return responseBody, nil
}
