package webapp

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// robotsHandler serves /robots.txt: every route is open to crawling, there
// is no route this site wants hidden, plus the sitemap's own location.
func (a *App) robotsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fmt.Fprintf(w, "User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n", a.siteBase())
	})
}

// sitemapHandler serves /sitemap.xml: one <url> per locale of every route
// routeFiles knows, each carrying an hreflang alternate for every locale
// plus x-default. routeFiles and locales are the exact route table
// pageHandler serves from, so this sitemap can never list a route the
// server does not actually answer, nor miss one it does.
func (a *App) sitemapHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")

		slugs := make([]string, 0, len(routeFiles))
		for slug := range routeFiles {
			slugs = append(slugs, slug)
		}
		sort.Strings(slugs)

		base := a.siteBase()

		var buf strings.Builder
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
		buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">` + "\n")
		for _, slug := range slugs {
			for _, locale := range locales {
				buf.WriteString("  <url>\n")
				fmt.Fprintf(&buf, "    <loc>%s</loc>\n", routeURL(base, locale, slug))
				for _, alt := range locales {
					fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"%s\" href=\"%s\"/>\n", alt, routeURL(base, alt, slug))
				}
				fmt.Fprintf(&buf, "    <xhtml:link rel=\"alternate\" hreflang=\"x-default\" href=\"%s\"/>\n", routeURL(base, defaultLocale, slug))
				buf.WriteString("  </url>\n")
			}
		}
		buf.WriteString("</urlset>\n")
		_, _ = w.Write([]byte(buf.String()))
	})
}

// siteBase is the configured site origin with no trailing slash, the base
// every sitemap and robots.txt URL is built from.
func (a *App) siteBase() string {
	return strings.TrimSuffix(a.siteURL.String(), "/")
}

// routeURL is the absolute URL of slug (a routeFiles key) under locale.
func routeURL(base, locale, slug string) string {
	if slug == "" {
		return base + "/" + locale + "/"
	}
	return base + "/" + locale + "/" + slug + "/"
}
