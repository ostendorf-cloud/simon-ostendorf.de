<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { GovueCard, GovuePage } from '@ostendorf-cloud/govue'

const { t } = useI18n()

// The two ways on. Plain links to the sibling sites: nothing is requested
// from them until a visitor follows one.
const ways = [
  { key: 'dev', href: 'https://ostendorf.dev' },
  { key: 'tech', href: 'https://ostendorf.tech' },
] as const
</script>

<template>
  <GovuePage :title="t('home.heading')" :lead="t('home.lead')">
    <ul class="flex flex-col gap-6 sm:flex-row" data-testid="ways">
      <li v-for="way in ways" :key="way.key" class="flex-1">
        <GovueCard as="section" :title="t(`home.ways.${way.key}.title`)" class="h-full">
          <p class="text-text-muted">{{ t(`home.ways.${way.key}.text`) }}</p>
          <template #footer>
            <a
              :href="way.href"
              rel="noopener"
              class="inline-flex min-h-11 items-center font-medium text-accent-text hover:underline"
            >
              {{ t(`home.ways.${way.key}.link`) }}
            </a>
          </template>
        </GovueCard>
      </li>
    </ul>
  </GovuePage>
</template>
