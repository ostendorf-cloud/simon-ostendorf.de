<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { GovuePage } from '@ostendorf-cloud/govue'

const { t } = useI18n()
</script>

<template>
  <!-- The contact form itself is a separate issue; see AGENTS.md's "Scope
       of this milestone". This shell only has to carry a heading. -->
  <GovuePage :title="t('contact.heading')" />
</template>
