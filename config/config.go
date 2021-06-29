package config

import (
	"net/http"
	"time"
)

// Version is the version of this program :|
const Version = "1.7.0"

// GlobalHttpClient is an http client which all request must be done through it
var GlobalHttpClient = &http.Client{Timeout: time.Second * 10}
