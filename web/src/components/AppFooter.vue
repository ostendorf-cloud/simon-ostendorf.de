<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { GovueFooter, type LegalPage } from '@ostendorf-cloud/govue'

const { t } = useI18n()

// This site has no terms, so no terms link: govue-check-legal requires the
// imprint and the privacy link on every prerendered page, in its own
// language, at the exact path /<code>/<page>.
const legal: readonly LegalPage[] = ['imprint', 'privacy']

const note = computed(() => t('footer.copyright', { year: new Date().getFullYear() }))
</script>

<template>
  <GovueFooter :legal="legal" :note="note" :aria-label="t('legal.eyebrow')" />
</template>
