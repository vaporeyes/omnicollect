<script lang="ts" setup>
// ABOUTME: Reusable glassmorphism summary card for the insights dashboard.
// ABOUTME: Displays a label, value, and optional subtitle with click support.

defineProps<{
  label: string
  value: string
  subtitle?: string
  clickable?: boolean
}>()

defineEmits<{
  (e: 'click'): void
}>()
</script>

<template>
  <div
    class="metric-card"
    :class="{clickable}"
    @click="clickable ? $emit('click') : undefined"
  >
    <span class="metric-label">{{ label }}</span>
    <span class="metric-value">{{ value }}</span>
    <span v-if="subtitle" class="metric-subtitle">{{ subtitle }}</span>
  </div>
</template>

<style scoped>
.metric-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-lg);
  padding: var(--space-lg) var(--space-lg);
  display: flex;
  flex-direction: column;
  gap: 4px;
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.metric-card.clickable {
  cursor: pointer;
}

.metric-card.clickable:hover {
  transform: scale(1.02);
  box-shadow: var(--shadow-md);
}

.metric-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
}

.metric-value {
  font-family: var(--font-mono);
  font-size: 26px;
  font-weight: 500;
  color: var(--text-primary);
  line-height: var(--leading-dense);
}

.metric-subtitle {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}
</style>
