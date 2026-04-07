<!-- ABOUTME: Collapsible filter bar dynamically generated from the active module's schema. -->
<!-- ABOUTME: Supports enum multi-select pills, boolean tri-state toggles, and number range inputs. -->
<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import {main} from '../../wailsjs/go/models'
import type {AttributeFilter} from '../stores/collectionStore'

const props = defineProps<{
  schema: main.ModuleSchema | null
  filters: Record<string, AttributeFilter[]>
}>()

const emit = defineEmits<{
  update: [filters: Record<string, AttributeFilter[]>]
  clear: []
}>()

const expanded = ref(false)

// Debounce timers for number inputs
const numberTimers: Record<string, ReturnType<typeof setTimeout>> = {}

// Extract filterable attributes from schema
const enumAttrs = computed(() =>
  props.schema?.attributes.filter(a => a.type === 'enum' && a.options?.length) ?? []
)
const boolAttrs = computed(() =>
  props.schema?.attributes.filter(a => a.type === 'boolean') ?? []
)
const numberAttrs = computed(() =>
  props.schema?.attributes.filter(a => a.type === 'number') ?? []
)

const hasFilterableAttrs = computed(() =>
  enumAttrs.value.length > 0 || boolAttrs.value.length > 0 || numberAttrs.value.length > 0
)

const activeFilterCount = computed(() => Object.keys(props.filters).length)
const hasActiveFilters = computed(() => activeFilterCount.value > 0)

// Helper: check if an enum value is selected for a field
function isEnumSelected(field: string, value: string): boolean {
  const f = props.filters[field]
  if (!f) return false
  const inFilter = f.find(x => x.op === 'in')
  return inFilter?.values?.includes(value) ?? false
}

// Helper: get boolean filter state for a field (null=off, true, false)
function boolState(field: string): boolean | null {
  const f = props.filters[field]
  if (!f) return null
  const eqFilter = f.find(x => x.op === 'eq')
  if (!eqFilter) return null
  return eqFilter.value as boolean
}

// Helper: get number range value for a field
function numberValue(field: string, op: 'gte' | 'lte'): string {
  const f = props.filters[field]
  if (!f) return ''
  const rangeFilter = f.find(x => x.op === op)
  if (!rangeFilter || rangeFilter.value === undefined || rangeFilter.value === null) return ''
  return String(rangeFilter.value)
}

function toggleEnum(field: string, value: string) {
  const next = {...props.filters}
  const current = next[field]?.find(x => x.op === 'in')
  const selected = current?.values ? [...current.values] : []

  const idx = selected.indexOf(value)
  if (idx >= 0) {
    selected.splice(idx, 1)
  } else {
    selected.push(value)
  }

  if (selected.length === 0) {
    delete next[field]
  } else {
    next[field] = [{field, op: 'in', values: selected}]
  }
  emit('update', next)
}

function toggleBool(field: string) {
  const next = {...props.filters}
  const state = boolState(field)
  if (state === null) {
    next[field] = [{field, op: 'eq', value: true}]
  } else if (state === true) {
    next[field] = [{field, op: 'eq', value: false}]
  } else {
    delete next[field]
  }
  emit('update', next)
}

function onNumberInput(field: string, op: 'gte' | 'lte', rawValue: string) {
  if (numberTimers[field + op]) clearTimeout(numberTimers[field + op])
  numberTimers[field + op] = setTimeout(() => {
    const next = {...props.filters}
    const existing = (next[field] ?? []).filter(f => f.op !== op)
    if (rawValue !== '') {
      existing.push({field, op, value: parseFloat(rawValue)})
    }
    if (existing.length === 0) {
      delete next[field]
    } else {
      next[field] = existing
    }
    emit('update', next)
  }, 400)
}

function attrLabel(attr: {name: string; display?: {label?: string} | null}): string {
  return attr.display?.label || attr.name
}

// Reset expanded state when schema changes
watch(() => props.schema?.id, () => {
  expanded.value = false
})
</script>

<template>
  <div v-if="schema && hasFilterableAttrs" class="filter-bar">
    <!-- Collapsed row -->
    <div class="filter-header">
      <button class="expand-toggle" @click="expanded = !expanded">
        <svg
          :class="['toggle-icon', {rotated: expanded}]"
          width="14" height="14" viewBox="0 0 24 24"
          fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"
        >
          <path d="M9 18l6-6-6-6"/>
        </svg>
        <span class="filter-label">Filters</span>
        <span v-if="hasActiveFilters" class="filter-count">{{ activeFilterCount }} active</span>
      </button>
      <button v-if="hasActiveFilters" class="clear-btn" @click="emit('clear')">Clear all</button>
    </div>

    <!-- Expanded controls -->
    <div v-if="expanded" class="filter-controls">
      <!-- Enum facets -->
      <div v-for="attr in enumAttrs" :key="attr.name" class="facet-group">
        <div class="facet-label">{{ attrLabel(attr) }}</div>
        <div class="pill-row">
          <button
            v-for="opt in attr.options"
            :key="opt"
            :class="['pill', {active: isEnumSelected(attr.name, opt)}]"
            @click="toggleEnum(attr.name, opt)"
          >{{ opt }}</button>
        </div>
      </div>

      <!-- Boolean facets -->
      <div v-for="attr in boolAttrs" :key="attr.name" class="facet-group">
        <div class="facet-label">{{ attrLabel(attr) }}</div>
        <div class="pill-row">
          <button
            :class="['pill', 'pill-bool', {
              'active-yes': boolState(attr.name) === true,
              'active-no': boolState(attr.name) === false
            }]"
            @click="toggleBool(attr.name)"
          >
            {{ boolState(attr.name) === null ? attrLabel(attr) : boolState(attr.name) ? 'Yes' : 'No' }}
          </button>
        </div>
      </div>

      <!-- Number range facets (schema attributes) -->
      <div v-for="attr in numberAttrs" :key="attr.name" class="facet-group">
        <div class="facet-label">{{ attrLabel(attr) }}</div>
        <div class="range-row">
          <input
            type="number"
            class="range-input"
            placeholder="Min"
            :value="numberValue(attr.name, 'gte')"
            @input="onNumberInput(attr.name, 'gte', ($event.target as HTMLInputElement).value)"
          />
          <span class="range-sep">-</span>
          <input
            type="number"
            class="range-input"
            placeholder="Max"
            :value="numberValue(attr.name, 'lte')"
            @input="onNumberInput(attr.name, 'lte', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <!-- Purchase Price (always present when a module is active) -->
      <div class="facet-group">
        <div class="facet-label">Purchase Price</div>
        <div class="range-row">
          <input
            type="number"
            class="range-input"
            placeholder="Min"
            :value="numberValue('purchasePrice', 'gte')"
            @input="onNumberInput('purchasePrice', 'gte', ($event.target as HTMLInputElement).value)"
          />
          <span class="range-sep">-</span>
          <input
            type="number"
            class="range-input"
            placeholder="Max"
            :value="numberValue('purchasePrice', 'lte')"
            @input="onNumberInput('purchasePrice', 'lte', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  margin-bottom: 12px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  backdrop-filter: blur(var(--glass-blur));
  -webkit-backdrop-filter: blur(var(--glass-blur));
  overflow: hidden;
}

/* Header row */
.filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
}
.expand-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary);
  font-family: 'Outfit', sans-serif;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  padding: 0;
}
.expand-toggle:hover {
  color: var(--text-primary);
}
.toggle-icon {
  transition: transform var(--transition-fast);
}
.toggle-icon.rotated {
  transform: rotate(90deg);
}
.filter-count {
  font-weight: 500;
  color: var(--accent-blue);
  text-transform: none;
  letter-spacing: normal;
  font-size: 11px;
}
.clear-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-family: 'Outfit', sans-serif;
  font-size: 11px;
  color: var(--error-text, #ef4444);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}
.clear-btn:hover {
  background: var(--error-bg, rgba(239, 68, 68, 0.1));
}

/* Controls area */
.filter-controls {
  padding: 4px 12px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  border-top: 1px solid var(--border-primary);
}
.facet-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.facet-label {
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: var(--tracking-wide);
  color: var(--text-muted);
}

/* Pill row for enums and booleans */
.pill-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.pill {
  padding: 4px 10px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: transparent;
  color: var(--text-secondary);
  font-family: 'Outfit', sans-serif;
  font-size: 12px;
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast), border-color var(--transition-fast);
}
.pill:hover {
  border-color: var(--accent-blue);
  color: var(--text-primary);
}
.pill.active {
  background: var(--accent-blue);
  border-color: var(--accent-blue);
  color: var(--text-on-accent);
}

/* Boolean tri-state pills */
.pill-bool.active-yes {
  background: var(--success-text, #22c55e);
  border-color: var(--success-text, #22c55e);
  color: #fff;
}
.pill-bool.active-no {
  background: var(--error-text, #ef4444);
  border-color: var(--error-text, #ef4444);
  color: #fff;
}

/* Number range inputs */
.range-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
.range-input {
  width: 80px;
  padding: 4px 8px;
  border: 1px solid var(--border-input);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  font-family: 'Outfit', sans-serif;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}
.range-input:focus {
  outline: none;
  border-color: var(--accent-blue);
  box-shadow: 0 0 0 2px var(--accent-blue-light);
}
.range-sep {
  color: var(--text-muted);
  font-size: 12px;
}
</style>
