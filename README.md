
# GO Podcast Sponsorblock RSS

  

This is a GO application that will take any podcast that is on Youtube and will generate a RSS feed with the audio only and all sponsored sections auto removed.

  

This app uses the following
* [Irstanley/go-ytdlp](https://github.com/lrstanley/go-ytdlp) for downloading Youtube videos (includes Sponsorblock for removing sponsored segments)
* [google-api-go-client](https://github.com/googleapis/google-api-go-client) Used to get all information for Youtube Podcast

  
  

# Docker Variables
|Variable| Description | Required |
|--|--|--|
| `-v config:/` | Where the audio files will be stored | Yes |
| `-e GOOGLE_API_KEY` | YouTube v3 API Key | Yes |
| `-e TOKEN` | Used for securing the endpoints. If using this you must add the query param `token` to the end of the URL for the `/rss` endpoint request ex.`?token=mySecureToken` | No |
| `-e TRUSTED_HOSTS` | If you want to limit what host this service can be called from. Can be a list of hosts separated by a `,` Ex: `localhost:8080,https://podcast.com` | No |

## Docker Run Command Templates
> Docker run command (only required parameters)

    docker run -p 8080:8080 -e GOOGLE_API_KEY={api key here} -v /config:/{audio download path here} ikoyhn/go-podcast-sponsor-block
    
> Docker run command (all parameters)

    docker run -p 8080:8080 -e GOOGLE_API_KEY={api key here} -v /config:/{audio download path here} -e TRUSTED_HOSTS={add hosts here} -e TOKEN={add secure token here} ikoyhn/go-podcast-sponsor-block


  
  

# How To Use

1. Search on Youtube for Podcast you want, normally found under the channels `Podcasts` section

* Once in the Podcast playlist look to the URL to find the playlist ID ex. for the TigerBelly Podcast URL is `https://www.youtube.com/playlist?list=PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222` so the ID would be `PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222`

2. Once you have the ID add the the ID to the end of the main endpoint ie. `{url this is running on}/rss/{id of playlist}`

* Following the TigerBelly example where this app is running on `http://localhost:8080` the url would be `http://localhost:8080/rss/PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222`
	* **NOTE:** If you have the docker var `-e TOKEN` set you must add the token as a query param to this url. Ex: `http://localhost:8080/rss/PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222?token=secureToken`

3. With this URL you can now add this to any of your favorite podcast apps that accept custom RSS feeds (Apple Podcasts app, VLC Media Player, etc)

4. Listen and enjoy ad/sponsor free podcasts :)

  

# Documentation of process behind the scenes

1. First step is get the Youtube playlist ID from the `/rss` endpoint

2. Get data for the Youtube playlist data using the Youtube v3 API key

3. Call out to Apple Search API to get more metadata for the podcast

* The way the podcast is found is using the title returned from the youtube v3 api response object, sometimes there are more than one result so a secondary check is done by comparing the total number of episodes between both the youtube v3 api response and the Apple search response. The one with the closest number will be selected from the Apple search api response.

4. Now all data needed for podcast has been gathered, now generate a podcast XML RSS and return it to the caller of the `/rss` endpoint

5. The podcast app the user is using (VLC, Apple Podcasts, etc.) should now have an RSS of the podcast.

6. When the user requests a podcast is will hit the `/media` endpoint

7. We will first check if the podcast has already been downloaded if so serve that file back to the client

8. If it is not present download the youtube video with audio only and with Sponsorblock category "sponsor" removed

9. Serve the file back to the client
