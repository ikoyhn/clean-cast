## DockerHub
https://hub.docker.com/r/ikoyhn/clean-cast

## Docker Run Command Templates

Docker run command (only required parameters)
```
docker run -p 8080:8080 -e GOOGLE_API_KEY=<api-key-here> -v /<config-path-here>:/config ikoyhn/clean-cast
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
- GOOGLE_API_KEY=<api-key-here>
volumes:
- /<config-path-here>:/config
```

## Docker Variables
|Variable| Description | Required |
|--|--|--|
| `-v <host-path>:/config` | Where config files will be stored (includes properties.yml and sqlite.db) | Yes |
| `-v <host-path>:/audio` | Where the audio files will be stored on a separate volume/drive. If not set, audio will default to `/config/audio` | No |
| `-e GOOGLE_API_KEY=<api-key>` | YouTube v3 API Key. Get your own api key [here](https://developers.google.com/youtube/v3/getting-started)| YES (must be set either here or in properties.yml) |

## Audio Storage Examples

**Option 1: Audio stored in config folder (default)**
```bash
docker run -p 8080:8080 -e GOOGLE_API_KEY=<api-key-here> -v /path/to/config:/config ikoyhn/clean-cast
```
Audio files will be stored in: `/path/to/config/audio`

**Option 2: Audio stored in separate volume/drive**
```bash
docker run -p 8080:8080 -e GOOGLE_API_KEY=<api-key-here> -v /path/to/config:/config -v /path/to/audio:/audio ikoyhn/clean-cast
```
Audio files will be stored in: `/path/to/audio`

