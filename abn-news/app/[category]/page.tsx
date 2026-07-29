import { Hash, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { getArticles, getCategories } from "@/lib/api";
import ArticleGrid from "@/components/ArticleGrid";
import { getLocale } from "@/lib/locale";
import { t } from "@/lib/i18n";

export const revalidate = 3600;

export async function generateStaticParams() {
  try {
    const categories = await getCategories();
    return categories.map((cat: { slug: string }) => ({ category: cat.slug }));
  } catch {
    return [];
  }
}

export default async function CategoryPage({
  params,
}: {
  params: Promise<{ category: string }>;
}) {
  const { category } = await params;
  const locale = await getLocale();

  let articles: import("@/lib/types").Article[] = [];
  let total = 0;
  let catName = t(locale, `cat.${category}` as any);

  try {
    const data = await getArticles(category, 1, 10, locale);
    articles = data.articles;
    total = data.total;
    const categories = await getCategories();
    const found = categories.find(
      (c: { slug: string; name: string }) => c.slug === category
    );
    if (found) catName = t(locale, `cat.${found.slug}` as any);
  } catch {
    // ISR serves cached version
  }

  return (
    <div>
      <Link
        href="/"
        className="inline-flex items-center gap-1 text-sm text-gray-500 hover:text-blue-600 mb-6 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        {t(locale, "article.back_home")}
      </Link>

      <div className="flex items-center gap-3 mb-2">
        <Hash className="w-6 h-6 text-blue-600" />
        <h1 className="text-3xl font-bold text-gray-900 capitalize">
          {catName}
        </h1>
      </div>
      {total > 0 && (
        <p className="text-gray-500 mb-8 ml-9">
          {t(locale, "articles.in_category", { count: total })}
        </p>
      )}

      <ArticleGrid articles={articles} locale={locale} />
    </div>
  );
}
