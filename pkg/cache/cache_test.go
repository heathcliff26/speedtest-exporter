package cache

import (
	"os"
	"testing"
	"time"

	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	tMatrix := []struct {
		Name             string
		Persist          bool
		Path             string
		ExpectedPersist  bool
		ShouldHaveResult bool
	}{
		{
			Name:             "EmptyPath",
			Persist:          true,
			Path:             "",
			ExpectedPersist:  false,
			ShouldHaveResult: false,
		},
		{
			Name:             "NoPersist",
			Persist:          false,
			Path:             "testdata/result.json",
			ExpectedPersist:  false,
			ShouldHaveResult: false,
		},
		{
			Name:             "PathDoesNotExist",
			Persist:          true,
			Path:             "/nonexistent/path/result.json",
			ExpectedPersist:  false,
			ShouldHaveResult: false,
		},
		{
			Name:             "InvalidJSON",
			Persist:          true,
			Path:             "testdata/not-json.txt",
			ExpectedPersist:  true,
			ShouldHaveResult: false,
		},
		{
			Name:             "InitializeFromFile",
			Persist:          true,
			Path:             "testdata/result.json",
			ExpectedPersist:  true,
			ShouldHaveResult: true,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)

			cache := NewCache(tCase.Persist, tCase.Path, time.Minute+cacheTimeGracePeriod)

			assert.Equal(tCase.ExpectedPersist, cache.persist, "Persist flag should be set correctly")
			assert.Equal(tCase.Path, cache.path, "Path should be set correctly")
			assert.Equal(time.Minute, cache.cacheTime, "Cache time should be set correctly")
			if tCase.ShouldHaveResult {
				assert.NotNil(cache.cachedResult, "Cached result should be initialized")
			} else {
				assert.Nil(cache.cachedResult, "Cached result should be empty")
			}
		})
	}

	t.Run("CacheTimeGracePeriod", func(t *testing.T) {
		assert := assert.New(t)

		cacheTime := cacheTimeGracePeriod
		cache := NewCache(false, "", cacheTime)
		assert.Equal(cacheTime, cache.cacheTime, "Cache time should be used directly when equal to grace period")

		cacheTime = cacheTimeGracePeriod - time.Second
		cache = NewCache(false, "", cacheTime)
		assert.Equal(cacheTime, cache.cacheTime, "Cache time should be used directly when less than grace period")

		cacheTime = cacheTimeGracePeriod + time.Second
		cache = NewCache(false, "", cacheTime)
		assert.Equal(time.Second, cache.cacheTime, "Cache time should be reduced by grace period when greater than grace period")
	})
}

func TestCacheNil(t *testing.T) {
	c := (*Cache)(nil)

	t.Run("Read", func(t *testing.T) {
		assert := assert.New(t)
		assert.NotPanics(func() {
			_, _ = c.Read()
		}, "Read should not panic on nil Cache")

		result, valid := c.Read()
		assert.Nil(result, "Cache should not return a result")
		assert.False(valid, "Cache should not be valid")
	})
	t.Run("Save", func(t *testing.T) {
		assert := assert.New(t)
		assert.NotPanics(func() {
			c.Save(speedtest.NewFailedSpeedtestResult())
		}, "Save should not panic on nil Cache")
	})
	t.Run("ExpiresAt", func(t *testing.T) {
		assert := assert.New(t)
		assert.NotPanics(func() {
			_ = c.ExpiresAt()
		}, "ExpiresAt should not panic on nil Cache")
		assert.Zero(c.ExpiresAt(), "Cache should return zero time")
	})
}

func TestRead(t *testing.T) {
	t.Run("EmptyCache", func(t *testing.T) {
		assert := assert.New(t)

		c := &Cache{
			cacheTime: time.Minute,
		}

		result, valid := c.Read()
		assert.Nil(result, "Should not return a result")
		assert.False(valid, "Cache should not be valid")
	})
	t.Run("ValidResult", func(t *testing.T) {
		assert := assert.New(t)

		expectedResult := speedtest.NewFailedSpeedtestResult()
		c := &Cache{
			cacheTime:    time.Minute,
			cachedResult: expectedResult,
		}

		result, valid := c.Read()
		assert.Equal(expectedResult, result, "Should return cached result")
		assert.True(valid, "Cache should be valid")
	})
	t.Run("ExpiredCache", func(t *testing.T) {
		assert := assert.New(t)

		expectedResult := speedtest.MockSpeedtestResult(time.Now().Add(-10 * time.Minute).UnixMilli())
		c := &Cache{
			cacheTime:    time.Minute,
			cachedResult: expectedResult,
		}

		result, valid := c.Read()
		assert.Equal(expectedResult, result, "Should return cached result")
		assert.False(valid, "Cache should not be valid")
	})
}

func TestSave(t *testing.T) {
	t.Run("DoNotPersist", func(t *testing.T) {
		assert := assert.New(t)

		c := &Cache{
			path:      t.TempDir() + "/cache_test_save.json",
			cacheTime: time.Minute,
			persist:   false,
		}

		expectedResult := speedtest.NewFailedSpeedtestResult()
		c.Save(expectedResult)

		_, err := os.Stat(c.path)
		assert.True(os.IsNotExist(err), "Cache file should not be created when persist is false")
		assert.Equal(expectedResult, c.cachedResult, "Should cache the result")
	})
	t.Run("PersistToDisk", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		c := &Cache{
			path:      t.TempDir() + "/cache_test_save.json",
			cacheTime: time.Minute,
			persist:   true,
		}

		expectedResult := speedtest.NewFailedSpeedtestResult()
		c.Save(expectedResult)

		assert.Equal(expectedResult, c.cachedResult, "Should cache the result")

		data, err := os.ReadFile(c.path)
		require.NoError(err, "Cache file should be created when persist is true")

		diskResult := &speedtest.SpeedtestResult{}
		err = diskResult.UnmarshalJSON(data)
		require.NoError(err, "Should unmarshal cached result from disk")
		assert.Equal(expectedResult, diskResult, "Cached result on disk should match expected result")
	})
}

func TestExpiresAt(t *testing.T) {
	assert := assert.New(t)

	expectedResult := speedtest.NewFailedSpeedtestResult()
	c := &Cache{
		cacheTime:    time.Minute,
		cachedResult: expectedResult,
	}
	expectedExpiry := expectedResult.TimestampAsTime().Add(c.cacheTime)

	assert.Equal(expectedExpiry, c.ExpiresAt(), "ExpiresAt should return correct expiry time")
}
