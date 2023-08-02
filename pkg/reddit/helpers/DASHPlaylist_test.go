package helpers

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestVideoQuality(t *testing.T) {
	tests := []struct {
		Name     string
		Data     AvailableVideo
		Expected string
	}{
		{
			Name:     "new",
			Data:     "DASH_220.mp4",
			Expected: "220",
		},
		{
			Name:     "old",
			Data:     "DASH_1080",
			Expected: "1080",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.Expected, test.Data.Quality())
		})
	}
}

func TestStructParser(t *testing.T) {
	tests := []struct {
		Name     string
		Data     string
		Expected AvailableMedia
		Error    error
	}{
		{ // From https://v.redd.it/pw4v2kzgg0fb1/DASHPlaylist.mpd
			Name: "audio_video_new",
			Data: "<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" mediaPresentationDuration=\"PT13S\" minBufferTime=\"PT4S\" profiles=\"urn:mpeg:dash:profile:isoff-on-demand:2011\" type=\"static\" xsi:schemaLocation=\"urn:mpeg:dash:schema:mpd:2011 DASH-MPD.xsd\">\n  <Period duration=\"PT13S\" id=\"0\">\n    <AdaptationSet contentType=\"video\" id=\"0\" maxFrameRate=\"15360/512\" maxHeight=\"480\" maxWidth=\"582\" par=\"23:19\" sar=\"1:1\" segmentAlignment=\"true\" startWithSAP=\"1\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n      <Representation bandwidth=\"188176\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"220\" id=\"1\" mimeType=\"video/mp4\" width=\"266\">\n        <BaseURL>DASH_220.mp4</BaseURL>\n        <SegmentBase indexRange=\"845-912\" timescale=\"15360\">\n          <Initialization range=\"0-844\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"245208\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"270\" id=\"2\" mimeType=\"video/mp4\" width=\"328\">\n        <BaseURL>DASH_270.mp4</BaseURL>\n        <SegmentBase indexRange=\"847-914\" timescale=\"15360\">\n          <Initialization range=\"0-846\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"360766\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"360\" id=\"3\" mimeType=\"video/mp4\" width=\"436\">\n        <BaseURL>DASH_360.mp4</BaseURL>\n        <SegmentBase indexRange=\"847-914\" timescale=\"15360\">\n          <Initialization range=\"0-846\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"528108\" codecs=\"avc1.4d401f\" frameRate=\"15360/512\" height=\"480\" id=\"4\" mimeType=\"video/mp4\" width=\"582\">\n        <BaseURL>DASH_480.mp4</BaseURL>\n        <SegmentBase indexRange=\"847-914\" timescale=\"15360\">\n          <Initialization range=\"0-846\" />\n        </SegmentBase>\n      </Representation>\n    </AdaptationSet>\n    <AdaptationSet contentType=\"audio\" id=\"1\" segmentAlignment=\"true\" startWithSAP=\"1\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n      <Representation audioSamplingRate=\"48000\" bandwidth=\"67281\" codecs=\"mp4a.40.2\" id=\"5\" mimeType=\"audio/mp4\">\n        <AudioChannelConfiguration schemeIdUri=\"urn:mpeg:dash:23003:3:audio_channel_configuration:2011\" value=\"2\" />\n        <BaseURL>DASH_AUDIO_64.mp4</BaseURL>\n        <SegmentBase indexRange=\"820-887\" timescale=\"48000\">\n          <Initialization range=\"0-819\" />\n        </SegmentBase>\n      </Representation>\n      <Representation audioSamplingRate=\"48000\" bandwidth=\"134610\" codecs=\"mp4a.40.2\" id=\"6\" mimeType=\"audio/mp4\">\n        <AudioChannelConfiguration schemeIdUri=\"urn:mpeg:dash:23003:3:audio_channel_configuration:2011\" value=\"2\" />\n        <BaseURL>DASH_AUDIO_128.mp4</BaseURL>\n        <SegmentBase indexRange=\"820-887\" timescale=\"48000\">\n          <Initialization range=\"0-819\" />\n        </SegmentBase>\n      </Representation>\n    </AdaptationSet>\n  </Period>\n</MPD>",
			Expected: AvailableMedia{
				AvailableVideos: []AvailableVideo{"DASH_220.mp4", "DASH_270.mp4", "DASH_360.mp4", "DASH_480.mp4"},
				AvailableAudios: []AvailableAudio{"DASH_AUDIO_64.mp4", "DASH_AUDIO_128.mp4"},
			},
		},
		{ // From https://v.redd.it/dbelx9ulpacb1/DASHPlaylist.mpd
			Name: "audio_video_old",
			Data: "<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" mediaPresentationDuration=\"PT9S\" minBufferTime=\"PT4S\" profiles=\"urn:mpeg:dash:profile:isoff-on-demand:2011\" type=\"static\" xsi:schemaLocation=\"urn:mpeg:dash:schema:mpd:2011 DASH-MPD.xsd\">\n  <Period duration=\"PT9S\" id=\"0\">\n    <AdaptationSet contentType=\"video\" id=\"0\" maxFrameRate=\"15360/512\" maxHeight=\"1080\" maxWidth=\"608\" par=\"9:16\" sar=\"1:1\" segmentAlignment=\"true\" startWithSAP=\"1\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n      <Representation bandwidth=\"89938\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"220\" id=\"1\" mimeType=\"video/mp4\" width=\"124\">\n        <BaseURL>DASH_220.mp4</BaseURL>\n        <SegmentBase indexRange=\"825-880\" timescale=\"15360\">\n          <Initialization range=\"0-824\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"96794\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"240\" id=\"2\" mimeType=\"video/mp4\" width=\"136\">\n        <BaseURL>DASH_240.mp4</BaseURL>\n        <SegmentBase indexRange=\"825-880\" timescale=\"15360\">\n          <Initialization range=\"0-824\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"159264\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"360\" id=\"3\" mimeType=\"video/mp4\" width=\"202\">\n        <BaseURL>DASH_360.mp4</BaseURL>\n        <SegmentBase indexRange=\"826-881\" timescale=\"15360\">\n          <Initialization range=\"0-825\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"225456\" codecs=\"avc1.4d401f\" frameRate=\"15360/512\" height=\"480\" id=\"4\" mimeType=\"video/mp4\" width=\"270\">\n        <BaseURL>DASH_480.mp4</BaseURL>\n        <SegmentBase indexRange=\"824-879\" timescale=\"15360\">\n          <Initialization range=\"0-823\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"366750\" codecs=\"avc1.4d401f\" frameRate=\"15360/512\" height=\"720\" id=\"5\" mimeType=\"video/mp4\" width=\"406\">\n        <BaseURL>DASH_720.mp4</BaseURL>\n        <SegmentBase indexRange=\"826-881\" timescale=\"15360\">\n          <Initialization range=\"0-825\" />\n        </SegmentBase>\n      </Representation>\n      </AdaptationSet>\n    <AdaptationSet contentType=\"audio\" id=\"1\" segmentAlignment=\"true\" startWithSAP=\"1\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n      <Representation audioSamplingRate=\"48000\" bandwidth=\"135442\" codecs=\"mp4a.40.2\" id=\"7\" mimeType=\"audio/mp4\">\n        <AudioChannelConfiguration schemeIdUri=\"urn:mpeg:dash:23003:3:audio_channel_configuration:2011\" value=\"2\" />\n        <BaseURL>DASH_audio.mp4</BaseURL>\n        <SegmentBase indexRange=\"820-875\" timescale=\"48000\">\n          <Initialization range=\"0-819\" />\n        </SegmentBase>\n      </Representation>\n    </AdaptationSet>\n  </Period>\n</MPD>",
			Expected: AvailableMedia{
				AvailableVideos: []AvailableVideo{"DASH_220.mp4", "DASH_240.mp4", "DASH_360.mp4", "DASH_480.mp4", "DASH_720.mp4"},
				AvailableAudios: []AvailableAudio{"DASH_audio.mp4"},
			},
		},
		{ // From https://v.redd.it/jzsvg42m78eb1/DASHPlaylist.mpd
			Name: "video_old",
			Data: "<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" mediaPresentationDuration=\"PT2M5S\" minBufferTime=\"PT4S\" profiles=\"urn:mpeg:dash:profile:isoff-on-demand:2011\" type=\"static\" xsi:schemaLocation=\"urn:mpeg:dash:schema:mpd:2011 DASH-MPD.xsd\">\n  <Period duration=\"PT2M5S\" id=\"0\">\n    <AdaptationSet contentType=\"video\" id=\"0\" maxFrameRate=\"15360/512\" maxHeight=\"1080\" maxWidth=\"1920\" par=\"16:9\" sar=\"1:1\" segmentAlignment=\"true\" startWithSAP=\"1\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n      <Representation bandwidth=\"263590\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"220\" id=\"1\" mimeType=\"video/mp4\" width=\"392\">\n        <BaseURL>DASH_220.mp4</BaseURL>\n        <SegmentBase indexRange=\"845-1248\" timescale=\"15360\">\n          <Initialization range=\"0-844\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"707294\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"240\" id=\"2\" mimeType=\"video/mp4\" width=\"426\">\n        <BaseURL>DASH_240.mp4</BaseURL>\n        <SegmentBase indexRange=\"845-1248\" timescale=\"15360\">\n          <Initialization range=\"0-844\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"916668\" codecs=\"avc1.4d401e\" frameRate=\"15360/512\" height=\"360\" id=\"3\" mimeType=\"video/mp4\" width=\"640\">\n        <BaseURL>DASH_360.mp4</BaseURL>\n        <SegmentBase indexRange=\"827-1230\" timescale=\"15360\">\n          <Initialization range=\"0-826\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"1398892\" codecs=\"avc1.4d401f\" frameRate=\"15360/512\" height=\"480\" id=\"4\" mimeType=\"video/mp4\" width=\"854\">\n        <BaseURL>DASH_480.mp4</BaseURL>\n        <SegmentBase indexRange=\"847-1250\" timescale=\"15360\">\n          <Initialization range=\"0-846\" />\n        </SegmentBase>\n      </Representation>\n      <Representation bandwidth=\"2790330\" codecs=\"avc1.4d401f\" frameRate=\"15360/512\" height=\"720\" id=\"5\" mimeType=\"video/mp4\" width=\"1280\">\n        <BaseURL>DASH_720.mp4</BaseURL>\n        <SegmentBase indexRange=\"825-1228\" timescale=\"15360\">\n          <Initialization range=\"0-824\" />\n        </SegmentBase>\n      </Representation>\n      </AdaptationSet>\n  </Period>\n</MPD>",
			Expected: AvailableMedia{
				AvailableVideos: []AvailableVideo{"DASH_220.mp4", "DASH_240.mp4", "DASH_360.mp4", "DASH_480.mp4", "DASH_720.mp4"},
				AvailableAudios: nil,
			},
		},
		{ // From https://v.redd.it/l81cm9bcwtp41/DASHPlaylist.mpd
			Name: "audio_video_very_old",
			Data: "<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" mediaPresentationDuration=\"PT28.5S\" minBufferTime=\"PT1.500S\" profiles=\"urn:mpeg:dash:profile:isoff-on-demand:2011\" type=\"static\">\n    <Period duration=\"PT28.5S\">\n        <AdaptationSet segmentAlignment=\"true\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n            <Representation bandwidth=\"2264565\" codecs=\"avc1.4d401f\" frameRate=\"30\" height=\"720\" id=\"VIDEO-1\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"404\">\n                <BaseURL>DASH_720</BaseURL>\n                <SegmentBase indexRange=\"978-1093\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-977\" />\n                </SegmentBase>\n            </Representation>\n            <Representation bandwidth=\"1138311\" codecs=\"avc1.4d401f\" frameRate=\"30\" height=\"480\" id=\"VIDEO-2\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"270\">\n                <BaseURL>DASH_480</BaseURL>\n                <SegmentBase indexRange=\"975-1090\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-974\" />\n                </SegmentBase>\n            </Representation>\n            <Representation bandwidth=\"756236\" codecs=\"avc1.4d401e\" frameRate=\"30\" height=\"360\" id=\"VIDEO-3\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"202\">\n                <BaseURL>DASH_360</BaseURL>\n                <SegmentBase indexRange=\"978-1093\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-977\" />\n                </SegmentBase>\n            </Representation>\n            <Representation bandwidth=\"568151\" codecs=\"avc1.4d401e\" frameRate=\"30\" height=\"240\" id=\"VIDEO-4\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"134\">\n                <BaseURL>DASH_240</BaseURL>\n                <SegmentBase indexRange=\"978-1093\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-977\" />\n                </SegmentBase>\n            </Representation>\n            </AdaptationSet>\n        <AdaptationSet>\n            <Representation audioSamplingRate=\"48000\" bandwidth=\"130325\" codecs=\"mp4a.40.2\" id=\"AUDIO-1\" mimeType=\"audio/mp4\" startWithSAP=\"1\">\n                <AudioChannelConfiguration schemeIdUri=\"urn:mpeg:dash:23003:3:audio_channel_configuration:2011\" value=\"2\" />\n                <BaseURL>audio</BaseURL>\n                <SegmentBase indexRange=\"892-995\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-891\" />\n                </SegmentBase>\n            </Representation>\n        </AdaptationSet>\n    </Period>\n</MPD>",
			Expected: AvailableMedia{
				AvailableVideos: []AvailableVideo{"DASH_720", "DASH_480", "DASH_360", "DASH_240"},
				AvailableAudios: []AvailableAudio{"audio"},
			},
		},
		{ // From https://v.redd.it/o8y2x0z8jsq41/DASHPlaylist.mpd
			Name: "video_very_old",
			Data: "<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" mediaPresentationDuration=\"PT30.9S\" minBufferTime=\"PT1.500S\" profiles=\"urn:mpeg:dash:profile:isoff-on-demand:2011\" type=\"static\">\n    <Period duration=\"PT30.9S\">\n        <AdaptationSet segmentAlignment=\"true\" subsegmentAlignment=\"true\" subsegmentStartsWithSAP=\"1\">\n            <Representation bandwidth=\"1168724\" codecs=\"avc1.4d401f\" frameRate=\"30\" height=\"480\" id=\"VIDEO-1\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"480\">\n                <BaseURL>DASH_480</BaseURL>\n                <SegmentBase indexRange=\"918-1045\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-917\" />\n                </SegmentBase>\n            </Representation>\n            <Representation bandwidth=\"780377\" codecs=\"avc1.4d401e\" frameRate=\"30\" height=\"360\" id=\"VIDEO-2\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"360\">\n                <BaseURL>DASH_360</BaseURL>\n                <SegmentBase indexRange=\"919-1046\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-918\" />\n                </SegmentBase>\n            </Representation>\n            <Representation bandwidth=\"589445\" codecs=\"avc1.4d401e\" frameRate=\"30\" height=\"240\" id=\"VIDEO-3\" mimeType=\"video/mp4\" startWithSAP=\"1\" width=\"240\">\n                <BaseURL>DASH_240</BaseURL>\n                <SegmentBase indexRange=\"917-1044\" indexRangeExact=\"true\">\n                    <Initialization range=\"0-916\" />\n                </SegmentBase>\n            </Representation>\n            </AdaptationSet>\n    </Period>\n</MPD>",
			Expected: AvailableMedia{
				AvailableVideos: []AvailableVideo{"DASH_480", "DASH_360", "DASH_240"},
				AvailableAudios: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			gotResult, err := parseDashPlaylist(strings.NewReader(test.Data))
			assert.ErrorIs(t, err, test.Error)
			assert.Equal(t, test.Expected, gotResult)
		})
	}
}
