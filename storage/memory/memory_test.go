package memory

import (
	"testing"

	"github.com/eqweqr/urlshortener/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	t.Run("should create new storage with initialized maps", func(t *testing.T) {
		store := NewStorage()

		assert.NotNil(t, store)
		assert.NotNil(t, store.shorten)
		assert.NotNil(t, store.originals)
		assert.Equal(t, uint64(1), store.nextId)
		assert.Empty(t, store.shorten)
		assert.Empty(t, store.originals)
	})
}

func TestMemoryStorage_SaveId(t *testing.T) {
	t.Run("should save new URL mapping successfully", func(t *testing.T) {
		store := NewStorage()
		originalURL := "https://example.com"
		shortenURL := "abc123"

		err := store.SaveId(originalURL, shortenURL)

		assert.NoError(t, err)
		assert.Equal(t, originalURL, store.originals[shortenURL])
		assert.Equal(t, shortenURL, store.shorten[originalURL])
	})

	t.Run("should return nil when saving duplicate original URL", func(t *testing.T) {
		store := NewStorage()
		originalURL := "https://example.com"
		shortenURL1 := "abc123"
		shortenURL2 := "def456"

		err1 := store.SaveId(originalURL, shortenURL1)
		assert.NoError(t, err1)

		err2 := store.SaveId(originalURL, shortenURL2)

		assert.NoError(t, err2)
		assert.Equal(t, originalURL, store.originals[shortenURL1])
		assert.Equal(t, shortenURL1, store.shorten[originalURL])
		assert.Empty(t, store.originals[shortenURL2])
	})

	t.Run("should handle concurrent saves safely", func(t *testing.T) {
		store := NewStorage()

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func(i int) {
				url := "https://example.com/" + string(rune(i))
				_ = store.SaveId(url, "short"+string(rune(i)))
				done <- true
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}

		assert.Equal(t, 100, len(store.shorten))
		assert.Equal(t, 100, len(store.originals))
	})
}

func TestMemoryStorage_GetOriginalId(t *testing.T) {
	t.Run("should return original URL for existing shorten URL", func(t *testing.T) {
		store := NewStorage()
		originalURL := "https://example.com"
		shortenURL := "abc123"

		_ = store.SaveId(originalURL, shortenURL)

		result, err := store.GetOriginalId(shortenURL)

		assert.NoError(t, err)
		assert.Equal(t, originalURL, result)
	})

	t.Run("should return error for non-existent shorten URL", func(t *testing.T) {
		store := NewStorage()

		result, err := store.GetOriginalId("nonexistent")

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "No associated value")
	})

	t.Run("should handle concurrent reads safely", func(t *testing.T) {
		store := NewStorage()
		originalURL := "https://example.com"
		shortenURL := "abc123"
		_ = store.SaveId(originalURL, shortenURL)

		done := make(chan bool)
		for i := 0; i < 50; i++ {
			go func() {
				result, err := store.GetOriginalId(shortenURL)
				assert.NoError(t, err)
				assert.Equal(t, originalURL, result)
				done <- true
			}()
		}

		for i := 0; i < 50; i++ {
			<-done
		}
	})
}

func TestMemoryStorage_GetShortenId(t *testing.T) {
	t.Run("should return shorten URL for existing original URL", func(t *testing.T) {
		store := NewStorage()
		originalURL := "https://example.com"
		shortenURL := "abc123"

		_ = store.SaveId(originalURL, shortenURL)

		result, err := store.GetShortenId(originalURL)

		assert.NoError(t, err)
		assert.Equal(t, shortenURL, result)
	})

	t.Run("should return storage.ErrNotFound for non-existent original URL", func(t *testing.T) {
		store := NewStorage()

		result, err := store.GetShortenId("https://nonexistent.com")

		assert.Error(t, err)
		assert.Equal(t, storage.ErrNotFound, err)
		assert.Empty(t, result)
	})

	t.Run("should handle concurrent reads safely", func(t *testing.T) {
		store := NewStorage()
		originalURL := "https://example.com"
		shortenURL := "abc123"
		_ = store.SaveId(originalURL, shortenURL)

		done := make(chan bool)
		for i := 0; i < 50; i++ {
			go func() {
				result, err := store.GetShortenId(originalURL)
				assert.NoError(t, err)
				assert.Equal(t, shortenURL, result)
				done <- true
			}()
		}

		for i := 0; i < 50; i++ {
			<-done
		}
	})
}
