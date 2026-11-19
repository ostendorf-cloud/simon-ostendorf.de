<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { GovueLegalPage } from '@ostendorf-cloud/govue'

import ImprintTextDe from '../legal/ImprintTextDe.vue'
import ImprintTextEn from '../legal/ImprintTextEn.vue'

const { t, locale } = useI18n()
const text = computed(() => (locale.value === 'en' ? ImprintTextEn : ImprintTextDe))
</script>

<template>
  <!-- The header already carries the language switch and App.vue the skip
       link, so the shell's own are off. -->
  <GovueLegalPage
    page="imprint"
    :title="t('legal.imprint.title')"
    :language-switch="false"
    :skip-link="false"
  >
    <component :is="text" />
  </GovueLegalPage>
</template>
