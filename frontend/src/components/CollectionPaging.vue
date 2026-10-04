<!-- ABOUTME: Bounded collection navigation with explicit page-local scope. -->
<!-- ABOUTME: Refresh restarts offset paging; sorting and selection apply only to the visible page. -->
<script setup lang="ts">
import {useCollectionStore, COLLECTION_PAGE_SIZE} from '../stores/collectionStore'
const store = useCollectionStore()
</script>
<template>
  <nav class="collection-paging" aria-label="Collection pages">
    <p>Pages contain at most {{ COLLECTION_PAGE_SIZE }} items. Selection, sorting and page Insights apply to the current page only. Collection-wide totals and sidebar counts are separate snapshots.</p>
    <p>Changes can shift pages. Refresh starts again from page one.</p>
    <p v-if="store.loaded && !store.loading && !store.error" role="status">
      Page {{ store.offset / COLLECTION_PAGE_SIZE + 1 }} — {{ store.items.length }} item(s).
      {{ store.hasMore ? 'More results available.' : 'End of results.' }}
    </p>
    <div>
      <button :disabled="store.loading || store.offset === 0" @click="store.goToPage(store.offset - COLLECTION_PAGE_SIZE)">Previous page</button>
      <button :disabled="store.loading" @click="store.fetchItems()">Refresh from first page</button>
      <button :disabled="store.loading || !!store.error || !store.hasMore || store.offset >= 1000000" @click="store.goToPage(store.offset + COLLECTION_PAGE_SIZE)">Next page</button>
    </div>
    <p v-if="store.offset >= 1000000 && store.hasMore">Page limit reached. Narrow the search or filters to find remaining items.</p>
    <p v-if="store.offset > 0 && !store.loading && !store.error && !store.items.length">This page is empty after collection changes. Return to the first page; the collection may still contain items.</p>
  </nav>
</template>
<style scoped>
.collection-paging { padding: 12px 24px; color: var(--text-secondary); font-size: 13px; }
p { margin: 4px 0; }
.collection-paging div { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 8px; }
button { padding: 6px 12px; border: 1px solid var(--border-primary); border-radius: var(--radius-sm); color: var(--text-primary); background: var(--bg-secondary); cursor: pointer; }
button:disabled { opacity: .6; cursor: not-allowed; }
</style>
