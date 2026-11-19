<script setup lang="ts">
import { OPERATOR } from './operator'
</script>

<template>
  <div class="flex flex-col gap-6 text-text">
    <p>
      {{ OPERATOR.name }}<br />
      {{ OPERATOR.street }}<br />
      {{ OPERATOR.postalCode }} {{ OPERATOR.city }}
    </p>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">Kontakt</h2>
      <p>
        Telefon:
        <a class="text-accent-text hover:underline" :href="OPERATOR.phoneHref">{{
          OPERATOR.phone
        }}</a><br />
        E-Mail:
        <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
          OPERATOR.email
        }}</a>
      </p>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">Umsatzsteuer-ID</h2>
      <p>Umsatzsteuer-Identifikationsnummer gemäß § 27a Umsatzsteuergesetz: {{ OPERATOR.vatId }}</p>
    </section>
  </div>
</template>
