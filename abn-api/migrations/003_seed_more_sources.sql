-- Add more RSS sources per category
INSERT INTO sources (name, site_url, feed_url, category_id, interval_sec) VALUES
    ('BBC News', 'https://bbc.com/news', 'https://feeds.bbci.co.uk/news/rss.xml', (SELECT id FROM categories WHERE slug='general'), 3600),
    ('BBC Sport', 'https://bbc.com/sport', 'https://feeds.bbci.co.uk/sport/rss.xml', (SELECT id FROM categories WHERE slug='deportes'), 3600),
    ('BBC Business', 'https://bbc.com/news/business', 'https://feeds.bbci.co.uk/news/business/rss.xml', (SELECT id FROM categories WHERE slug='economia'), 3600),
    ('BBC Entertainment', 'https://bbc.com/news/entertainment_and_arts', 'https://feeds.bbci.co.uk/news/entertainment_and_arts/rss.xml', (SELECT id FROM categories WHERE slug='cultura'), 3600),
    ('BBC Science', 'https://bbc.com/news/science_and_environment', 'https://feeds.bbci.co.uk/news/science_and_environment/rss.xml', (SELECT id FROM categories WHERE slug='ciencia'), 3600),
    ('NYT World', 'https://nytimes.com', 'https://rss.nytimes.com/services/xml/rss/nyt/World.xml', (SELECT id FROM categories WHERE slug='general'), 3600)
ON CONFLICT (feed_url) DO NOTHING;
