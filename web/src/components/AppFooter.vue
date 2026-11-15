<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { GovueFooter, type FooterLink } from '@ostendorf-cloud/govue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const links = computed<FooterLink[]>(() => {
  const locale = route.params.locale
  const resolve = (name: string) => router.resolve({ name, params: { locale } }).href
  return [
    { href: resolve('privacy'), label: t('legal.privacy.title') },
    { href: resolve('imprint'), label: t('legal.imprint.title') },
  ]
})

const note = computed(() => t('footer.copyright', { year: new Date().getFullYear() }))
</script>

<template>
  <GovueFooter :links="links" :note="note" :aria-label="t('legal.eyebrow')" />
</template>
