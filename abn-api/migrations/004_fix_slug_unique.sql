-- Remove UNIQUE constraint on slug (we dedup by hash)
ALTER TABLE articles DROP CONSTRAINT IF EXISTS articles_slug_key;
CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
