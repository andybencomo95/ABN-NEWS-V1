import { Sparkles } from "lucide-react";
import { getArticles } from "@/lib/api";
import HeroSection from "@/components/HeroSection";
import ArticleGrid from "@/components/ArticleGrid";
import { getLocale } from "@/lib/locale";
import { t } from "@/lib/i18n";

export const revalidate = 3600;

export default async function HomePage() {
  const locale = await getLocale();
  let articles: import("@/lib/types").Article[] = [];

  try {
    const data = await getArticles(undefined, 1, 10, locale);
    articles = data.articles;
  } catch {
    // ISR serves cached version on error
  }

  const hero = articles[0];
  const rest = articles.slice(1);

  return (
    <div>
      {hero && <HeroSection article={hero} locale={locale} />}

      <div className="flex items-center gap-2 mb-6">
        <Sparkles className="w-5 h-5 text-blue-600" />
        <h2 className="text-2xl font-bold text-gray-900">
          {t(locale, "home.latest")}
        </h2>
        <span className="text-sm text-gray-400 ml-auto">
          {t(locale, "articles.count", { count: articles.length })}
        </span>
      </div>

      <ArticleGrid articles={rest} locale={locale} />
    </div>
  );
}
