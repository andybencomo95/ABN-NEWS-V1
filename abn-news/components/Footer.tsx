import { Newspaper } from "lucide-react";
import Link from "next/link";
import { type Locale, t } from "@/lib/i18n";

interface Props {
  locale: Locale;
}

const footerCategories = [
  { slug: "general", icon: Newspaper },
  { slug: "tecnologia", icon: null },
  { slug: "deportes", icon: null },
  { slug: "economia", icon: null },
  { slug: "cultura", icon: null },
  { slug: "ciencia", icon: null },
];

export default function Footer({ locale }: Props) {
  const year = new Date().getFullYear();

  return (
    <footer className="border-t border-gray-200 bg-white mt-16">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 py-10">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <Link href="/" className="flex items-center gap-2 text-lg font-bold text-blue-700">
              <Newspaper className="w-6 h-6" />
              {t(locale, "site.name")}
            </Link>
            <p className="mt-2 text-sm text-gray-500">
              {t(locale, "site.description")}
            </p>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900 mb-3">
              {t(locale, "nav.categories")}
            </h3>
            <div className="flex flex-col gap-2 text-sm text-gray-500">
              {["general", "tecnologia", "deportes", "economia"].map((slug) => (
                <Link
                  key={slug}
                  href={`/${slug}`}
                  className="hover:text-blue-600"
                >
                  {t(locale, `cat.${slug}` as any)}
                </Link>
              ))}
            </div>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900 mb-3">
              {t(locale, "footer.about")}
            </h3>
            <p className="text-sm text-gray-500">
              {t(locale, "footer.about_text")}
            </p>
          </div>
        </div>
        <div className="mt-8 pt-6 border-t border-gray-100 text-center text-sm text-gray-400">
          {t(locale, "footer.copyright", { year })}
        </div>
      </div>
    </footer>
  );
}
