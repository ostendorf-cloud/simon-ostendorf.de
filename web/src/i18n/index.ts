import { createI18n } from 'vue-i18n'

import de from './locales/de.json'
import en from './locales/en.json'

/** The only two locales this site serves. German is the source of truth. */
export const SUPPORTED_LOCALES = ['de', 'en'] as const
export type Locale = (typeof SUPPORTED_LOCALES)[number]
export const DEFAULT_LOCALE: Locale = 'de'

export function isLocale(value: string): value is Locale {
  return (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

/**
 * detectPreferredLocale picks English when it is the visitor's highest
 * weighted browser language, German otherwise. This mirrors the server's
 * own Accept-Language based redirect at "/", for the rare client-side
 * navigation that reaches it instead, for example during local
 * development or a direct in-app link to "/".
 */
export function detectPreferredLocale(languages: readonly string[]): Locale {
  const primary = languages[0]?.split('-')[0]?.toLowerCase()
  return primary === 'en' ? 'en' : DEFAULT_LOCALE
}

/**
 * createAppI18n builds one vue-i18n instance for locale. English falls
 * back to German for any key that is not yet translated, rather than
 * showing the raw key: fallbackLocale is what gives that for free, since
 * German is authored as the complete set of messages.
 */
export function createAppI18n(locale: Locale = DEFAULT_LOCALE) {
  return createI18n({
    legacy: false,
    locale,
    fallbackLocale: DEFAULT_LOCALE,
    missingWarn: true,
    fallbackWarn: false,
    messages: { de, en },
  })
}
