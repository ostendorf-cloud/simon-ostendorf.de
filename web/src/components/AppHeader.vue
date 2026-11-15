<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { GovueHeader, GovueNav, type NavItem } from '@ostendorf-cloud/govue'

import LocaleSwitch from './LocaleSwitch.vue'
import ThemeToggle from './ThemeToggle.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const items = computed<NavItem[]>(() => {
  const locale = route.params.locale
  const resolve = (name: string) => router.resolve({ name, params: { locale } }).href
  return [
    { href: resolve('home'), label: t('nav.home'), current: route.name === 'home' },
    { href: resolve('cv'), label: t('nav.cv'), current: route.name === 'cv' },
    { href: resolve('contact'), label: t('nav.contact'), current: route.name === 'contact' },
  ]
})
</script>

<template>
  <GovueHeader>
    <template #brand>
      <router-link
        :to="{ name: 'home', params: { locale: route.params.locale } }"
        class="inline-flex min-h-11 items-center"
      >
        {{ t('common.siteName') }}
      </router-link>
    </template>
    <GovueNav :items="items" :aria-label="t('nav.main')" />
    <template #actions>
      <LocaleSwitch />
      <ThemeToggle />
    </template>
  </GovueHeader>
</template>
