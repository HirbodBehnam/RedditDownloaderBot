package reddit

import (
	"RedditDownloaderBot/pkg/util"
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

func TestGetVideoIDFromVReddit(t *testing.T) {
	tests := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{
			TestName: "normal",
			Input:    "https://v.redd.it/dbelx9ulpacb1/DASH_1080.mp4?source=fallback",
			Expected: "dbelx9ulpacb1",
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			assert.Equal(t, test.Expected, getVideoIDFromVReddit(test.Input))
		})
	}
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

func TestGetCommentFromRoot(t *testing.T) {
	tests := []struct {
		TestName string
		Root     string
		Expected interface{}
	}{
		// From https://www.reddit.com/r/gtaonline/comments/ww9qw1/comment/iljyela/?utm_source=share&utm_medium=web2x&context=3
		{
			TestName: "Text Comment",
			Root:     `{"data":{"after":null,"before":null,"children":[{"data":{"all_awardings":[],"approved_at_utc":null,"approved_by":null,"archived":false,"associated_award":null,"author":"MD9564","author_flair_background_color":"#46d160","author_flair_css_class":"xbx","author_flair_richtext":[{"a":":XBX1:","e":"emoji","u":"https://emoji.redditmedia.com/wts1rb0yicq71_t5_2xrd1/XBX1"},{"a":":XBX2:","e":"emoji","u":"https://emoji.redditmedia.com/8wwoun2yicq71_t5_2xrd1/XBX2"}],"author_flair_template_id":"303594a4-2bbc-11eb-b5d1-0e259b87ccd1","author_flair_text":":XBX1::XBX2:","author_flair_text_color":"light","author_flair_type":"richtext","author_fullname":"t2_1eygor7t","author_is_blocked":false,"author_patreon_flair":false,"author_premium":false,"awarders":[],"banned_at_utc":null,"banned_by":null,"body":"The Plane Door is closed.","body_html":"\u0026lt;div class=\"md\"\u0026gt;\u0026lt;p\u0026gt;The Plane Door is closed.\u0026lt;/p\u0026gt;\n\u0026lt;/div\u0026gt;","can_gild":true,"can_mod_post":false,"collapsed":false,"collapsed_because_crowd_control":null,"collapsed_reason":null,"collapsed_reason_code":null,"comment_type":null,"controversiality":0,"created":1661315239,"created_utc":1661315239,"distinguished":null,"downs":0,"edited":false,"gilded":0,"gildings":{},"id":"iljyela","is_submitter":false,"likes":null,"link_id":"t3_ww9qw1","locked":false,"mod_note":null,"mod_reason_by":null,"mod_reason_title":null,"mod_reports":[],"name":"t1_iljyela","no_follow":false,"num_reports":null,"parent_id":"t3_ww9qw1","permalink":"/r/gtaonline/comments/ww9qw1/if_you_are_a_real_cayo_grinder_then_tell_me_whats/iljyela/","removal_reason":null,"replies":"","report_reasons":null,"saved":false,"score":31,"score_hidden":false,"send_replies":true,"stickied":false,"subreddit":"gtaonline","subreddit_id":"t5_2xrd1","subreddit_name_prefixed":"r/gtaonline","subreddit_type":"public","top_awarded_type":null,"total_awards_received":0,"treatment_tags":[],"unrepliable_reason":null,"ups":31,"user_reports":[]},"kind":"t1"}],"dist":1,"geo_filter":"","modhash":""},"kind":"Listing"}`,
			Expected: FetchResultComment{"The Plane Door is closed."},
		},
		// From https://www.reddit.com/r/whenthe/comments/wq2fpi/comment/ikkn4sr/?utm_source=share&utm_medium=web2x&context=3
		{
			TestName: "Gif Comment",
			Root:     `{"data":{"after":null,"before":null,"children":[{"data":{"all_awardings":[],"approved_at_utc":null,"approved_by":null,"archived":false,"associated_award":null,"author":"FuckYeahPhotography","author_flair_background_color":"#800080","author_flair_css_class":null,"author_flair_richtext":[{"e":"text","t":"My Profile Posts are the Hottest Party 📸"}],"author_flair_template_id":"4c85ef62-c37d-11e9-9242-0eb1ea29758e","author_flair_text":"My Profile Posts are the Hottest Party 📸","author_flair_text_color":"light","author_flair_type":"richtext","author_fullname":"t2_73yjd","author_is_blocked":false,"author_patreon_flair":false,"author_premium":true,"awarders":[],"banned_at_utc":null,"banned_by":null,"body":"![gif](giphy|gVoBC0SuaHStq)","body_html":"\u0026lt;div class=\"md\"\u0026gt;\u0026lt;p\u0026gt;\u0026lt;a href=\"https://giphy.com/gifs/gVoBC0SuaHStq\" target=\"_blank\"\u0026gt;\u0026lt;img src=\"https://external-preview.redd.it/F1xkfBzKhzUkqP558H1pT2WMhX6O2XRrmPWyJMC7Q3I.gif?width=196\u0026amp;height=200\u0026amp;s=901a7ba82c3fea1b2736817d69cd76287270c1f5\" width=\"196\" height=\"200\"\u0026gt;\u0026lt;/a\u0026gt;\u0026lt;/p\u0026gt;\n\u0026lt;/div\u0026gt;","can_gild":true,"can_mod_post":false,"collapsed":false,"collapsed_because_crowd_control":null,"collapsed_reason":null,"collapsed_reason_code":null,"comment_type":null,"controversiality":0,"created":1660684723,"created_utc":1660684723,"distinguished":null,"downs":0,"edited":false,"gilded":0,"gildings":{},"id":"ikkn4sr","is_submitter":false,"likes":null,"link_id":"t3_wq2fpi","locked":false,"media_metadata":{"giphy|gVoBC0SuaHStq":{"e":"AnimatedImage","ext":"https://giphy.com/gifs/gVoBC0SuaHStq","id":"giphy|gVoBC0SuaHStq","m":"image/gif","p":[{"u":"https://b.thumbs.redditmedia.com/qDt_ZorM1vG2y-05l4R-m5H0APry8psej7IC4HVOL8Q.jpg","x":140,"y":140}],"s":{"gif":"https://external-preview.redd.it/F1xkfBzKhzUkqP558H1pT2WMhX6O2XRrmPWyJMC7Q3I.gif?width=196\u0026amp;height=200\u0026amp;s=901a7ba82c3fea1b2736817d69cd76287270c1f5","mp4":"https://external-preview.redd.it/F1xkfBzKhzUkqP558H1pT2WMhX6O2XRrmPWyJMC7Q3I.gif?width=196\u0026amp;height=200\u0026amp;format=mp4\u0026amp;s=7df4f115ddc8d384f3f406461d508752e7de97f7","x":196,"y":200},"status":"valid","t":"giphy"}},"mod_note":null,"mod_reason_by":null,"mod_reason_title":null,"mod_reports":[],"name":"t1_ikkn4sr","no_follow":false,"num_reports":null,"parent_id":"t1_ikkm540","permalink":"/r/whenthe/comments/wq2fpi/oh_boy_a_new_dad/ikkn4sr/","removal_reason":null,"replies":"","report_reasons":null,"saved":false,"score":46,"score_hidden":false,"send_replies":true,"stickied":false,"subreddit":"whenthe","subreddit_id":"t5_23gidu","subreddit_name_prefixed":"r/whenthe","subreddit_type":"public","top_awarded_type":null,"total_awards_received":0,"treatment_tags":[],"unrepliable_reason":null,"ups":46,"user_reports":[]},"kind":"t1"}],"dist":1,"geo_filter":"","modhash":""},"kind":"Listing"}`,
			Expected: FetchResultMedia{
				Medias: []FetchResultMediaEntry{{
					Link:    "https://i.giphy.com/media/gVoBC0SuaHStq/giphy.gif",
					Quality: "giphy",
				}},
				Type:  FetchResultMediaTypeGif,
				Title: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			var root map[string]interface{}
			err := json.NewDecoder(strings.NewReader(test.Root)).Decode(&root)
			assert.NoError(t, err, "not expecting error when decoding sample root")
			result := getCommentFromRoot(root)
			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestGetPost(t *testing.T) {
	tests := []struct {
		TestName       string
		PostUrl        string
		Root           string
		WebserverUrls  []string // null if no webserver is needed
		ExpectedResult interface{}
		ExpectedError  *FetchError
	}{
		{
			TestName:       "Text 1",
			PostUrl:        "https://www.reddit.com/r/Showerthoughts/comments/ww6stq/we_are_all_taught_to_walk_calming_out_of_a/?utm_source=share&utm_medium=web2x&context=3",
			Root:           `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "2dwcqob84k94764bfe9471424541c51ea59f9d367f91a55ff8", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "Showerthoughts", "selftext": "", "author_fullname": "t2_8hdaf78b", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "We are all taught to walk calming out of a building during an emergency. Yet on every exit sign, that MF is running out the door.", "link_flair_richtext": [], "subreddit_name_prefixed": "r/Showerthoughts", "hidden": false, "pwls": 6, "link_flair_css_class": null, "downs": 0, "thumbnail_height": null, "top_awarded_type": null, "hide_score": false, "name": "t3_ww6stq", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.97, "author_flair_background_color": null, "subreddit_type": "public", "ups": 4101, "total_awards_received": 0, "media_embed": {}, "thumbnail_width": null, "author_flair_template_id": null, "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": null, "can_mod_post": false, "score": 4101, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "self", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [], "gildings": {}, "content_categories": null, "is_self": true, "mod_note": null, "created": 1661306308.0, "link_flair_type": "text", "wls": 6, "removed_by_category": null, "banned_by": null, "author_flair_type": "text", "domain": "self.Showerthoughts", "allow_live_comments": false, "selftext_html": null, "likes": true, "suggested_sort": null, "banned_at_utc": null, "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "all_awardings": [], "awarders": [], "media_only": false, "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": null, "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2szyo", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "ww6stq", "is_robot_indexable": true, "report_reasons": null, "author": "LBK0909", "discussion_type": null, "num_comments": 45, "send_replies": true, "whitelist_status": "all_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": null, "permalink": "/r/Showerthoughts/comments/ww6stq/we_are_all_taught_to_walk_calming_out_of_a/", "parent_whitelist_status": "all_ads", "stickied": false, "url": "https://www.reddit.com/r/Showerthoughts/comments/ww6stq/we_are_all_taught_to_walk_calming_out_of_a/", "subreddit_subscribers": 25463223, "created_utc": 1661306308.0, "num_crossposts": 0, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultText{Title: "We are all taught to walk calming out of a building during an emergency. Yet on every exit sign, that MF is running out the door."},
			ExpectedError:  nil,
		},
		{
			TestName: "Text 2",
			PostUrl:  "https://www.reddit.com/r/csharp/comments/ww48d3/crossplatform_library_for_managing_windows/?utm_source=share&utm_medium=web2x&context=3",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "bi256brxl5a0386b54a4467c43ae8a64facb494b5a6cc7655f", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "csharp", "selftext": "Hi there!\n\nI am searching for a simple cross-platform library that can simply render a window given the pixels of the image (like a two-dimentional array). Is there something like that? The simpler the library, the better.\n\nThanks!", "author_fullname": "t2_c95g9k73", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "Cross-platform library for managing windows", "link_flair_richtext": [], "subreddit_name_prefixed": "r/csharp", "hidden": false, "pwls": 6, "link_flair_css_class": "discussion", "downs": 0, "thumbnail_height": null, "top_awarded_type": null, "hide_score": false, "name": "t3_ww48d3", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.86, "author_flair_background_color": null, "subreddit_type": "public", "ups": 5, "total_awards_received": 0, "media_embed": {}, "thumbnail_width": null, "author_flair_template_id": null, "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": "Discussion", "can_mod_post": false, "score": 5, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "self", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [], "gildings": {}, "content_categories": null, "is_self": true, "mod_note": null, "created": 1661299306.0, "link_flair_type": "text", "wls": 6, "removed_by_category": null, "banned_by": null, "author_flair_type": "text", "domain": "self.csharp", "allow_live_comments": false, "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;Hi there!&lt;/p&gt;\n\n&lt;p&gt;I am searching for a simple cross-platform library that can simply render a window given the pixels of the image (like a two-dimentional array). Is there something like that? The simpler the library, the better.&lt;/p&gt;\n\n&lt;p&gt;Thanks!&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "likes": null, "suggested_sort": null, "banned_at_utc": null, "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "all_awardings": [], "awarders": [], "media_only": false, "link_flair_template_id": "0ab15834-e357-11e4-8da2-22000bc1889b", "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": null, "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2qhdf", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "ww48d3", "is_robot_indexable": true, "report_reasons": null, "author": "rafaellintz", "discussion_type": null, "num_comments": 2, "send_replies": true, "whitelist_status": "all_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": null, "permalink": "/r/csharp/comments/ww48d3/crossplatform_library_for_managing_windows/", "parent_whitelist_status": "all_ads", "stickied": false, "url": "https://www.reddit.com/r/csharp/comments/ww48d3/crossplatform_library_for_managing_windows/", "subreddit_subscribers": 205784, "created_utc": 1661299306.0, "num_crossposts": 0, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultText{
				Title: "Cross-platform library for managing windows",
				Text:  "Hi there!\n\nI am searching for a simple cross-platform library that can simply render a window given the pixels of the image (like a two-dimentional array). Is there something like that? The simpler the library, the better.\n\nThanks!",
			},
			ExpectedError: nil,
		},
		{
			TestName: "Text 3 (With code)",
			PostUrl:  "https://www.reddit.com/r/csharp/comments/ww748q/how_to_return_multiple_columns_when_comparing_two/?utm_source=share&utm_medium=web2x&context=3",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "7vw64yd4mec318d520b9dc40cce03e69b0500055b628c45659", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "csharp", "selftext": "I apologize if I don't explain this properly. I'll do my best to make sense of it. I'm fairly new into LINQ.\n\nI have two queries, the first is the original query that has the information I need, the second is the query that contains all the IDs. I want to compare those two queries using an Except or equivalent, to remove the IDs that are present in the second query from the first.\n\nThis is what I have so far.\n\nHere is my first query:\n\n    var QueryOne = (from T1 in Table1\n                    join T2 in Table2 on T1.ID equals T2.ID\n                    join T3 in Table3 on T2.Name equals T3.Name\n                    join T4 in Table4 on T1.File equals T4.File\n                    where\n                    ( WHERE CLAUSE HERE )\n                    select new\n                    {\n                    ID = T1.ID.ToString(),\n                    Name = T1.Name(),\n                    Loc = T1.LocationOfDocument.ToString(),\n                    Email = T1.EmailDomain.ToString(),\n                    EmailName = T1.EmailName.ToString(),\n                    OtherID = T2.ID  \n                    }\n                    ).Distinct().ToList();\n    \n    QueryOne.Dump();\n\nHere is my second query:\n\n    var QueryTwo = (from T5 in Table5\n                    join T3 in Table3 on T5.ID equals T3.ID\n                    join T2 in Table2 on T3.Name equals T2.Name\n                    select new\n                    {\n                    ID = (int)T5.ID,\n                    OtherID = T2.ID \n                    }\n                    ).Distinct().ToList();\n    \n    QueryTwo.Dump();\n\nNow I want to compare those two, remove all IDs that are present in both queries. The only issue I have is that I don't know how to return two selects from the third variable.\n\nThis is what the third part looks like:\n\n    var result = QueryOne.Select(x =&gt; new { x.OtherID, x.ID }).Except(QueryTwo.Select( x =&gt; new { x.OtherID, x.ID }));\n    \n    result.Dump();\n\nI'm testing all of this in LINQPad, which is why it's formatted that way.\n\nThe error I receive is as follows:\n\n    CS1061: 'List&lt;anonymous type: string ID, string Name, string etc...&gt;' does not contain a definition for 'Table' and no accessible extension method ' ' accepting a first argument of type 'List&lt;&gt;' could be found (press f4 to add an assembly reference or import a namespace)\n\nWhen I press F4 it then throws an error CS1929 IEnumerable&lt;anonymous type&gt; does not contain a definition for 'Except'.\n\nAm I missing something obvious, or am I trying to do something that isn't possible?\n\nAny helps is greatly appreciated. I've been stuck on this and I've read multiple StackOverflow questions, articles, pretty much everything I could find online. I'm just not sure how to search for this specific question.\n\nThank you all.", "author_fullname": "t2_u8vyj", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "How to return multiple columns when comparing two LINQ queries?", "link_flair_richtext": [], "subreddit_name_prefixed": "r/csharp", "hidden": false, "pwls": 6, "link_flair_css_class": "help", "downs": 0, "thumbnail_height": null, "top_awarded_type": null, "hide_score": false, "name": "t3_ww748q", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 1.0, "author_flair_background_color": null, "subreddit_type": "public", "ups": 1, "total_awards_received": 0, "media_embed": {}, "thumbnail_width": null, "author_flair_template_id": null, "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": "Help", "can_mod_post": false, "score": 1, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "self", "edited": 1661307456.0, "author_flair_css_class": null, "author_flair_richtext": [], "gildings": {}, "content_categories": null, "is_self": true, "mod_note": null, "created": 1661307205.0, "link_flair_type": "text", "wls": 6, "removed_by_category": null, "banned_by": null, "author_flair_type": "text", "domain": "self.csharp", "allow_live_comments": false, "selftext_html": "&lt;!-- SC_OFF --&gt;&lt;div class=\"md\"&gt;&lt;p&gt;I apologize if I don&amp;#39;t explain this properly. I&amp;#39;ll do my best to make sense of it. I&amp;#39;m fairly new into LINQ.&lt;/p&gt;\n\n&lt;p&gt;I have two queries, the first is the original query that has the information I need, the second is the query that contains all the IDs. I want to compare those two queries using an Except or equivalent, to remove the IDs that are present in the second query from the first.&lt;/p&gt;\n\n&lt;p&gt;This is what I have so far.&lt;/p&gt;\n\n&lt;p&gt;Here is my first query:&lt;/p&gt;\n\n&lt;pre&gt;&lt;code&gt;var QueryOne = (from T1 in Table1\n                join T2 in Table2 on T1.ID equals T2.ID\n                join T3 in Table3 on T2.Name equals T3.Name\n                join T4 in Table4 on T1.File equals T4.File\n                where\n                ( WHERE CLAUSE HERE )\n                select new\n                {\n                ID = T1.ID.ToString(),\n                Name = T1.Name(),\n                Loc = T1.LocationOfDocument.ToString(),\n                Email = T1.EmailDomain.ToString(),\n                EmailName = T1.EmailName.ToString(),\n                OtherID = T2.ID  \n                }\n                ).Distinct().ToList();\n\nQueryOne.Dump();\n&lt;/code&gt;&lt;/pre&gt;\n\n&lt;p&gt;Here is my second query:&lt;/p&gt;\n\n&lt;pre&gt;&lt;code&gt;var QueryTwo = (from T5 in Table5\n                join T3 in Table3 on T5.ID equals T3.ID\n                join T2 in Table2 on T3.Name equals T2.Name\n                select new\n                {\n                ID = (int)T5.ID,\n                OtherID = T2.ID \n                }\n                ).Distinct().ToList();\n\nQueryTwo.Dump();\n&lt;/code&gt;&lt;/pre&gt;\n\n&lt;p&gt;Now I want to compare those two, remove all IDs that are present in both queries. The only issue I have is that I don&amp;#39;t know how to return two selects from the third variable.&lt;/p&gt;\n\n&lt;p&gt;This is what the third part looks like:&lt;/p&gt;\n\n&lt;pre&gt;&lt;code&gt;var result = QueryOne.Select(x =&amp;gt; new { x.OtherID, x.ID }).Except(QueryTwo.Select( x =&amp;gt; new { x.OtherID, x.ID }));\n\nresult.Dump();\n&lt;/code&gt;&lt;/pre&gt;\n\n&lt;p&gt;I&amp;#39;m testing all of this in LINQPad, which is why it&amp;#39;s formatted that way.&lt;/p&gt;\n\n&lt;p&gt;The error I receive is as follows:&lt;/p&gt;\n\n&lt;pre&gt;&lt;code&gt;CS1061: &amp;#39;List&amp;lt;anonymous type: string ID, string Name, string etc...&amp;gt;&amp;#39; does not contain a definition for &amp;#39;Table&amp;#39; and no accessible extension method &amp;#39; &amp;#39; accepting a first argument of type &amp;#39;List&amp;lt;&amp;gt;&amp;#39; could be found (press f4 to add an assembly reference or import a namespace)\n&lt;/code&gt;&lt;/pre&gt;\n\n&lt;p&gt;When I press F4 it then throws an error CS1929 IEnumerable&amp;lt;anonymous type&amp;gt; does not contain a definition for &amp;#39;Except&amp;#39;.&lt;/p&gt;\n\n&lt;p&gt;Am I missing something obvious, or am I trying to do something that isn&amp;#39;t possible?&lt;/p&gt;\n\n&lt;p&gt;Any helps is greatly appreciated. I&amp;#39;ve been stuck on this and I&amp;#39;ve read multiple StackOverflow questions, articles, pretty much everything I could find online. I&amp;#39;m just not sure how to search for this specific question.&lt;/p&gt;\n\n&lt;p&gt;Thank you all.&lt;/p&gt;\n&lt;/div&gt;&lt;!-- SC_ON --&gt;", "likes": null, "suggested_sort": null, "banned_at_utc": null, "view_count": null, "archived": false, "no_follow": true, "is_crosspostable": true, "pinned": false, "over_18": false, "all_awardings": [], "awarders": [], "media_only": false, "link_flair_template_id": "e2a3c0c0-e356-11e4-93d9-22000b6f0317", "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": null, "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2qhdf", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "ww748q", "is_robot_indexable": true, "report_reasons": null, "author": "mister_peachmango", "discussion_type": null, "num_comments": 1, "send_replies": true, "whitelist_status": "all_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": null, "permalink": "/r/csharp/comments/ww748q/how_to_return_multiple_columns_when_comparing_two/", "parent_whitelist_status": "all_ads", "stickied": false, "url": "https://www.reddit.com/r/csharp/comments/ww748q/how_to_return_multiple_columns_when_comparing_two/", "subreddit_subscribers": 205784, "created_utc": 1661307205.0, "num_crossposts": 0, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultText{
				Title: "How to return multiple columns when comparing two LINQ queries?",
				Text:  "I apologize if I don't explain this properly. I'll do my best to make sense of it. I'm fairly new into LINQ.\n\nI have two queries, the first is the original query that has the information I need, the second is the query that contains all the IDs. I want to compare those two queries using an Except or equivalent, to remove the IDs that are present in the second query from the first.\n\nThis is what I have so far.\n\nHere is my first query:\n\n    var QueryOne = (from T1 in Table1\n                    join T2 in Table2 on T1.ID equals T2.ID\n                    join T3 in Table3 on T2.Name equals T3.Name\n                    join T4 in Table4 on T1.File equals T4.File\n                    where\n                    ( WHERE CLAUSE HERE )\n                    select new\n                    {\n                    ID = T1.ID.ToString(),\n                    Name = T1.Name(),\n                    Loc = T1.LocationOfDocument.ToString(),\n                    Email = T1.EmailDomain.ToString(),\n                    EmailName = T1.EmailName.ToString(),\n                    OtherID = T2.ID  \n                    }\n                    ).Distinct().ToList();\n    \n    QueryOne.Dump();\n\nHere is my second query:\n\n    var QueryTwo = (from T5 in Table5\n                    join T3 in Table3 on T5.ID equals T3.ID\n                    join T2 in Table2 on T3.Name equals T2.Name\n                    select new\n                    {\n                    ID = (int)T5.ID,\n                    OtherID = T2.ID \n                    }\n                    ).Distinct().ToList();\n    \n    QueryTwo.Dump();\n\nNow I want to compare those two, remove all IDs that are present in both queries. The only issue I have is that I don't know how to return two selects from the third variable.\n\nThis is what the third part looks like:\n\n    var result = QueryOne.Select(x => new { x.OtherID, x.ID }).Except(QueryTwo.Select( x => new { x.OtherID, x.ID }));\n    \n    result.Dump();\n\nI'm testing all of this in LINQPad, which is why it's formatted that way.\n\nThe error I receive is as follows:\n\n    CS1061: 'List<anonymous type: string ID, string Name, string etc...>' does not contain a definition for 'Table' and no accessible extension method ' ' accepting a first argument of type 'List<>' could be found (press f4 to add an assembly reference or import a namespace)\n\nWhen I press F4 it then throws an error CS1929 IEnumerable<anonymous type> does not contain a definition for 'Except'.\n\nAm I missing something obvious, or am I trying to do something that isn't possible?\n\nAny helps is greatly appreciated. I've been stuck on this and I've read multiple StackOverflow questions, articles, pretty much everything I could find online. I'm just not sure how to search for this specific question.\n\nThank you all.",
			},
			ExpectedError: nil,
		},
		{
			TestName: "Image (Reddit Hosted)",
			PostUrl:  "https://www.reddit.com/r/dankmemes/comments/wvuvup/the_truth_has_been_spoken/",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "lymh2sfbiw2418274285ce9cc377261177eb76ed2936359ed5", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "dankmemes", "selftext": "", "author_fullname": "t2_4vwy0yf8", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "The truth has been spoken", "link_flair_richtext": [], "subreddit_name_prefixed": "r/dankmemes", "hidden": false, "pwls": 0, "link_flair_css_class": null, "downs": 0, "thumbnail_height": 140, "top_awarded_type": null, "hide_score": false, "name": "t3_wvuvup", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.99, "author_flair_background_color": null, "subreddit_type": "public", "ups": 4855, "total_awards_received": 1, "media_embed": {}, "thumbnail_width": 140, "author_flair_template_id": null, "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": true, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": null, "can_mod_post": false, "score": 4855, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "https://a.thumbs.redditmedia.com/UDGzZFhGg0-YZgQH8URyowLYyN9k1cz7QRa0iDRq9k0.jpg", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [], "gildings": {}, "post_hint": "image", "content_categories": null, "is_self": false, "mod_note": null, "created": 1661276100.0, "link_flair_type": "text", "wls": 0, "removed_by_category": null, "banned_by": null, "author_flair_type": "text", "domain": "i.redd.it", "allow_live_comments": false, "selftext_html": null, "likes": true, "suggested_sort": "top", "banned_at_utc": null, "url_overridden_by_dest": "https://i.redd.it/kk1x0xw81ij91.jpg", "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "preview": {"images": [{"source": {"url": "https://preview.redd.it/kk1x0xw81ij91.jpg?auto=webp&amp;s=1692ba46047ae5e4e2f315bde8a00f7e4a8c5759", "width": 750, "height": 929}, "resolutions": [{"url": "https://preview.redd.it/kk1x0xw81ij91.jpg?width=108&amp;crop=smart&amp;auto=webp&amp;s=d221522a953550ca2059ba069e8acd32c20fc7dc", "width": 108, "height": 133}, {"url": "https://preview.redd.it/kk1x0xw81ij91.jpg?width=216&amp;crop=smart&amp;auto=webp&amp;s=3caf98415dacb066ed3b3f30794e2df28078e8c7", "width": 216, "height": 267}, {"url": "https://preview.redd.it/kk1x0xw81ij91.jpg?width=320&amp;crop=smart&amp;auto=webp&amp;s=37245e7eab0a70474f0f797b5def47b527661b65", "width": 320, "height": 396}, {"url": "https://preview.redd.it/kk1x0xw81ij91.jpg?width=640&amp;crop=smart&amp;auto=webp&amp;s=36a64baba57f781bfb25e140ae91fca6a72ff63a", "width": 640, "height": 792}], "variants": {}, "id": "XRwHI5HyqbPWJ-b-Zkd7suFchJMIyUHiRbNl6yh2raI"}], "enabled": true}, "all_awardings": [{"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 125, "id": "award_5f123e3d-4f48-42f4-9c11-e98b566d5897", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=16&amp;height=16&amp;auto=webp&amp;s=92932f465d58e4c16b12b6eac4ca07d27e3d11c0", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=32&amp;height=32&amp;auto=webp&amp;s=d11484a208d68a318bf9d4fcf371171a1cb6a7ef", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=48&amp;height=48&amp;auto=webp&amp;s=febdf28b6f39f7da7eb1365325b85e0bb49a9f63", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=64&amp;height=64&amp;auto=webp&amp;s=b4406a2d88bf86fa3dc8a45aacf7e0c7bdccc4fb", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=128&amp;height=128&amp;auto=webp&amp;s=19555b13e3e196b62eeb9160d1ac1d1b372dcb0b", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "When you come across a feel-good thing.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 1, "static_icon_height": 2048, "name": "Wholesome", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=16&amp;height=16&amp;auto=webp&amp;s=92932f465d58e4c16b12b6eac4ca07d27e3d11c0", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=32&amp;height=32&amp;auto=webp&amp;s=d11484a208d68a318bf9d4fcf371171a1cb6a7ef", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=48&amp;height=48&amp;auto=webp&amp;s=febdf28b6f39f7da7eb1365325b85e0bb49a9f63", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=64&amp;height=64&amp;auto=webp&amp;s=b4406a2d88bf86fa3dc8a45aacf7e0c7bdccc4fb", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=128&amp;height=128&amp;auto=webp&amp;s=19555b13e3e196b62eeb9160d1ac1d1b372dcb0b", "width": 128, "height": 128}], "icon_format": null, "icon_height": 2048, "penny_price": null, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png"}], "awarders": [], "media_only": false, "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": null, "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2zmfe", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "wvuvup", "is_robot_indexable": true, "report_reasons": null, "author": "mitus376", "discussion_type": null, "num_comments": 27, "send_replies": true, "whitelist_status": "no_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": null, "permalink": "/r/dankmemes/comments/wvuvup/the_truth_has_been_spoken/", "parent_whitelist_status": "no_ads", "stickied": false, "url": "https://i.redd.it/kk1x0xw81ij91.jpg", "subreddit_subscribers": 5799187, "created_utc": 1661276100.0, "num_crossposts": 0, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultMedia{
				Medias: []FetchResultMediaEntry{
					{
						Link:    "https://i.redd.it/kk1x0xw81ij91.jpg",
						Quality: "Original",
					},
					{
						Link:    "https://preview.redd.it/kk1x0xw81ij91.jpg?auto=webp&s=1692ba46047ae5e4e2f315bde8a00f7e4a8c5759",
						Quality: "750×929",
					},
					{
						Link:    "https://preview.redd.it/kk1x0xw81ij91.jpg?width=640&crop=smart&auto=webp&s=36a64baba57f781bfb25e140ae91fca6a72ff63a",
						Quality: "640×792",
					},
					{
						Link:    "https://preview.redd.it/kk1x0xw81ij91.jpg?width=320&crop=smart&auto=webp&s=37245e7eab0a70474f0f797b5def47b527661b65",
						Quality: "320×396",
					},
					{
						Link:    "https://preview.redd.it/kk1x0xw81ij91.jpg?width=216&crop=smart&auto=webp&s=3caf98415dacb066ed3b3f30794e2df28078e8c7",
						Quality: "216×267",
					},
					{
						Link:    "https://preview.redd.it/kk1x0xw81ij91.jpg?width=108&crop=smart&auto=webp&s=d221522a953550ca2059ba069e8acd32c20fc7dc",
						Quality: "108×133",
					},
				},
				ThumbnailLink: "https://a.thumbs.redditmedia.com/UDGzZFhGg0-YZgQH8URyowLYyN9k1cz7QRa0iDRq9k0.jpg",
				Title:         "The truth has been spoken",
				Type:          FetchResultMediaTypePhoto,
			},
			ExpectedError: nil,
		},
		{
			TestName: "Image (From imgur)",
			PostUrl:  "https://www.reddit.com/r/dankmemes/comments/iozykz/you_what_now/?utm_source=share&utm_medium=web2x&context=3",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "3v4ass0fnmb9091e87ff9fd856e22409c7e55f76437407bd84", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "dankmemes", "selftext": "", "author_fullname": "t2_od9dtcj", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "you what now?", "link_flair_richtext": [{"e": "text", "t": "Halal Meme"}], "subreddit_name_prefixed": "r/dankmemes", "hidden": false, "pwls": 0, "link_flair_css_class": "", "downs": 0, "thumbnail_height": 101, "top_awarded_type": null, "hide_score": false, "name": "t3_iozykz", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.91, "author_flair_background_color": "#46d160", "ups": 29, "domain": "i.imgur.com", "media_embed": {}, "thumbnail_width": 140, "author_flair_template_id": "83725ee6-b662-11e6-8296-0e342ab1a566", "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": "Halal Meme", "can_mod_post": false, "score": 29, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "https://a.thumbs.redditmedia.com/Jpsd715ZqmYgIl5pJFjLw8FieykuRVbPBqs1rmqrYY8.jpg", "edited": false, "author_flair_css_class": "green", "author_flair_richtext": [{"e": "text", "t": "Green"}], "gildings": {}, "post_hint": "image", "content_categories": null, "is_self": false, "subreddit_type": "public", "created": 1599592252.0, "link_flair_type": "richtext", "wls": 0, "removed_by_category": null, "banned_by": null, "author_flair_type": "richtext", "total_awards_received": 0, "allow_live_comments": false, "selftext_html": null, "likes": true, "suggested_sort": "top", "banned_at_utc": null, "url_overridden_by_dest": "https://i.imgur.com/cP5n0Kz.jpg", "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "preview": {"images": [{"source": {"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?auto=webp&amp;s=94e9b63ae66d85297049171604a9f7e8ec872326", "width": 2787, "height": 2022}, "resolutions": [{"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=108&amp;crop=smart&amp;auto=webp&amp;s=549de4c6a376f37ca5e5b09bb91caa13f6340246", "width": 108, "height": 78}, {"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=216&amp;crop=smart&amp;auto=webp&amp;s=6dd4625922fd88f8916ffa0f9db75802390c0fa8", "width": 216, "height": 156}, {"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=320&amp;crop=smart&amp;auto=webp&amp;s=29a4e06679c433c68b53486a0c99a00dd3bf5ab7", "width": 320, "height": 232}, {"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=640&amp;crop=smart&amp;auto=webp&amp;s=cec0e9e13e9ee60a83154b1c8ef506dc3232354a", "width": 640, "height": 464}, {"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=960&amp;crop=smart&amp;auto=webp&amp;s=89d0687a852e5a539e77595da9ae972ee3a6f17a", "width": 960, "height": 696}, {"url": "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=1080&amp;crop=smart&amp;auto=webp&amp;s=32fa1cc26cd06eb1dcbf75adc01547ed77c5b59f", "width": 1080, "height": 783}], "variants": {}, "id": "C0S7mtilTdy1tlRZxWIJXCAgzdWqaqelKanuxasC_d0"}], "enabled": true}, "all_awardings": [], "awarders": [], "media_only": false, "link_flair_template_id": "937e65dc-8bd9-11ea-8aeb-0e0baa35d8ff", "can_gild": false, "spoiler": false, "locked": false, "author_flair_text": "Green", "treatment_tags": [], "rte_mode": "markdown", "visited": false, "removed_by": null, "mod_note": null, "distinguished": null, "subreddit_id": "t5_2zmfe", "author_is_blocked": false, "mod_reason_by": null, "num_reports": null, "removal_reason": null, "link_flair_background_color": "#349e48", "id": "iozykz", "is_robot_indexable": true, "report_reasons": null, "author": "HirbodBehnam", "discussion_type": null, "num_comments": 3, "send_replies": false, "whitelist_status": "no_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": "dark", "permalink": "/r/dankmemes/comments/iozykz/you_what_now/", "parent_whitelist_status": "no_ads", "stickied": false, "url": "https://i.imgur.com/cP5n0Kz.jpg", "subreddit_subscribers": 5799225, "created_utc": 1599592252.0, "num_crossposts": 0, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultMedia{
				Medias: []FetchResultMediaEntry{
					{
						Link:    "https://i.imgur.com/cP5n0Kz.jpg",
						Quality: "Original",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?auto=webp&s=94e9b63ae66d85297049171604a9f7e8ec872326",
						Quality: "2787×2022",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=1080&crop=smart&auto=webp&s=32fa1cc26cd06eb1dcbf75adc01547ed77c5b59f",
						Quality: "1080×783",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=960&crop=smart&auto=webp&s=89d0687a852e5a539e77595da9ae972ee3a6f17a",
						Quality: "960×696",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=640&crop=smart&auto=webp&s=cec0e9e13e9ee60a83154b1c8ef506dc3232354a",
						Quality: "640×464",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=320&crop=smart&auto=webp&s=29a4e06679c433c68b53486a0c99a00dd3bf5ab7",
						Quality: "320×232",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=216&crop=smart&auto=webp&s=6dd4625922fd88f8916ffa0f9db75802390c0fa8",
						Quality: "216×156",
					},
					{
						Link:    "https://external-preview.redd.it/gJthFa1BIb3_ku0tYIgyrQr22I4oDKIl7QlUa-CzJak.jpg?width=108&crop=smart&auto=webp&s=549de4c6a376f37ca5e5b09bb91caa13f6340246",
						Quality: "108×78",
					},
				},
				ThumbnailLink: "https://a.thumbs.redditmedia.com/Jpsd715ZqmYgIl5pJFjLw8FieykuRVbPBqs1rmqrYY8.jpg",
				Title:         "you what now?",
				Type:          FetchResultMediaTypePhoto,
			},
		},
		{
			TestName: "Imgur Gif",
			PostUrl:  "https://www.reddit.com/r/dankmemes/comments/gag117/you_daughter_of_a_bitch_im_in/",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "cdmabtqd4s1d552db032821df930cdc239d90a26f0f2bb79bb", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "dankmemes", "selftext": "", "author_fullname": "t2_3phxk2ir", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "You daughter of a bitch, I'm in.", "link_flair_richtext": [], "subreddit_name_prefixed": "r/dankmemes", "hidden": false, "pwls": 0, "link_flair_css_class": null, "downs": 0, "thumbnail_height": 140, "top_awarded_type": null, "hide_score": false, "name": "t3_gag117", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.97, "author_flair_background_color": "#7193ff", "subreddit_type": "public", "ups": 11199, "total_awards_received": 0, "media_embed": {}, "thumbnail_width": 140, "author_flair_template_id": "c9b5e252-37e1-11ea-99b6-0e79a5debf93", "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": null, "can_mod_post": false, "score": 11199, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": true, "thumbnail": "https://b.thumbs.redditmedia.com/OZVSbT-X1eTPZADOoF4l8ZoYcKC_dWxQ-DTBbdcINLU.jpg", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [{"a": ":bruh:", "e": "emoji", "u": "https://emoji.redditmedia.com/vwanncag81b41_t5_2zmfe/bruh"}, {"e": "text", "t": " makes good maymays"}], "gildings": {}, "post_hint": "image", "content_categories": null, "is_self": false, "mod_note": null, "created": 1588189037.0, "link_flair_type": "text", "wls": 0, "removed_by_category": null, "banned_by": null, "author_flair_type": "richtext", "domain": "i.imgur.com", "allow_live_comments": true, "selftext_html": null, "likes": true, "suggested_sort": "top", "banned_at_utc": null, "url_overridden_by_dest": "https://i.imgur.com/QdBe1Vw.gif", "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "preview": {"images": [{"source": {"url": "https://external-preview.redd.it/BFq8Eg5rbtrOYxMrF63cP3t30AwAnWgXch-OpLAvO4E.jpg?auto=webp&amp;s=9dcb8c2318b68ebb97867ee161d8428bee871b65", "width": 686, "height": 854}, "resolutions": [{"url": "https://external-preview.redd.it/BFq8Eg5rbtrOYxMrF63cP3t30AwAnWgXch-OpLAvO4E.jpg?width=108&amp;crop=smart&amp;auto=webp&amp;s=de260faefaf40d443495605a8c26ade63db488cf", "width": 108, "height": 134}, {"url": "https://external-preview.redd.it/BFq8Eg5rbtrOYxMrF63cP3t30AwAnWgXch-OpLAvO4E.jpg?width=216&amp;crop=smart&amp;auto=webp&amp;s=07548bd941fe6d33790be90c5785f5c60a758447", "width": 216, "height": 268}, {"url": "https://external-preview.redd.it/BFq8Eg5rbtrOYxMrF63cP3t30AwAnWgXch-OpLAvO4E.jpg?width=320&amp;crop=smart&amp;auto=webp&amp;s=27f6532299427eb2f18d2a3dd78a600c8b8128d8", "width": 320, "height": 398}, {"url": "https://external-preview.redd.it/BFq8Eg5rbtrOYxMrF63cP3t30AwAnWgXch-OpLAvO4E.jpg?width=640&amp;crop=smart&amp;auto=webp&amp;s=fcfe2343362ea55e8d808b37cf76ea99f8374517", "width": 640, "height": 796}], "variants": {}, "id": "VkYaG4vVpjhHqr0UAM-8HwmAONfamiDk3UPxhdqjMk0"}], "enabled": true}, "all_awardings": [], "awarders": [], "media_only": false, "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": ":bruh: makes good maymays", "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2zmfe", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "gag117", "is_robot_indexable": true, "report_reasons": null, "author": "mijuzz7", "discussion_type": null, "num_comments": 36, "send_replies": true, "whitelist_status": "no_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": "dark", "permalink": "/r/dankmemes/comments/gag117/you_daughter_of_a_bitch_im_in/", "parent_whitelist_status": "no_ads", "stickied": false, "url": "https://i.imgur.com/QdBe1Vw.gif", "subreddit_subscribers": 5799232, "created_utc": 1588189037.0, "num_crossposts": 5, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultMedia{
				Medias: []FetchResultMediaEntry{{
					Link:    "https://i.imgur.com/download/QdBe1Vw.gif",
					Quality: "imgur",
				}},
				ThumbnailLink: "https://b.thumbs.redditmedia.com/OZVSbT-X1eTPZADOoF4l8ZoYcKC_dWxQ-DTBbdcINLU.jpg",
				Title:         "You daughter of a bitch, I'm in.",
				Type:          FetchResultMediaTypeGif,
			},
			ExpectedError: nil,
		},
		{
			TestName: "Reddit Gif",
			PostUrl:  "https://www.reddit.com/r/dankmemes/comments/wvz9qd/i_gotta_do_this_more_often/",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "bja5wplpjs8b4465f708caaf21d628831d903abd2e679e8503", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "dankmemes", "selftext": "", "author_fullname": "t2_8axfzuwn", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "I gotta do this more often", "link_flair_richtext": [], "subreddit_name_prefixed": "r/dankmemes", "hidden": false, "pwls": 0, "link_flair_css_class": null, "downs": 0, "thumbnail_height": 140, "top_awarded_type": null, "hide_score": false, "name": "t3_wvz9qd", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.98, "author_flair_background_color": "#000000", "subreddit_type": "public", "ups": 11369, "total_awards_received": 2, "media_embed": {}, "thumbnail_width": 140, "author_flair_template_id": "88bb5232-43c0-11eb-a884-0ee48fee3991", "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": true, "is_meta": false, "category": null, "secure_media_embed": {}, "link_flair_text": null, "can_mod_post": false, "score": 11369, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": true, "thumbnail": "https://b.thumbs.redditmedia.com/hSd7rznf5UKJMeVMdR-2u06tCdbfTWqgZsTgkLuibPk.jpg", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [{"a": ":maymay:", "e": "emoji", "u": "https://emoji.redditmedia.com/m8s21j1513p51_t5_2zmfe/maymay"}, {"e": "text", "t": " MayMayMakers "}, {"a": ":maymay:", "e": "emoji", "u": "https://emoji.redditmedia.com/m8s21j1513p51_t5_2zmfe/maymay"}], "gildings": {"gid_1": 1}, "post_hint": "image", "content_categories": null, "is_self": false, "mod_note": null, "created": 1661286775.0, "link_flair_type": "text", "wls": 0, "removed_by_category": null, "banned_by": null, "author_flair_type": "richtext", "domain": "i.redd.it", "allow_live_comments": false, "selftext_html": null, "likes": true, "suggested_sort": "top", "banned_at_utc": null, "url_overridden_by_dest": "https://i.redd.it/jp4owaxuwij91.gif", "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "preview": {"images": [{"source": {"url": "https://preview.redd.it/jp4owaxuwij91.gif?format=png8&amp;s=6e600c08043863f074d48763846ed961f832d401", "width": 600, "height": 602}, "resolutions": [{"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=108&amp;crop=smart&amp;format=png8&amp;s=abcd150c22696766c3fee111c420a6355c60a9ea", "width": 108, "height": 108}, {"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=216&amp;crop=smart&amp;format=png8&amp;s=566447ed95a701c1fa681e312647b04020e00415", "width": 216, "height": 216}, {"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=320&amp;crop=smart&amp;format=png8&amp;s=2c7122140fe310ba474844edc76cec066d0f2ddd", "width": 320, "height": 321}], "variants": {"gif": {"source": {"url": "https://preview.redd.it/jp4owaxuwij91.gif?s=3f418e92089220ee45652866da4b096a3f3e6d72", "width": 600, "height": 602}, "resolutions": [{"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=108&amp;crop=smart&amp;s=690185a6df4a4da0a7c1e7fb1cd838241226c34d", "width": 108, "height": 108}, {"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=216&amp;crop=smart&amp;s=f8fbadec4ae6f5a3966351b8c991377d26de849b", "width": 216, "height": 216}, {"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=320&amp;crop=smart&amp;s=dd0771918e3f7fd781dd59ef92449b6a1c1b7fba", "width": 320, "height": 321}]}, "mp4": {"source": {"url": "https://preview.redd.it/jp4owaxuwij91.gif?format=mp4&amp;s=8275fede28d85ec9c97a17d0c9f74ded167c3bf3", "width": 600, "height": 602}, "resolutions": [{"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=108&amp;format=mp4&amp;s=da93ba9b859ec1b417006ba4eb9094e062ee2681", "width": 108, "height": 108}, {"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=216&amp;format=mp4&amp;s=1b9b3cafecb83bbf373e9bad12220507dbf263d5", "width": 216, "height": 216}, {"url": "https://preview.redd.it/jp4owaxuwij91.gif?width=320&amp;format=mp4&amp;s=17c618edb312b236a3d530ab55c62bbb5e7fef6f", "width": 320, "height": 321}]}}, "id": "dPPU3U2CL2_bZX2VLmMxsjxhf00W0yzLtTn2XBzr_6o"}], "enabled": true}, "all_awardings": [{"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 100, "id": "gid_1", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://www.redditstatic.com/gold/awards/icon/silver_512.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://www.redditstatic.com/gold/awards/icon/silver_16.png", "width": 16, "height": 16}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_32.png", "width": 32, "height": 32}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_48.png", "width": 48, "height": 48}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_64.png", "width": 64, "height": 64}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_128.png", "width": 128, "height": 128}], "icon_width": 512, "static_icon_width": 512, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "Shows the Silver Award... and that's it.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 1, "static_icon_height": 512, "name": "Silver", "resized_static_icons": [{"url": "https://www.redditstatic.com/gold/awards/icon/silver_16.png", "width": 16, "height": 16}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_32.png", "width": 32, "height": 32}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_48.png", "width": 48, "height": 48}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_64.png", "width": 64, "height": 64}, {"url": "https://www.redditstatic.com/gold/awards/icon/silver_128.png", "width": 128, "height": 128}], "icon_format": null, "icon_height": 512, "penny_price": null, "award_type": "global", "static_icon_url": "https://www.redditstatic.com/gold/awards/icon/silver_512.png"}, {"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 125, "id": "award_5f123e3d-4f48-42f4-9c11-e98b566d5897", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=16&amp;height=16&amp;auto=webp&amp;s=92932f465d58e4c16b12b6eac4ca07d27e3d11c0", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=32&amp;height=32&amp;auto=webp&amp;s=d11484a208d68a318bf9d4fcf371171a1cb6a7ef", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=48&amp;height=48&amp;auto=webp&amp;s=febdf28b6f39f7da7eb1365325b85e0bb49a9f63", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=64&amp;height=64&amp;auto=webp&amp;s=b4406a2d88bf86fa3dc8a45aacf7e0c7bdccc4fb", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=128&amp;height=128&amp;auto=webp&amp;s=19555b13e3e196b62eeb9160d1ac1d1b372dcb0b", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "When you come across a feel-good thing.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 1, "static_icon_height": 2048, "name": "Wholesome", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=16&amp;height=16&amp;auto=webp&amp;s=92932f465d58e4c16b12b6eac4ca07d27e3d11c0", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=32&amp;height=32&amp;auto=webp&amp;s=d11484a208d68a318bf9d4fcf371171a1cb6a7ef", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=48&amp;height=48&amp;auto=webp&amp;s=febdf28b6f39f7da7eb1365325b85e0bb49a9f63", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=64&amp;height=64&amp;auto=webp&amp;s=b4406a2d88bf86fa3dc8a45aacf7e0c7bdccc4fb", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=128&amp;height=128&amp;auto=webp&amp;s=19555b13e3e196b62eeb9160d1ac1d1b372dcb0b", "width": 128, "height": 128}], "icon_format": null, "icon_height": 2048, "penny_price": null, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png"}], "awarders": [], "media_only": false, "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": ":maymay: MayMayMakers :maymay:", "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2zmfe", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "wvz9qd", "is_robot_indexable": true, "report_reasons": null, "author": "Alarmed-Ad-436", "discussion_type": null, "num_comments": 93, "send_replies": true, "whitelist_status": "no_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": "light", "permalink": "/r/dankmemes/comments/wvz9qd/i_gotta_do_this_more_often/", "parent_whitelist_status": "no_ads", "stickied": false, "url": "https://i.redd.it/jp4owaxuwij91.gif", "subreddit_subscribers": 5799235, "created_utc": 1661286775.0, "num_crossposts": 2, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultMedia{
				Medias: []FetchResultMediaEntry{
					{
						Link:    "https://preview.redd.it/jp4owaxuwij91.gif?format=mp4&s=8275fede28d85ec9c97a17d0c9f74ded167c3bf3",
						Quality: "600×602",
					},
					{
						Link:    "https://preview.redd.it/jp4owaxuwij91.gif?width=320&format=mp4&s=17c618edb312b236a3d530ab55c62bbb5e7fef6f",
						Quality: "320×321",
					},
					{
						Link:    "https://preview.redd.it/jp4owaxuwij91.gif?width=216&format=mp4&s=1b9b3cafecb83bbf373e9bad12220507dbf263d5",
						Quality: "216×216",
					},
					{
						Link:    "https://preview.redd.it/jp4owaxuwij91.gif?width=108&format=mp4&s=da93ba9b859ec1b417006ba4eb9094e062ee2681",
						Quality: "108×108",
					},
				},
				ThumbnailLink: "https://b.thumbs.redditmedia.com/hSd7rznf5UKJMeVMdR-2u06tCdbfTWqgZsTgkLuibPk.jpg",
				Title:         "I gotta do this more often",
				Type:          FetchResultMediaTypeGif,
			},
			ExpectedError: nil,
		},
		{
			TestName:       "Youtube",
			PostUrl:        "https://www.reddit.com/r/csharp/comments/wvt9an/i_really_hope_this_helps_demystify_authentication/?utm_source=share&utm_medium=web2x&context=3",
			Root:           `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "fhmeg6g02ibf3750193636f836cdcf0c603612b55414e27259", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "csharp", "selftext": "", "author_fullname": "t2_ricjx3od", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "title": "I really hope this helps demystify authentication!", "link_flair_richtext": [], "subreddit_name_prefixed": "r/csharp", "hidden": false, "pwls": 6, "link_flair_css_class": null, "downs": 0, "thumbnail_height": 105, "top_awarded_type": null, "hide_score": false, "name": "t3_wvt9an", "quarantine": false, "link_flair_text_color": "dark", "upvote_ratio": 0.91, "author_flair_background_color": null, "subreddit_type": "public", "ups": 27, "total_awards_received": 0, "media_embed": {"content": "&lt;iframe width=\"356\" height=\"200\" src=\"https://www.youtube.com/embed/7ILCRfPmQxQ?feature=oembed&amp;enablejsapi=1\" frameborder=\"0\" allow=\"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture\" allowfullscreen title=\"JWT Bearer Authentication in ASP.NET 6 |   CLEAN ARCHITECTURE &amp; DDD From Scratch Tutorial | Part 9\"&gt;&lt;/iframe&gt;", "width": 356, "scrolling": false, "height": 200}, "thumbnail_width": 140, "author_flair_template_id": null, "is_original_content": false, "user_reports": [], "secure_media": {"type": "youtube.com", "oembed": {"provider_url": "https://www.youtube.com/", "version": "1.0", "title": "JWT Bearer Authentication in ASP.NET 6 |   CLEAN ARCHITECTURE &amp; DDD From Scratch Tutorial | Part 9", "type": "video", "thumbnail_width": 480, "height": 200, "width": 356, "html": "&lt;iframe width=\"356\" height=\"200\" src=\"https://www.youtube.com/embed/7ILCRfPmQxQ?feature=oembed&amp;enablejsapi=1\" frameborder=\"0\" allow=\"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture\" allowfullscreen title=\"JWT Bearer Authentication in ASP.NET 6 |   CLEAN ARCHITECTURE &amp; DDD From Scratch Tutorial | Part 9\"&gt;&lt;/iframe&gt;", "author_name": "Amichai Mantinband", "provider_name": "YouTube", "thumbnail_url": "https://i.ytimg.com/vi/7ILCRfPmQxQ/hqdefault.jpg", "thumbnail_height": 360, "author_url": "https://www.youtube.com/c/AmichaiMantinband"}}, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {"content": "&lt;iframe width=\"356\" height=\"200\" src=\"https://www.youtube.com/embed/7ILCRfPmQxQ?feature=oembed&amp;enablejsapi=1\" frameborder=\"0\" allow=\"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture\" allowfullscreen title=\"JWT Bearer Authentication in ASP.NET 6 |   CLEAN ARCHITECTURE &amp; DDD From Scratch Tutorial | Part 9\"&gt;&lt;/iframe&gt;", "width": 356, "scrolling": false, "media_domain_url": "https://www.redditmedia.com/mediaembed/wvt9an", "height": 200}, "link_flair_text": null, "can_mod_post": false, "score": 27, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "https://b.thumbs.redditmedia.com/gCIsq_MRJ614E19XCI3TMQaEszAYouEWWkxUYcQt7TY.jpg", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [], "gildings": {}, "post_hint": "rich:video", "content_categories": null, "is_self": false, "mod_note": null, "created": 1661272160.0, "link_flair_type": "text", "wls": 6, "removed_by_category": null, "banned_by": null, "author_flair_type": "text", "domain": "youtu.be", "allow_live_comments": false, "selftext_html": null, "likes": null, "suggested_sort": null, "banned_at_utc": null, "url_overridden_by_dest": "https://youtu.be/7ILCRfPmQxQ", "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": false, "preview": {"images": [{"source": {"url": "https://external-preview.redd.it/PLfg_6dYzLk9cIwE2luebtIDqZFtEmezt933Up0Ye7M.jpg?auto=webp&amp;s=6dc4c4364e71b98e7a88df308aead13626fed89d", "width": 480, "height": 360}, "resolutions": [{"url": "https://external-preview.redd.it/PLfg_6dYzLk9cIwE2luebtIDqZFtEmezt933Up0Ye7M.jpg?width=108&amp;crop=smart&amp;auto=webp&amp;s=fa6f4819fdf33b5e226624c07a4b8a036cc70bd1", "width": 108, "height": 81}, {"url": "https://external-preview.redd.it/PLfg_6dYzLk9cIwE2luebtIDqZFtEmezt933Up0Ye7M.jpg?width=216&amp;crop=smart&amp;auto=webp&amp;s=d846c0685bd3af741a18d6ffa617df95a3555d36", "width": 216, "height": 162}, {"url": "https://external-preview.redd.it/PLfg_6dYzLk9cIwE2luebtIDqZFtEmezt933Up0Ye7M.jpg?width=320&amp;crop=smart&amp;auto=webp&amp;s=84020a0439b65681531c559a70978a6fd42573ca", "width": 320, "height": 240}], "variants": {}, "id": "DB9l3Nr5LlkFm_zCCzNn-XjdLz2WRa1CHCuwFAhFME0"}], "enabled": false}, "all_awardings": [], "awarders": [], "media_only": false, "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": null, "treatment_tags": [], "visited": false, "removed_by": null, "num_reports": null, "distinguished": null, "subreddit_id": "t5_2qhdf", "author_is_blocked": false, "mod_reason_by": null, "removal_reason": null, "link_flair_background_color": "", "id": "wvt9an", "is_robot_indexable": true, "report_reasons": null, "author": "amantinband", "discussion_type": null, "num_comments": 0, "send_replies": false, "whitelist_status": "all_ads", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": null, "permalink": "/r/csharp/comments/wvt9an/i_really_hope_this_helps_demystify_authentication/", "parent_whitelist_status": "all_ads", "stickied": false, "url": "https://youtu.be/7ILCRfPmQxQ", "subreddit_subscribers": 205793, "created_utc": 1661272160.0, "num_crossposts": 0, "media": {"type": "youtube.com", "oembed": {"provider_url": "https://www.youtube.com/", "version": "1.0", "title": "JWT Bearer Authentication in ASP.NET 6 |   CLEAN ARCHITECTURE &amp; DDD From Scratch Tutorial | Part 9", "type": "video", "thumbnail_width": 480, "height": 200, "width": 356, "html": "&lt;iframe width=\"356\" height=\"200\" src=\"https://www.youtube.com/embed/7ILCRfPmQxQ?feature=oembed&amp;enablejsapi=1\" frameborder=\"0\" allow=\"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture\" allowfullscreen title=\"JWT Bearer Authentication in ASP.NET 6 |   CLEAN ARCHITECTURE &amp; DDD From Scratch Tutorial | Part 9\"&gt;&lt;/iframe&gt;", "author_name": "Amichai Mantinband", "provider_name": "YouTube", "thumbnail_url": "https://i.ytimg.com/vi/7ILCRfPmQxQ/hqdefault.jpg", "thumbnail_height": 360, "author_url": "https://www.youtube.com/c/AmichaiMantinband"}}, "is_video": false}}], "before": null}}`,
			ExpectedResult: nil,
			ExpectedError: &FetchError{
				NormalError: "",
				BotError:    "This bot does not support downloading from youtu.be\nThe url field in json is https://youtu.be/7ILCRfPmQxQ",
			},
		},
		{
			TestName:      "Video (Reddit Hosted + Audio)",
			PostUrl:       "https://www.reddit.com/r/gtaonline/comments/wwc1to/when_youre_showing_a_low_level_around/",
			Root:          `{"kind":"Listing","data":{"after":null,"dist":1,"modhash":"oblfmd19kc33cd5a98f14b1ef49018c259d83514950d66197f","geo_filter":"","children":[{"kind":"t3","data":{"approved_at_utc":null,"subreddit":"gtaonline","selftext":"","author_fullname":"t2_11aksu","saved":false,"mod_reason_title":null,"gilded":0,"clicked":false,"title":"When you’re showing a low level around","link_flair_richtext":[{"a":":VID1:","e":"emoji","u":"https://emoji.redditmedia.com/77ocp6gzc8h81_t5_2xrd1/VID1"},{"a":":VID2:","e":"emoji","u":"https://emoji.redditmedia.com/7gyfpnhzc8h81_t5_2xrd1/VID2"},{"a":":VID3:","e":"emoji","u":"https://emoji.redditmedia.com/bf61dvjzc8h81_t5_2xrd1/VID3"}],"subreddit_name_prefixed":"r/gtaonline","hidden":false,"pwls":6,"link_flair_css_class":"VIDEO","downs":0,"thumbnail_height":140,"top_awarded_type":null,"hide_score":false,"name":"t3_wwc1to","quarantine":false,"link_flair_text_color":"light","upvote_ratio":0.99,"author_flair_background_color":null,"ups":405,"total_awards_received":1,"media_embed":{},"thumbnail_width":140,"author_flair_template_id":null,"is_original_content":false,"user_reports":[],"secure_media":{"reddit_video":{"bitrate_kbps":1200,"fallback_url":"https://v.redd.it/5scwdfq0wlj91/DASH_480.mp4?source=fallback","height":480,"width":424,"scrubber_media_url":"https://v.redd.it/5scwdfq0wlj91/DASH_96.mp4","dash_url":"https://v.redd.it/5scwdfq0wlj91/DASHPlaylist.mpd?a=1663925795%2CMDg3NzRhYzc2MmYwMjg2NTk1MDVmY2ZjYWNiNGM0MjA0MjcxYzRmMDIxYTIxNWRiMWEzMmNkZTU1NjYyNDBlZQ%3D%3D&amp;v=1&amp;f=hd","duration":5,"hls_url":"https://v.redd.it/5scwdfq0wlj91/HLSPlaylist.m3u8?a=1663925795%2CNzMyODVkZmJmOGJiNTE2MTYxM2Y4ZGU1OWI5YzZiZWFhN2VlMzczN2Y4OGM5OWQ5NjhjNTFjZDMxODA5OTY5ZA%3D%3D&amp;v=1&amp;f=hd","is_gif":false,"transcoding_status":"completed"}},"is_reddit_media_domain":true,"is_meta":false,"category":null,"secure_media_embed":{},"link_flair_text":":VID1::VID2::VID3:","can_mod_post":false,"score":405,"approved_by":null,"is_created_from_ads_ui":false,"author_premium":false,"thumbnail":"https://b.thumbs.redditmedia.com/8O1iFRVHNR-8Y83_OHCoSh-ME_oF1DQz3Bv8q7LjxQE.jpg","edited":false,"author_flair_css_class":null,"author_flair_richtext":[],"gildings":{"gid_1":1},"post_hint":"hosted:video","content_categories":null,"is_self":false,"subreddit_type":"public","created":1661322769.0,"link_flair_type":"richtext","wls":6,"removed_by_category":null,"banned_by":null,"author_flair_type":"text","domain":"v.redd.it","allow_live_comments":false,"selftext_html":null,"likes":true,"suggested_sort":null,"banned_at_utc":null,"url_overridden_by_dest":"https://v.redd.it/5scwdfq0wlj91","view_count":null,"archived":false,"no_follow":false,"is_crosspostable":true,"pinned":false,"over_18":false,"preview":{"images":[{"source":{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?format=pjpg&amp;auto=webp&amp;s=96d2b469bec8267d2ecf3f81fc94e37458afbc7d","width":432,"height":488},"resolutions":[{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?width=108&amp;crop=smart&amp;format=pjpg&amp;auto=webp&amp;s=29ffa06c84be32c10250e01ecf6f719e153fd52f","width":108,"height":121},{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?width=216&amp;crop=smart&amp;format=pjpg&amp;auto=webp&amp;s=54c41243cf27638af3b963a6e29e075b18c1430d","width":216,"height":243},{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?width=320&amp;crop=smart&amp;format=pjpg&amp;auto=webp&amp;s=86026d2a1948e66d8630070dad1764069c6ca033","width":320,"height":361}],"variants":{},"id":"Gtm9A6XbVxK0VKuKvysa9HmI0qmfMTAk9103eFZqTmE"}],"enabled":false},"all_awardings":[{"giver_coin_reward":null,"subreddit_id":null,"is_new":false,"days_of_drip_extension":null,"coin_price":100,"id":"gid_1","penny_donate":null,"award_sub_type":"GLOBAL","coin_reward":0,"icon_url":"https://www.redditstatic.com/gold/awards/icon/silver_512.png","days_of_premium":null,"tiers_by_required_awardings":null,"resized_icons":[{"url":"https://www.redditstatic.com/gold/awards/icon/silver_16.png","width":16,"height":16},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_32.png","width":32,"height":32},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_48.png","width":48,"height":48},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_64.png","width":64,"height":64},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_128.png","width":128,"height":128}],"icon_width":512,"static_icon_width":512,"start_date":null,"is_enabled":true,"awardings_required_to_grant_benefits":null,"description":"Shows the Silver Award... and that's it.","end_date":null,"sticky_duration_seconds":null,"subreddit_coin_reward":0,"count":1,"static_icon_height":512,"name":"Silver","resized_static_icons":[{"url":"https://www.redditstatic.com/gold/awards/icon/silver_16.png","width":16,"height":16},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_32.png","width":32,"height":32},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_48.png","width":48,"height":48},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_64.png","width":64,"height":64},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_128.png","width":128,"height":128}],"icon_format":null,"icon_height":512,"penny_price":null,"award_type":"global","static_icon_url":"https://www.redditstatic.com/gold/awards/icon/silver_512.png"}],"awarders":[],"media_only":false,"link_flair_template_id":"c7456812-be7a-11e3-a70f-12313d06c56f","can_gild":true,"spoiler":false,"locked":false,"author_flair_text":null,"treatment_tags":[],"visited":false,"removed_by":null,"mod_note":null,"distinguished":null,"subreddit_id":"t5_2xrd1","author_is_blocked":false,"mod_reason_by":null,"num_reports":null,"removal_reason":null,"link_flair_background_color":"#0079d3","id":"wwc1to","is_robot_indexable":true,"report_reasons":null,"author":"GreuDeFumat","discussion_type":null,"num_comments":12,"send_replies":true,"whitelist_status":"all_ads","contest_mode":false,"mod_reports":[],"author_patreon_flair":false,"author_flair_text_color":null,"permalink":"/r/gtaonline/comments/wwc1to/when_youre_showing_a_low_level_around/","parent_whitelist_status":"all_ads","stickied":false,"url":"https://v.redd.it/5scwdfq0wlj91","subreddit_subscribers":1281240,"created_utc":1661322769.0,"num_crossposts":0,"media":{"reddit_video":{"bitrate_kbps":1200,"fallback_url":"<WEBSERVER_URL>/5scwdfq0wlj91/DASH_480.mp4?source=fallback","height":480,"width":424,"scrubber_media_url":"https://v.redd.it/5scwdfq0wlj91/DASH_96.mp4","dash_url":"https://v.redd.it/5scwdfq0wlj91/DASHPlaylist.mpd?a=1663925795%2CMDg3NzRhYzc2MmYwMjg2NTk1MDVmY2ZjYWNiNGM0MjA0MjcxYzRmMDIxYTIxNWRiMWEzMmNkZTU1NjYyNDBlZQ%3D%3D&amp;v=1&amp;f=hd","duration":5,"hls_url":"https://v.redd.it/5scwdfq0wlj91/HLSPlaylist.m3u8?a=1663925795%2CNzMyODVkZmJmOGJiNTE2MTYxM2Y4ZGU1OWI5YzZiZWFhN2VlMzczN2Y4OGM5OWQ5NjhjNTFjZDMxODA5OTY5ZA%3D%3D&amp;v=1&amp;f=hd","is_gif":false,"transcoding_status":"completed"}},"is_video":true}}],"before":null}}`,
			WebserverUrls: []string{"/5scwdfq0wlj91/DASH_audio.mp4"},
			ExpectedResult: FetchResultMedia{
				Medias: []FetchResultMediaEntry{
					{
						Link:    "%s/5scwdfq0wlj91/DASH_480.mp4",
						Quality: "480p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_360.mp4",
						Quality: "360p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_240.mp4",
						Quality: "240p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_96.mp4",
						Quality: "96p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_audio.mp4",
						Quality: DownloadAudioQuality,
					},
				},
				ThumbnailLink: "https://b.thumbs.redditmedia.com/8O1iFRVHNR-8Y83_OHCoSh-ME_oF1DQz3Bv8q7LjxQE.jpg",
				Title:         "When you’re showing a low level around",
				Duration:      5,
				Type:          FetchResultMediaTypeVideo,
			},
			ExpectedError: nil,
		},
		{
			TestName:      "Video (Reddit Hosted - Audio)",
			PostUrl:       "https://www.reddit.com/r/gtaonline/comments/wwc1to/when_youre_showing_a_low_level_around/",
			Root:          `{"kind":"Listing","data":{"after":null,"dist":1,"modhash":"oblfmd19kc33cd5a98f14b1ef49018c259d83514950d66197f","geo_filter":"","children":[{"kind":"t3","data":{"approved_at_utc":null,"subreddit":"gtaonline","selftext":"","author_fullname":"t2_11aksu","saved":false,"mod_reason_title":null,"gilded":0,"clicked":false,"title":"When you’re showing a low level around","link_flair_richtext":[{"a":":VID1:","e":"emoji","u":"https://emoji.redditmedia.com/77ocp6gzc8h81_t5_2xrd1/VID1"},{"a":":VID2:","e":"emoji","u":"https://emoji.redditmedia.com/7gyfpnhzc8h81_t5_2xrd1/VID2"},{"a":":VID3:","e":"emoji","u":"https://emoji.redditmedia.com/bf61dvjzc8h81_t5_2xrd1/VID3"}],"subreddit_name_prefixed":"r/gtaonline","hidden":false,"pwls":6,"link_flair_css_class":"VIDEO","downs":0,"thumbnail_height":140,"top_awarded_type":null,"hide_score":false,"name":"t3_wwc1to","quarantine":false,"link_flair_text_color":"light","upvote_ratio":0.99,"author_flair_background_color":null,"ups":405,"total_awards_received":1,"media_embed":{},"thumbnail_width":140,"author_flair_template_id":null,"is_original_content":false,"user_reports":[],"secure_media":{"reddit_video":{"bitrate_kbps":1200,"fallback_url":"https://v.redd.it/5scwdfq0wlj91/DASH_480.mp4?source=fallback","height":480,"width":424,"scrubber_media_url":"https://v.redd.it/5scwdfq0wlj91/DASH_96.mp4","dash_url":"https://v.redd.it/5scwdfq0wlj91/DASHPlaylist.mpd?a=1663925795%2CMDg3NzRhYzc2MmYwMjg2NTk1MDVmY2ZjYWNiNGM0MjA0MjcxYzRmMDIxYTIxNWRiMWEzMmNkZTU1NjYyNDBlZQ%3D%3D&amp;v=1&amp;f=hd","duration":5,"hls_url":"https://v.redd.it/5scwdfq0wlj91/HLSPlaylist.m3u8?a=1663925795%2CNzMyODVkZmJmOGJiNTE2MTYxM2Y4ZGU1OWI5YzZiZWFhN2VlMzczN2Y4OGM5OWQ5NjhjNTFjZDMxODA5OTY5ZA%3D%3D&amp;v=1&amp;f=hd","is_gif":false,"transcoding_status":"completed"}},"is_reddit_media_domain":true,"is_meta":false,"category":null,"secure_media_embed":{},"link_flair_text":":VID1::VID2::VID3:","can_mod_post":false,"score":405,"approved_by":null,"is_created_from_ads_ui":false,"author_premium":false,"thumbnail":"https://b.thumbs.redditmedia.com/8O1iFRVHNR-8Y83_OHCoSh-ME_oF1DQz3Bv8q7LjxQE.jpg","edited":false,"author_flair_css_class":null,"author_flair_richtext":[],"gildings":{"gid_1":1},"post_hint":"hosted:video","content_categories":null,"is_self":false,"subreddit_type":"public","created":1661322769.0,"link_flair_type":"richtext","wls":6,"removed_by_category":null,"banned_by":null,"author_flair_type":"text","domain":"v.redd.it","allow_live_comments":false,"selftext_html":null,"likes":true,"suggested_sort":null,"banned_at_utc":null,"url_overridden_by_dest":"https://v.redd.it/5scwdfq0wlj91","view_count":null,"archived":false,"no_follow":false,"is_crosspostable":true,"pinned":false,"over_18":false,"preview":{"images":[{"source":{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?format=pjpg&amp;auto=webp&amp;s=96d2b469bec8267d2ecf3f81fc94e37458afbc7d","width":432,"height":488},"resolutions":[{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?width=108&amp;crop=smart&amp;format=pjpg&amp;auto=webp&amp;s=29ffa06c84be32c10250e01ecf6f719e153fd52f","width":108,"height":121},{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?width=216&amp;crop=smart&amp;format=pjpg&amp;auto=webp&amp;s=54c41243cf27638af3b963a6e29e075b18c1430d","width":216,"height":243},{"url":"https://external-preview.redd.it/ZSjB6Jgw9bTQiihEUsKu7SWlCMs0Px4HxbgtUsv8BmQ.png?width=320&amp;crop=smart&amp;format=pjpg&amp;auto=webp&amp;s=86026d2a1948e66d8630070dad1764069c6ca033","width":320,"height":361}],"variants":{},"id":"Gtm9A6XbVxK0VKuKvysa9HmI0qmfMTAk9103eFZqTmE"}],"enabled":false},"all_awardings":[{"giver_coin_reward":null,"subreddit_id":null,"is_new":false,"days_of_drip_extension":null,"coin_price":100,"id":"gid_1","penny_donate":null,"award_sub_type":"GLOBAL","coin_reward":0,"icon_url":"https://www.redditstatic.com/gold/awards/icon/silver_512.png","days_of_premium":null,"tiers_by_required_awardings":null,"resized_icons":[{"url":"https://www.redditstatic.com/gold/awards/icon/silver_16.png","width":16,"height":16},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_32.png","width":32,"height":32},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_48.png","width":48,"height":48},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_64.png","width":64,"height":64},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_128.png","width":128,"height":128}],"icon_width":512,"static_icon_width":512,"start_date":null,"is_enabled":true,"awardings_required_to_grant_benefits":null,"description":"Shows the Silver Award... and that's it.","end_date":null,"sticky_duration_seconds":null,"subreddit_coin_reward":0,"count":1,"static_icon_height":512,"name":"Silver","resized_static_icons":[{"url":"https://www.redditstatic.com/gold/awards/icon/silver_16.png","width":16,"height":16},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_32.png","width":32,"height":32},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_48.png","width":48,"height":48},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_64.png","width":64,"height":64},{"url":"https://www.redditstatic.com/gold/awards/icon/silver_128.png","width":128,"height":128}],"icon_format":null,"icon_height":512,"penny_price":null,"award_type":"global","static_icon_url":"https://www.redditstatic.com/gold/awards/icon/silver_512.png"}],"awarders":[],"media_only":false,"link_flair_template_id":"c7456812-be7a-11e3-a70f-12313d06c56f","can_gild":true,"spoiler":false,"locked":false,"author_flair_text":null,"treatment_tags":[],"visited":false,"removed_by":null,"mod_note":null,"distinguished":null,"subreddit_id":"t5_2xrd1","author_is_blocked":false,"mod_reason_by":null,"num_reports":null,"removal_reason":null,"link_flair_background_color":"#0079d3","id":"wwc1to","is_robot_indexable":true,"report_reasons":null,"author":"GreuDeFumat","discussion_type":null,"num_comments":12,"send_replies":true,"whitelist_status":"all_ads","contest_mode":false,"mod_reports":[],"author_patreon_flair":false,"author_flair_text_color":null,"permalink":"/r/gtaonline/comments/wwc1to/when_youre_showing_a_low_level_around/","parent_whitelist_status":"all_ads","stickied":false,"url":"https://v.redd.it/5scwdfq0wlj91","subreddit_subscribers":1281240,"created_utc":1661322769.0,"num_crossposts":0,"media":{"reddit_video":{"bitrate_kbps":1200,"fallback_url":"<WEBSERVER_URL>/5scwdfq0wlj91/DASH_480.mp4?source=fallback","height":480,"width":424,"scrubber_media_url":"https://v.redd.it/5scwdfq0wlj91/DASH_96.mp4","dash_url":"https://v.redd.it/5scwdfq0wlj91/DASHPlaylist.mpd?a=1663925795%2CMDg3NzRhYzc2MmYwMjg2NTk1MDVmY2ZjYWNiNGM0MjA0MjcxYzRmMDIxYTIxNWRiMWEzMmNkZTU1NjYyNDBlZQ%3D%3D&amp;v=1&amp;f=hd","duration":5,"hls_url":"https://v.redd.it/5scwdfq0wlj91/HLSPlaylist.m3u8?a=1663925795%2CNzMyODVkZmJmOGJiNTE2MTYxM2Y4ZGU1OWI5YzZiZWFhN2VlMzczN2Y4OGM5OWQ5NjhjNTFjZDMxODA5OTY5ZA%3D%3D&amp;v=1&amp;f=hd","is_gif":false,"transcoding_status":"completed"}},"is_video":true}}],"before":null}}`,
			WebserverUrls: []string{},
			ExpectedResult: FetchResultMedia{
				Medias: []FetchResultMediaEntry{
					{
						Link:    "%s/5scwdfq0wlj91/DASH_480.mp4",
						Quality: "480p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_360.mp4",
						Quality: "360p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_240.mp4",
						Quality: "240p",
					},
					{
						Link:    "%s/5scwdfq0wlj91/DASH_96.mp4",
						Quality: "96p",
					},
				},
				ThumbnailLink: "https://b.thumbs.redditmedia.com/8O1iFRVHNR-8Y83_OHCoSh-ME_oF1DQz3Bv8q7LjxQE.jpg",
				Title:         "When you’re showing a low level around",
				Duration:      5,
				Type:          FetchResultMediaTypeVideo,
			},
			ExpectedError: nil,
		},
		{
			TestName: "Album",
			PostUrl:  "https://www.reddit.com/r/gtaonline/comments/wuid83/caught_these_2_under_the_pier_they_went_3_rounds/?utm_source=share&utm_medium=web2x&context=3",
			Root:     `{"kind": "Listing", "data": {"after": null, "dist": 1, "modhash": "zxskjp1e0te19c2ea5bd71ee2a18e341614e15ea17e3fe2df2", "geo_filter": "", "children": [{"kind": "t3", "data": {"approved_at_utc": null, "subreddit": "gtaonline", "selftext": "", "author_fullname": "t2_3sh37jio", "saved": false, "mod_reason_title": null, "gilded": 0, "clicked": false, "is_gallery": true, "title": "Caught these 2 under the pier. They went 3 rounds.", "link_flair_richtext": [{"a": ":SNAP1:", "e": "emoji", "u": "https://emoji.redditmedia.com/4kwwti5na8h81_t5_2xrd1/SNAP1"}, {"a": ":SNAP2:", "e": "emoji", "u": "https://emoji.redditmedia.com/0tdbtc7na8h81_t5_2xrd1/SNAP2"}, {"a": ":SNAP3:", "e": "emoji", "u": "https://emoji.redditmedia.com/g0vs9d9na8h81_t5_2xrd1/SNAP3"}, {"a": ":SNAP4:", "e": "emoji", "u": "https://emoji.redditmedia.com/r4r5wgbna8h81_t5_2xrd1/SNAP4"}, {"a": ":SNAP5:", "e": "emoji", "u": "https://emoji.redditmedia.com/6tcm9jdna8h81_t5_2xrd1/SNAP5"}], "subreddit_name_prefixed": "r/gtaonline", "hidden": false, "pwls": 6, "link_flair_css_class": "SNAPMATIC", "downs": 0, "thumbnail_height": 78, "top_awarded_type": null, "hide_score": false, "media_metadata": {"175srd9hm6j91": {"status": "valid", "e": "Image", "m": "image/jpg", "o": [{"y": 1080, "x": 1920, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=1080&amp;blur=40&amp;format=pjpg&amp;auto=webp&amp;s=580eccf250b6fd30344aaed372b370ee07361bfa"}], "p": [{"y": 60, "x": 108, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=108&amp;crop=smart&amp;auto=webp&amp;s=4ad372d3033811e5307efcb5e3628cc3c3af5a36"}, {"y": 121, "x": 216, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=216&amp;crop=smart&amp;auto=webp&amp;s=c94ee14dd4d8acca278f1632307473226772fed2"}, {"y": 180, "x": 320, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=320&amp;crop=smart&amp;auto=webp&amp;s=e0c636bff8d555975a2eec4720b52a738530ffc5"}, {"y": 360, "x": 640, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=640&amp;crop=smart&amp;auto=webp&amp;s=603904050c1b07ee5a6ad33a0130498a0878cf93"}, {"y": 540, "x": 960, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=960&amp;crop=smart&amp;auto=webp&amp;s=6984617dd4e9c739e1b96d96df6e440ac8fcf282"}, {"y": 607, "x": 1080, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=1080&amp;crop=smart&amp;auto=webp&amp;s=4be0eb1f89cb23c9180e8e862fe8568a2561abe0"}], "s": {"y": 1080, "x": 1920, "u": "https://preview.redd.it/175srd9hm6j91.jpg?width=1920&amp;format=pjpg&amp;auto=webp&amp;s=8e8be25e2233f83b9ef621bd9cc9768b1b8ac5b7"}, "id": "175srd9hm6j91"}, "rdpwee9hm6j91": {"status": "valid", "e": "Image", "m": "image/jpg", "o": [{"y": 1080, "x": 1920, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=1080&amp;blur=40&amp;format=pjpg&amp;auto=webp&amp;s=f79c8a3eecbe65aad603f4ab48520d46a19449a1"}], "p": [{"y": 60, "x": 108, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=108&amp;crop=smart&amp;auto=webp&amp;s=3aa63ed07f7242096f270eb2947620c08b3152e0"}, {"y": 121, "x": 216, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=216&amp;crop=smart&amp;auto=webp&amp;s=72b19927946624fb263e1e3ccf9a12ba4931cee9"}, {"y": 180, "x": 320, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=320&amp;crop=smart&amp;auto=webp&amp;s=9e3d268f422c40640db3dd7c52702ed4772758a7"}, {"y": 360, "x": 640, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=640&amp;crop=smart&amp;auto=webp&amp;s=fe9112ca32b00cfd12a947ef4f827047f50a223b"}, {"y": 540, "x": 960, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=960&amp;crop=smart&amp;auto=webp&amp;s=7e5a575950d48271f6e072be43f42ffb9b2110dd"}, {"y": 607, "x": 1080, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=1080&amp;crop=smart&amp;auto=webp&amp;s=cad4b97b852bffbd5f8cf5209b949dee86182b7f"}], "s": {"y": 1080, "x": 1920, "u": "https://preview.redd.it/rdpwee9hm6j91.jpg?width=1920&amp;format=pjpg&amp;auto=webp&amp;s=81b50126588ec9ba0c3d1d354e78377db8d887aa"}, "id": "rdpwee9hm6j91"}}, "name": "t3_wuid83", "quarantine": false, "link_flair_text_color": "light", "upvote_ratio": 0.98, "author_flair_background_color": null, "ups": 2390, "domain": "reddit.com", "media_embed": {}, "thumbnail_width": 140, "author_flair_template_id": null, "is_original_content": false, "user_reports": [], "secure_media": null, "is_reddit_media_domain": false, "is_meta": false, "category": null, "secure_media_embed": {}, "gallery_data": {"items": [{"media_id": "175srd9hm6j91", "id": 178597657}, {"media_id": "rdpwee9hm6j91", "id": 178597658}]}, "link_flair_text": ":SNAP1::SNAP2::SNAP3::SNAP4::SNAP5:", "can_mod_post": false, "score": 2390, "approved_by": null, "is_created_from_ads_ui": false, "author_premium": false, "thumbnail": "nsfw", "edited": false, "author_flair_css_class": null, "author_flair_richtext": [], "gildings": {}, "content_categories": null, "is_self": false, "subreddit_type": "public", "created": 1661137957.0, "link_flair_type": "richtext", "wls": 3, "removed_by_category": null, "banned_by": null, "author_flair_type": "text", "total_awards_received": 8, "allow_live_comments": true, "selftext_html": null, "likes": true, "suggested_sort": null, "banned_at_utc": null, "url_overridden_by_dest": "https://www.reddit.com/gallery/wuid83", "view_count": null, "archived": false, "no_follow": false, "is_crosspostable": true, "pinned": false, "over_18": true, "all_awardings": [{"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 125, "id": "award_5f123e3d-4f48-42f4-9c11-e98b566d5897", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=16&amp;height=16&amp;auto=webp&amp;s=92932f465d58e4c16b12b6eac4ca07d27e3d11c0", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=32&amp;height=32&amp;auto=webp&amp;s=d11484a208d68a318bf9d4fcf371171a1cb6a7ef", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=48&amp;height=48&amp;auto=webp&amp;s=febdf28b6f39f7da7eb1365325b85e0bb49a9f63", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=64&amp;height=64&amp;auto=webp&amp;s=b4406a2d88bf86fa3dc8a45aacf7e0c7bdccc4fb", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=128&amp;height=128&amp;auto=webp&amp;s=19555b13e3e196b62eeb9160d1ac1d1b372dcb0b", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "When you come across a feel-good thing.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 3, "static_icon_height": 2048, "name": "Wholesome", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=16&amp;height=16&amp;auto=webp&amp;s=92932f465d58e4c16b12b6eac4ca07d27e3d11c0", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=32&amp;height=32&amp;auto=webp&amp;s=d11484a208d68a318bf9d4fcf371171a1cb6a7ef", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=48&amp;height=48&amp;auto=webp&amp;s=febdf28b6f39f7da7eb1365325b85e0bb49a9f63", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=64&amp;height=64&amp;auto=webp&amp;s=b4406a2d88bf86fa3dc8a45aacf7e0c7bdccc4fb", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png?width=128&amp;height=128&amp;auto=webp&amp;s=19555b13e3e196b62eeb9160d1ac1d1b372dcb0b", "width": 128, "height": 128}], "icon_format": null, "icon_height": 2048, "penny_price": null, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_22cerq/5izbv4fn0md41_Wholesome.png"}, {"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 50, "id": "award_02d9ab2c-162e-4c01-8438-317a016ed3d9", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_q0gj4/p4yzxkaed5f61_oldtakemyenergy.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_q0gj4/p4yzxkaed5f61_oldtakemyenergy.png?width=16&amp;height=16&amp;auto=webp&amp;s=10034f3fdf8214c8377134bb60c5b832d4bbf588", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/p4yzxkaed5f61_oldtakemyenergy.png?width=32&amp;height=32&amp;auto=webp&amp;s=100f785bf261fa9452a5d82ee0ef0793369dbfa5", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/p4yzxkaed5f61_oldtakemyenergy.png?width=48&amp;height=48&amp;auto=webp&amp;s=b15d030fdfbbe4af4a5b34ab9dc90a174df40a23", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/p4yzxkaed5f61_oldtakemyenergy.png?width=64&amp;height=64&amp;auto=webp&amp;s=601c75be6ee30dc4b47a5c65d64dea9a185502a1", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/p4yzxkaed5f61_oldtakemyenergy.png?width=128&amp;height=128&amp;auto=webp&amp;s=540f36e65c0e2f1347fe32020e4a1565e3680437", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "I'm in this with you.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 2, "static_icon_height": 2048, "name": "Take My Energy", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_q0gj4/jtw7x06j68361_TakeMyEnergyElf.png?width=16&amp;height=16&amp;auto=webp&amp;s=045db73f47a9513c44823d132b4c393ab9241b6a", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/jtw7x06j68361_TakeMyEnergyElf.png?width=32&amp;height=32&amp;auto=webp&amp;s=298a02e0edbb5b5e293087eeede63802cbe1d2c7", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/jtw7x06j68361_TakeMyEnergyElf.png?width=48&amp;height=48&amp;auto=webp&amp;s=7d06d606eb23dbcd6dbe39ee0e60588c5eb89065", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/jtw7x06j68361_TakeMyEnergyElf.png?width=64&amp;height=64&amp;auto=webp&amp;s=ecd9854b14104a36a210028c43420f0dababd96b", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/jtw7x06j68361_TakeMyEnergyElf.png?width=128&amp;height=128&amp;auto=webp&amp;s=0d5d7b92c1d66aff435f2ad32e6330ca2b971f6d", "width": 128, "height": 128}], "icon_format": "PNG", "icon_height": 2048, "penny_price": 0, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_q0gj4/jtw7x06j68361_TakeMyEnergyElf.png"}, {"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 70, "id": "award_99d95969-6100-45b2-b00c-0ec45ae19596", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=16&amp;height=16&amp;auto=webp&amp;s=ff94d9e3eb38878a038b2568c06b58e809d7f0f5", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=32&amp;height=32&amp;auto=webp&amp;s=2dcdf8ac6a205b6e93b0fb31012044b66f3f4186", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=48&amp;height=48&amp;auto=webp&amp;s=3d8d317fd0e68c3f2696425efb7a5bc85b6f7603", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=64&amp;height=64&amp;auto=webp&amp;s=a54e710bdf1bc88eb1bb2da67d1ecf813f1707be", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=128&amp;height=128&amp;auto=webp&amp;s=b564b07d31245f583542d97aa99f58e9dadaed2f", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "A smol, delicate danger noodle.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 1, "static_icon_height": 2048, "name": "Snek", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=16&amp;height=16&amp;auto=webp&amp;s=ff94d9e3eb38878a038b2568c06b58e809d7f0f5", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=32&amp;height=32&amp;auto=webp&amp;s=2dcdf8ac6a205b6e93b0fb31012044b66f3f4186", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=48&amp;height=48&amp;auto=webp&amp;s=3d8d317fd0e68c3f2696425efb7a5bc85b6f7603", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=64&amp;height=64&amp;auto=webp&amp;s=a54e710bdf1bc88eb1bb2da67d1ecf813f1707be", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png?width=128&amp;height=128&amp;auto=webp&amp;s=b564b07d31245f583542d97aa99f58e9dadaed2f", "width": 128, "height": 128}], "icon_format": "PNG", "icon_height": 2048, "penny_price": 0, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_22cerq/rc5iesz2z8t41_Snek.png"}, {"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 50, "id": "award_69c94eb4-d6a3-48e7-9cf2-0f39fed8b87c", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=16&amp;height=16&amp;auto=webp&amp;s=bb033b3352b6ece0954d279a56f99e16c67abe14", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=32&amp;height=32&amp;auto=webp&amp;s=a8e1d0c2994e6e0b254fab1611d539a4fb94e38a", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=48&amp;height=48&amp;auto=webp&amp;s=723e4e932c9692ac61cf5b7509424c6ae1b5d220", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=64&amp;height=64&amp;auto=webp&amp;s=b7f0640e403ac0ef31236a4a0b7f3dc25de6046c", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=128&amp;height=128&amp;auto=webp&amp;s=ac954bb1a06af66bf9295bbfee4550443fb6f21d", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "Listen, get educated, and get involved.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 1, "static_icon_height": 2048, "name": "Ally", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=16&amp;height=16&amp;auto=webp&amp;s=bb033b3352b6ece0954d279a56f99e16c67abe14", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=32&amp;height=32&amp;auto=webp&amp;s=a8e1d0c2994e6e0b254fab1611d539a4fb94e38a", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=48&amp;height=48&amp;auto=webp&amp;s=723e4e932c9692ac61cf5b7509424c6ae1b5d220", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=64&amp;height=64&amp;auto=webp&amp;s=b7f0640e403ac0ef31236a4a0b7f3dc25de6046c", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png?width=128&amp;height=128&amp;auto=webp&amp;s=ac954bb1a06af66bf9295bbfee4550443fb6f21d", "width": 128, "height": 128}], "icon_format": "PNG", "icon_height": 2048, "penny_price": 0, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_22cerq/5nswjpyy44551_Ally.png"}, {"giver_coin_reward": null, "subreddit_id": null, "is_new": false, "days_of_drip_extension": null, "coin_price": 50, "id": "award_80d4d339-95d0-43ac-b051-bc3fe0a9bab8", "penny_donate": null, "award_sub_type": "GLOBAL", "coin_reward": 0, "icon_url": "https://i.redd.it/award_images/t5_q0gj4/9auzllkyd5f61_oldwearing.png", "days_of_premium": null, "tiers_by_required_awardings": null, "resized_icons": [{"url": "https://preview.redd.it/award_images/t5_q0gj4/9auzllkyd5f61_oldwearing.png?width=16&amp;height=16&amp;auto=webp&amp;s=7530150c82cb32627e80f409d92bacd95b4b6f89", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/9auzllkyd5f61_oldwearing.png?width=32&amp;height=32&amp;auto=webp&amp;s=8960e957206d6214bc7a5ba3db21ac70aff76e73", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/9auzllkyd5f61_oldwearing.png?width=48&amp;height=48&amp;auto=webp&amp;s=a1853cd01a345600cdf8589476e3fdfb66b53936", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/9auzllkyd5f61_oldwearing.png?width=64&amp;height=64&amp;auto=webp&amp;s=611185fbe83a4c1b658bc08dc4bd4fb711a4db65", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/9auzllkyd5f61_oldwearing.png?width=128&amp;height=128&amp;auto=webp&amp;s=dff2aa35972f73905377622832fc2c70df360617", "width": 128, "height": 128}], "icon_width": 2048, "static_icon_width": 2048, "start_date": null, "is_enabled": true, "awardings_required_to_grant_benefits": null, "description": "Keep the community and yourself healthy and happy.", "end_date": null, "sticky_duration_seconds": null, "subreddit_coin_reward": 0, "count": 1, "static_icon_height": 2048, "name": "Wearing is Caring", "resized_static_icons": [{"url": "https://preview.redd.it/award_images/t5_q0gj4/0mxct3p878361_WearingIsCaringElf.png?width=16&amp;height=16&amp;auto=webp&amp;s=0349ceebb30e25e913f1ebc8cde78807d2f94cfe", "width": 16, "height": 16}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/0mxct3p878361_WearingIsCaringElf.png?width=32&amp;height=32&amp;auto=webp&amp;s=07cc6b9c14c3755605148f2240ac582a44a78596", "width": 32, "height": 32}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/0mxct3p878361_WearingIsCaringElf.png?width=48&amp;height=48&amp;auto=webp&amp;s=d89451c2145881c3d525a6b78742a11546feea3c", "width": 48, "height": 48}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/0mxct3p878361_WearingIsCaringElf.png?width=64&amp;height=64&amp;auto=webp&amp;s=1513567b75db31adff4e4a7a157e6cab8a3e41ad", "width": 64, "height": 64}, {"url": "https://preview.redd.it/award_images/t5_q0gj4/0mxct3p878361_WearingIsCaringElf.png?width=128&amp;height=128&amp;auto=webp&amp;s=234cb3d8f90476a6e38e2105c52f0f7281585176", "width": 128, "height": 128}], "icon_format": "PNG", "icon_height": 2048, "penny_price": 0, "award_type": "global", "static_icon_url": "https://i.redd.it/award_images/t5_q0gj4/0mxct3p878361_WearingIsCaringElf.png"}], "awarders": [], "media_only": false, "link_flair_template_id": "cb72ffa8-be7a-11e3-9116-12313d056a4d", "can_gild": true, "spoiler": false, "locked": false, "author_flair_text": null, "treatment_tags": [], "visited": false, "removed_by": null, "mod_note": null, "distinguished": null, "subreddit_id": "t5_2xrd1", "author_is_blocked": false, "mod_reason_by": null, "num_reports": null, "removal_reason": null, "link_flair_background_color": "#24a0ed", "id": "wuid83", "is_robot_indexable": true, "report_reasons": null, "author": "AlphaMale3Percent", "discussion_type": null, "num_comments": 115, "send_replies": true, "whitelist_status": "promo_adult_nsfw", "contest_mode": false, "mod_reports": [], "author_patreon_flair": false, "author_flair_text_color": null, "permalink": "/r/gtaonline/comments/wuid83/caught_these_2_under_the_pier_they_went_3_rounds/", "parent_whitelist_status": "all_ads", "stickied": false, "url": "https://www.reddit.com/gallery/wuid83", "subreddit_subscribers": 1281253, "created_utc": 1661137957.0, "num_crossposts": 0, "media": null, "is_video": false}}], "before": null}}`,
			ExpectedResult: FetchResultAlbum{Album: []FetchResultAlbumEntry{
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
			ExpectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			// Spawn the webserver if needed
			if test.WebserverUrls != nil {
				server := newSimpleWebserver(test.WebserverUrls...)
				defer server.Close()
				test.Root = strings.ReplaceAll(test.Root, "<WEBSERVER_URL>", server.URL)
				if media, isMedia := test.ExpectedResult.(FetchResultMedia); isMedia {
					for i := range media.Medias {
						media.Medias[i].Link = fmt.Sprintf(media.Medias[i].Link, server.URL)
					}
				}
			}
			var root map[string]interface{}
			err := json.NewDecoder(strings.NewReader(test.Root)).Decode(&root)
			assert.NoError(t, err, "not expecting error when decoding sample root")
			result, fetchError := getPost(test.PostUrl, root)
			if fetchError != nil && test.ExpectedError != nil {
				assert.Equal(t, *test.ExpectedError, *fetchError)
			} else if fetchError != nil && test.ExpectedError == nil {
				assert.Fail(t, "Unexpected error:", *fetchError)
			} else if fetchError == nil && test.ExpectedError != nil {
				assert.Fail(t, "Expected error:", *test.ExpectedError)
			} else {
				assert.Equal(t, test.ExpectedResult, result)
			}
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
