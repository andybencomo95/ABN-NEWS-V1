import Link from "next/link";
import {
  Newspaper,
  Monitor,
  Trophy,
  TrendingUp,
  Palette,
  FlaskConical,
  type LucideIcon,
} from "lucide-react";
import { getCategories } from "@/lib/api";
import { t, type Locale } from "@/lib/i18n";

interface Category {
  slug: string;
  name: string;
}

const iconMap: Record<string, LucideIcon> = {
  general: Newspaper,
  tecnologia: Monitor,
  deportes: Trophy,
  economia: TrendingUp,
  cultura: Palette,
  ciencia: FlaskConical,
};

export default async function CategoryNav({ locale }: { locale: Locale }) {
  let categories: Category[] = [];
  try {
    categories = await getCategories();
  } catch {
    return (
      <nav className="flex items-center gap-1">
        <Link
          href="/"
          className="flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-blue-700 bg-blue-50 rounded-full hover:bg-blue-100 transition-colors"
        >
          <Newspaper className="w-4 h-4" />
          {t(locale, "nav.home")}
        </Link>
      </nav>
    );
  }

  return (
    <nav className="flex items-center gap-1 overflow-x-auto pb-1 scrollbar-none">
      <Link
        href="/"
        className="flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-gray-600 hover:text-blue-700 hover:bg-blue-50 rounded-full whitespace-nowrap transition-colors"
      >
        <Newspaper className="w-4 h-4" />
        {t(locale, "nav.home")}
      </Link>
      {categories.map((cat) => {
        const Icon = iconMap[cat.slug] || Newspaper;
        const translatedName = t(locale, `cat.${cat.slug}` as any);
        return (
          <Link
            key={cat.slug}
            href={`/${cat.slug}`}
            className="flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-gray-600 hover:text-blue-700 hover:bg-blue-50 rounded-full whitespace-nowrap transition-colors"
          >
            <Icon className="w-4 h-4" />
            {translatedName}
          </Link>
        );
      })}
    </nav>
  );
}
