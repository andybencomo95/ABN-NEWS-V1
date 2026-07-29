import Link from "next/link";
import { Newspaper } from "lucide-react";
import CategoryNav from "./CategoryNav";
import LangSwitcher from "./LangSwitcher";
import { getLocale } from "@/lib/locale";
import { t } from "@/lib/i18n";

export default async function Header() {
  const locale = await getLocale();

  return (
    <header className="sticky top-0 z-50 bg-white/95 backdrop-blur-sm border-b border-gray-200 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="flex items-center justify-between h-16">
          <Link
            href="/"
            className="flex items-center gap-2 text-xl sm:text-2xl font-bold text-blue-700 hover:text-blue-800"
          >
            <Newspaper className="w-7 h-7" />
            <span>{t(locale, "site.name")}</span>
          </Link>
          <div className="flex items-center">
            <CategoryNav locale={locale} />
            <LangSwitcher current={locale} />
          </div>
        </div>
      </div>
    </header>
  );
}
