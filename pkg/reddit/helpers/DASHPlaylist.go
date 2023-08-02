package helpers

import (
	"RedditDownloaderBot/pkg/common"
	"encoding/xml"
	"github.com/go-faster/errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// numberRegex will only match numbers in a string
var numberRegex = regexp.MustCompile("(\\d+)")

// DashPlaylistXML is the root of
type DashPlaylistXML struct {
	XMLName xml.Name `xml:"MPD"`
	Period  struct {
		XMLName    xml.Name                     `xml:"Period"`
		MediaTypes []DashPlaylistApplicationSet `xml:"AdaptationSet"`
	}
}

// DashPlaylistApplicationSet represents the audio or video urls of current video
type DashPlaylistApplicationSet struct {
	XMLName     xml.Name                     `xml:"AdaptationSet"`
	ContentType string                       `xml:"contentType,attr"`
	Qualities   []DashPlaylistRepresentation `xml:"Representation"`
}

// DashPlaylistRepresentation represents the link to each media type
type DashPlaylistRepresentation struct {
	XMLName xml.Name `xml:"Representation"`
	BaseURL string   `xml:"BaseURL"`
	ID      string   `xml:"id,attr"`
}

// AvailableVideo represents a single available video quality for a video on reddit
type AvailableVideo string

// Quality gets the quality of a video
func (v AvailableVideo) Quality() string {
	numbers := numberRegex.FindStringSubmatch(string(v))
	if len(numbers) < 2 {
		return "NA"
	}
	return numbers[1]
}

// AvailableAudio represents a single available audio quality for a video on reddit
type AvailableAudio string

// AvailableMedia represents the available medias for a video on reddit
type AvailableMedia struct {
	AvailableVideos []AvailableVideo
	AvailableAudios []AvailableAudio
}

// parseDashPlaylist will parse the DashPlaylist file from Reddit
func parseDashPlaylist(r io.Reader) (AvailableMedia, error) {
	// Parse XML
	var parsedXML DashPlaylistXML
	err := xml.NewDecoder(r).Decode(&parsedXML)
	if err != nil {
		return AvailableMedia{}, errors.Wrap(err, "cannot parse XML")
	}
	// Convert to result
	var result AvailableMedia
	for _, media := range parsedXML.Period.MediaTypes {
		switch media.ContentType {
		case "video":
			result.AvailableVideos = make([]AvailableVideo, len(media.Qualities))
			for i, video := range media.Qualities {
				result.AvailableVideos[i] = AvailableVideo(video.BaseURL)
			}
		case "audio":
			result.AvailableAudios = make([]AvailableAudio, len(media.Qualities))
			for i, audio := range media.Qualities {
				result.AvailableAudios[i] = AvailableAudio(audio.BaseURL)
			}
		case "": // Used in very old videos. See tests
			for _, m := range media.Qualities {
				if strings.HasPrefix(m.ID, "VIDEO") {
					result.AvailableVideos = append(result.AvailableVideos, AvailableVideo(m.BaseURL))
				} else if strings.HasPrefix(m.ID, "AUDIO") {
					result.AvailableAudios = append(result.AvailableAudios, AvailableAudio(m.BaseURL))
				}
			}
		}
	}
	return result, nil
}

// ParseDashPlaylistFromID will parse the dash playlist file for a video ID
func ParseDashPlaylistFromID(vidID string) (AvailableMedia, error) {
	// Check if vidID is empty
	if vidID == "" {
		return AvailableMedia{}, errors.New("empty vidID")
	}
	// Request the dash file
	resp, err := common.GlobalHttpClient.Get("https://v.redd.it/" + vidID + "/DASHPlaylist.mpd")
	if err != nil {
		return AvailableMedia{}, errors.Wrap(err, "cannot get url")
	}
	defer resp.Body.Close()
	// Check status
	if resp.StatusCode != http.StatusOK {
		return AvailableMedia{}, errors.Errorf("status code of page is not OK: it is %d (%s)", resp.StatusCode, resp.Status)
	}
	// Parse body
	return parseDashPlaylist(resp.Body)
}
