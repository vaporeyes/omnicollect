<!-- ABOUTME: Tag input component with autocomplete, removable chips, and v-model binding. -->
<!-- ABOUTME: Normalizes tags to lowercase, trims whitespace, enforces max 50 chars and no duplicates. -->
<script lang="ts" setup>
import {ref, computed} from 'vue'
import type {TagCount} from '../api/types'

const props = defineProps<{
  modelValue: string[]
  allTags: TagCount[]
}>()

const emit = defineEmits<{
  'update:modelValue': [tags: string[]]
}>()

const inputValue = ref('')
const showDropdown = ref(false)

const filteredSuggestions = computed(() => {
  const q = inputValue.value.trim().toLowerCase()
  if (!q) return []
  return props.allTags
    .filter(t => t.name.includes(q) && !props.modelValue.includes(t.name))
    .slice(0, 10)
})

function addTag(raw: string) {
  const tag = raw.trim().toLowerCase().slice(0, 50)
  if (!tag || props.modelValue.includes(tag)) return
  emit('update:modelValue', [...props.modelValue, tag])
  inputValue.value = ''
  showDropdown.value = false
}

function removeTag(tag: string) {
  emit('update:modelValue', props.modelValue.filter(t => t !== tag))
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    addTag(inputValue.value)
  }
}

function onInput() {
  showDropdown.value = inputValue.value.trim().length > 0
}

function selectSuggestion(name: string) {
  addTag(name)
}

function onBlur() {
  // Delay to allow click on suggestion
  setTimeout(() => { showDropdown.value = false }, 150)
}
</script>

<template>
  <div class="tag-input-wrap">
    <label class="field-label">Tags</label>
    <div class="tag-chips" v-if="modelValue.length > 0">
      <span v-for="tag in modelValue" :key="tag" class="tag-chip">
        {{ tag }}
        <button type="button" class="tag-remove" @click="removeTag(tag)" title="Remove tag">&times;</button>
      </span>
    </div>
    <div class="tag-input-container">
      <input
        type="text"
        v-model="inputValue"
        @keydown="onKeydown"
        @input="onInput"
        @focus="onInput"
        @blur="onBlur"
        placeholder="Add a tag..."
        class="field-input tag-text-input"
        maxlength="50"
      />
      <div v-if="showDropdown && filteredSuggestions.length > 0" class="tag-dropdown">
        <div
          v-for="s in filteredSuggestions"
          :key="s.name"
          class="tag-suggestion"
          @mousedown.prevent="selectSuggestion(s.name)"
        >
          <span class="suggestion-name">{{ s.name }}</span>
          <span class="suggestion-count">{{ s.count }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tag-input-wrap {
  margin-bottom: 12px;
}
.field-label {
  display: block;
  font-weight: 600;
  margin-bottom: 4px;
  font-size: 14px;
}
.tag-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 6px;
}
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  background: var(--accent-blue-light, rgba(59,130,246,0.12));
  color: var(--accent-blue);
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  font-family: 'Outfit', sans-serif;
}
.tag-remove {
  background: none;
  border: none;
  color: var(--accent-blue);
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0 2px;
  opacity: 0.7;
}
.tag-remove:hover {
  opacity: 1;
}
.tag-input-container {
  position: relative;
}
.tag-text-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid var(--border-input);
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}
.tag-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--bg-secondary, #2a2a3e);
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  max-height: 200px;
  overflow-y: auto;
  z-index: 100;
  margin-top: 2px;
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0,0,0,0.2));
}
.tag-suggestion {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.08s;
}
.tag-suggestion:hover {
  background: var(--bg-hover, rgba(255,255,255,0.06));
}
.suggestion-name {
  color: var(--text-primary);
}
.suggestion-count {
  color: var(--text-muted);
  font-size: 11px;
}
</style>
