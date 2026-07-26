package service

import (
	"errors"
	"testing"

	"github.com/eqweqr/urlshortener/encoder"
	"github.com/eqweqr/urlshortener/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetShortenId(originUrl string) (string, error) {
	args := m.Called(originUrl)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetNextId() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockStorage) SaveId(originUrl string, shortenId string) error {
	args := m.Called(originUrl, shortenId)
	return args.Error(0)
}

func (m *MockStorage) GetOriginalId(shortenUrl string) (string, error) {
	args := m.Called(shortenUrl)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Close() error {
	return nil
}

func TestUrlService_Shorten_ExistingURL(t *testing.T) {
	tests := []struct {
		name      string
		originUrl string
		shortenId string
		mockError error
		wantError bool
	}{
		{
			name:      "existing URL returns stored ID",
			originUrl: "https://example.com",
			shortenId: "abc1234567",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "existing URL with error",
			originUrl: "https://example.com",
			shortenId: "",
			mockError: errors.New("database connection error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			service := NewUrlService(mockStorage)

			mockStorage.On("GetShortenId", tt.originUrl).Return(tt.shortenId, tt.mockError)

			result, err := service.Shorten(tt.originUrl)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.shortenId, result)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUrlService_Shorten_NewURL(t *testing.T) {
	tests := []struct {
		name              string
		originUrl         string
		nextId            uint64
		expectedId        string
		getShortenIdError error
		getNextIdError    error
		saveError         error
		wantError         bool
	}{
		{
			name:              "new URL successfully shortened",
			originUrl:         "https://new-example.com",
			nextId:            1,
			getShortenIdError: storage.ErrNotFound,
			getNextIdError:    nil,
			saveError:         nil,
			wantError:         false,
		},
		{
			name:              "new URL with save error",
			originUrl:         "https://new-example.com",
			nextId:            1,
			getShortenIdError: storage.ErrNotFound,
			getNextIdError:    nil,
			saveError:         errors.New("save failed"),
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			service := NewUrlService(mockStorage)

			mockStorage.On("GetShortenId", tt.originUrl).Return("", tt.getShortenIdError)

			if tt.getShortenIdError == storage.ErrNotFound {
				if tt.getNextIdError != nil {
					mockStorage.On("GetNextId").Return(int64(0), tt.getNextIdError)
				} else {
					mockStorage.On("GetNextId").Return(tt.nextId, nil)

					if tt.saveError != nil {
						enc := encoder.NewBase62Encoder()
						encodedId := enc.Encode10WithPadding(tt.nextId, 10)
						mockStorage.On("SaveId", tt.originUrl, encodedId).Return(tt.saveError)
					} else {
						enc := encoder.NewBase62Encoder()
						encodedId := enc.Encode10WithPadding(tt.nextId, 10)
						mockStorage.On("SaveId", tt.originUrl, encodedId).Return(nil)
					}
				}
			}

			result, err := service.Shorten(tt.originUrl)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUrlService_Resolve(t *testing.T) {
	tests := []struct {
		name        string
		shortenUrl  string
		originalUrl string
		mockError   error
		wantError   bool
	}{
		{
			name:        "existing short URL resolved successfully",
			shortenUrl:  "abc1234567",
			originalUrl: "https://example.com",
			mockError:   nil,
			wantError:   false,
		},
		{
			name:        "non-existent short URL",
			shortenUrl:  "nonexistent",
			originalUrl: "",
			mockError:   storage.ErrNotFound,
			wantError:   true,
		},
		{
			name:        "storage error during resolution",
			shortenUrl:  "abc1234567",
			originalUrl: "",
			mockError:   errors.New("database error"),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			service := NewUrlService(mockStorage)

			mockStorage.On("GetOriginalId", tt.shortenUrl).Return(tt.originalUrl, tt.mockError)

			result, err := service.Resolve(tt.shortenUrl)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.originalUrl, result)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
