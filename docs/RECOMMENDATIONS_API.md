# Recommendations API Documentation

## Overview

The Recommendations API provides intelligent episode and podcast recommendations based on various algorithms including:
- Content similarity analysis
- Trending content based on analytics
- Related episodes from the same source
- Personalized recommendations based on listening history

## Implementation Files

### 1. Recommendation Service (`internal/services/recommendations/recommendations.go`)

Core service implementing recommendation algorithms:

- **SimilarPodcasts**: Finds podcasts similar to a given podcast based on:
  - Category matching (50% weight)
  - Description keyword overlap using Jaccard similarity (50% weight)

- **TrendingEpisodes**: Returns most played episodes in the last 7 days based on playback analytics

- **RelatedEpisodes**: Gets episodes from the same channel/playlist

- **PersonalizedRecommendations**: Generates recommendations based on:
  - User's listening history (last 30 days)
  - Most frequently listened podcasts
  - Similar podcasts to user's favorites
  - Falls back to trending content for new users

### 2. API Handlers (`internal/api/recommendations.go`)

HTTP handlers with full error handling and validation:

- `GET /api/recommendations/similar/:podcastId?limit=10`
- `GET /api/recommendations/trending?limit=10`
- `GET /api/recommendations/related/:videoId?limit=10`
- `GET /api/recommendations/for-you?limit=10`

All endpoints include:
- Input validation
- Rate limiting (30 requests/minute)
- Error handling with structured responses
- Swagger/OpenAPI documentation comments

### 3. OpenAPI Specification (`docs/openapi.yaml`)

Comprehensive OpenAPI 3.0 documentation including:
- All API endpoints with parameters and responses
- Request/response schemas
- Authentication requirements
- Error response formats
- Example requests and responses

### 4. Route Registration (`internal/app/controller.go`)

Routes registered in the main application:
- Recommendations API group with rate limiting
- Swagger UI at `/docs/*`

## API Endpoints

### Similar Podcasts

```bash
GET /api/recommendations/similar/:podcastId?limit=10
```

**Parameters:**
- `podcastId` (path, required): YouTube channel or playlist ID
- `limit` (query, optional): Number of results (1-50, default: 10)

**Response:**
```json
{
  "podcast_id": "UCXuqSBlHAE6Xw-yeJA0Tunw",
  "limit": 10,
  "results": 5,
  "similar": [
    {
      "podcast": {
        "id": "UC...",
        "podcast_name": "Tech Talks",
        "description": "...",
        "category": "Technology",
        ...
      },
      "similarity_score": 0.85
    }
  ]
}
```

### Trending Episodes

```bash
GET /api/recommendations/trending?limit=10
```

**Parameters:**
- `limit` (query, optional): Number of results (1-50, default: 10)

**Response:**
```json
{
  "period": "last_7_days",
  "limit": 10,
  "results": 10,
  "episodes": [
    {
      "episode": {
        "id": 12345,
        "youtube_video_id": "dQw4w9WgXcQ",
        "episode_name": "Latest Episode",
        ...
      },
      "play_count": 150
    }
  ]
}
```

### Related Episodes

```bash
GET /api/recommendations/related/:videoId?limit=10
```

**Parameters:**
- `videoId` (path, required): YouTube video ID
- `limit` (query, optional): Number of results (1-50, default: 10)

**Response:**
```json
{
  "video_id": "dQw4w9WgXcQ",
  "limit": 10,
  "results": 8,
  "episodes": [
    {
      "id": 12346,
      "youtube_video_id": "abc123",
      "episode_name": "Related Episode",
      "published_date": "2024-01-15T10:00:00Z",
      ...
    }
  ]
}
```

### Personalized Recommendations

```bash
GET /api/recommendations/for-you?limit=10
```

**Parameters:**
- `limit` (query, optional): Number of results (1-50, default: 10)

**Response:**
```json
{
  "strategy": "personalized_based_on_history",
  "limit": 10,
  "results": 10,
  "episodes": [
    {
      "id": 12347,
      "youtube_video_id": "xyz789",
      "episode_name": "Recommended Episode",
      ...
    }
  ]
}
```

## Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/docs/
```

The Swagger UI provides:
- Interactive API explorer
- Request/response examples
- Schema definitions
- Try-it-out functionality

## Recommendation Algorithms

### 1. Content Similarity

The similarity algorithm uses two factors:

1. **Category Match** (50% weight):
   - Exact match: 0.5
   - No match: 0.0

2. **Keyword Overlap** (50% weight):
   - Extracts keywords from podcast descriptions
   - Filters out common stop words
   - Uses Jaccard similarity: `intersection / union`
   - Maximum contribution: 0.5

**Final Score**: Category match + Keyword overlap (0.0 to 1.0)

### 2. Trending Algorithm

Based on playback analytics:
- Tracks episode plays in the last 7 days
- Aggregates play counts per episode
- Sorts by play count (highest first)
- Returns top N results

### 3. Personalized Recommendations

Multi-stage algorithm:

1. **History Analysis** (last 30 days):
   - Identifies most frequently listened podcasts
   - Tracks which episodes have been played

2. **Recommendation Generation**:
   - First: Unplayed episodes from favorite podcasts
   - Second: Episodes from similar podcasts
   - Fallback: Trending episodes for new users

## Rate Limiting

All recommendation endpoints are rate-limited to **30 requests per minute** per client.

Exceeding this limit returns:
```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded",
  "request_id": "req-123456789"
}
```

## Error Handling

All endpoints use structured error responses:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": {
    "field": "additional_context"
  },
  "request_id": "req-123456789"
}
```

Common error codes:
- `INVALID_PARAM`: Invalid parameter value
- `NOT_FOUND`: Resource not found
- `INTERNAL_ERROR`: Server error
- `BAD_REQUEST`: Malformed request
- `RATE_LIMIT_EXCEEDED`: Too many requests

## Dependencies

Added packages:
- `github.com/swaggo/swag` - Swagger generation
- `github.com/swaggo/echo-swagger` - Echo Swagger middleware
- `github.com/swaggo/files` - Static file serving for Swagger UI

## Testing Examples

### cURL Examples

```bash
# Get similar podcasts
curl "http://localhost:8080/api/recommendations/similar/UCXuqSBlHAE6Xw-yeJA0Tunw?limit=5"

# Get trending episodes
curl "http://localhost:8080/api/recommendations/trending?limit=10"

# Get related episodes
curl "http://localhost:8080/api/recommendations/related/dQw4w9WgXcQ"

# Get personalized recommendations
curl "http://localhost:8080/api/recommendations/for-you?limit=20"
```

### Using httpie

```bash
# Get similar podcasts
http GET localhost:8080/api/recommendations/similar/UCXuqSBlHAE6Xw-yeJA0Tunw limit==5

# Get trending episodes
http GET localhost:8080/api/recommendations/trending limit==10

# Get related episodes
http GET localhost:8080/api/recommendations/related/dQw4w9WgXcQ

# Get personalized recommendations
http GET localhost:8080/api/recommendations/for-you limit==20
```

## Future Enhancements

Potential improvements:
- Machine learning-based recommendations
- Collaborative filtering (users who liked X also liked Y)
- Time-based recommendations (time of day, day of week)
- Genre and tag-based filtering
- User feedback integration (likes/dislikes)
- Podcast popularity scoring
- Cross-podcast episode similarity
- Seasonal and trending topic detection

## Performance Considerations

- Similarity calculations are performed in-memory (suitable for moderate dataset sizes)
- Consider caching similarity scores for large podcast catalogs
- Trending calculations query the last 7 days of data only
- Personalized recommendations limit history to 30 days and 100 most recent plays
- All queries use indexed database columns for performance

## Database Schema

The recommendations service uses existing tables:
- `podcasts`: Podcast metadata (id, name, description, category)
- `podcast_episodes`: Episode information
- `episode_playback_history`: Play tracking with timestamps

No additional tables required.
