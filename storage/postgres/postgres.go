package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/eqweqr/urlshortener/config"
	"github.com/eqweqr/urlshortener/storage"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(cfg config.DbConfig) (*PostgresStorage, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) GetNextId() (uint64, error) {
	var id uint64
	err := p.db.QueryRow("SELECT nextval('urls_id_seq')").Scan(&id)
	return id, err
}

func (p *PostgresStorage) GetShortenId(originalUrl string) (string, error) {
	var shortenUrl string

	err := p.db.QueryRow(
		"SELECT shortened_url FROM urls WHERE original_url = $1",
		originalUrl,
	).Scan(&shortenUrl)

	if err != nil {
		return "", storage.ErrNotFound
	}

	return shortenUrl, nil
}

func (p *PostgresStorage) SaveId(originalUrl string, shortenUrl string) error {
	_, err := p.db.Exec("INSERT INTO urls (shortened_url, original_url) VALUES ($1, $2)",
		shortenUrl, originalUrl,
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetOriginalId(shortenedUrl string) (string, error) {
	var id string
	err := p.db.QueryRow(
		"SELECT original_url FROM urls WHERE shortened_url = $1 ",
		shortenedUrl,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (p *PostgresStorage) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
