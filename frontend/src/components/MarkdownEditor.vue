<!-- ABOUTME: CodeMirror-based Markdown editor with a minimal formatting toolbar. -->
<!-- ABOUTME: Replaces plain textarea for schema attributes with widget: "textarea". -->
<script lang="ts" setup>
import {ref} from 'vue'
import {Codemirror} from 'vue-codemirror'
import {markdown} from '@codemirror/lang-markdown'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const extensions = [markdown()]
const cmView = ref<any>(null)

function onReady(payload: any) {
  cmView.value = payload.view
}

// Insert Markdown syntax at cursor or wrap selection
function insertSyntax(before: string, after: string, placeholder: string) {
  const view = cmView.value
  if (!view) return
  const {from, to} = view.state.selection.main
  const selected = view.state.sliceDoc(from, to)
  const text = selected || placeholder
  view.dispatch({
    changes: {from, to, insert: before + text + after},
    selection: {anchor: from + before.length, head: from + before.length + text.length},
  })
  view.focus()
}

function insertLinePrefix(prefix: string, placeholder: string) {
  const view = cmView.value
  if (!view) return
  const {from} = view.state.selection.main
  const line = view.state.doc.lineAt(from)
  const text = line.text.trim() || placeholder
  view.dispatch({
    changes: {from: line.from, to: line.to, insert: prefix + text},
    selection: {anchor: line.from + prefix.length, head: line.from + prefix.length + text.length},
  })
  view.focus()
}

// Strip HTML from pasted content to insert as plain text
function handlePaste(event: ClipboardEvent) {
  const html = event.clipboardData?.getData('text/html')
  if (html) {
    event.preventDefault()
    const plain = event.clipboardData?.getData('text/plain') ?? ''
    const view = cmView.value
    if (!view) return
    const {from, to} = view.state.selection.main
    view.dispatch({changes: {from, to, insert: plain}})
  }
}

function bold() { insertSyntax('**', '**', 'bold text') }
function italic() { insertSyntax('*', '*', 'italic text') }
function heading() { insertLinePrefix('## ', 'Heading') }
function bulletList() { insertLinePrefix('- ', 'List item') }
function numberedList() { insertLinePrefix('1. ', 'List item') }
function link() { insertSyntax('[', '](url)', 'link text') }
</script>

<template>
  <div class="md-editor" @paste="handlePaste">
    <div class="md-toolbar">
      <button type="button" class="tb-btn" @click="bold" title="Bold">
        <strong>B</strong>
      </button>
      <button type="button" class="tb-btn" @click="italic" title="Italic">
        <em>I</em>
      </button>
      <button type="button" class="tb-btn" @click="heading" title="Heading">
        H
      </button>
      <button type="button" class="tb-btn" @click="bulletList" title="Bullet List">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><circle cx="3.5" cy="6" r="1"/><circle cx="3.5" cy="12" r="1"/><circle cx="3.5" cy="18" r="1"/></svg>
      </button>
      <button type="button" class="tb-btn" @click="numberedList" title="Numbered List">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="10" y1="6" x2="21" y2="6"/><line x1="10" y1="12" x2="21" y2="12"/><line x1="10" y1="18" x2="21" y2="18"/><text x="1" y="8" font-size="7" fill="currentColor" stroke="none" font-family="sans-serif">1</text><text x="1" y="14" font-size="7" fill="currentColor" stroke="none" font-family="sans-serif">2</text><text x="1" y="20" font-size="7" fill="currentColor" stroke="none" font-family="sans-serif">3</text></svg>
      </button>
      <button type="button" class="tb-btn" @click="link" title="Link">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>
      </button>
    </div>
    <Codemirror
      :modelValue="modelValue"
      @update:modelValue="val => emit('update:modelValue', val)"
      :extensions="extensions"
      :style="{minHeight: '120px', fontSize: '14px'}"
      @ready="onReady"
    />
  </div>
</template>

<style scoped>
.md-editor {
  border: 1px solid var(--border-input);
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  background: var(--bg-input, #fff);
}
.md-toolbar {
  display: flex;
  gap: 2px;
  padding: 4px 6px;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-secondary);
}
.tb-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 26px;
  border: none;
  border-radius: var(--radius-sm, 4px);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-family: 'Outfit', sans-serif;
  font-size: 13px;
  transition: background 0.1s, color 0.1s;
}
.tb-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
</style>
