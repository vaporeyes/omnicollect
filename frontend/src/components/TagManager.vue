<!-- ABOUTME: Tag management panel listing all tags with counts, inline rename, and delete. -->
<!-- ABOUTME: Emits rename and delete events for the parent to handle via API calls. -->
<script lang="ts" setup>
import {ref} from 'vue'
import type {TagCount} from '../api/types'

const props = defineProps<{
  tags: TagCount[]
}>()

const emit = defineEmits<{
  rename: [payload: {oldName: string, newName: string}]
  delete: [name: string]
  close: []
}>()

const editingTag = ref<string | null>(null)
const editValue = ref('')
const confirmDelete = ref<string | null>(null)

function startEdit(name: string) {
  editingTag.value = name
  editValue.value = name
}

function cancelEdit() {
  editingTag.value = null
  editValue.value = ''
}

function submitEdit() {
  const newName = editValue.value.trim().toLowerCase()
  if (!newName || newName === editingTag.value) {
    cancelEdit()
    return
  }
  emit('rename', {oldName: editingTag.value!, newName})
  editingTag.value = null
  editValue.value = ''
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
  confirmDelete.value = name
}

function cancelDelete() {
  confirmDelete.value = null
}

function executeDelete() {
  if (confirmDelete.value) {
    emit('delete', confirmDelete.value)
    confirmDelete.value = null
  }
}
</script>

<template>
  <div class="tag-manager">
    <div class="tag-manager-header">
      <h3>Manage Tags</h3>
      <button class="close-btn" @click="emit('close')">&times;</button>
    </div>

    <div v-if="tags.length === 0" class="empty-state">
      No tags yet. Add tags to items using the edit form.
    </div>

    <div v-else class="tag-list">
      <div v-for="tag in tags" :key="tag.name" class="tag-row">
        <template v-if="editingTag === tag.name">
          <input
            v-model="editValue"
            class="tag-edit-input"
            maxlength="50"
            @keydown="onEditKeydown"
            ref="editInput"
            autofocus
          />
          <button class="tag-action-btn save-btn" @click="submitEdit">Save</button>
          <button class="tag-action-btn cancel-btn" @click="cancelEdit">Cancel</button>
        </template>
        <template v-else>
          <span class="tag-name" @click="startEdit(tag.name)">{{ tag.name }}</span>
          <span class="tag-item-count">{{ tag.count }} item{{ tag.count === 1 ? '' : 's' }}</span>
          <button class="tag-action-btn rename-btn" @click="startEdit(tag.name)">Rename</button>
          <button class="tag-action-btn delete-btn" @click="requestDelete(tag.name)">Delete</button>
        </template>
      </div>
    </div>

    <!-- Delete confirmation -->
    <Teleport to="body">
      <div v-if="confirmDelete" class="confirm-overlay" @click.self="cancelDelete">
        <div class="confirm-dialog">
          <p class="confirm-title">Delete tag "{{ confirmDelete }}"?</p>
          <p class="confirm-message">This will remove the tag from all items. Items will not be deleted.</p>
          <div class="confirm-actions">
            <button class="confirm-cancel-btn" @click="cancelDelete">Cancel</button>
            <button class="confirm-delete-btn" @click="executeDelete">Delete</button>
          </div>
        </div>
      </div>
    </Teleport>
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
  font-family: 'Instrument Serif', serif;
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
  flex: 1;
  font-family: 'Outfit', sans-serif;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  cursor: pointer;
}
.tag-name:hover {
  text-decoration: underline;
}
.tag-item-count {
  font-family: 'Outfit', sans-serif;
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
  font-family: 'Outfit', sans-serif;
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

/* Confirmation dialog */
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 4000;
}
.confirm-dialog {
  background: var(--bg-primary, #1e1e2e);
  border: 1px solid var(--border-primary, #333);
  border-radius: var(--radius-md);
  padding: 28px;
  max-width: 360px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}
.confirm-title {
  margin: 0 0 4px;
  font-family: 'Instrument Serif', serif;
  font-size: 18px;
  font-weight: 400;
  color: var(--text-primary);
}
.confirm-message {
  margin: 0 0 20px;
  font-size: 13px;
  color: var(--text-secondary);
}
.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.confirm-cancel-btn {
  padding: 8px 18px;
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}
.confirm-cancel-btn:hover {
  background: var(--bg-hover);
}
.confirm-delete-btn {
  padding: 8px 18px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--error-border, #dc2626);
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
}
.confirm-delete-btn:hover {
  background: #b91c1c;
}
</style>
