# Controller.go RSS Endpoint Patches

The following changes need to be applied to `/home/user/clean-cast/internal/app/controller.go` to add format support to RSS endpoints:

## Patch 1: Channel RSS Endpoint

**Location:** Around line 85-95

**Replace:**
```go
e.GET("/channel/:channelId", func(c echo.Context) error {
    rssRequestParams, err := validateQueryParams(c, c.Param("channelId"))
    if err != nil {
        return err
    }
    data := rss.BuildChannelRssFeed(c.Param("channelId"), rssRequestParams, handler(c.Request()))
    c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
    c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
    c.Response().Header().Del("Transfer-Encoding")
    return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
```

**With:**
```go
e.GET("/channel/:channelId", func(c echo.Context) error {
    rssRequestParams, err := validateQueryParams(c, c.Param("channelId"))
    if err != nil {
        return err
    }
    // Get format from query parameter or use default
    formatParam := c.QueryParam("format")
    if formatParam == "" {
        formatParam = config.Config.AudioFormat
    }
    audioFormat := downloader.GetAudioFormat(formatParam)

    data := rss.BuildChannelRssFeed(c.Param("channelId"), rssRequestParams, handler(c.Request()), audioFormat)
    c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
    c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
    c.Response().Header().Del("Transfer-Encoding")
    return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
```

## Patch 2: Playlist RSS Endpoint

**Location:** Around line 97-107

**Replace:**
```go
e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
    rssRequestParams, err := validateQueryParams(c, c.Param("youtubePlaylistId"))
    if err != nil {
        return err
    }
    data := playlist.BuildPlaylistRssFeed(c.Param("youtubePlaylistId"), rssRequestParams, handler(c.Request()))
    c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
    c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
    c.Response().Header().Del("Transfer-Encoding")
    return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
```

**With:**
```go
e.GET("/rss/:youtubePlaylistId", func(c echo.Context) error {
    rssRequestParams, err := validateQueryParams(c, c.Param("youtubePlaylistId"))
    if err != nil {
        return err
    }
    // Get format from query parameter or use default
    formatParam := c.QueryParam("format")
    if formatParam == "" {
        formatParam = config.Config.AudioFormat
    }
    audioFormat := downloader.GetAudioFormat(formatParam)

    data := playlist.BuildPlaylistRssFeed(c.Param("youtubePlaylistId"), rssRequestParams, handler(c.Request()), audioFormat)
    c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
    c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
    c.Response().Header().Del("Transfer-Encoding")
    return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", data)
}, middleware.AuthMiddleware(), middleware.RateLimitMiddleware(10))
```

## Summary of Changes

Both patches add the following logic before calling the RSS build functions:

1. Get the `format` query parameter from the request
2. If not provided, use the default from configuration (`config.Config.AudioFormat`)
3. Convert the format string to an `AudioFormat` struct using `downloader.GetAudioFormat()`
4. Pass the `audioFormat` as the 4th parameter to the RSS build functions

This allows users to request RSS feeds with specific audio formats:
- `/channel/CHANNEL_ID?format=mp3`
- `/rss/PLAYLIST_ID?format=opus`

If no format is specified, it uses the system default configured via the `AUDIO_FORMAT` environment variable.
