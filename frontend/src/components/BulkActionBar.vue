<!-- ABOUTME: Floating glassmorphism action bar for bulk operations on selected items. -->
<!-- ABOUTME: Shows selected count and offers Delete, Export CSV, Edit Module, and Deselect actions. -->
<script lang="ts" setup>
defineProps<{
  count: number
}>()

const emit = defineEmits<{
  delete: []
  export: []
  editModule: []
  deselectAll: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="bar">
      <div v-if="count > 0" class="bulk-bar">
        <span class="bar-count">{{ count }} item{{ count === 1 ? '' : 's' }} selected</span>
        <div class="bar-actions">
          <button class="bar-btn bar-btn-danger" @click="emit('delete')">Delete Selected</button>
          <button class="bar-btn" @click="emit('export')">Export CSV</button>
          <button class="bar-btn" @click="emit('editModule')">Edit Module</button>
          <button class="bar-btn bar-btn-muted" @click="emit('deselectAll')">Deselect All</button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.bulk-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 2500;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 20px;
  background: var(--bg-secondary, rgba(30, 30, 46, 0.85));
  backdrop-filter: blur(var(--glass-blur, 20px));
  -webkit-backdrop-filter: blur(var(--glass-blur, 20px));
  border: 1px solid var(--border-primary, rgba(255,255,255,0.08));
  border-radius: var(--radius-lg, 16px);
  box-shadow: var(--shadow-lg, 0 16px 48px rgba(0,0,0,0.3));
}
.bar-count {
  font-family: 'Outfit', sans-serif;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
}
.bar-actions {
  display: flex;
  gap: 6px;
}
.bar-btn {
  padding: 6px 14px;
  border: none;
  border-radius: var(--radius-sm, 6px);
  background: var(--accent-blue);
  color: var(--text-on-accent);
  font-family: 'Outfit', sans-serif;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: background var(--transition-fast);
}
.bar-btn:hover {
  background: var(--accent-blue-hover);
}
.bar-btn-danger {
  background: var(--error-bg, rgba(239, 68, 68, 0.15));
  color: var(--error-text, #ef4444);
}
.bar-btn-danger:hover {
  background: var(--error-border, #dc2626);
  color: #fff;
}
.bar-btn-muted {
  background: transparent;
  color: var(--text-muted);
  border: 1px solid var(--border-primary);
}
.bar-btn-muted:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* Slide-up animation */
.bar-enter-active {
  transition: transform 0.2s ease-out, opacity 0.2s ease-out;
}
.bar-leave-active {
  transition: transform 0.15s ease-in, opacity 0.15s ease-in;
}
.bar-enter-from {
  transform: translateX(-50%) translateY(20px);
  opacity: 0;
}
.bar-leave-to {
  transform: translateX(-50%) translateY(20px);
  opacity: 0;
}
</style>
