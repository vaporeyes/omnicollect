<!-- ABOUTME: Tag management panel listing all tags with counts, inline rename, and delete. -->
<!-- ABOUTME: Emits rename and delete events for the parent to handle via API calls. -->
<script lang="ts" setup>
import {ref, computed} from 'vue'
import {useEditorGuard} from '../composables/editorGuard'
import type {TagCount} from '../api/types'

const props = defineProps<{
  tags: TagCount[]
  disabled?: boolean
  loadError?: string
}>()

const emit = defineEmits<{
  rename: [payload: {oldName: string, newName: string}]
  delete: [name: string]
  close: []
  reload: []
}>()

const editingTag = ref<string | null>(null)
const editValue = ref('')
const guard = useEditorGuard(computed(() => editingTag.value !== null && editValue.value !== editingTag.value), ref(false))
defineExpose({canLeave: guard.canLeave, cancelEdit})

function startEdit(name: string) {
  if (editingTag.value !== name && !guard.canLeave()) return
  editingTag.value = name
  editValue.value = name
}

function cancelEdit() {
  editingTag.value = null
  editValue.value = ''
}

function submitEdit() {
  if (props.disabled || props.loadError) return
  const newName = editValue.value.trim().toLowerCase()
  if (!newName || newName === editingTag.value) {
    cancelEdit()
    return
  }
  emit('rename', {oldName: editingTag.value!, newName})
}

function onEditKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    submitEdit()
  } else if (e.key === 'Escape') {
    cancelEdit()
  }
}

function requestDelete(name: string) {
  if (!props.disabled && !props.loadError) emit('delete', name)
}
</script>

<template>
  <div class="tag-manager">
    <div class="tag-manager-header">
      <h3>Manage Tags</h3>
      <button class="close-btn" aria-label="Close tag manager" @click="emit('close')">&times;</button>
    </div>

    <p v-if="loadError" role="alert">{{ loadError }} <button @click="emit('reload')">Retry loading tags</button></p>
    <div v-else-if="tags.length === 0" class="empty-state">
      No tags yet. Add tags to items using the edit form.
    </div>

    <div v-else class="tag-list">
      <div v-for="tag in tags" :key="tag.name" class="tag-row">
        <template v-if="editingTag === tag.name">
          <input
            v-model="editValue"
            class="tag-edit-input"
            maxlength="50"
            aria-label="New tag name"
            @keydown.stop="onEditKeydown"
            ref="editInput"
            autofocus
          />
          <button class="tag-action-btn save-btn" :disabled="disabled || !!loadError" @click="submitEdit">Save</button>
          <button class="tag-action-btn cancel-btn" @click="cancelEdit">Cancel</button>
        </template>
        <template v-else>
          <button class="tag-name" @click="startEdit(tag.name)">{{ tag.name }}</button>
          <span class="tag-item-count">{{ tag.count }} item{{ tag.count === 1 ? '' : 's' }}</span>
          <button class="tag-action-btn rename-btn" @click="startEdit(tag.name)">Rename</button>
          <button class="tag-action-btn delete-btn" :disabled="disabled || !!loadError" @click="requestDelete(tag.name)">Delete</button>
        </template>
      </div>
    </div>


  </div>
</template>

<style scoped>
.tag-manager {
  padding: 16px;
  max-width: 600px;
}
.tag-manager-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.tag-manager-header h3 {
  margin: 0;
  font-family: var(--font-heading);
  font-size: 22px;
  font-weight: 400;
  color: var(--text-primary);
}
.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 20px;
  padding: 4px 8px;
}
.close-btn:hover {
  color: var(--text-primary);
}
.empty-state {
  color: var(--text-muted);
  font-size: 14px;
  text-align: center;
  padding: 32px 0;
}
.tag-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.tag-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  transition: background 0.08s;
}
.tag-row:hover {
  background: var(--bg-hover, rgba(255,255,255,0.04));
}
.tag-name {
  background: transparent;
  border: 0;
  text-align: left;
  flex: 1;
  font-family: var(--font-body);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  cursor: pointer;
}
.tag-name:hover {
  text-decoration: underline;
}
.tag-item-count {
  font-family: var(--font-body);
  font-size: 12px;
  color: var(--text-muted);
  margin-right: 8px;
}
.tag-edit-input {
  flex: 1;
  padding: 4px 8px;
  border: 1px solid var(--border-input);
  border-radius: 4px;
  font-size: 14px;
  background: var(--bg-input, transparent);
  color: var(--text-primary);
  box-sizing: border-box;
}
.tag-action-btn {
  padding: 4px 10px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-body);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  transition: background 0.1s;
}
.rename-btn {
  background: transparent;
  color: var(--text-muted);
}
.rename-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
.delete-btn {
  background: transparent;
  color: var(--error-text, #ef4444);
}
.delete-btn:hover {
  background: var(--error-bg, rgba(239,68,68,0.12));
}
.save-btn {
  background: var(--accent-blue);
  color: var(--text-on-accent);
}
.cancel-btn {
  background: transparent;
  color: var(--text-muted);
}

</style>
