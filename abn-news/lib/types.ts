export interface Article {
  id: number;
  source_id: number;
  category_id: number | null;
  title: string;
  slug: string;
  url: string;
  content: string;
  excerpt: string;
  image_url: string;
  author: string;
  published_at: string;
  fetched_at: string;
}

export interface PaginatedArticles {
  articles: Article[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface Category {
  id: number;
  slug: string;
  name: string;
  description: string;
}

export interface SourceHealth {
  name: string;
  status: string;
  last_fetched_at: string | null;
  backoff_until: string | null;
  last_error: string | null;
  articles_today: number;
}
