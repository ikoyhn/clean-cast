package api

import (
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/services/transcript"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TranscriptHandler handles transcript-related HTTP requests
type TranscriptHandler struct {
	service *transcript.Service
}

// NewTranscriptHandler creates a new transcript handler
func NewTranscriptHandler() *TranscriptHandler {
	youtubeService := youtube.SetupYoutubeService()
	return &TranscriptHandler{
		service: transcript.NewService(youtubeService),
	}
}

// GetTranscript handles GET /api/transcript/:videoId
// Query params: lang (default: en)
func (h *TranscriptHandler) GetTranscript(c echo.Context) error {
	videoId := c.Param("videoId")
	language := c.QueryParam("lang")

	if videoId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Video ID is required",
		})
	}

	if language == "" {
		language = "en"
	}

	logger.Logger.Debug().
		Str("video_id", videoId).
		Str("language", language).
		Msg("Fetching transcript")

	// Try to get from database first
	transcriptData, err := h.service.GetTranscript(videoId, language)
	if err != nil {
		// If not found, try to fetch from YouTube
		logger.Logger.Debug().
			Str("video_id", videoId).
			Str("language", language).
			Msg("Transcript not in database, fetching from YouTube")

		transcriptData, err = h.service.FetchTranscriptSimple(videoId, language)
		if err != nil {
			logger.Logger.Error().
				Err(err).
				Str("video_id", videoId).
				Str("language", language).
				Msg("Error fetching transcript")
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Transcript not available for this video",
			})
		}
	}

	return c.JSON(http.StatusOK, transcriptData)
}

// GetAllTranscripts handles GET /api/transcript/:videoId/all
// Returns all available transcripts for a video
func (h *TranscriptHandler) GetAllTranscripts(c echo.Context) error {
	videoId := c.Param("videoId")

	if videoId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Video ID is required",
		})
	}

	logger.Logger.Debug().
		Str("video_id", videoId).
		Msg("Fetching all transcripts")

	transcripts, err := h.service.GetAllTranscripts(videoId)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Msg("Error fetching transcripts")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch transcripts",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"video_id":    videoId,
		"transcripts": transcripts,
		"count":       len(transcripts),
	})
}

// GetAvailableLanguages handles GET /api/transcript/:videoId/languages
// Returns available languages for a video's transcripts
func (h *TranscriptHandler) GetAvailableLanguages(c echo.Context) error {
	videoId := c.Param("videoId")

	if videoId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Video ID is required",
		})
	}

	logger.Logger.Debug().
		Str("video_id", videoId).
		Msg("Fetching available languages")

	languages, err := h.service.GetAvailableLanguages(videoId)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Msg("Error fetching available languages")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch available languages",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"video_id":  videoId,
		"languages": languages,
		"count":     len(languages),
	})
}

// FetchTranscript handles POST /api/transcript/:videoId/fetch
// Explicitly fetches a transcript from YouTube
// Query params: lang (default: en)
func (h *TranscriptHandler) FetchTranscript(c echo.Context) error {
	videoId := c.Param("videoId")
	language := c.QueryParam("lang")

	if videoId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Video ID is required",
		})
	}

	if language == "" {
		language = "en"
	}

	logger.Logger.Info().
		Str("video_id", videoId).
		Str("language", language).
		Msg("Explicitly fetching transcript from YouTube")

	transcriptData, err := h.service.FetchTranscriptSimple(videoId, language)
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Str("language", language).
			Msg("Error fetching transcript")
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Failed to fetch transcript from YouTube",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Transcript fetched and saved successfully",
		"transcript": transcriptData,
	})
}
