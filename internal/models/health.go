package models

// ComponentStatus represents the status of a single component
type ComponentStatus struct {
	Status  string `json:"status"`  // "ok", "degraded", "down"
	Message string `json:"message,omitempty"`
}

// HealthResponse represents the overall health status
type HealthResponse struct {
	Status     string                      `json:"status"` // "healthy", "degraded", "unhealthy"
	Components map[string]ComponentStatus  `json:"components"`
}

// ReadinessResponse represents the readiness status
type ReadinessResponse struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message,omitempty"`
}
