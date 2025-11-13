# Quick Start Guide - Recommendations API

## Getting Started

### 1. Build and Run

Once network connectivity is restored, run:

```bash
# Install dependencies
go mod download
go mod tidy

# Build the application
go build -o cleancast ./cmd/app/main.go

# Run the server
./cleancast
```

### 2. Access the API

The server will start on `http://localhost:8080` (or the configured HOST:PORT).

## Swagger UI

Visit the interactive API documentation:

```
http://localhost:8080/docs/
```

This provides:
- Full API documentation
- Interactive "Try it out" functionality
- Request/response examples
- Schema definitions

## Testing the Recommendations API

### Example 1: Get Similar Podcasts

Find podcasts similar to "Linus Tech Tips" (UCXuqSBlHAE6Xw-yeJA0Tunw):

```bash
curl http://localhost:8080/api/recommendations/similar/UCXuqSBlHAE6Xw-yeJA0Tunw?limit=5
```

Response:
```json
{
  "podcast_id": "UCXuqSBlHAE6Xw-yeJA0Tunw",
  "limit": 5,
  "results": 5,
  "similar": [
    {
      "podcast": {
        "id": "UC...",
        "podcast_name": "Similar Tech Podcast",
        "category": "Technology",
        ...
      },
      "similarity_score": 0.85
    }
  ]
}
```

### Example 2: Get Trending Episodes

Get the 10 most played episodes in the last 7 days:

```bash
curl http://localhost:8080/api/recommendations/trending?limit=10
```

Response:
```json
{
  "period": "last_7_days",
  "limit": 10,
  "results": 10,
  "episodes": [
    {
      "episode": {
        "youtube_video_id": "dQw4w9WgXcQ",
        "episode_name": "Popular Episode",
        ...
      },
      "play_count": 150
    }
  ]
}
```

### Example 3: Get Related Episodes

Get episodes from the same channel/playlist:

```bash
curl http://localhost:8080/api/recommendations/related/dQw4w9WgXcQ?limit=10
```

Response:
```json
{
  "video_id": "dQw4w9WgXcQ",
  "limit": 10,
  "results": 8,
  "episodes": [
    {
      "youtube_video_id": "abc123",
      "episode_name": "Another Episode",
      ...
    }
  ]
}
```

### Example 4: Get Personalized Recommendations

Get recommendations based on your listening history:

```bash
curl http://localhost:8080/api/recommendations/for-you?limit=20
```

Response:
```json
{
  "strategy": "personalized_based_on_history",
  "limit": 20,
  "results": 20,
  "episodes": [
    {
      "youtube_video_id": "xyz789",
      "episode_name": "Recommended for You",
      ...
    }
  ]
}
```

## Using with JavaScript/Fetch

```javascript
// Get trending episodes
fetch('http://localhost:8080/api/recommendations/trending?limit=10')
  .then(response => response.json())
  .then(data => {
    console.log('Trending episodes:', data.episodes);
  });

// Get similar podcasts
fetch('http://localhost:8080/api/recommendations/similar/UCXuqSBlHAE6Xw-yeJA0Tunw?limit=5')
  .then(response => response.json())
  .then(data => {
    console.log('Similar podcasts:', data.similar);
  });
```

## Using with Python

```python
import requests

# Get trending episodes
response = requests.get('http://localhost:8080/api/recommendations/trending',
                       params={'limit': 10})
trending = response.json()
print(f"Found {trending['results']} trending episodes")

# Get similar podcasts
response = requests.get('http://localhost:8080/api/recommendations/similar/UCXuqSBlHAE6Xw-yeJA0Tunw',
                       params={'limit': 5})
similar = response.json()
for item in similar['similar']:
    print(f"{item['podcast']['podcast_name']}: {item['similarity_score']}")
```

## Rate Limiting

All recommendation endpoints are rate-limited to **30 requests per minute**.

If you exceed this limit, you'll receive a 429 response:

```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded",
  "request_id": "req-123456789"
}
```

## OpenAPI Specification

The full OpenAPI specification is available at:

```
/home/user/clean-cast/docs/openapi.yaml
```

You can use this with tools like:
- Swagger UI
- Postman (import OpenAPI spec)
- API clients (generate code)
- Documentation generators

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Controller                          │
│              (internal/app/controller.go)               │
└──────────────────────┬──────────────────────────────────┘
                       │
                       │ Routes
                       ▼
┌─────────────────────────────────────────────────────────┐
│                  API Handlers                           │
│          (internal/api/recommendations.go)              │
│  • RegisterRecommendationRoutes()                       │
│  • GetSimilarPodcasts()                                 │
│  • GetTrendingEpisodes()                                │
│  • GetRelatedEpisodes()                                 │
│  • GetPersonalizedRecommendations()                     │
└──────────────────────┬──────────────────────────────────┘
                       │
                       │ Business Logic
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Recommendation Service                      │
│    (internal/services/recommendations/                  │
│              recommendations.go)                         │
│  • Similarity calculation                               │
│  • Trending analysis                                    │
│  • Personalization algorithms                           │
└──────────────────────┬──────────────────────────────────┘
                       │
                       │ Data Access
                       ▼
┌─────────────────────────────────────────────────────────┐
│                   Database Layer                         │
│          (internal/database/*.go)                        │
│  • Podcasts                                             │
│  • Episodes                                             │
│  • Playback History                                     │
└─────────────────────────────────────────────────────────┘
```

## Next Steps

1. Start the server
2. Visit `/docs/` for interactive documentation
3. Test the endpoints with the examples above
4. Integrate into your application
5. Monitor rate limits and analytics

## Support

For more detailed information, see:
- `/home/user/clean-cast/docs/RECOMMENDATIONS_API.md` - Full API documentation
- `/home/user/clean-cast/docs/openapi.yaml` - OpenAPI specification
- `/home/user/clean-cast/internal/services/recommendations/recommendations.go` - Algorithm implementation
