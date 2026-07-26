package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUrlService struct {
	mock.Mock
}

func (m *MockUrlService) Shorten(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockUrlService) Resolve(shortCode string) (string, error) {
	args := m.Called(shortCode)
	return args.String(0), args.Error(1)
}

func setupTest() (*UrlHandler, *MockUrlService) {
	mockService := new(MockUrlService)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	handler := NewHandler(mockService, logger)
	return handler, mockService
}

func TestShorten_Success(t *testing.T) {
	handler, mockService := setupTest()

	originalURL := "https://example.com"
	shortCode := "abc123"
	expectedShortURL := "http://example.com/abc123"

	mockService.On("Shorten", originalURL).Return(shortCode, nil)

	reqBody := map[string]string{"url": originalURL}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	req.Host = "example.com"
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedShortURL, response["short_url"])
	assert.Equal(t, originalURL, response["original_url"])

	mockService.AssertExpectations(t)
}

func TestShorten_InvalidMethod(t *testing.T) {
	handler, _ := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestShorten_InvalidJSON(t *testing.T) {
	handler, _ := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString("{invalid json}"))
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestShorten_EmptyURL(t *testing.T) {
	handler, _ := setupTest()

	reqBody := map[string]string{"url": ""}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "URL is required")
}

func TestShorten_ServiceError(t *testing.T) {
	handler, mockService := setupTest()

	originalURL := "https://example.com"
	expectedErr := errors.New("database error")

	mockService.On("Shorten", originalURL).Return("", expectedErr)

	reqBody := map[string]string{"url": originalURL}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to shorten URL")

	mockService.AssertExpectations(t)
}

func TestResolve_Success(t *testing.T) {
	handler, mockService := setupTest()

	shortCode := "abc123"
	originalURL := "https://example.com"

	mockService.On("Resolve", shortCode).Return(originalURL, nil)

	req := httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, originalURL, w.Header().Get("Location"))

	mockService.AssertExpectations(t)
}

func TestResolve_InvalidMethod(t *testing.T) {
	handler, _ := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/abc123", nil)
	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestResolve_EmptyShortCode(t *testing.T) {
	handler, _ := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Short code is required")
}

func TestResolve_NotFound(t *testing.T) {
	handler, mockService := setupTest()

	shortCode := "notfound"
	expectedErr := errors.New("not found")

	mockService.On("Resolve", shortCode).Return("", expectedErr)

	req := httptest.NewRequest(http.MethodGet, "/"+shortCode, nil)
	w := httptest.NewRecorder()

	handler.Resolve(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Short URL not found")

	mockService.AssertExpectations(t)
}
