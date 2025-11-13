// Clean Cast Admin Panel - API Client

class CleanCastAPI {
  constructor(baseURL = '') {
    this.baseURL = baseURL;
    this.authToken = localStorage.getItem('authToken') || '';
  }

  /**
   * Set authentication token
   */
  setAuthToken(token) {
    this.authToken = token;
    localStorage.setItem('authToken', token);
  }

  /**
   * Get authentication token
   */
  getAuthToken() {
    return this.authToken;
  }

  /**
   * Clear authentication token
   */
  clearAuthToken() {
    this.authToken = '';
    localStorage.removeItem('authToken');
  }

  /**
   * Make HTTP request
   */
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers
    };

    if (this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    const config = {
      ...options,
      headers
    };

    try {
      const response = await fetch(url, config);

      // Handle non-JSON responses
      const contentType = response.headers.get('content-type');
      if (contentType && !contentType.includes('application/json')) {
        if (response.ok) {
          return await response.text();
        }
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();

      if (!response.ok) {
        throw {
          status: response.status,
          message: data.error || data.message || response.statusText,
          response: { data }
        };
      }

      return data;
    } catch (error) {
      console.error('API Request Error:', error);
      throw error;
    }
  }

  /**
   * GET request
   */
  async get(endpoint, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const url = queryString ? `${endpoint}?${queryString}` : endpoint;
    return this.request(url, { method: 'GET' });
  }

  /**
   * POST request
   */
  async post(endpoint, data = {}) {
    return this.request(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  /**
   * PUT request
   */
  async put(endpoint, data = {}) {
    return this.request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  /**
   * PATCH request
   */
  async patch(endpoint, data = {}) {
    return this.request(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(data)
    });
  }

  /**
   * DELETE request
   */
  async delete(endpoint) {
    return this.request(endpoint, { method: 'DELETE' });
  }

  // ==================== PODCAST ENDPOINTS ====================

  /**
   * Search podcasts
   */
  async searchPodcasts(query, limit = 20, offset = 0) {
    return this.get('/search/podcasts', { query, limit, offset });
  }

  /**
   * Get all podcasts (paginated)
   */
  async getPodcasts(limit = 20, offset = 0) {
    return this.get('/search/podcasts', { limit, offset });
  }

  /**
   * Batch add podcasts
   */
  async batchAddPodcasts(podcasts) {
    return this.post('/api/batch/podcasts/add', { podcasts });
  }

  /**
   * Batch refresh podcasts
   */
  async batchRefreshPodcasts(podcastIds) {
    return this.post('/api/batch/refresh', { podcast_ids: podcastIds });
  }

  // ==================== EPISODE ENDPOINTS ====================

  /**
   * Search episodes
   */
  async searchEpisodes(params = {}) {
    return this.get('/search/episodes', params);
  }

  /**
   * Batch delete episodes
   */
  async batchDeleteEpisodes(episodeIds) {
    return this.post('/api/batch/episodes/delete', { episode_ids: episodeIds });
  }

  // ==================== BATCH JOB ENDPOINTS ====================

  /**
   * Get batch job status
   */
  async getBatchJobStatus(jobId) {
    return this.get(`/api/batch/status/${jobId}`);
  }

  // ==================== ANALYTICS ENDPOINTS ====================

  /**
   * Get dashboard analytics
   */
  async getDashboardAnalytics(period = '7d') {
    return this.get('/api/analytics/dashboard', { period });
  }

  /**
   * Get analytics summary
   */
  async getAnalyticsSummary(period = '7d') {
    return this.get('/api/analytics/summary', { period });
  }

  /**
   * Get popular episodes
   */
  async getPopularEpisodes(limit = 10, period = '7d') {
    return this.get('/api/analytics/popular', { limit, period });
  }

  /**
   * Get episode analytics
   */
  async getEpisodeAnalytics(videoId, period = '7d') {
    return this.get(`/api/analytics/episode/${videoId}`, { period });
  }

  /**
   * Get geographic distribution
   */
  async getGeographicDistribution(period = '7d') {
    return this.get('/api/analytics/geographic', { period });
  }

  // ==================== WEBHOOK ENDPOINTS ====================

  /**
   * Get all webhooks
   */
  async getWebhooks() {
    return this.get('/api/webhooks');
  }

  /**
   * Get webhook by ID
   */
  async getWebhook(id) {
    return this.get(`/api/webhooks/${id}`);
  }

  /**
   * Create webhook
   */
  async createWebhook(webhook) {
    return this.post('/api/webhooks', webhook);
  }

  /**
   * Update webhook
   */
  async updateWebhook(id, webhook) {
    return this.put(`/api/webhooks/${id}`, webhook);
  }

  /**
   * Delete webhook
   */
  async deleteWebhook(id) {
    return this.delete(`/api/webhooks/${id}`);
  }

  /**
   * Get webhook deliveries
   */
  async getWebhookDeliveries(id, limit = 50) {
    return this.get(`/api/webhooks/${id}/deliveries`, { limit });
  }

  // ==================== BACKUP ENDPOINTS ====================

  /**
   * Create backup
   */
  async createBackup(includeAudio = false) {
    return this.post('/api/backup/create', { include_audio: includeAudio });
  }

  /**
   * List backups
   */
  async listBackups() {
    return this.get('/api/backup/list');
  }

  /**
   * Restore backup
   */
  async restoreBackup(filename) {
    return this.post('/api/backup/restore', { filename });
  }

  /**
   * Delete backup
   */
  async deleteBackup(filename) {
    return this.delete(`/api/backup/${filename}`);
  }

  /**
   * Download backup
   */
  async downloadBackup(filename) {
    const url = `${this.baseURL}/api/backup/download/${filename}`;
    window.open(url, '_blank');
  }

  // ==================== PREFERENCES ENDPOINTS ====================

  /**
   * Get user preferences
   */
  async getUserPreferences() {
    return this.get('/api/preferences');
  }

  /**
   * Update user preferences
   */
  async updateUserPreferences(preferences) {
    return this.put('/api/preferences', preferences);
  }

  /**
   * Get all feed preferences
   */
  async getAllFeedPreferences() {
    return this.get('/api/preferences/feeds');
  }

  /**
   * Get feed preferences
   */
  async getFeedPreferences(feedId) {
    return this.get(`/api/preferences/feed/${feedId}`);
  }

  /**
   * Update feed preferences
   */
  async updateFeedPreferences(feedId, preferences) {
    return this.put(`/api/preferences/feed/${feedId}`, preferences);
  }

  /**
   * Delete feed preferences
   */
  async deleteFeedPreferences(feedId) {
    return this.delete(`/api/preferences/feed/${feedId}`);
  }

  // ==================== CONTENT FILTER ENDPOINTS ====================

  /**
   * Get all filters
   */
  async getFilters() {
    return this.get('/api/filters');
  }

  /**
   * Get filter by ID
   */
  async getFilter(id) {
    return this.get(`/api/filters/${id}`);
  }

  /**
   * Create filter
   */
  async createFilter(filter) {
    return this.post('/api/filters', filter);
  }

  /**
   * Update filter
   */
  async updateFilter(id, filter) {
    return this.put(`/api/filters/${id}`, filter);
  }

  /**
   * Delete filter
   */
  async deleteFilter(id) {
    return this.delete(`/api/filters/${id}`);
  }

  /**
   * Toggle filter
   */
  async toggleFilter(id) {
    return this.patch(`/api/filters/${id}/toggle`);
  }

  // ==================== SMART PLAYLIST ENDPOINTS ====================

  /**
   * Get all smart playlists
   */
  async getSmartPlaylists() {
    return this.get('/api/playlist/smart');
  }

  /**
   * Get smart playlist by ID
   */
  async getSmartPlaylist(id) {
    return this.get(`/api/playlist/smart/${id}`);
  }

  /**
   * Create smart playlist
   */
  async createSmartPlaylist(playlist) {
    return this.post('/api/playlist/smart', playlist);
  }

  /**
   * Update smart playlist
   */
  async updateSmartPlaylist(id, playlist) {
    return this.put(`/api/playlist/smart/${id}`, playlist);
  }

  /**
   * Delete smart playlist
   */
  async deleteSmartPlaylist(id) {
    return this.delete(`/api/playlist/smart/${id}`);
  }

  /**
   * Get smart playlist RSS feed URL
   */
  getSmartPlaylistRSSUrl(id) {
    return `${this.baseURL}/rss/smart/${id}`;
  }

  // ==================== OPML ENDPOINTS ====================

  /**
   * Import OPML
   */
  async importOPML(opmlContent) {
    return this.post('/api/opml/import', { opml: opmlContent });
  }

  /**
   * Export OPML
   */
  async exportOPML() {
    const url = `${this.baseURL}/api/opml/export`;
    window.open(url, '_blank');
  }

  // ==================== TRANSCRIPT ENDPOINTS ====================

  /**
   * Get transcript for video
   */
  async getTranscript(videoId, language = '') {
    return this.get(`/api/transcript/${videoId}`, language ? { language } : {});
  }

  /**
   * Get all transcripts for video
   */
  async getAllTranscripts(videoId) {
    return this.get(`/api/transcript/${videoId}/all`);
  }

  /**
   * Get available transcript languages
   */
  async getTranscriptLanguages(videoId) {
    return this.get(`/api/transcript/${videoId}/languages`);
  }

  /**
   * Fetch transcript from YouTube
   */
  async fetchTranscript(videoId, language = '') {
    return this.post(`/api/transcript/${videoId}/fetch`, language ? { language } : {});
  }

  // ==================== RECOMMENDATIONS ENDPOINTS ====================

  /**
   * Get recommended episodes
   */
  async getRecommendations(limit = 10) {
    return this.get('/api/recommendations', { limit });
  }

  /**
   * Get similar episodes
   */
  async getSimilarEpisodes(videoId, limit = 5) {
    return this.get(`/api/recommendations/similar/${videoId}`, { limit });
  }

  // ==================== HEALTH ENDPOINTS ====================

  /**
   * Health check
   */
  async healthCheck() {
    return this.get('/health');
  }

  /**
   * Readiness check
   */
  async readinessCheck() {
    return this.get('/ready');
  }
}

// Create global API instance
const api = new CleanCastAPI();

// Export for use in modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = CleanCastAPI;
}
