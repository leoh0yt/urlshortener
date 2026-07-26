CREATE SEQUENCE IF NOT EXISTS urls_id_seq START WITH 1 INCREMENT BY 1;

CREATE TABLE IF NOT EXISTS urls (
    shortened_url VARCHAR(10) PRIMARY KEY,
    original_url TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);