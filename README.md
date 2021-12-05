# Reddit Downloader Bot

A telegram bot to download the reddit posts.

## What can this bot do?

* Send posts as text in telegram
* Download and send photos in `i.redd.it`
* Download and send GIFs
* Download videos on `v.redd.it` and merge the audio with them using [FFmpeg](https://www.ffmpeg.org/)
* Download galleries
* Choose the quality of photos or videos (except galleries, always the best resolution is used in them)
* Download comments
* Limit the users which can use it

## What this bot can't do?

* Send deleted posts
* Send polls
* Send text posts with more than 4096 characters or complex markdown (like tables)
* Upload files that are larger than 50MB
* Download any videos or photos that are not hosted on `x.redd.it` (for example youtube)

### List of non `x.redd.it` hosts that this bot can download from them

1. imgur (gifs and pictures)
2. gfycat (Note that some of them may not work because they are not hosted on reddit. Also, they are soundless)
3. streamable

## Setup

### Build

To start, please at first install [FFmpeg](https://www.ffmpeg.org/) in your path. On Ubuntu, `apt install ffmpeg` is
enough.

Then download and build this project:

```bash
git clone https://github.com/HirbodBehnam/RedditDownloaderBot
cd RedditDownloaderBot
go build
```

### Reddit Token

To use this bot, you need to have a registered Reddit application. To do so, you can
use [this](https://github.com/reddit-archive/reddit/wiki/OAuth2#getting-started) guide by Reddit itself.
Choose `Script app` as application type. Doing so, will give you two tokens: A client id and a client secret. (The
client id is the `one personal use script` text on top left)

### Running

For running, you have to set the environmental variables like this:

```bash
export CLIENT_ID=p-jcoLKBynTLew
export CLIENT_SECRET=gko_LXELoV07ZBNUXrvWZfzE3aI
export BOT_TOKEN=1234567:4TT8bAc8GHUspu3ERYn-KGcvsvGB9u_n4ddy
export ALLOWED_USERS=1,2,3 # Optional; You can simply ignore this line
./bot
```