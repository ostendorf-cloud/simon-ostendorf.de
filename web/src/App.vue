<script setup lang="ts">
import { GovueFooter, GovueHeader, GovuePage, GovueSkipLink } from '@ostendorf-cloud/govue'
</script>

<template>
  <GovueSkipLink>Skip to content</GovueSkipLink>
  <GovueHeader>
    <template #brand>Simon Ostendorf</template>
  </GovueHeader>
  <GovuePage title="Simon Ostendorf" lead="This site is under construction.">
    <p>
      The personal hub at simon-ostendorf.de is still a scaffold. See the repository's AGENTS.md for
      what is planned and not yet built.
    </p>
  </GovuePage>
  <GovueFooter note="© 2026 Simon Ostendorf" />
</template>
