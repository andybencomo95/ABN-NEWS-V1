import { Newspaper } from "lucide-react";
import ArticleCard from "./ArticleCard";
import type { Article } from "@/lib/types";
import { type Locale, t } from "@/lib/i18n";

interface Props {
  articles: Article[];
  locale: Locale;
}

export default function ArticleGrid({ articles, locale }: Props) {
  if (articles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-gray-400">
        <Newspaper className="w-16 h-16 mb-4" />
        <p className="text-lg font-medium">{t(locale, "empty.title")}</p>
        <p className="text-sm mt-1">{t(locale, "empty.subtitle")}</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {articles.map((article) => (
        <ArticleCard key={article.id} article={article} locale={locale} />
      ))}
    </div>
  );
}
