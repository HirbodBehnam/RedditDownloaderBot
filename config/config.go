package config

import (
	"net/http"
	"time"
)

// Version is the version of this program :|
const Version = "1.8.1"

// GlobalHttpClient is an http client which all request must be done through it
var GlobalHttpClient = &http.Client{Timeout: time.Second * 10}
