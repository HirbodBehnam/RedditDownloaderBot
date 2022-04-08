package config

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Version is the version of this program :|
const Version = "2.3.0"

// GlobalHttpClient is a http client which all request must be done through it
var GlobalHttpClient = &http.Client{
	Timeout: time.Second * 10,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			//CipherSuites: []uint16{tls.TLS_AES_128_GCM_SHA256},
		},
	},
}
