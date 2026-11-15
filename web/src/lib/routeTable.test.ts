/**
 * validateRouteTable is this file's own copy of
 * internal/routes/routes.go's Parse, needed because Vite's JSON import
 * in routeTable.ts resolves routes.json directly into a JS value at this
 * package's own build time, never through the Go binary: there is no Go
 * validation to inherit here. These tests run the exact same fixtures
 * under internal/routes/testdata that routes_test.go's
 * TestParseAgainstSharedFixtures does: a fixture added to either
 * directory is exercised by both suites without touching either test
 * file.
 */
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

import { routeTable, type RouteEntry, validateRouteTable } from './routeTable'

const fixturesDir = join(process.cwd(), '../internal/routes/testdata')

function readFixtures(sub: string): Array<{ name: string; table: RouteEntry[] }> {
  const dir = join(fixturesDir, sub)
  return readdirSync(dir).map((name) => ({
    name,
    table: JSON.parse(readFileSync(join(dir, name), 'utf-8')) as RouteEntry[],
  }))
}

describe('validateRouteTable against the fixtures shared with internal/routes', () => {
  const invalid = readFixtures('invalid')
  const valid = readFixtures('valid')

  it('found fixtures of both kinds', () => {
    expect(invalid.length).toBeGreaterThan(0)
    expect(valid.length).toBeGreaterThan(0)
  })

  it.each(invalid)('rejects $name', ({ table }) => {
    expect(() => validateRouteTable(table)).toThrow()
  })

  it.each(valid)('accepts $name', ({ table }) => {
    expect(() => validateRouteTable(table)).not.toThrow()
  })
})

describe('validateRouteTable', () => {
  it('names the duplicate in its error message', () => {
    const table: RouteEntry[] = [
      { name: 'dup', slug: 'a', file: 'a.html' },
      { name: 'dup', slug: 'b', file: 'b.html' },
    ]
    expect(() => validateRouteTable(table)).toThrow(/dup/)
  })

  it('accepts the checked-in routes.json', () => {
    expect(() => validateRouteTable(routeTable)).not.toThrow()
  })
})
