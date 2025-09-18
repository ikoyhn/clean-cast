## DockerHub
https://hub.docker.com/r/ikoyhn/clean-cast

## Docker Run Command Templates

Docker run command (only required parameters)
```
docker run -p 8080:8080 -e GOOGLE_API_KEY=<api key here> -v /<audio download path here>:/config ikoyhn/go-podcast-sponsor-block
```


Docker run command (all parameters)
```
docker run -p 8080:8080 -e GOOGLE_API_KEY=<api key here> -v /<audio download path here>:/config -e TRUSTED_HOSTS=<add hosts here> -e TOKEN=<add secure token here> -e CRON="0 0 * * 0" -e SPONSORBLOCK_CATEGORIES="sponsor" -e COOKIES_FILE <cookies file path here> -e MIN_DURATION=<min duration of video to grab> ikoyhn/go-podcast-sponsor-block
```

## Docker Compose Templates

> Docker compose template (only required parameters)

```yaml

services:
podcast-sponsor-block:
image: ikoyhn/go-podcast-sponsor-block
ports:
- "8080:8080"
environment:
- GOOGLE_API_KEY=<api key here>
volumes:
- /<audio download path here>:/config
```

  

> Docker compose template (all parameters)

```yaml
services:
podcast-sponsor-block:
image: ikoyhn/go-podcast-sponsor-block
ports:
- "8080:8080"
environment:
- GOOGLE_API_KEY=<api key here>
- TRUSTED_HOSTS=<add hosts here>
- TOKEN=<add secure token here>
- CRON=0 0 * * 0
- SPONSORBLOCK_CATEGORIES=sponsor
- COOKIES_FILE=<filename here>
volumes:
- /<audio download path here>:/config

```

## Docker Variables
|Variable| Description | Required |
|--|--|--|
| `-v <container path>:/config` | Where the audio files and config files will be stored | Yes |
| `-e GOOGLE_API_KEY=<api key>` | YouTube v3 API Key. Get your own api key [here](https://developers.google.com/youtube/v3/getting-started)| Yes |
| `-e TOKEN=<secure key>` | Used for securing the endpoints. If using this you must add the query param `token` to the end of the URL for the `/rss` endpoint request ex.`?token=mySecureToken` | No |
| `-e TRUSTED_HOSTS=<list of hosts>` | If you want to limit what host this service can be called from. Can be a list of hosts separated by a `,` Ex: `localhost:8080,https://podcast.com` | No |
| `-e CRON` | By default a cron job will be run weekly to delete any podcast episode files that havent been access in over a week, if you want to modify when this runs you can set the cron here ([CRON examples](https://crontab.guru/))| No |
| `-e SPONSORBLOCK_CATEGORIES` | Customize the categories that you would like to remove from your podcasts. String separated by `,` with possible values `sponsor,selfpromo,interaction,intro,outro,preview,music_offtopic,filler`. Default: `sponsor` | No |
| `-e COOKIES_FILE` | Run the app once for the config folder to be created then store your cookies folder in the root of the config folder and set the filename for the docker var. Set this if you want to use custom cookies for YT-DLP| No |
| `-e MIN_DURATION` | To filter out YT shorts there is a minimum duration a video has to be in order to grab it. The default is 5min, modify as needed. Example values: (30s, 5m, 1hr) | No |
