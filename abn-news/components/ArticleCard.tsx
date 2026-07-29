import Link from "next/link";
import { Clock, User } from "lucide-react";
import type { Article } from "@/lib/types";
import { type Locale, t } from "@/lib/i18n";
import { formatDateShort } from "@/lib/locale";

interface Props {
  article: Article;
  locale: Locale;
}

export default function ArticleCard({ article, locale }: Props) {
  const date = formatDateShort(article.published_at, locale);

  return (
    <article className="group bg-white rounded-xl border border-gray-200 overflow-hidden hover:shadow-lg hover:border-blue-200 transition-all duration-300">
      {article.image_url ? (
        <div className="relative h-48 overflow-hidden">
          <img
            src={article.image_url}
            alt={article.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
          />
        </div>
      ) : (
        <div className="h-48 bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center">
          <span className="text-4xl font-bold text-blue-200">ABN</span>
        </div>
      )}
      <div className="p-5">
        <div className="flex items-center gap-3 text-xs text-gray-500 mb-3">
          {article.author && (
            <span className="flex items-center gap-1">
              <User className="w-3 h-3" />
              {t(locale, "article.by_author", { author: article.author })}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {date}
          </span>
        </div>
        <Link href={`/articles/${article.slug}`}>
          <h2 className="text-lg font-semibold text-gray-900 mb-2 group-hover:text-blue-700 transition-colors line-clamp-2">
            {article.title}
          </h2>
        </Link>
        <p className="text-sm text-gray-600 line-clamp-3 leading-relaxed">
          {article.excerpt}
        </p>
        <Link
          href={`/articles/${article.slug}`}
          className="inline-flex items-center gap-1 mt-4 text-sm font-medium text-blue-600 hover:text-blue-800 transition-colors"
        >
          {t(locale, "article.read_more")}
          <span className="text-lg leading-none">→</span>
        </Link>
      </div>
    </article>
  );
}
