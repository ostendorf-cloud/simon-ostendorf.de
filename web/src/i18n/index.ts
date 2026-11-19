import type { App } from 'vue'
import { createI18n } from 'vue-i18n'
import { createGovueI18n } from '@ostendorf-cloud/govue'

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
 * installI18n installs the app's own vue-i18n instance and govue's language
 * layer on top of it. The layer merges govue's catalogues (the legal page
 * titles, the footer link labels, the back link) into the same instance and
 * follows its locale, which the router keeps in step with the URL. Nothing is
 * detected and nothing is stored: storage is off, so the layer writes no
 * cookie and no localStorage entry, and the locale stays the URL's.
 */
export function installI18n(app: App, i18n: ReturnType<typeof createAppI18n>): void {
  app.use(i18n)
  app.use(
    createGovueI18n({
      i18n,
      locale: i18n.global.locale.value,
      detect: false,
      storage: { kind: 'none' },
    }),
  )
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
