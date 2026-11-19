import { describe, expect, it } from 'vitest'
import { createMemoryHistory } from 'vue-router'

import { routeTable } from '../lib/routeTable'
import { createAppRouter, NOT_FOUND_ROUTE_NAME, views } from './index'

// This is the router side of the same check
// cmd/website/routes_test.go's TestPrerenderedOutputMatchesRouteTable runs
// for the Go side: routeTable is the single source of truth for which
// pages exist, and nothing today checks the router against it by hand.
// The failure mode this catches is a page added to routes.json but never
// routed, which builds and prerenders fine, then 404s client-side once
// the SPA hydrates over the server-rendered HTML.

describe('router', () => {
  const router = createAppRouter(createMemoryHistory())

  it.each(routeTable)('registers a route named $name for routes.json entry $name', (entry) => {
    expect(router.hasRoute(entry.name)).toBe(true)
  })

  it.each(routeTable)('routes $name to the path implied by its routes.json slug', (entry) => {
    const record = router.getRoutes().find((r) => r.name === entry.name)
    const trailing = entry.legal ? '' : '/'
    const wantPath =
      entry.slug === '' ? '/:locale(de|en)/' : `/:locale(de|en)/${entry.slug}${trailing}`
    expect(record?.path).toBe(wantPath)
  })

  it.each(routeTable)('has a view component registered for routes.json entry $name', (entry) => {
    expect(views[entry.name]).toBeDefined()
  })

  it('registers exactly the page routes routeTable names, plus the deliberate exceptions', () => {
    const tableNames = new Set(routeTable.map((entry) => entry.name))
    const registeredNames = router
      .getRoutes()
      .map((r) => String(r.name ?? ''))
      .filter((name) => name !== '' && name !== NOT_FOUND_ROUTE_NAME)

    for (const name of registeredNames) {
      expect(
        tableNames.has(name),
        `router registers route ${name}, which routes.json does not list`,
      ).toBe(true)
    }
    expect(registeredNames).toHaveLength(tableNames.size)
  })
})
