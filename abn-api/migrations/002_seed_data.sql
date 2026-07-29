-- Seed categories
INSERT INTO categories (slug, name, description) VALUES
    ('general', 'General', 'Noticias generales de actualidad'),
    ('tecnologia', 'Tecnología', 'Noticias de tecnología e innovación'),
    ('deportes', 'Deportes', 'Noticias deportivas'),
    ('economia', 'Economía', 'Noticias de economía y finanzas'),
    ('cultura', 'Cultura', 'Noticias culturales y de entretenimiento'),
    ('ciencia', 'Ciencia', 'Noticias científicas')
ON CONFLICT (slug) DO NOTHING;

-- Seed RSS sources
INSERT INTO sources (name, site_url, feed_url, category_id) VALUES
    ('TechCrunch', 'https://techcrunch.com', 'https://techcrunch.com/feed/', (SELECT id FROM categories WHERE slug='tecnologia'))
ON CONFLICT (feed_url) DO NOTHING;
