<!-- ABOUTME: Independent collection-wide counts and recorded purchase totals. -->
<!-- ABOUTME: Clearly separates global snapshots from filtered/page Insights and absent/error values. -->
<script setup lang="ts">
import {useSummaryStore} from '../stores/summaryStore'
const store = useSummaryStore()
</script>
<template>
  <section class="collection-summary" aria-labelledby="global-summary-title">
    <h2 id="global-summary-title">Collection-wide totals</h2>
    <p>All collections, independent of search, filters and page. Snapshot, not live market valuation.</p>
    <p v-if="store.loading" role="status">Loading totals…</p>
    <p v-else-if="store.error" role="alert">Totals unavailable: {{ store.error }}</p>
    <div v-else-if="store.summary">
      <dl>
        <div><dt>Items</dt><dd>{{ store.summary.items.toLocaleString() }}</dd></div>
        <div><dt>Recorded purchase total</dt><dd>{{ !store.summary.valueAvailable ? 'Unavailable (overflow)' : store.summary.purchaseTotal === null ? 'No recorded prices' : store.summary.purchaseTotal.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2}) }}</dd></div>
      </dl>
      <p>{{ store.summary.pricedItems }} priced; {{ store.summary.items - store.summary.pricedItems - store.summary.invalidPrices }} missing prices; {{ store.summary.invalidPrices }} invalid prices excluded.</p>
      <p v-if="store.summary.modulesTruncated">Some module counts are omitted from the sidebar. Overall totals remain complete.</p>
      <p>Snapshot refreshed at {{ store.updatedAt }}. Changes elsewhere require refresh.</p>
    </div>
    <p v-else>Totals have not loaded.</p>
    <button :disabled="store.loading" @click="store.refresh()">Refresh collection-wide totals</button>
  </section>
</template>
<style scoped>
.collection-summary { margin: 12px 24px; padding: 16px; border: 1px solid var(--border-primary); border-radius: var(--radius-md); color: var(--text-primary); background: var(--bg-secondary); overflow-wrap: anywhere; }
h2 { font-family: var(--font-heading); font-size: 18px; margin: 0 0 8px; }
p { font-size: 13px; color: var(--text-secondary); margin: 6px 0; }
dl { display: flex; gap: 24px; flex-wrap: wrap; }
dt { font-size: 13px; }
dd { font-size: 22px; margin: 4px 0; }
[role=alert] { color: var(--error-text); }
button { padding: 6px 12px; border: 1px solid var(--border-primary); border-radius: var(--radius-sm); color: var(--text-primary); background: var(--bg-primary); cursor: pointer; }
button:disabled { opacity: .6; cursor: not-allowed; }
</style>
