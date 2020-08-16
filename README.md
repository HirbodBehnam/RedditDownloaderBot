# Reddit Downloader Bot
A telegram bot to download the reddit posts.
## What can this bot do?
* Send posts as text in telegram
* Download and send photos in `i.redd.it`
* Download and send GIFs
* Download videos on `v.redd.it` and merge the audio with them using [FFmpeg](https://www.ffmpeg.org/)
* Donwload galleries (Not well tested, also max file size is 10 MB)
* Choose the quality of photos or videos (Except galleries)
## What this bot can't do?
* Upload files that are larger than 50MB
* Download any videos or photos that are not hosted on `x.redd.it` (for example youtube)
### List of non `x.redd.it` hosts that this bot can download from them
1. imgur (gifs and pictures)
2. gfycat (Note that some of them may not work because they are not hosted on reddit. Also they are soundless)
3. streamable
## Setup
To start, please at first install [FFmpeg](https://www.ffmpeg.org/) in your path. On Ubuntu, `apt install ffmpeg` is enough.

Then download and build this project:
```bash
git clone https://github.com/HirbodBehnam/RedditDownloaderBot
go get github.com/go-telegram-bot-api/telegram-bot-api github.com/google/uuid github.com/patrickmn/go-cache github.com/PuerkitoBio/goquery
go build main.go
```

For running, pass the bot token as the first argument: `./bot 1234567:4TT8bAc8GHUspu3ERYn-KGcvsvGB9u_n4ddy`
