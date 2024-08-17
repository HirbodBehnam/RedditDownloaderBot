package common

import (
	"net/http"
	"time"
)

// Version is the version of this program :|
const Version = "4.0.0-beta"

// GlobalHttpClient is a http client which all request must be done through it
var GlobalHttpClient = http.Client{
	Timeout: time.Second * 10,
}
