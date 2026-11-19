<script setup lang="ts">
import { OPERATOR } from './operator'
</script>

<template>
  <div class="flex flex-col gap-6 text-text">
    <p class="text-text-muted">
      This is a translation. The German version of this page is the legally binding one.
    </p>

    <p>
      {{ OPERATOR.name }}<br />
      {{ OPERATOR.street }}<br />
      {{ OPERATOR.postalCode }} {{ OPERATOR.city }}, {{ OPERATOR.country }}
    </p>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">Contact</h2>
      <p>
        Phone:
        <a class="text-accent-text hover:underline" :href="OPERATOR.phoneHref">{{
          OPERATOR.phone
        }}</a><br />
        Email:
        <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
          OPERATOR.email
        }}</a>
      </p>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">VAT ID</h2>
      <p>VAT identification number per Section 27a of the German VAT Act: {{ OPERATOR.vatId }}</p>
    </section>
  </div>
</template>
