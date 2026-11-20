<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { GovueBadge, GovueCard } from '@ostendorf-cloud/govue'

import { isLocale, SUPPORTED_LOCALES, type Locale } from '../i18n'
import { formatDate, isExpired, type CvDocument } from '../lib/cv'

const props = defineProps<{
  document: CvDocument
  /** The URL of each language's PDF. */
  pdfs: Record<Locale, string>
  /** The reference for the expiry mark of a certificate. */
  now: Date
}>()

const i18n = useI18n()
const { t, locale } = i18n

const uiLocale = computed<Locale>(() => (isLocale(locale.value) ? locale.value : 'de'))

// The page's own language first, so the likely download comes first.
const pdfLocales = computed(() => [
  uiLocale.value,
  ...SUPPORTED_LOCALES.filter((l) => l !== uiLocale.value),
])

const hasEngagements = computed(() => props.document.engagements.length > 0)

function date(value: string): string {
  return formatDate(value, uiLocale.value)
}

function period(start: string | null, end: string | null): string {
  if (!start && !end) {
    return ''
  }
  const from = start ? date(start) : ''
  const to = end ? date(end) : t('cv.present')
  return `${from} – ${to}`
}

function level(value: string): string {
  const key = `cv.levels.${value}`
  return i18n.te(key) ? t(key) : value
}

function employerOf(experienceId: string): string | null {
  return props.document.experience.find((e) => e.id === experienceId)?.employer ?? null
}
</script>

<template>
  <GovueCard as="section" :title="t('cv.pdf.title')" data-testid="cv-pdf">
    <ul class="flex flex-col gap-2">
      <li v-for="l in pdfLocales" :key="l">
        <a
          :href="pdfs[l]"
          :hreflang="l"
          :lang="l"
          :download="`simon-ostendorf-cv.${l}.pdf`"
          type="application/pdf"
          class="inline-flex min-h-11 items-center font-medium text-accent-text hover:underline"
        >
          {{ t(`cv.pdf.link.${l}`) }}
        </a>
      </li>
    </ul>
  </GovueCard>

  <GovueCard
    v-if="document.experience.length > 0"
    as="section"
    :title="t('cv.sections.experience')"
    data-testid="cv-experience"
  >
    <ol class="flex flex-col gap-5">
      <li v-for="e in document.experience" :key="e.id" class="flex flex-col gap-1">
        <h3 class="font-semibold">{{ e.title }}, {{ e.employer }}</h3>
        <p class="text-sm text-text-muted">
          {{ period(e.start, e.end) }}
          <template v-if="e.location">, {{ e.location }}</template>
          <template v-if="e.employment_type">, {{ e.employment_type }}</template>
        </p>
        <p v-if="e.summary">{{ e.summary }}</p>
      </li>
    </ol>
  </GovueCard>

  <GovueCard
    v-if="document.education.length > 0"
    as="section"
    :title="t('cv.sections.education')"
    data-testid="cv-education"
  >
    <ol class="flex flex-col gap-5">
      <li v-for="e in document.education" :key="e.id" class="flex flex-col gap-1">
        <h3 class="font-semibold">{{ e.degree }}, {{ e.institution }}</h3>
        <p class="text-sm text-text-muted">
          {{ period(e.start, e.end) }}
          <template v-if="e.grade">, {{ t('cv.grade') }}: {{ e.grade }}</template>
        </p>
        <p v-if="e.summary">{{ e.summary }}</p>
      </li>
    </ol>
  </GovueCard>

  <GovueCard
    v-if="document.certificates.length > 0"
    as="section"
    :title="t('cv.sections.certificates')"
    data-testid="cv-certificates"
  >
    <ul class="flex flex-col gap-4">
      <li v-for="c in document.certificates" :key="c.id" class="flex flex-col gap-1">
        <h3 class="font-semibold">
          {{ c.name }}
          <GovueBadge v-if="isExpired(c.valid_until, now)" kind="danger" data-testid="expired">
            {{ t('cv.expired') }}
          </GovueBadge>
        </h3>
        <p class="text-sm text-text-muted">
          {{ c.issuer }}, {{ t('cv.issued') }} {{ date(c.issued) }}
          <template v-if="c.valid_until">
            , {{ t('cv.validUntil') }} {{ date(c.valid_until) }}
          </template>
        </p>
      </li>
    </ul>
  </GovueCard>

  <GovueCard
    v-if="document.skills.length > 0"
    as="section"
    :title="t('cv.sections.skills')"
    data-testid="cv-skills"
  >
    <ul class="flex flex-wrap gap-2">
      <li v-for="s in document.skills" :key="s.id">
        <GovueBadge>{{ s.name }}: {{ level(s.level) }}</GovueBadge>
      </li>
    </ul>
  </GovueCard>

  <GovueCard
    v-if="hasEngagements"
    as="section"
    :title="t('cv.sections.engagements')"
    data-testid="cv-engagements"
  >
    <ul class="flex flex-col gap-5">
      <li v-for="e in document.engagements" :key="e.id" class="flex flex-col gap-1">
        <h3 class="font-semibold">{{ e.title }}</h3>
        <p class="text-sm text-text-muted">
          <template v-if="employerOf(e.experience)">{{ employerOf(e.experience) }}</template>
          <template v-if="e.industry">, {{ e.industry }}</template>
          <template v-if="e.start || e.end">, {{ period(e.start, e.end) }}</template>
        </p>
        <p>{{ e.summary }}</p>
      </li>
    </ul>
  </GovueCard>

  <GovueCard
    v-if="document.languages.length > 0"
    as="section"
    :title="t('cv.sections.languages')"
    data-testid="cv-languages"
  >
    <ul class="flex flex-col gap-1">
      <li v-for="l in document.languages" :key="l.id">{{ l.name }}: {{ level(l.level) }}</li>
    </ul>
  </GovueCard>
</template>
