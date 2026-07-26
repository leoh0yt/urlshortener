package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/eqweqr/urlshortener/encoder"
	"github.com/eqweqr/urlshortener/storage"
)

type UrlService struct {
	storage storage.Storage
	encoder *encoder.Encoder
	logger  slog.Logger
}

func NewUrlService(s storage.Storage) *UrlService {
	return &UrlService{
		storage: s,
		encoder: encoder.NewBase62Encoder(),
	}
}

func (s *UrlService) Shorten(originUrl string) (string, error) {
	id, err := s.storage.GetShortenId(originUrl)
	if err == nil {
		return id, nil
	}

	if !errors.Is(err, storage.ErrNotFound) {
		return "", fmt.Errorf("failed to check existing URL: %w", err)
	}

	nextId, err := s.storage.GetNextId()
	if err != nil {
		return "", err
	}

	encodedId := s.encoder.Encode10WithPadding(nextId, 10)

	err = s.storage.SaveId(originUrl, encodedId)
	if err != nil {
		return "", err
	}
	return encodedId, nil
}

func (s *UrlService) Resolve(shortenUrl string) (string, error) {
	originalUrl, err := s.storage.GetOriginalId(shortenUrl)
	if err != nil {
		return "", err
	}
	return originalUrl, err
}
