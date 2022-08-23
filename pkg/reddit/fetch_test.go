package reddit

import (
	"RedditDownloaderBot/pkg/util"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractLinkAndRes(t *testing.T) {
	assertion := assert.New(t)
	var parsedJson interface{}
	dataString := `{ "url":"https://preview.redd.it/utx00pfe4cp41.jpg?auto=webp&amp;s=de4ff82478b12df6369b8d7eeca3894f094e87e1", "width":624, "height":960 }`
	err := json.NewDecoder(strings.NewReader(dataString)).Decode(&parsedJson)
	assertion.NoError(err, "unexpected error when parsing test json")
	url, width, height := extractLinkAndRes(parsedJson)
	assertion.Equal("624", width, "unexpected width")
	assertion.Equal("960", height, "unexpected height")
	assertion.Equal("https://preview.redd.it/utx00pfe4cp41.jpg?auto=webp&s=de4ff82478b12df6369b8d7eeca3894f094e87e1", url, "unexpected url")
}

func TestExtractVideoQualities(t *testing.T) {
	id := randomId()
	t.Run("480p With New Audio", func(t *testing.T) {
		server := newSimpleWebserver(
			fmt.Sprintf("/%s/DASH_480.mp4", id),
			fmt.Sprintf("/%s/DASH_audio.mp4", id),
		)
		defer server.Close()
		url := fmt.Sprintf("%s/%s/DASH_480.mp4?source=fallback", server.URL, id)
		result := extractVideoQualities(url)
		assert.Equal(t, []FetchResultMediaEntry{
			{
				Link:    fmt.Sprintf("%s/%s/DASH_480.mp4", server.URL, id),
				Quality: "480p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_360.mp4", server.URL, id),
				Quality: "360p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_240.mp4", server.URL, id),
				Quality: "240p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_96.mp4", server.URL, id),
				Quality: "96p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_audio.mp4", server.URL, id),
				Quality: DownloadAudioQuality,
			},
		}, result)
	})
	t.Run("480p With Old Audio", func(t *testing.T) {
		server := newSimpleWebserver(
			fmt.Sprintf("/%s/DASH_480", id),
			fmt.Sprintf("/%s/audio", id),
		)
		defer server.Close()
		url := fmt.Sprintf("%s/%s/DASH_480?source=fallback", server.URL, id)
		result := extractVideoQualities(url)
		assert.Equal(t, []FetchResultMediaEntry{
			{
				Link:    fmt.Sprintf("%s/%s/DASH_480", server.URL, id),
				Quality: "480p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_360", server.URL, id),
				Quality: "360p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_240", server.URL, id),
				Quality: "240p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_96", server.URL, id),
				Quality: "96p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/audio", server.URL, id),
				Quality: DownloadAudioQuality,
			},
		}, result)
	})
	t.Run("480p Without Audio", func(t *testing.T) {
		server := newSimpleWebserver(
			fmt.Sprintf("/%s/DASH_480.mp4", id),
		)
		defer server.Close()
		url := fmt.Sprintf("%s/%s/DASH_480.mp4?source=fallback", server.URL, id)
		result := extractVideoQualities(url)
		assert.Equal(t, []FetchResultMediaEntry{
			{
				Link:    fmt.Sprintf("%s/%s/DASH_480.mp4", server.URL, id),
				Quality: "480p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_360.mp4", server.URL, id),
				Quality: "360p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_240.mp4", server.URL, id),
				Quality: "240p",
			},
			{
				Link:    fmt.Sprintf("%s/%s/DASH_96.mp4", server.URL, id),
				Quality: "96p",
			},
		}, result)
	})
	t.Run("Max Quality Without Audio", func(t *testing.T) {
		server := newSimpleWebserver()
		defer server.Close()
		url := fmt.Sprintf("%s/%s/DASH_%s.mp4?source=fallback", server.URL, id, qualities[0])
		expectedResult := make([]FetchResultMediaEntry, len(qualities))
		for i, quality := range qualities {
			expectedResult[i] = FetchResultMediaEntry{
				Link:    fmt.Sprintf("%s/%s/DASH_%s.mp4", server.URL, id, quality),
				Quality: quality + "p",
			}
		}
		result := extractVideoQualities(url)
		assert.Equal(t, expectedResult, result)
	})
}

func TestExtractPhotoGifQualities(t *testing.T) {
	tests := []struct {
		TestName string
		RawData  string
		Expected []FetchResultMediaEntry
	}{
		{
			// From https://www.reddit.com/r/blender/comments/vgvnt5/stray_cat/
			TestName: "Simple Test",
			RawData:  `{"id":"Hg38I_MeQa1F_iYHaLwkQDeUGvJacQnAsHwm4rnsvak","resolutions":[{"height":192,"url":"https://preview.redd.it/trv29s0abu691.jpg?width=108\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=1cf2bf34ea4cfb637e2796feb06c6d1bee0a69ad","width":108},{"height":384,"url":"https://preview.redd.it/trv29s0abu691.jpg?width=216\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=6f0d22a06f478444bffa6ae3f56a22b4c372eae2","width":216},{"height":568,"url":"https://preview.redd.it/trv29s0abu691.jpg?width=320\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=507fcb794698d27f419da0b4e8ca1c0c55261acc","width":320},{"height":1137,"url":"https://preview.redd.it/trv29s0abu691.jpg?width=640\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=c5bb5c9097c68e614b06f365c4de84e64fc5f308","width":640},{"height":1706,"url":"https://preview.redd.it/trv29s0abu691.jpg?width=960\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=d7e0ff2025d1938585e71707208b074cc5db64b4","width":960},{"height":1920,"url":"https://preview.redd.it/trv29s0abu691.jpg?width=1080\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=dd6b8e8b31e689c6b93f30aa69eaa888763f5d53","width":1080}],"source":{"height":2880,"url":"https://preview.redd.it/trv29s0abu691.jpg?auto=webp\u0026amp;s=79a056cb82e2ef8f0c90b5e509c0aa2e8d11da9f","width":1620},"variants":{}}`,
			Expected: []FetchResultMediaEntry{
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?auto=webp&s=79a056cb82e2ef8f0c90b5e509c0aa2e8d11da9f",
					Quality: "1620×2880",
				},
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?width=1080&crop=smart&auto=webp&s=dd6b8e8b31e689c6b93f30aa69eaa888763f5d53",
					Quality: "1080×1920",
				},
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?width=960&crop=smart&auto=webp&s=d7e0ff2025d1938585e71707208b074cc5db64b4",
					Quality: "960×1706",
				},
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?width=640&crop=smart&auto=webp&s=c5bb5c9097c68e614b06f365c4de84e64fc5f308",
					Quality: "640×1137",
				},
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?width=320&crop=smart&auto=webp&s=507fcb794698d27f419da0b4e8ca1c0c55261acc",
					Quality: "320×568",
				},
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?width=216&crop=smart&auto=webp&s=6f0d22a06f478444bffa6ae3f56a22b4c372eae2",
					Quality: "216×384",
				},
				{
					Link:    "https://preview.redd.it/trv29s0abu691.jpg?width=108&crop=smart&auto=webp&s=1cf2bf34ea4cfb637e2796feb06c6d1bee0a69ad",
					Quality: "108×192",
				},
			},
		},
		{
			// From https://www.reddit.com/r/dankmemes/comments/vqphiy/more_than_bargain_for/
			TestName: "Has Same Sizes As Original",
			RawData:  `{"resolutions":[{"height":75,"url":"https://preview.redd.it/gaqrixuhqe991.gif?width=108\u0026amp;format=mp4\u0026amp;s=8996a081c0078a4b9ef946efa00962017a60ca99","width":108},{"height":151,"url":"https://preview.redd.it/gaqrixuhqe991.gif?width=216\u0026amp;format=mp4\u0026amp;s=077098992450e5b2406ac655fba74016ac35e515","width":216},{"height":224,"url":"https://preview.redd.it/gaqrixuhqe991.gif?width=320\u0026amp;format=mp4\u0026amp;s=243d619b455c3abf088063bbf5910ef0809150c1","width":320},{"height":448,"url":"https://preview.redd.it/gaqrixuhqe991.gif?width=640\u0026amp;format=mp4\u0026amp;s=df4b65307976aec57c16cd980633bf09fcae2d66","width":640}],"source":{"height":448,"url":"https://preview.redd.it/gaqrixuhqe991.gif?format=mp4\u0026amp;s=0e3eb80b311b783b615226093f00ceadd7d8a881","width":640}}`,
			Expected: []FetchResultMediaEntry{
				{
					Link:    "https://preview.redd.it/gaqrixuhqe991.gif?format=mp4&s=0e3eb80b311b783b615226093f00ceadd7d8a881",
					Quality: "640×448",
				},
				{
					Link:    "https://preview.redd.it/gaqrixuhqe991.gif?width=320&format=mp4&s=243d619b455c3abf088063bbf5910ef0809150c1",
					Quality: "320×224",
				},
				{
					Link:    "https://preview.redd.it/gaqrixuhqe991.gif?width=216&format=mp4&s=077098992450e5b2406ac655fba74016ac35e515",
					Quality: "216×151",
				},
				{
					Link:    "https://preview.redd.it/gaqrixuhqe991.gif?width=108&format=mp4&s=8996a081c0078a4b9ef946efa00962017a60ca99",
					Quality: "108×75",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			var parsedData map[string]interface{}
			err := json.NewDecoder(strings.NewReader(test.RawData)).Decode(&parsedData)
			assert.NoError(t, err, "sample data must be parsed without errors")
			result := extractPhotoGifQualities(parsedData)
			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestGetGalleryData(t *testing.T) {
	tests := []struct {
		TestName         string
		Files            string
		GalleryDataItems string
		ExpectedAlbum    FetchResultAlbum
	}{
		// From https://www.reddit.com/r/gtaonline/comments/wuid83/caught_these_2_under_the_pier_they_went_3_rounds/?utm_source=share&utm_medium=web2x&context=3
		{
			TestName:         "Sample Gallery",
			Files:            `{"175srd9hm6j91":{"e":"Image","id":"175srd9hm6j91","m":"image/jpg","o":[{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=1080\u0026amp;blur=40\u0026amp;format=pjpg\u0026amp;auto=webp\u0026amp;s=580eccf250b6fd30344aaed372b370ee07361bfa","x":1920,"y":1080}],"p":[{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=108\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=4ad372d3033811e5307efcb5e3628cc3c3af5a36","x":108,"y":60},{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=216\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=c94ee14dd4d8acca278f1632307473226772fed2","x":216,"y":121},{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=320\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=e0c636bff8d555975a2eec4720b52a738530ffc5","x":320,"y":180},{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=640\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=603904050c1b07ee5a6ad33a0130498a0878cf93","x":640,"y":360},{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=960\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=6984617dd4e9c739e1b96d96df6e440ac8fcf282","x":960,"y":540},{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=1080\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=4be0eb1f89cb23c9180e8e862fe8568a2561abe0","x":1080,"y":607}],"s":{"u":"https://preview.redd.it/175srd9hm6j91.jpg?width=1920\u0026amp;format=pjpg\u0026amp;auto=webp\u0026amp;s=8e8be25e2233f83b9ef621bd9cc9768b1b8ac5b7","x":1920,"y":1080},"status":"valid"},"rdpwee9hm6j91":{"e":"Image","id":"rdpwee9hm6j91","m":"image/jpg","o":[{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=1080\u0026amp;blur=40\u0026amp;format=pjpg\u0026amp;auto=webp\u0026amp;s=f79c8a3eecbe65aad603f4ab48520d46a19449a1","x":1920,"y":1080}],"p":[{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=108\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=3aa63ed07f7242096f270eb2947620c08b3152e0","x":108,"y":60},{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=216\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=72b19927946624fb263e1e3ccf9a12ba4931cee9","x":216,"y":121},{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=320\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=9e3d268f422c40640db3dd7c52702ed4772758a7","x":320,"y":180},{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=640\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=fe9112ca32b00cfd12a947ef4f827047f50a223b","x":640,"y":360},{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=960\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=7e5a575950d48271f6e072be43f42ffb9b2110dd","x":960,"y":540},{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=1080\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=cad4b97b852bffbd5f8cf5209b949dee86182b7f","x":1080,"y":607}],"s":{"u":"https://preview.redd.it/rdpwee9hm6j91.jpg?width=1920\u0026amp;format=pjpg\u0026amp;auto=webp\u0026amp;s=81b50126588ec9ba0c3d1d354e78377db8d887aa","x":1920,"y":1080},"status":"valid"}}`,
			GalleryDataItems: `[{"id":178597657,"media_id":"175srd9hm6j91"},{"id":178597658,"media_id":"rdpwee9hm6j91"}]`,
			ExpectedAlbum: FetchResultAlbum{Album: []FetchResultAlbumEntry{
				{
					Link:    "https://preview.redd.it/175srd9hm6j91.jpg?width=1920&format=pjpg&auto=webp&s=8e8be25e2233f83b9ef621bd9cc9768b1b8ac5b7",
					Caption: "",
					Type:    FetchResultMediaTypePhoto,
				},
				{
					Link:    "https://preview.redd.it/rdpwee9hm6j91.jpg?width=1920&format=pjpg&auto=webp&s=81b50126588ec9ba0c3d1d354e78377db8d887aa",
					Caption: "",
					Type:    FetchResultMediaTypePhoto,
				},
			}},
		},
		// From https://www.reddit.com/r/gaming/comments/vdrdxu/225_years_of_game_reviews/
		{
			TestName:         "Gallery With Caption",
			Files:            `{"k1u47jv0r0691":{"e":"Image","id":"k1u47jv0r0691","m":"image/png","p":[{"u":"https://preview.redd.it/k1u47jv0r0691.png?width=108\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=ae775025d25e26f65d07d2a65f265c1950085465","x":108,"y":81},{"u":"https://preview.redd.it/k1u47jv0r0691.png?width=216\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=21a6ef7f7beb776ad3db94b1ca8f985c9ced1125","x":216,"y":162},{"u":"https://preview.redd.it/k1u47jv0r0691.png?width=320\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=e6806cc7d95025ab21d674642f2e20fd42b4dccd","x":320,"y":240},{"u":"https://preview.redd.it/k1u47jv0r0691.png?width=640\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=06bc4bea6474e5f2ac9c3ddcaf413b4257dbe45c","x":640,"y":480}],"s":{"u":"https://preview.redd.it/k1u47jv0r0691.png?width=640\u0026amp;format=png\u0026amp;auto=webp\u0026amp;s=49da7bbf74a98eb3fa634ab1ff5ad2ae7aef834e","x":640,"y":480},"status":"valid"},"la2xoko0r0691":{"e":"Image","id":"la2xoko0r0691","m":"image/png","p":[{"u":"https://preview.redd.it/la2xoko0r0691.png?width=108\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=857be761023e7a5f682b16ef3d205f5fe1f30c8c","x":108,"y":81},{"u":"https://preview.redd.it/la2xoko0r0691.png?width=216\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=deba1208ce288def790e839159f060fad1c3405d","x":216,"y":162},{"u":"https://preview.redd.it/la2xoko0r0691.png?width=320\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=bfa7cdad1484d40aa8e2f844626d366049d5e521","x":320,"y":240},{"u":"https://preview.redd.it/la2xoko0r0691.png?width=640\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=70bd221613760550ba7370ac97088e0c7406b8e3","x":640,"y":480}],"s":{"u":"https://preview.redd.it/la2xoko0r0691.png?width=640\u0026amp;format=png\u0026amp;auto=webp\u0026amp;s=7659e4925b6fb948e1328338e814bd5bec17ee1d","x":640,"y":480},"status":"valid"},"xjsrk4u0r0691":{"e":"Image","id":"xjsrk4u0r0691","m":"image/png","p":[{"u":"https://preview.redd.it/xjsrk4u0r0691.png?width=108\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=24307afed39e05015a9bf6d54aad0c7d441cf6c9","x":108,"y":81},{"u":"https://preview.redd.it/xjsrk4u0r0691.png?width=216\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=a1898705ccb2543bb4667cec13265189aa614714","x":216,"y":162},{"u":"https://preview.redd.it/xjsrk4u0r0691.png?width=320\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=8863d795d015349ca80593b874cbe9cb856a9534","x":320,"y":240},{"u":"https://preview.redd.it/xjsrk4u0r0691.png?width=640\u0026amp;crop=smart\u0026amp;auto=webp\u0026amp;s=50d89f26ea70065673247eacd63be42f14494e85","x":640,"y":480}],"s":{"u":"https://preview.redd.it/xjsrk4u0r0691.png?width=640\u0026amp;format=png\u0026amp;auto=webp\u0026amp;s=8e7b575a52f4dd2adb03eae458c91ead7bd86292","x":640,"y":480},"status":"valid"}}`,
			GalleryDataItems: `[{"caption":"wtf happened in 2007 to universally change the minds of critics, but not the users? Swipe for console breakdown.","id":153545018,"media_id":"la2xoko0r0691"},{"id":153545019,"media_id":"xjsrk4u0r0691"},{"id":153545020,"media_id":"k1u47jv0r0691"}]`,
			ExpectedAlbum: FetchResultAlbum{Album: []FetchResultAlbumEntry{
				{
					Link:    "https://preview.redd.it/la2xoko0r0691.png?width=640&format=png&auto=webp&s=7659e4925b6fb948e1328338e814bd5bec17ee1d",
					Caption: "wtf happened in 2007 to universally change the minds of critics, but not the users? Swipe for console breakdown.",
					Type:    FetchResultMediaTypePhoto,
				},
				{
					Link:    "https://preview.redd.it/xjsrk4u0r0691.png?width=640&format=png&auto=webp&s=8e7b575a52f4dd2adb03eae458c91ead7bd86292",
					Caption: "",
					Type:    FetchResultMediaTypePhoto,
				},
				{
					Link:    "https://preview.redd.it/k1u47jv0r0691.png?width=640&format=png&auto=webp&s=49da7bbf74a98eb3fa634ab1ff5ad2ae7aef834e",
					Caption: "",
					Type:    FetchResultMediaTypePhoto,
				},
			}},
		},
		{
			TestName:         "GIF Test",
			Files:            `{"f83dt8n6q5861":{"e":"AnimatedImage","id":"f83dt8n6q5861","m":"image/gif","p":[{"u":"https://preview.redd.it/f83dt8n6q5861.gif?width=108\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=f3d61dc7a4c5e13193de875928bd772018432d7a","x":108,"y":108},{"u":"https://preview.redd.it/f83dt8n6q5861.gif?width=216\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=b12f7bdf549fe027a4641d680cc1121ce3fb57e2","x":216,"y":216},{"u":"https://preview.redd.it/f83dt8n6q5861.gif?width=320\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=cec53b8442a884bbdb44bcbf9288597c1ff87ecc","x":320,"y":320},{"u":"https://preview.redd.it/f83dt8n6q5861.gif?width=640\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=56d517b161c90771cff655fb4797fa1308d6a75f","x":640,"y":640}],"s":{"gif":"https://i.redd.it/f83dt8n6q5861.gif","mp4":"https://preview.redd.it/f83dt8n6q5861.gif?format=mp4\u0026amp;s=b3213eed2507622e46f5614d03734d80f7b81f58","x":720,"y":720},"status":"valid"},"j4p941i6q5861":{"e":"AnimatedImage","id":"j4p941i6q5861","m":"image/gif","p":[{"u":"https://preview.redd.it/j4p941i6q5861.gif?width=108\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=848db65f847057f0d97708ed58ee9ae24ebd946f","x":108,"y":108},{"u":"https://preview.redd.it/j4p941i6q5861.gif?width=216\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=459788c30dae59c8f4be166dfaafec79581e7cef","x":216,"y":216},{"u":"https://preview.redd.it/j4p941i6q5861.gif?width=320\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=5b9844a3513a0c5ddb4f0868ea479e99d742b6a4","x":320,"y":320},{"u":"https://preview.redd.it/j4p941i6q5861.gif?width=640\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=8a00073b9a09a0ea1333653581205fe990b15e50","x":640,"y":640}],"s":{"gif":"https://i.redd.it/j4p941i6q5861.gif","mp4":"https://preview.redd.it/j4p941i6q5861.gif?format=mp4\u0026amp;s=26728fa136f4f4f1681fbb59d49eefa0c646874e","x":720,"y":720},"status":"valid"},"xx9knep6q5861":{"e":"AnimatedImage","id":"xx9knep6q5861","m":"image/gif","p":[{"u":"https://preview.redd.it/xx9knep6q5861.gif?width=108\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=9669cab64b0830ab47d8d78f6e9660255aec658b","x":108,"y":108},{"u":"https://preview.redd.it/xx9knep6q5861.gif?width=216\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=2da4c5cd485494c7ee4f591dadf5c9bfbca1b05b","x":216,"y":216},{"u":"https://preview.redd.it/xx9knep6q5861.gif?width=320\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=1d34e38c44779bf682afe24e942b9a27d9c82761","x":320,"y":320},{"u":"https://preview.redd.it/xx9knep6q5861.gif?width=640\u0026amp;crop=smart\u0026amp;format=png8\u0026amp;s=51155c61179515f90c2bb8143dab7ff4bb677258","x":640,"y":640}],"s":{"gif":"https://i.redd.it/xx9knep6q5861.gif","mp4":"https://preview.redd.it/xx9knep6q5861.gif?format=mp4\u0026amp;s=5cb367822b02ff2b7c3333019feefefda15e5012","x":720,"y":720},"status":"valid"}}`,
			GalleryDataItems: `[{"id":19568528,"media_id":"j4p941i6q5861"},{"id":19568529,"media_id":"f83dt8n6q5861"},{"id":19568530,"media_id":"xx9knep6q5861"}]`,
			ExpectedAlbum: FetchResultAlbum{Album: []FetchResultAlbumEntry{
				{
					Link:    "https://preview.redd.it/j4p941i6q5861.gif?format=mp4&s=26728fa136f4f4f1681fbb59d49eefa0c646874e",
					Caption: "",
					Type:    FetchResultMediaTypeGif,
				},
				{
					Link:    "https://preview.redd.it/f83dt8n6q5861.gif?format=mp4&s=b3213eed2507622e46f5614d03734d80f7b81f58",
					Caption: "",
					Type:    FetchResultMediaTypeGif,
				},
				{
					Link:    "https://preview.redd.it/xx9knep6q5861.gif?format=mp4&s=5cb367822b02ff2b7c3333019feefefda15e5012",
					Caption: "",
					Type:    FetchResultMediaTypeGif,
				},
			}},
		},
		// TODO: add a video test
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			var files map[string]interface{}
			var galleryDataItems []interface{}
			err := json.NewDecoder(strings.NewReader(test.Files)).Decode(&files)
			assert.NoError(t, err, "not expecting error when decoding sample files")
			err = json.NewDecoder(strings.NewReader(test.GalleryDataItems)).Decode(&galleryDataItems)
			assert.NoError(t, err, "not expecting error when decoding sample gallery data items")
			result := getGalleryData(files, galleryDataItems)
			assert.Equal(t, test.ExpectedAlbum, result)
		})
	}
}

func TestGetPostId(t *testing.T) {
	tests := []struct {
		TestName          string
		Url               string
		NeedsInternet     bool
		ExpectedID        string
		ExpectedIsComment bool
		ExpectedError     string // is empty if no error must be thrown
	}{
		{
			TestName:          "Normal Post",
			Url:               "https://www.reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Normal Comment",
			Url:               "https://www.reddit.com/r/gaming/comments/vdrdxu/comment/icm3y72/?utm_source=share&utm_medium=web2x&context=3",
			NeedsInternet:     false,
			ExpectedID:        "icm3y72",
			ExpectedIsComment: true,
			ExpectedError:     "",
		},
		{
			TestName:          "redd.it Link",
			Url:               "https://redd.it/kmi4d3",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "v.redd.it Link",
			Url:               "https://v.redd.it/rhs0ixoyc7j91",
			NeedsInternet:     true,
			ExpectedID:        "wul62b",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Post With Other Lines",
			Url:               "Prop Hunt Was Fun\nhttps://www.reddit.com/r/Unexpected/comments/wul62b/prop_hunt_was_fun/\nhttps://google.com",
			NeedsInternet:     false,
			ExpectedID:        "wul62b",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Invalid Url",
			Url:               "",
			NeedsInternet:     false,
			ExpectedID:        "",
			ExpectedIsComment: false,
			ExpectedError:     "Cannot parse reddit the url. Does your text contain a reddit url?",
		},
		{
			TestName:          "Short Url",
			Url:               "https://www.reddit.com/r/Unexpected/comments",
			NeedsInternet:     false,
			ExpectedID:        "",
			ExpectedIsComment: false,
			ExpectedError:     "Cannot parse reddit the url. Does your text contain a reddit url?",
		},
		{
			TestName:          "Short Reddit Url",
			Url:               "https://www.reddit.com/wul62b",
			NeedsInternet:     false,
			ExpectedID:        "wul62b",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Old Post",
			Url:               "https://old.reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Normal Post 2",
			Url:               "https://reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
		{
			TestName:          "Normal Post No Transport",
			Url:               "reddit.com/r/dankmemes/comments/kmi4d3/invest_in_sliding_gif_memes/?utm_medium=android_app&utm_source=share",
			NeedsInternet:     false,
			ExpectedID:        "kmi4d3",
			ExpectedIsComment: false,
			ExpectedError:     "",
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			// Check internet if needed
			if test.NeedsInternet {
				if _, err := util.FollowRedirect(test.Url); err != nil {
					t.Skip("cannot connect to internet:", err)
					return
				}
			}
			// Get the id
			id, isComment, err := getPostID(test.Url)
			if err != nil {
				assert.Equal(t, test.ExpectedError, err.BotError)
			}
			assert.Equal(t, test.ExpectedIsComment, isComment)
			assert.Equal(t, test.ExpectedID, id)
		})
	}
}

func newSimpleWebserver(patterns ...string) *httptest.Server {
	mux := http.NewServeMux()
	for _, path := range patterns {
		mux.HandleFunc(path, func(http.ResponseWriter, *http.Request) {})
	}
	return httptest.NewServer(mux)
}

func randomId() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
