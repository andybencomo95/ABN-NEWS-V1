import { getArticle } from "@/lib/api";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, User, Clock } from "lucide-react";
import type { Metadata } from "next";
import { t } from "@/lib/i18n";
import { getLocale, formatDateFull } from "@/lib/locale";

export const revalidate = 3600;

interface Props {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  try {
    const { slug } = await params;
    const article = await getArticle(slug);
    return {
      title: article.title,
      description: article.excerpt,
      openGraph: {
        title: article.title,
        description: article.excerpt,
        images: article.image_url ? [{ url: article.image_url }] : [],
      },
    };
  } catch {
    return {};
  }
}

export default async function ArticlePage({ params }: Props) {
  const { slug } = await params;
  const locale = await getLocale();
  let article: import("@/lib/types").Article | null = null;

  try {
    article = await getArticle(slug, locale);
  } catch {
    notFound();
  }

  if (!article) notFound();

  const date = formatDateFull(article.published_at, locale);

  return (
    <article className="max-w-3xl mx-auto">
      <Link
        href="/"
        className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-blue-600 mb-8 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        {t(locale, "article.back_home")}
      </Link>

      {article.image_url && (
        <div className="rounded-2xl overflow-hidden mb-8 shadow-lg">
          <img
            src={article.image_url}
            alt={article.title}
            className="w-full h-64 md:h-96 object-cover"
          />
        </div>
      )}

      <h1 className="text-3xl md:text-4xl font-bold text-gray-900 mb-6 leading-tight">
        {article.title}
      </h1>

      <div className="flex flex-wrap items-center gap-4 text-sm text-gray-500 mb-8 pb-6 border-b border-gray-200">
        {article.author && (
          <span className="flex items-center gap-1.5">
            <User className="w-4 h-4" />
            {t(locale, "article.by_author", { author: article.author })}
          </span>
        )}
        <span className="flex items-center gap-1.5">
          <Clock className="w-4 h-4" />
          <time dateTime={article.published_at}>{date}</time>
        </span>
      </div>

      <div
        className="prose prose-lg max-w-none text-gray-800 leading-relaxed"
        dangerouslySetInnerHTML={
          article.content
            ? { __html: article.content }
            : { __html: `<p>${article.excerpt}</p>` }
        }
      />
    </article>
  );
}
