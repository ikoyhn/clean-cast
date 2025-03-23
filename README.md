
# GO Podcast Sponsorblock RSS

  

This is a GO application that will take any podcast that is on Youtube and will generate a RSS feed with the audio only and all sponsored sections auto removed. The actual podcasts episodes are downloaded on demand from youtube when the user requests the specific episode then it is served to the user seamlessly. By default a CRON job is run weekly to delete any episode files that havent been accessed in over a week, this can be modified with docker variables. This project requires a Youtube v3 API Key which you can generate [here](https://developers.google.com/youtube/v3/getting-started).

  

  

This app uses the following

* [Irstanley/go-ytdlp](https://github.com/lrstanley/go-ytdlp) for downloading Youtube videos (includes Sponsorblock for removing sponsored segments)

* [google-api-go-client](https://github.com/googleapis/google-api-go-client) Used to get all information for Youtube Podcast

  

# Docker Variables
|Variable| Description | Required |
|--|--|--|
| `-v <container path>:/config` | Where the audio files will be stored | Yes |
| `-e GOOGLE_API_KEY=<api key>` | YouTube v3 API Key. Get your own api key [here](https://developers.google.com/youtube/v3/getting-started)| Yes |
| `-e TOKEN=<secure key>` | Used for securing the endpoints. If using this you must add the query param `token` to the end of the URL for the `/rss` endpoint request ex.`?token=mySecureToken` | No |
| `-e TRUSTED_HOSTS=<list of hosts>` | If you want to limit what host this service can be called from. Can be a list of hosts separated by a `,` Ex: `localhost:8080,https://podcast.com` | No |
| `-e CRON` | By default a cron job will be run weekly to delete any podcast episode files that havent been access in over a week, if you want to modify when this runs you can set the cron here | No |
| `-e SPONSORBLOCK_CATEGORIES` | Customize the categories that you would like to remove from your podcasts. String separated by `,` with possible values `sponsor,selfpromo,interaction,intro,outro,preview,music_offtopic,filler`. Default: `sponsor` | No |
| `-e COOKIES_FILE` | Run the app once for the config folder to be created then store your cookies folder in the root of the config folder and set the filename for the docker var. Set this if you want to use custom cookies for YT-DLP| No |
  

## Docker Run Command Templates

> Docker run command (only required parameters)

  

docker run -p 8080:8080 -e GOOGLE_API_KEY=<api key here> -v /<audio download path here>:/config ikoyhn/go-podcast-sponsor-block

> Docker run command (all parameters)

  

docker run -p 8080:8080 -e GOOGLE_API_KEY=<api key here> -v /<audio download path here>:/config -e TRUSTED_HOSTS=<add hosts here> -e TOKEN=<add secure token here> -e CRON="0 0 * * 0" -e SPONSORBLOCK_CATEGORIES="sponsor" ikoyhn/go-podcast-sponsor-block

  

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

  

# How To Use

  

 1. Find your ID
	-  **Playlist ID**: To find this navigate to the channel and either click the podcast tab or playlist tab and click on the playlist you want to add. Ex TigerBelly the url you should see after clicking it will be www.youtube.com/playlist?list=PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222` so the ID would be `PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222`
			
	 - **Channel ID**: If you want to create a podcast from all videos from a channel use this. You can get the channel ID using websites such as https://www.tunepocket.com/youtube-channel-id-finder/.

  

2. Build your URL
	-  **Playlist**: If you are building a podcast URL using a playlist use the `/rss`endpoint. * Following the TigerBelly example where this app is running on `http://localhost:8080` the url would be `http://localhost:8080/rss/PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222`
			
	 - **Channel**: If you are building a podcast URL using a channel ID use the `/channel` endpoint. An example would be `http://localhost:8080/channel/UCoj1ZgGoSBoonNZqMsVUfAA`

*  **NOTE:** If you have the docker var `-e TOKEN=<secure token>` set you must add the token as a query param to this url. Ex: `http://localhost:8080/rss/PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222?token=secureToken`

  

3. With this URL you can now add this to any of your favorite podcast apps that accept custom RSS feeds (Apple Podcasts app, VLC Media Player, etc)

  

4. Listen and enjoy ad/sponsor free podcasts :)


# Please report any issues you run into!
