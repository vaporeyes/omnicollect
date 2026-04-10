<script lang="ts" setup>
import {ref, onMounted} from 'vue'
import type {ModuleSchema, Showcase} from '../api/types'
import {toggleShowcase, listShowcases, getAIStatus} from '../api/client'

defineProps<{
  modules: ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [module: ModuleSchema]
  edit: [module: ModuleSchema]
  createSchema: []
}>()

// Showcase state per module
const showcaseMap = ref<Record<string, Showcase>>({})
const copiedSlug = ref<string | null>(null)
const isCloudMode = ref(false)

onMounted(async () => {
  // Check AI status to determine if we're in cloud mode (showcase feature requires it)
  try {
    const status = await getAIStatus()
    isCloudMode.value = status.cloudMode === true
    if (isCloudMode.value) {
      const showcases = await listShowcases()
      for (const sc of showcases) {
        showcaseMap.value[sc.moduleId] = sc
      }
    }
  } catch {
    isCloudMode.value = false
  }
})

async function onToggleShowcase(mod: ModuleSchema, event: Event) {
  event.stopPropagation()
  const current = showcaseMap.value[mod.id]
  const newEnabled = !current?.enabled
  try {
    const result = await toggleShowcase(mod.id, newEnabled)
    showcaseMap.value[mod.id] = result
  } catch (e) {
    console.error('Failed to toggle showcase:', e)
  }
}

function copyShowcaseUrl(mod: ModuleSchema, event: Event) {
  event.stopPropagation()
  const sc = showcaseMap.value[mod.id]
  if (!sc?.url) return
  const url = window.location.origin + sc.url
  navigator.clipboard.writeText(url)
  copiedSlug.value = sc.slug
  setTimeout(() => { copiedSlug.value = null }, 2000)
}
</script>

<template>
  <div class="module-selector">
    <h3>Collection Types</h3>
    <div v-if="modules.length === 0" class="empty-state">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
        <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
        <line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/>
      </svg>
      <p>No collection types yet</p>
    </div>
    <ul v-else class="module-list">
      <li
        v-for="(mod, index) in modules"
        :key="mod.id"
        class="module-item animate-slide-up"
        :style="{ animationDelay: `${index * 0.05}s` }"
        @click="emit('select', mod)"
      >
        <span class="module-name-row">
          <span class="module-name">{{ mod.displayName }}</span>
          <span class="module-actions">
            <button
              v-if="isCloudMode"
              class="share-btn"
              :class="{active: showcaseMap[mod.id]?.enabled}"
              @click.stop="onToggleShowcase(mod, $event)"
              :title="showcaseMap[mod.id]?.enabled ? 'Make private' : 'Make public'"
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M4 12v8a2 2 0 002 2h12a2 2 0 002-2v-8"/>
                <polyline points="16 6 12 2 8 6"/>
                <line x1="12" y1="2" x2="12" y2="15"/>
              </svg>
            </button>
            <button
              v-if="isCloudMode && showcaseMap[mod.id]?.enabled"
              class="copy-btn"
              @click.stop="copyShowcaseUrl(mod, $event)"
              :title="copiedSlug === showcaseMap[mod.id]?.slug ? 'Copied!' : 'Copy link'"
            >
              {{ copiedSlug === showcaseMap[mod.id]?.slug ? '&#10003;' : '&#128279;' }}
            </button>
            <button class="edit-btn" @click.stop="emit('edit', mod)" title="Edit schema">&#9998;</button>
          </span>
        </span>
        <span v-if="mod.description" class="module-desc">{{ mod.description }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.module-selector h3 {
  margin: 0 0 6px 0;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
}
.module-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.module-item {
  padding: 7px 12px;
  cursor: pointer;
  border-radius: 0;
  margin-bottom: 1px;
  border-left: 2px solid transparent;
  transition: background var(--transition-fast), border-color var(--transition-fast);
}
.module-item:hover {
  background: var(--bg-hover);
}
.module-item:hover .edit-btn {
  opacity: 1;
}
.module-item.active {
  border-left-color: var(--accent-blue);
  background: var(--bg-hover);
}
.module-name-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.module-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  line-height: var(--leading-tight);
}
.module-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}
.edit-btn, .share-btn, .copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-muted);
  padding: 0 2px;
  opacity: 0;
  transition: opacity var(--transition-fast), color var(--transition-fast);
}
.edit-btn:hover, .share-btn:hover, .copy-btn:hover {
  color: var(--accent-blue);
}
.share-btn.active {
  color: var(--accent-blue);
  opacity: 0.8;
}
.module-item:hover .share-btn,
.module-item:hover .copy-btn {
  opacity: 1;
}
.module-desc {
  font-size: 11px;
  color: var(--text-muted);
  line-height: var(--leading-dense);
  margin-top: 1px;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 20px 12px;
  color: var(--text-muted);
}
.empty-state svg {
  margin-bottom: 8px;
  opacity: 0.5;
}
.empty-state p {
  margin: 0 0 12px 0;
  font-size: 13px;
}
.cta-btn {
  padding: 8px 16px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent-blue);
  color: var(--text-on-accent);
  cursor: pointer;
  font-size: 13px;
}
.cta-btn:hover {
  background: var(--accent-blue-hover);
}
</style>
