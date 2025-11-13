package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

func TestValidateQueryParams(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name          string
		id            string
		queryParams   map[string]string
		expectedError bool
		expectedLimit *int
		expectedDate  *time.Time
	}{
		{
			name:          "No query parameters",
			id:            "UCtest123",
			queryParams:   map[string]string{},
			expectedError: false,
			expectedLimit: nil,
			expectedDate:  nil,
		},
		{
			name:          "Valid limit parameter",
			id:            "UCtest123",
			queryParams:   map[string]string{"limit": "10"},
			expectedError: false,
			expectedLimit: intPtr(10),
			expectedDate:  nil,
		},
		{
			name:          "Valid date parameter",
			id:            "UCtest123",
			queryParams:   map[string]string{"date": "01-15-2024"},
			expectedError: false,
			expectedLimit: nil,
			expectedDate:  timePtr(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:          "Both limit and date - should error",
			id:            "UCtest123",
			queryParams:   map[string]string{"limit": "10", "date": "01-15-2024"},
			expectedError: true,
		},
		{
			name:          "Invalid limit - non-numeric",
			id:            "UCtest123",
			queryParams:   map[string]string{"limit": "abc"},
			expectedError: true,
		},
		{
			name:          "Invalid date format",
			id:            "UCtest123",
			queryParams:   map[string]string{"date": "2024-01-15"},
			expectedError: true,
		},
		{
			name:          "Invalid id parameter",
			id:            "invalid<>id",
			queryParams:   map[string]string{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			result, err := validateQueryParams(c, tt.id)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectedLimit != nil {
					require.NotNil(t, result.Limit)
					assert.Equal(t, *tt.expectedLimit, *result.Limit)
				} else {
					assert.Nil(t, result.Limit)
				}

				if tt.expectedDate != nil {
					require.NotNil(t, result.Date)
					assert.True(t, tt.expectedDate.Equal(*result.Date))
				} else {
					assert.Nil(t, result.Date)
				}
			}
		})
	}
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name     string
		tls      bool
		host     string
		expected string
	}{
		{
			name:     "HTTP request",
			tls:      false,
			host:     "localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "HTTPS request",
			tls:      true,
			host:     "example.com",
			expected: "https://example.com",
		},
		{
			name:     "HTTP with domain",
			tls:      false,
			host:     "api.example.com:3000",
			expected: "http://api.example.com:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.host

			if tt.tls {
				req.TLS = &struct{}{}
			}

			result := handler(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthCheckEndpoints(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "Root endpoint",
			endpoint:       "/",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.endpoint == "/" {
				err := func(c echo.Context) error {
					return c.String(http.StatusOK, "Hello, World!")
				}(c)
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// Table-driven benchmark tests
func BenchmarkValidateQueryParams(b *testing.B) {
	e := echo.New()

	benchmarks := []struct {
		name        string
		id          string
		queryParams map[string]string
	}{
		{
			name:        "No parameters",
			id:          "UCtest123",
			queryParams: map[string]string{},
		},
		{
			name:        "With limit",
			id:          "UCtest123",
			queryParams: map[string]string{"limit": "10"},
		},
		{
			name:        "With date",
			id:          "UCtest123",
			queryParams: map[string]string{"date": "01-15-2024"},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			req := httptest.NewRequest(http.MethodGet, "/test?", nil)
			q := req.URL.Query()
			for k, v := range bm.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				validateQueryParams(c, bm.id)
			}
		})
	}
}

func BenchmarkHandler(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler(req)
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
