CREATE TABLE IF NOT EXISTS urls (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
