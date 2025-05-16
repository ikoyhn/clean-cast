## Docker Run Command Templates

Docker run command (only required parameters)

  

> docker run -p 8080:8080 -e GOOGLE_API_KEY=<api key here> -v /<audio download path here>:/config ikoyhn/go-podcast-sponsor-block

Docker run command (all parameters)

  

> docker run -p 8080:8080 -e GOOGLE_API_KEY=<api key here> -v /<audio download path here>:/config -e TRUSTED_HOSTS=<add hosts here> -e TOKEN=<add secure token here> -e CRON="0 0 * * 0" -e SPONSORBLOCK_CATEGORIES="sponsor" -e COOKIES_FILE <cookies file path here> ikoyhn/go-podcast-sponsor-block

  

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