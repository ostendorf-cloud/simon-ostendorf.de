import type { Ref } from 'vue'
import { createRouter, type RouteRecordRaw, type RouterHistory, type Router } from 'vue-router'

import { detectPreferredLocale, isLocale } from '../i18n'
import { routeTable } from '../lib/routeTable'

/** views maps each routes.json entry's name to the view component that
 *  renders it: a frontend-only concern routes.json has no use for, the
 *  same reason lib/seo.ts keeps titleKeys and descriptionKeys explicit
 *  rather than folding them into the table. router/index.test.ts checks
 *  that every name routes.json carries has an entry here; a name with
 *  none falls back to the 404 view below rather than crashing the app. */
export const views: Record<string, () => Promise<unknown>> = {
  home: () => import('../views/HomeView.vue'),
  cv: () => import('../views/CvView.vue'),
  contact: () => import('../views/ContactView.vue'),
  imprint: () => import('../views/ImprintView.vue'),
  privacy: () => import('../views/PrivacyView.vue'),
}

const notFoundView = () => import('../views/NotFoundView.vue')

/** NOT_FOUND_ROUTE_NAME is the 404 page's own router name: it answers no
 *  path of its own, so it is deliberately not part of routes.json (see
 *  internal/routes/routes.go) and is the one route this file still
 *  writes by hand instead of deriving from the table. lib/seo.ts has its
 *  own equivalent, NOT_FOUND. */
export const NOT_FOUND_ROUTE_NAME = 'not-found'

/** routePath is the router path pattern for slug, the home page's own
 *  empty slug included. */
function routePath(slug: string): string {
  if (slug === '') {
    return '/:locale(de|en)/'
  }
  // A page's canonical path is /<locale>/<slug>, with no trailing slash:
  // the server answers the slash form with a redirect.
  return `/:locale(de|en)/${slug}`
}

/** pageRoutes builds one vue-router route per routes.json entry: adding a
 *  page there is now enough to route to it here too, as long as a
 *  matching entry also exists in views above (router/index.test.ts is
 *  the check for that). Exported for that test. */
export const pageRoutes: RouteRecordRaw[] = routeTable.map((entry) => ({
  path: routePath(entry.slug),
  name: entry.name,
  component: views[entry.name] ?? notFoundView,
  meta: { legal: entry.legal === true },
}))

const routes: RouteRecordRaw[] = [
  ...pageRoutes,
  {
    path: '/:locale(de|en)/:pathMatch(.*)*',
    name: NOT_FOUND_ROUTE_NAME,
    component: notFoundView,
  },
  {
    // No locale prefix at all, "/" included: redirected to the preferred
    // locale's home page, the same rule the server applies to "/" itself.
    path: '/:pathMatch(.*)*',
    redirect: () => `/${detectPreferredLocale(navigatorLanguages())}/`,
  },
]

/** navigatorLanguages reads the browser's language preference, or an empty
 *  list when run outside a browser, such as the prerender script. */
function navigatorLanguages(): readonly string[] {
  if (typeof navigator === 'undefined') {
    return []
  }
  return navigator.languages ?? [navigator.language]
}

/** createAppRouter builds the router over history, so the client supplies
 *  a browser history and the prerender script supplies a memory one. */
export function createAppRouter(history: RouterHistory): Router {
  return createRouter({ history, routes })
}

/**
 * attachLocaleSync keeps the active locale ref, and the document's lang
 * attribute, in step with the route's own :locale param on every
 * navigation, including the first one. A locale switch is then nothing
 * more than a normal navigation to the same route under the other
 * locale prefix; nothing else has to set the active locale by hand.
 *
 * Takes the locale ref itself, i18n.global.locale, rather than the whole
 * Composer: the Composer's own generic type parameters are inferred from
 * the exact message schema it was built with, which makes "Composer" on
 * its own, with no type arguments, too narrow a parameter type to accept
 * the real instance createAppI18n returns.
 */
export function attachLocaleSync(router: Router, locale: Ref<string>): void {
  router.beforeEach((to) => {
    const next = to.params.locale
    if (typeof next === 'string' && isLocale(next)) {
      locale.value = next
      if (typeof document !== 'undefined') {
        document.documentElement.lang = next
      }
    }
  })
}
