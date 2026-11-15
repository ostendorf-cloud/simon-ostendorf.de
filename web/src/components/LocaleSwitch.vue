<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import { type Locale } from '../i18n'

/** The other locale's version of the current route: same name, same
 *  params and hash, only :locale swapped, so switching never drops a
 *  visitor back to the home page. */
const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const otherLocale = computed<Locale>(() => (route.params.locale === 'de' ? 'en' : 'de'))
// A repeatable param resolved with no entries, the not-found route's own
// pathMatch for instance, drops the trailing slash the route pattern
// otherwise always carries; restored here so the link is already
// canonical instead of relying on the server's own redirect for it.
const target = computed(() => {
  const href = router.resolve({
    name: route.name ?? undefined,
    params: { ...route.params, locale: otherLocale.value },
    hash: route.hash,
  }).href
  return href.endsWith('/') ? href : `${href}/`
})
</script>

<template>
  <router-link
    :to="target"
    class="inline-flex min-h-11 items-center justify-center gap-2 rounded-md px-4 text-base font-medium text-accent-text transition-colors hover:bg-accent-subtle"
  >
    {{ t('common.switchLocale') }}
  </router-link>
</template>
