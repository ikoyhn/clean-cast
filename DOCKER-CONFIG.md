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
| `-v <container path>:/config` | Where config files will be stored | Yes |
| `-e GOOGLE_API_KEY=<api-key>` | YouTube v3 API Key. Get your own api key [here](https://developers.google.com/youtube/v3/getting-started)| YES (must be set either here or in properties.yml) |
