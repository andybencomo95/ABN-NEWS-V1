import type { Article, PaginatedArticles, Category, SourceHealth } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function fetchAPI<T>(path: string, lang?: string): Promise<T> {
  const separator = path.includes("?") ? "&" : "?";
  const url = lang ? `${API_BASE}${path}${separator}lang=${lang}` : `${API_BASE}${path}`;
  const res = await fetch(url, { cache: "no-store" });

  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

export async function getArticles(
  category?: string,
  page = 1,
  limit = 10,
  lang?: string
): Promise<PaginatedArticles> {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (category) params.set("category", category);
  return fetchAPI<PaginatedArticles>(`/api/articles?${params}`, lang);
}

export async function getArticle(slug: string, lang?: string): Promise<Article> {
  return fetchAPI<Article>(`/api/articles/${slug}`, lang);
}

export async function getCategories(): Promise<Category[]> {
  return fetchAPI<Category[]>("/api/categories");
}
