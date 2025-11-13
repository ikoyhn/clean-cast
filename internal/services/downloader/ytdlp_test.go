package downloader

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ikoyhn/podcast-sponsorblock/internal/config"
)

func TestSetDownloadCallbacks(t *testing.T) {
	startCalled := false
	completeCalled := false

	onStart := func() {
		startCalled = true
	}

	onComplete := func() {
		completeCalled = true
	}

	SetDownloadCallbacks(onStart, onComplete)

	// Verify callbacks were set (we can't test them directly, but we can ensure no panic)
	assert.NotNil(t, onDownloadStart)
	assert.NotNil(t, onDownloadComplete)
}

func TestGetYoutubeVideo_FileAlreadyExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ytdlp_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up config
	oldAudioDir := config.Config.AudioDir
	config.Config.AudioDir = tempDir + "/"
	defer func() {
		config.Config.AudioDir = oldAudioDir
	}()

	// Create a test file
	videoId := "test_video_id"
	testFilePath := filepath.Join(tempDir, videoId+".m4a")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test
	result, done := GetYoutubeVideo(videoId + ".m4a")

	// Wait for completion
	<-done

	// Assertions
	assert.Equal(t, videoId, result)
}

func TestGetYoutubeVideo_ConcurrentRequests(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ytdlp_test_concurrent_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up config
	oldAudioDir := config.Config.AudioDir
	config.Config.AudioDir = tempDir + "/"
	defer func() {
		config.Config.AudioDir = oldAudioDir
	}()

	// Create a test file to simulate existing file
	videoId := "concurrent_test_video"
	testFilePath := filepath.Join(tempDir, videoId+".m4a")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test concurrent requests
	var wg sync.WaitGroup
	numGoroutines := 10
	results := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, done := GetYoutubeVideo(videoId + ".m4a")
			<-done
			results[index] = result
		}(i)
	}

	wg.Wait()

	// All results should be the same
	for _, result := range results {
		assert.Equal(t, videoId, result)
	}
}

func TestGetYoutubeVideo_StripM4aExtension(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ytdlp_test_strip_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up config
	oldAudioDir := config.Config.AudioDir
	config.Config.AudioDir = tempDir + "/"
	defer func() {
		config.Config.AudioDir = oldAudioDir
	}()

	videoId := "test_strip_extension"

	// Create file without .m4a initially
	testFilePath := filepath.Join(tempDir, videoId+".m4a")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test with .m4a extension - should be stripped
	result, done := GetYoutubeVideo(videoId + ".m4a")
	<-done

	assert.Equal(t, videoId, result)
}

func TestDownloadCallbacks(t *testing.T) {
	startCalled := false
	completeCalled := false
	var mu sync.Mutex

	onStart := func() {
		mu.Lock()
		defer mu.Unlock()
		startCalled = true
	}

	onComplete := func() {
		mu.Lock()
		defer mu.Unlock()
		completeCalled = true
	}

	SetDownloadCallbacks(onStart, onComplete)

	// We can't easily test the actual download without external dependencies,
	// but we can verify the callbacks are set properly
	assert.NotNil(t, onDownloadStart)
	assert.NotNil(t, onDownloadComplete)

	// Call them directly to test they work
	onDownloadStart()
	onDownloadComplete()

	mu.Lock()
	defer mu.Unlock()
	assert.True(t, startCalled)
	assert.True(t, completeCalled)
}

func TestYoutubeVideoMutexManagement(t *testing.T) {
	// Test that mutexes are properly managed
	videoId := "mutex_test_video"

	// Store a mutex
	mutex := &sync.Mutex{}
	youtubeVideoMutexes.Store(videoId, mutex)

	// Load it back
	loadedMutex, ok := youtubeVideoMutexes.Load(videoId)
	assert.True(t, ok)
	assert.Equal(t, mutex, loadedMutex)

	// Delete it
	youtubeVideoMutexes.Delete(videoId)

	// Verify it's gone
	_, ok = youtubeVideoMutexes.Load(videoId)
	assert.False(t, ok)
}

// Benchmark tests
func BenchmarkGetYoutubeVideo_ExistingFile(b *testing.B) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ytdlp_bench_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Set up config
	oldAudioDir := config.Config.AudioDir
	config.Config.AudioDir = tempDir + "/"
	defer func() {
		config.Config.AudioDir = oldAudioDir
	}()

	// Create a test file
	videoId := "bench_video_id"
	testFilePath := filepath.Join(tempDir, videoId+".m4a")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, done := GetYoutubeVideo(videoId + ".m4a")
		<-done
	}
}

func BenchmarkMutexOperations(b *testing.B) {
	videoIds := []string{"video1", "video2", "video3", "video4", "video5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		videoId := videoIds[i%len(videoIds)]

		// Store
		mutex := &sync.Mutex{}
		youtubeVideoMutexes.Store(videoId, mutex)

		// Load
		youtubeVideoMutexes.Load(videoId)

		// Delete
		youtubeVideoMutexes.Delete(videoId)
	}
}

func BenchmarkConcurrentMutexAccess(b *testing.B) {
	videoId := "concurrent_bench_video"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mutex, ok := youtubeVideoMutexes.Load(videoId)
			if !ok {
				mutex = &sync.Mutex{}
				youtubeVideoMutexes.Store(videoId, mutex)
			}

			// Simulate work
			mutex.(*sync.Mutex).Lock()
			time.Sleep(1 * time.Microsecond)
			mutex.(*sync.Mutex).Unlock()

			youtubeVideoMutexes.Delete(videoId)
		}
	})
}
