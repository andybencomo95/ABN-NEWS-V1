export type Locale = "es" | "en";

export const translations: Record<Locale, Record<string, string>> = {
  es: {
    "site.name": "ABN News",
    "site.description": "Tu fuente de noticias actualizadas de múltiples categorías.",
    "nav.home": "Home",
    "nav.categories": "Categorías",
    "hero.latest": "Última noticia",
    "home.latest": "Últimas Noticias",
    "articles.count": "{count} artículos",
    "articles.in_category": "{count} artículos en esta categoría",
    "article.read_more": "Leer más",
    "article.back_home": "Volver al inicio",
    "article.by_author": "Por {author}",
    "empty.title": "No hay artículos disponibles",
    "empty.subtitle": "Pronto tendremos nuevas noticias",
    "footer.about": "Acerca de",
    "footer.about_text": "ABN News agrega noticias de múltiples fuentes RSS para mantenerte informado. Las noticias se actualizan automáticamente cada hora.",
    "footer.copyright": "© {year} ABN News. Todos los derechos reservados.",
    "lang.es": "Español",
    "lang.en": "English",
    "source.live": "En vivo",
    "cat.general": "General",
    "cat.tecnologia": "Tecnología",
    "cat.deportes": "Deportes",
    "cat.economia": "Economía",
    "cat.cultura": "Cultura",
    "cat.ciencia": "Ciencia",
  },
  en: {
    "site.name": "ABN News",
    "site.description": "Your source for updated news across multiple categories.",
    "nav.home": "Home",
    "nav.categories": "Categories",
    "hero.latest": "Latest News",
    "home.latest": "Latest News",
    "articles.count": "{count} articles",
    "articles.in_category": "{count} articles in this category",
    "article.read_more": "Read more",
    "article.back_home": "Back to home",
    "article.by_author": "By {author}",
    "empty.title": "No articles available",
    "empty.subtitle": "We'll have new stories soon",
    "footer.about": "About",
    "footer.about_text": "ABN News aggregates news from multiple RSS sources to keep you informed. News updates automatically every hour.",
    "footer.copyright": "© {year} ABN News. All rights reserved.",
    "lang.es": "Español",
    "lang.en": "English",
    "source.live": "Live",
    "cat.general": "General",
    "cat.tecnologia": "Technology",
    "cat.deportes": "Sports",
    "cat.economia": "Economy",
    "cat.cultura": "Culture",
    "cat.ciencia": "Science",
  },
};

export function t(locale: Locale, key: string, vars?: Record<string, string | number>): string {
  let text = translations[locale]?.[key] ?? translations["es"]?.[key] ?? key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      text = text.replace(`{${k}}`, String(v));
    }
  }
  return text;
}
