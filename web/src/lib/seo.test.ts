import { describe, expect, it } from 'vitest'

import { routeTable } from './routeTable'
import { canonical, descriptionKeys, hreflangAlternates, NOT_FOUND, path, titleKeys } from './seo'

// titleKeys and descriptionKeys are kept separate from routes.json
// deliberately (see seo.ts), which only stays safe as long as this test
// fails the moment the two lists disagree: a route with no key, or a key
// for a name that is not a real route.

describe('titleKeys', () => {
  const wantNames = new Set([...routeTable.map((entry) => entry.name), NOT_FOUND])

  it.each(routeTable)('has a title key for routes.json entry $name', (entry) => {
    expect(titleKeys[entry.name]).toBeDefined()
  })

  it('has a title key for the 404 page', () => {
    expect(titleKeys[NOT_FOUND]).toBeDefined()
  })

  it('has no title key for a name that is not a real route', () => {
    for (const name of Object.keys(titleKeys)) {
      expect(wantNames.has(name), `titleKeys names "${name}", which is not a real route`).toBe(true)
    }
  })

  it('carries exactly one key per real route plus the 404 page', () => {
    expect(Object.keys(titleKeys)).toHaveLength(wantNames.size)
  })
})

describe('descriptionKeys', () => {
  const wantNames = new Set([...routeTable.map((entry) => entry.name), NOT_FOUND])

  it.each(routeTable)('has a description key for routes.json entry $name', (entry) => {
    expect(descriptionKeys[entry.name]).toBeDefined()
  })

  it('has a description key for the 404 page', () => {
    expect(descriptionKeys[NOT_FOUND]).toBeDefined()
  })

  it('has no description key for a name that is not a real route', () => {
    for (const name of Object.keys(descriptionKeys)) {
      expect(
        wantNames.has(name),
        `descriptionKeys names "${name}", which is not a real route`,
      ).toBe(true)
    }
  })

  it('carries exactly one key per real route plus the 404 page', () => {
    expect(Object.keys(descriptionKeys)).toHaveLength(wantNames.size)
  })
})

describe('path and canonical', () => {
  // The path must be the one the server answers with 200 and the sitemap
  // lists: a trailing slash, none for a legal page. Anything else is a 301.
  it.each(routeTable)('resolves routes.json entry $name to its served path', (entry) => {
    const want =
      entry.slug === '' ? '/de/' : entry.legal ? `/de/${entry.slug}` : `/de/${entry.slug}/`
    expect(path('de', entry.name)).toBe(want)
  })

  it('gives cv a trailing slash and the imprint none', () => {
    expect(canonical('en', 'cv')).toBe('https://simon-ostendorf.de/en/cv/')
    expect(canonical('en', 'imprint')).toBe('https://simon-ostendorf.de/en/imprint')
  })

  it.each(routeTable)(
    'builds a canonical URL for routes.json entry $name without throwing',
    (entry) => {
      expect(() => canonical('de', entry.name)).not.toThrow()
    },
  )

  it('resolves the 404 page to an empty slug', () => {
    expect(path('de', NOT_FOUND)).toBe('/de/')
  })

  it('builds a canonical URL for the 404 page without throwing', () => {
    expect(() => canonical('en', NOT_FOUND)).not.toThrow()
  })
})

describe('hreflangAlternates', () => {
  it('points x-default at the English page, the site default', () => {
    const alternates = hreflangAlternates('imprint')
    expect(alternates.find((a) => a.hreflang === 'x-default')?.href).toBe(
      'https://simon-ostendorf.de/en/imprint',
    )
  })
})
