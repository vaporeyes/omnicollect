<script lang="ts" setup>
import {Codemirror} from 'vue-codemirror'
import {json} from '@codemirror/lang-json'
import {EditorView} from '@codemirror/view'

defineProps<{
  modelValue: string
  error?: string | null
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const extensions = [json(), EditorView.contentAttributes.of({'aria-label': 'Schema JSON'})]
</script>

<template>
  <div class="code-editor">
    <Codemirror
      :modelValue="modelValue"
      :disabled="disabled"
      @update:modelValue="val => emit('update:modelValue', val)"
      :extensions="extensions"
      :style="{minHeight: '300px', fontSize: '13px'}"
    />
    <div v-if="error" class="parse-error" role="alert">{{ error }}</div>
  </div>
</template>

<style scoped>
.code-editor {
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  overflow: hidden;
}
.parse-error {
  background: var(--error-bg);
  color: var(--error-text);
  padding: 4px 8px;
  font-size: 12px;
  border-top: 1px solid var(--error-border);
}
</style>
