<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { GovuePage } from '@ostendorf-cloud/govue'

const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <GovuePage :title="t('notFound.title')" :lead="t('notFound.lead')">
    <router-link
      :to="{ name: 'home', params: { locale: route.params.locale } }"
      class="inline-flex min-h-11 items-center text-base font-medium text-accent-text transition-colors hover:underline"
    >
      {{ t('notFound.backHome') }}
    </router-link>
  </GovuePage>
</template>
