<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { GovuePage } from '@ostendorf-cloud/govue'

import CvDocument from '../components/CvDocument.vue'
import { isLocale } from '../i18n'
import { loadCv } from '../lib/cv'

const { t, locale } = useI18n()

// Build time data: loadCv throws when the build inputs are missing.
const cv = loadCv()

// The prerendered HTML marks an expired certificate against the build time.
// With JavaScript on, the mark moves to the visitor's own date after mount,
// so a page served long after its build is still right.
const now = ref(new Date(cv.dataset.generated_at))
onMounted(() => {
  now.value = new Date()
})

const cvDoc = computed(() => cv.dataset.documents[isLocale(locale.value) ? locale.value : 'de'])
</script>

<template>
  <GovuePage :title="t('cv.heading')" :lead="t('cv.lead')">
    <CvDocument :document="cvDoc" :pdfs="cv.pdfs" :now="now" />
  </GovuePage>
</template>
