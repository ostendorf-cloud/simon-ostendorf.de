<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { GovuePage } from '@ostendorf-cloud/govue'

const { t } = useI18n()
</script>

<template>
  <!-- This domain's own imprint content is a separate issue; see
       AGENTS.md's "Scope of this milestone". This shell only has to
       carry a heading. -->
  <GovuePage :title="t('legal.imprint.title')" />
</template>
