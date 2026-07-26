package memory

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/eqweqr/urlshortener/storage"
)

type MemoryStorage struct {
	mu        sync.RWMutex
	nextId    uint64
	shorten   map[string]string
	originals map[string]string
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		nextId:    1,
		shorten:   make(map[string]string),
		originals: make(map[string]string),
	}
}

func (m *MemoryStorage) SaveId(originalUrl string, shortenUrl string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.shorten[originalUrl]; ok {
		return nil
	}

	m.originals[shortenUrl] = originalUrl
	m.shorten[originalUrl] = shortenUrl

	return nil
}

func (m *MemoryStorage) GetOriginalId(shortenURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	originURL, ok := m.originals[shortenURL]
	if !ok {
		return "", fmt.Errorf("No associated value with shortenURL: %s", shortenURL)
	}
	return originURL, nil
}

func (m *MemoryStorage) GetShortenId(originalURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shortenURL, ok := m.shorten[originalURL]
	if !ok {
		return "", storage.ErrNotFound
	}
	return shortenURL, nil
}

func (m *MemoryStorage) GetNextId() (uint64, error) {
	return atomic.AddUint64(&m.nextId, 1), nil
}

func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shorten = nil
	m.originals = nil

	return nil
}
