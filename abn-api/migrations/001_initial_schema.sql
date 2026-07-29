CREATE TABLE IF NOT EXISTS categories (
    id          SERIAL PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS sources (
    id            SERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    site_url      TEXT NOT NULL,
    feed_url      TEXT NOT NULL UNIQUE,
    category_id   INT REFERENCES categories(id),
    interval_sec  INT DEFAULT 3600,
    health_status TEXT DEFAULT 'unknown',
    last_fetched_at  TIMESTAMPTZ,
    last_error    TEXT,
    backoff_until TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS articles (
    id            SERIAL PRIMARY KEY,
    source_id     INT REFERENCES sources(id) NOT NULL,
    category_id   INT REFERENCES categories(id),
    title         TEXT NOT NULL,
    slug          TEXT NOT NULL UNIQUE,
    url           TEXT NOT NULL,
    content       TEXT,
    excerpt       TEXT,
    image_url     TEXT,
    author        TEXT,
    published_at  TIMESTAMPTZ,
    fetched_at    TIMESTAMPTZ DEFAULT NOW(),
    hash          TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_published ON articles(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_hash ON articles(hash);

CREATE TABLE IF NOT EXISTS article_tags (
    article_id INT REFERENCES articles(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (article_id, tag)
);
