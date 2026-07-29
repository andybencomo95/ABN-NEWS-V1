import { cookies } from "next/headers";
import type { Locale } from "./i18n";

const LOCALE_COOKIE = "abn_locale";
const DEFAULT_LOCALE: Locale = "es";

export function getLocaleFromCookies(cookieStore: { get(name: string): { value: string } | undefined }): Locale {
  const val = cookieStore.get(LOCALE_COOKIE)?.value;
  if (val === "es" || val === "en") return val;
  return DEFAULT_LOCALE;
}

export async function getLocale(): Promise<Locale> {
  try {
    const store = await cookies();
    return getLocaleFromCookies(store);
  } catch {
    return DEFAULT_LOCALE;
  }
}

export function formatDate(dateStr: string, locale: Locale): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(locale === "es" ? "es-ES" : "en-US", {
    day: "numeric",
    month: locale === "es" ? "long" : "short",
    year: "numeric",
  });
}

export function formatDateShort(dateStr: string, locale: Locale): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(locale === "es" ? "es-ES" : "en-US", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

export function formatDateFull(dateStr: string, locale: Locale): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(locale === "es" ? "es-ES" : "en-US", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });
}
