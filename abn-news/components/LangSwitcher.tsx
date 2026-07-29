"use client";

import { usePathname } from "next/navigation";

export default function LangSwitcher({ current }: { current: string }) {
  const pathname = usePathname();

  function switchLocale(locale: string) {
    if (locale === current) return;
    document.cookie = `abn_locale=${locale}; path=/; max-age=31536000; SameSite=Lax`;
    // Force full page reload to re-render everything server-side with new locale
    window.location.href = pathname;
  }

  return (
    <div className="flex items-center gap-1 ml-2 border-l border-gray-200 pl-3">
      <button
        onClick={() => switchLocale("en")}
        className={`text-lg leading-none px-1.5 py-1 rounded transition-transform ${
          current === "en"
            ? "opacity-100 scale-110"
            : "opacity-50 hover:opacity-80"
        }`}
        title="English"
        aria-label="Switch to English"
      >
        🇺🇸
      </button>
      <button
        onClick={() => switchLocale("es")}
        className={`text-lg leading-none px-1.5 py-1 rounded transition-transform ${
          current === "es"
            ? "opacity-100 scale-110"
            : "opacity-50 hover:opacity-80"
        }`}
        title="Español"
        aria-label="Cambiar a Español"
      >
        🇨🇺
      </button>
    </div>
  );
}
