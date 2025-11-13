package api

import (
	"encoding/json"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/cache"
	"ikoyhn/podcast-sponsorblock/internal/database"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/channel"
	"ikoyhn/podcast-sponsorblock/internal/services/playlist"
	"ikoyhn/podcast-sponsorblock/internal/services/webhook"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/labstack/echo/v4"
)

// BatchRefreshPodcasts handles bulk refresh of podcast metadata
func BatchRefreshPodcasts(c echo.Context) error {
	var req models.BatchRefreshRequest
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	if len(req.PodcastIds) == 0 {
		return appErrors.NewBadRequestError("podcast_ids cannot be empty")
	}

	if len(req.PodcastIds) > 50 {
		return appErrors.NewBadRequestError("Cannot refresh more than 50 podcasts at once")
	}

	// Create batch job
	job := &models.BatchJobStatus{
		JobType:    "refresh",
		Status:     "pending",
		TotalItems: len(req.PodcastIds),
	}

	if err := database.SaveBatchJob(job); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create batch job")
		return appErrors.NewInternalServerError("Failed to create batch job")
	}

	// Process batch in background
	go processBatchRefresh(job.Id, req.PodcastIds)

	return c.JSON(http.StatusAccepted, models.BatchOperationResponse{
		JobId:   job.Id,
		Status:  "pending",
		Message: fmt.Sprintf("Batch refresh job created for %d podcasts", len(req.PodcastIds)),
	})
}

// BatchDeleteEpisodes handles bulk deletion of episodes
func BatchDeleteEpisodes(c echo.Context) error {
	var req models.BatchDeleteEpisodesRequest
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	if len(req.EpisodeIds) == 0 {
		return appErrors.NewBadRequestError("episode_ids cannot be empty")
	}

	if len(req.EpisodeIds) > 100 {
		return appErrors.NewBadRequestError("Cannot delete more than 100 episodes at once")
	}

	// Create batch job
	job := &models.BatchJobStatus{
		JobType:    "delete_episodes",
		Status:     "pending",
		TotalItems: len(req.EpisodeIds),
	}

	if err := database.SaveBatchJob(job); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create batch job")
		return appErrors.NewInternalServerError("Failed to create batch job")
	}

	// Process batch in background
	go processBatchDeleteEpisodes(job.Id, req.EpisodeIds)

	return c.JSON(http.StatusAccepted, models.BatchOperationResponse{
		JobId:   job.Id,
		Status:  "pending",
		Message: fmt.Sprintf("Batch delete job created for %d episodes", len(req.EpisodeIds)),
	})
}

// BatchAddPodcasts handles bulk addition of podcasts
func BatchAddPodcasts(c echo.Context) error {
	var req models.BatchAddPodcastsRequest
	if err := c.Bind(&req); err != nil {
		return appErrors.NewBadRequestError("Invalid request body")
	}

	if len(req.Podcasts) == 0 {
		return appErrors.NewBadRequestError("podcasts cannot be empty")
	}

	if len(req.Podcasts) > 20 {
		return appErrors.NewBadRequestError("Cannot add more than 20 podcasts at once")
	}

	// Validate podcast types
	for _, p := range req.Podcasts {
		if p.Type != "playlist" && p.Type != "channel" {
			return appErrors.NewBadRequestError(fmt.Sprintf("Invalid podcast type: %s. Must be 'playlist' or 'channel'", p.Type))
		}
	}

	// Create batch job
	metadata, _ := json.Marshal(req.Podcasts)
	job := &models.BatchJobStatus{
		JobType:    "add_podcasts",
		Status:     "pending",
		TotalItems: len(req.Podcasts),
		Metadata:   string(metadata),
	}

	if err := database.SaveBatchJob(job); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to create batch job")
		return appErrors.NewInternalServerError("Failed to create batch job")
	}

	// Process batch in background
	go processBatchAddPodcasts(job.Id, req.Podcasts)

	return c.JSON(http.StatusAccepted, models.BatchOperationResponse{
		JobId:   job.Id,
		Status:  "pending",
		Message: fmt.Sprintf("Batch add job created for %d podcasts", len(req.Podcasts)),
	})
}

// GetBatchJobStatus returns the status of a batch job
func GetBatchJobStatus(c echo.Context) error {
	jobIdStr := c.Param("jobId")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 32)
	if err != nil {
		return appErrors.NewInvalidParamError("jobId").WithDetail("value", jobIdStr)
	}

	job := database.GetBatchJobStatus(int32(jobId))
	if job == nil {
		return appErrors.NewNotFoundError("Batch job not found")
	}

	return c.JSON(http.StatusOK, models.BatchStatusResponse{
		JobId:          job.Id,
		JobType:        job.JobType,
		Status:         job.Status,
		TotalItems:     job.TotalItems,
		ProcessedItems: job.ProcessedItems,
		FailedItems:    job.FailedItems,
		ErrorMessage:   job.ErrorMessage,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
		CompletedAt:    job.CompletedAt,
	})
}

// Background processing functions

func processBatchRefresh(jobId int32, podcastIds []string) {
	logger.Logger.Info().
		Int32("job_id", jobId).
		Int("count", len(podcastIds)).
		Msg("Starting batch refresh job")

	// Update job status to running
	job := database.GetBatchJobStatus(jobId)
	if job == nil {
		return
	}
	job.Status = "running"
	database.UpdateBatchJob(job)

	service := youtube.SetupYoutubeService()
	processed := 0
	failed := 0

	for _, podcastId := range podcastIds {
		// Check if it's a playlist or channel
		podcast := database.GetPodcast(podcastId)
		if podcast == nil {
			logger.Logger.Warn().
				Str("podcast_id", podcastId).
				Msg("Podcast not found, skipping")
			failed++
			database.UpdateBatchJobProgress(jobId, processed, failed)
			continue
		}

		// Clear cache for this podcast
		rssFeedCache := cache.GetRSSFeedCache()
		rssFeedCache.InvalidateByPodcastId(podcast.Id)

		// Refresh metadata
		params := &models.RssRequestParams{Limit: nil, Date: nil}

		// Determine if it's a playlist or channel and refresh accordingly
		if hasPlaylistEpisodes(podcastId) {
			playlist.BuildPlaylistRssFeed(podcastId, params, "")
		} else {
			// It's a channel
			channel.GetChannelMetadataAndVideos(podcastId, service, params)
		}

		processed++
		database.UpdateBatchJobProgress(jobId, processed, failed)

		logger.Logger.Debug().
			Str("podcast_id", podcastId).
			Int("processed", processed).
			Int("total", len(podcastIds)).
			Msg("Refreshed podcast metadata")
	}

	// Mark job as completed
	if failed > 0 && processed == 0 {
		database.CompleteBatchJob(jobId, "failed", "All podcast refreshes failed")
	} else if failed > 0 {
		database.CompleteBatchJob(jobId, "completed", fmt.Sprintf("Completed with %d failures", failed))
	} else {
		database.CompleteBatchJob(jobId, "completed", "")
	}

	logger.Logger.Info().
		Int32("job_id", jobId).
		Int("processed", processed).
		Int("failed", failed).
		Msg("Batch refresh job completed")
}

func processBatchDeleteEpisodes(jobId int32, episodeIds []string) {
	logger.Logger.Info().
		Int32("job_id", jobId).
		Int("count", len(episodeIds)).
		Msg("Starting batch delete episodes job")

	// Update job status to running
	job := database.GetBatchJobStatus(jobId)
	if job == nil {
		return
	}
	job.Status = "running"
	database.UpdateBatchJob(job)

	processed := 0
	failed := 0

	for _, episodeId := range episodeIds {
		// Delete the audio file
		audioFile := path.Join("/config/audio", episodeId+".m4a")
		if err := os.Remove(audioFile); err != nil {
			if !os.IsNotExist(err) {
				logger.Logger.Error().
					Err(err).
					Str("episode_id", episodeId).
					Msg("Failed to delete episode file")
				failed++
				database.UpdateBatchJobProgress(jobId, processed, failed)
				continue
			}
		}

		// Delete from database (playback history)
		db := database.GetDB()
		if err := db.Delete(&models.EpisodePlaybackHistory{}, "youtube_video_id = ?", episodeId).Error; err != nil {
			logger.Logger.Error().
				Err(err).
				Str("episode_id", episodeId).
				Msg("Failed to delete episode from database")
			failed++
			database.UpdateBatchJobProgress(jobId, processed, failed)
			continue
		}

		processed++
		database.UpdateBatchJobProgress(jobId, processed, failed)

		logger.Logger.Debug().
			Str("episode_id", episodeId).
			Int("processed", processed).
			Int("total", len(episodeIds)).
			Msg("Deleted episode")
	}

	// Mark job as completed
	if failed > 0 && processed == 0 {
		database.CompleteBatchJob(jobId, "failed", "All episode deletions failed")
	} else if failed > 0 {
		database.CompleteBatchJob(jobId, "completed", fmt.Sprintf("Completed with %d failures", failed))
	} else {
		database.CompleteBatchJob(jobId, "completed", "")
	}

	logger.Logger.Info().
		Int32("job_id", jobId).
		Int("processed", processed).
		Int("failed", failed).
		Msg("Batch delete episodes job completed")
}

func processBatchAddPodcasts(jobId int32, podcasts []models.BatchPodcastItem) {
	logger.Logger.Info().
		Int32("job_id", jobId).
		Int("count", len(podcasts)).
		Msg("Starting batch add podcasts job")

	// Update job status to running
	job := database.GetBatchJobStatus(jobId)
	if job == nil {
		return
	}
	job.Status = "running"
	database.UpdateBatchJob(job)

	service := youtube.SetupYoutubeService()
	processed := 0
	failed := 0

	for _, p := range podcasts {
		// Check if podcast already exists
		exists, err := database.PodcastExists(p.Id)
		if err != nil {
			logger.Logger.Error().
				Err(err).
				Str("podcast_id", p.Id).
				Msg("Failed to check if podcast exists")
			failed++
			database.UpdateBatchJobProgress(jobId, processed, failed)
			continue
		}

		if exists {
			logger.Logger.Debug().
				Str("podcast_id", p.Id).
				Msg("Podcast already exists, skipping")
			processed++
			database.UpdateBatchJobProgress(jobId, processed, failed)
			continue
		}

		// Add the podcast
		params := &models.RssRequestParams{Limit: nil, Date: nil}

		if p.Type == "playlist" {
			// Get playlist data and save
			podcast := youtube.GetChannelData(p.Id, service, true)
			if podcast == nil {
				logger.Logger.Error().
					Str("podcast_id", p.Id).
					Msg("Failed to get playlist data")
				failed++
				database.UpdateBatchJobProgress(jobId, processed, failed)

				// Send error webhook
				webhook.SendWebhook(webhook.EventError, map[string]interface{}{
					"error":      "Failed to get playlist data",
					"podcast_id": p.Id,
					"context":    "batch_add_podcasts",
				})
				continue
			}
			database.SavePodcast(podcast)
			playlist.BuildPlaylistRssFeed(p.Id, params, "")
		} else {
			// It's a channel
			podcast := youtube.GetChannelData(p.Id, service, false)
			if podcast == nil {
				logger.Logger.Error().
					Str("podcast_id", p.Id).
					Msg("Failed to get channel data")
				failed++
				database.UpdateBatchJobProgress(jobId, processed, failed)

				// Send error webhook
				webhook.SendWebhook(webhook.EventError, map[string]interface{}{
					"error":      "Failed to get channel data",
					"podcast_id": p.Id,
					"context":    "batch_add_podcasts",
				})
				continue
			}
			database.SavePodcast(podcast)
			channel.GetChannelMetadataAndVideos(p.Id, service, params)
		}

		processed++
		database.UpdateBatchJobProgress(jobId, processed, failed)

		logger.Logger.Debug().
			Str("podcast_id", p.Id).
			Str("type", p.Type).
			Int("processed", processed).
			Int("total", len(podcasts)).
			Msg("Added podcast")
	}

	// Mark job as completed
	if failed > 0 && processed == 0 {
		database.CompleteBatchJob(jobId, "failed", "All podcast additions failed")
	} else if failed > 0 {
		database.CompleteBatchJob(jobId, "completed", fmt.Sprintf("Completed with %d failures", failed))
	} else {
		database.CompleteBatchJob(jobId, "completed", "")
	}

	logger.Logger.Info().
		Int32("job_id", jobId).
		Int("processed", processed).
		Int("failed", failed).
		Msg("Batch add podcasts job completed")
}

// Helper function to determine if a podcast has playlist episodes
func hasPlaylistEpisodes(podcastId string) bool {
	db := database.GetDB()
	var count int64
	db.Model(&models.PodcastEpisode{}).
		Where("podcast_id = ? AND type = ?", podcastId, "PLAYLIST").
		Count(&count)
	return count > 0
}
