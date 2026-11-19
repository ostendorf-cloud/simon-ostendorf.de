import { createSSRApp } from 'vue'
import { createWebHistory } from 'vue-router'
import { useTheme } from '@ostendorf-cloud/govue'

import App from './App.vue'
import { createAppI18n, detectPreferredLocale, installI18n, isLocale } from './i18n'
import { THEME_NAME } from './lib/theme'
import { attachLocaleSync, createAppRouter } from './router'
import './styles.css'

// The theme comes from src/lib/theme.ts, which the prerender script reads
// too, so both ends open on the exact same value. Mode is left at
// useTheme's own default, "system": see that file's own comment for why
// this site, unlike ostendorf.tech, pins no mode of its own.
useTheme({ theme: THEME_NAME })

const firstSegment = window.location.pathname.split('/')[1] ?? ''
const initialLocale = isLocale(firstSegment)
  ? firstSegment
  : detectPreferredLocale(navigator.languages ?? [navigator.language])

const i18n = createAppI18n(initialLocale)
const router = createAppRouter(createWebHistory())
attachLocaleSync(router, i18n.global.locale)

const app = createSSRApp(App)
app.use(router)
installI18n(app, i18n)

// router.isReady() resolves after the initial navigation, including the
// async route components above, which is what the hydration this app
// mounts into needs: the DOM mount has to match what the server rendered
// for this exact path, not a transient empty state.
void router.isReady().then(() => {
  app.mount('#app')
})
