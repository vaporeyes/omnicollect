<!-- ABOUTME: Collapsible filter bar dynamically generated from the active module's schema. -->
<!-- ABOUTME: Supports enum multi-select pills, boolean tri-state toggles, and number range inputs. -->
<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import type {ModuleSchema} from '../api/types'
import type {AttributeFilter} from '../stores/collectionStore'

const props = defineProps<{
  schema: ModuleSchema | null
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

function setBool(field: string, state: boolean | null) {
  const next = {...props.filters}
  if (state === null) {
    delete next[field]
  } else {
    next[field] = [{field, op: 'eq', value: state}]
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

function clearFilter(field: string) {
  const next = {...props.filters}
  delete next[field]
  emit('update', next)
}

function attrLabel(attr: {name: string; display?: {label?: string} | null}): string {
  return attr.display?.label || attr.name
}

// Map of active filters for chips
const activeChips = computed(() => {
  const chips: Array<{id: string, label: string, valueStr: string}> = []
  for (const [field, filters] of Object.entries(props.filters)) {
    let label = field
    // Try to find display label
    const attr = props.schema?.attributes.find(a => a.name === field)
    if (attr && attr.display?.label) label = attr.display.label
    if (field === 'purchasePrice') label = 'Purchase Price'

    // Format value string based on ops
    let valueStr = ''
    const inFilter = filters.find(f => f.op === 'in')
    if (inFilter && inFilter.values) {
      valueStr = inFilter.values.join(', ')
    } else {
      const eqFilter = filters.find(f => f.op === 'eq')
      if (eqFilter) {
        valueStr = String(eqFilter.value)
      } else {
        const minFilter = filters.find(f => f.op === 'gte')
        const maxFilter = filters.find(f => f.op === 'lte')
        if (minFilter && maxFilter) {
          valueStr = `${minFilter.value} - ${maxFilter.value}`
        } else if (minFilter) {
          valueStr = `>= ${minFilter.value}`
        } else if (maxFilter) {
          valueStr = `<= ${maxFilter.value}`
        }
      }
    }
    chips.push({id: field, label, valueStr})
  }
  return chips
})

// Reset expanded state when schema changes
watch(() => props.schema?.id, () => {
  expanded.value = false
})
</script>

<template>
  <div v-if="schema && hasFilterableAttrs" class="power-filter-bar">
    
    <!-- Active Filter Chips Row -->
    <div class="active-filters-row">
      <div class="active-filters-left">
        <button class="expand-btn" @click="expanded = !expanded">
          <svg class="filter-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3"></polygon>
          </svg>
          Power Filters
          <svg :class="['chevron-icon', {rotated: expanded}]" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <path d="M9 18l6-6-6-6"/>
          </svg>
        </button>

        <div v-if="activeChips.length > 0" class="chips-container">
          <div 
            v-for="chip in activeChips" 
            :key="chip.id"
            class="filter-chip"
            @click="clearFilter(chip.id)"
          >
            <span class="chip-label">{{ chip.label }}:</span>
            <span class="chip-value">{{ chip.valueStr }}</span>
            <svg class="chip-close" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </div>
          <button class="clear-all-btn" @click="emit('clear')">Clear All</button>
        </div>
        <div v-else class="no-filters-hint">
          No filters active
        </div>
      </div>
    </div>

    <!-- Facets Area (Bento Layout) -->
    <div v-if="expanded" class="facets-area">
      <!-- Enum facets -->
      <div v-for="attr in enumAttrs" :key="attr.name" class="facet-bento">
        <div class="facet-title">{{ attrLabel(attr) }}</div>
        <div class="facet-content enum-pills">
          <button
            v-for="opt in attr.options"
            :key="opt"
            :class="['bento-pill', {active: isEnumSelected(attr.name, opt)}]"
            @click="toggleEnum(attr.name, opt)"
          >{{ opt }}</button>
        </div>
      </div>

      <!-- Boolean facets (Segmented Control) -->
      <div v-for="attr in boolAttrs" :key="attr.name" class="facet-bento">
        <div class="facet-title">{{ attrLabel(attr) }}</div>
        <div class="facet-content segmented-control">
          <button 
            :class="['segment-btn', {'active': boolState(attr.name) === true}]" 
            @click="setBool(attr.name, true)"
          >Yes</button>
          <button 
            :class="['segment-btn', {'active': boolState(attr.name) === null}]" 
            @click="setBool(attr.name, null)"
          >All</button>
          <button 
            :class="['segment-btn', {'active': boolState(attr.name) === false}]" 
            @click="setBool(attr.name, false)"
          >No</button>
        </div>
      </div>

      <!-- Number range facets -->
      <div v-for="attr in numberAttrs" :key="attr.name" class="facet-bento">
        <div class="facet-title">{{ attrLabel(attr) }}</div>
        <div class="facet-content range-inputs">
          <input
            type="number"
            class="bento-input"
            placeholder="Min"
            :value="numberValue(attr.name, 'gte')"
            @input="onNumberInput(attr.name, 'gte', ($event.target as HTMLInputElement).value)"
          />
          <span class="range-dash">-</span>
          <input
            type="number"
            class="bento-input"
            placeholder="Max"
            :value="numberValue(attr.name, 'lte')"
            @input="onNumberInput(attr.name, 'lte', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <!-- Purchase Price -->
      <div class="facet-bento">
        <div class="facet-title">Purchase Price</div>
        <div class="facet-content range-inputs">
          <input
            type="number"
            class="bento-input"
            placeholder="Min"
            :value="numberValue('purchasePrice', 'gte')"
            @input="onNumberInput('purchasePrice', 'gte', ($event.target as HTMLInputElement).value)"
          />
          <span class="range-dash">-</span>
          <input
            type="number"
            class="bento-input"
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
.power-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  border-bottom: 1px solid var(--border-primary);
  padding-bottom: 16px;
  margin-bottom: 16px;
}

/* Active Filters Row */
.active-filters-row {
  display: flex;
  align-items: center;
  min-height: 32px;
}

.active-filters-left {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.expand-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  padding: 6px 10px;
  color: var(--text-primary);
  font-family: 'Outfit', sans-serif;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.expand-btn:hover {
  background: var(--bg-secondary);
  border-color: var(--accent-blue);
}
.filter-icon {
  color: var(--text-muted);
}
.chevron-icon {
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}
.chevron-icon.rotated {
  transform: rotate(90deg);
}

.no-filters-hint {
  font-size: 13px;
  color: var(--text-muted);
  font-style: italic;
}

.chips-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 16px;
  background: var(--bg-hover);
  border: 1px solid var(--border-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.filter-chip:hover {
  border-color: var(--error-text);
  background: var(--error-bg);
}
.chip-label {
  color: var(--text-muted);
  margin-right: 4px;
}
.chip-value {
  color: var(--text-primary);
  font-weight: 500;
}
.chip-close {
  margin-left: 6px;
  color: var(--text-muted);
}
.filter-chip:hover .chip-close {
  color: var(--error-text);
}

.clear-all-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
}
.clear-all-btn:hover {
  color: var(--text-primary);
}

/* Facets Area */
.facets-area {
  display: flex;
  gap: 20px;
  overflow-x: auto;
  padding-bottom: 8px;
}
.facets-area::-webkit-scrollbar {
  height: 4px;
}

.facet-bento {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-w: 160px;
}

.facet-title {
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.facet-content {
  display: flex;
}

/* Enums */
.enum-pills {
  flex-wrap: wrap;
  gap: 6px;
}
.bento-pill {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.bento-pill:hover {
  border-color: var(--accent-blue-light);
  color: var(--text-primary);
}
.bento-pill.active {
  background: var(--accent-blue);
  border-color: var(--accent-blue);
  color: var(--text-on-accent);
  box-shadow: var(--shadow-sm);
}

/* Segmented Control */
.segmented-control {
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  padding: 2px;
  border: 1px solid var(--border-primary);
  display: inline-flex;
}
.segment-btn {
  padding: 4px 12px;
  font-size: 12px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  border-radius: 4px;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.segment-btn:hover {
  color: var(--text-primary);
}
.segment-btn.active {
  background: var(--bg-primary);
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
  font-weight: 500;
}

/* Ranges */
.range-inputs {
  align-items: center;
  gap: 6px;
}
.bento-input {
  width: 80px;
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  outline: none;
  transition: border-color var(--transition-fast);
}
.bento-input:focus {
  border-color: var(--accent-blue);
  background: var(--bg-primary);
}
.range-dash {
  color: var(--text-muted);
}
</style>
