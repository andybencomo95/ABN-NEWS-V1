import Link from "next/link";
import { Clock, ArrowRight } from "lucide-react";
import type { Article } from "@/lib/types";
import { type Locale, t } from "@/lib/i18n";
import { formatDateFull } from "@/lib/locale";

interface Props {
  article: Article;
  locale: Locale;
}

export default function HeroSection({ article, locale }: Props) {
  const date = formatDateFull(article.published_at, locale);

  return (
    <div className="relative rounded-2xl overflow-hidden mb-10 bg-gradient-to-br from-blue-900 via-blue-800 to-indigo-900 shadow-xl">
      {article.image_url && (
        <div className="absolute inset-0">
          <img
            src={article.image_url}
            alt={article.title}
            className="w-full h-full object-cover opacity-30"
          />
        </div>
      )}
      <div className="relative px-6 py-12 md:px-10 md:py-16 lg:px-16 lg:py-20">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-white/20 backdrop-blur-sm rounded-full text-xs font-medium text-white mb-4">
          <span className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
          {t(locale, "source.live")}
        </div>
        <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold text-white mb-4 leading-tight max-w-3xl">
          <Link
            href={`/articles/${article.slug}`}
            className="hover:text-blue-200 transition-colors"
          >
            {article.title}
          </Link>
        </h1>
        <p className="text-blue-100 text-base md:text-lg mb-6 max-w-2xl line-clamp-2">
          {article.excerpt}
        </p>
        <div className="flex items-center gap-4 text-sm text-blue-200">
          <span className="flex items-center gap-1.5">
            <Clock className="w-4 h-4" />
            {date}
          </span>
          <Link
            href={`/articles/${article.slug}`}
            className="inline-flex items-center gap-2 px-4 py-2 bg-white text-blue-900 rounded-lg font-medium hover:bg-blue-50 transition-colors"
          >
            {t(locale, "article.read_more")}
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>
    </div>
  );
}
