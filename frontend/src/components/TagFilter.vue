<!-- ABOUTME: Clickable tag chip filter bar for cross-collection tag filtering. -->
<!-- ABOUTME: Renders all tags as toggleable chips; selected tags filter the collection view. -->
<script lang="ts" setup>
import type {TagCount} from '../api/types'

const props = defineProps<{
  allTags: TagCount[]
  selectedTags: string[]
}>()

const emit = defineEmits<{
  update: [tags: string[]]
}>()

function toggleTag(name: string) {
  const selected = props.selectedTags.includes(name)
  if (selected) {
    emit('update', props.selectedTags.filter(t => t !== name))
  } else {
    emit('update', [...props.selectedTags, name])
  }
}
</script>

<template>
  <div v-if="allTags.length > 0" class="tag-filter-bar">
    <span class="tag-filter-label">Tags</span>
    <button
      v-for="tag in allTags"
      :key="tag.name"
      :class="['tag-filter-chip', {active: selectedTags.includes(tag.name)}]"
      @click="toggleTag(tag.name)"
    >
      {{ tag.name }}
      <span class="tag-count">{{ tag.count }}</span>
    </button>
  </div>
</template>

<style scoped>
.tag-filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}
.tag-filter-label {
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide, 0.08em);
  color: var(--text-muted);
  margin-right: 4px;
}
.tag-filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  font-family: 'Outfit', sans-serif;
  transition: background 0.12s, color 0.12s, border-color 0.12s;
}
.tag-filter-chip:hover {
  border-color: var(--accent-blue);
  color: var(--text-primary);
}
.tag-filter-chip.active {
  background: var(--accent-blue);
  color: var(--text-on-accent);
  border-color: var(--accent-blue);
}
.tag-count {
  font-size: 10px;
  opacity: 0.7;
}
</style>
