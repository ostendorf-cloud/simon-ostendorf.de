/**
 * Which theme this site ships, read from here by the two sides that have
 * to agree on it: src/main.ts hands it to useTheme, and
 * web/scripts/prerender.ts writes the same value into every page's
 * data-theme attribute. This site ships exactly one theme, the
 * placeholder palette AGENTS.md describes; see
 * src/theme/simon-ostendorf-de.css for why it is a placeholder rather than
 * a real brand colour.
 *
 * The mode is left at useTheme's own default, "system": unlike
 * ostendorf.tech this site has no reason to open dark by default, so
 * there is no pinned mode constant to keep prerender.ts and main.ts in
 * step on, and no data-mode attribute written into the prerendered
 * markup either; useTheme's own documented behaviour for an absent
 * attribute, "follow the operating system", is exactly the starting
 * point this site wants.
 */
export const THEME_NAME = 'simon-ostendorf-de'
