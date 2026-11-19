<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { GovueLegalPage } from '@ostendorf-cloud/govue'

import PrivacyTextDe from '../legal/PrivacyTextDe.vue'
import PrivacyTextEn from '../legal/PrivacyTextEn.vue'

const { t, locale } = useI18n()
const text = computed(() => (locale.value === 'en' ? PrivacyTextEn : PrivacyTextDe))
</script>

<template>
  <!-- The header already carries the language switch and App.vue the skip
       link, so the shell's own are off. -->
  <GovueLegalPage
    page="privacy"
    :title="t('legal.privacy.title')"
    :language-switch="false"
    :skip-link="false"
  >
    <component :is="text" />
  </GovueLegalPage>
</template>
