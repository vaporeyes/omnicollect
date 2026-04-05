<script lang="ts" setup>
import {main} from '../../wailsjs/go/models'

defineProps<{
  modules: main.ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [module: main.ModuleSchema]
  edit: [module: main.ModuleSchema]
}>()
</script>

<template>
  <div class="module-selector">
    <h3>Collection Types</h3>
    <div v-if="modules.length === 0" class="empty-state">
      No collection types available. Add JSON schema files to
      ~/.omnicollect/modules/
    </div>
    <ul v-else class="module-list">
      <li
        v-for="mod in modules"
        :key="mod.id"
        class="module-item"
        @click="emit('select', mod)"
      >
        <span class="module-name-row">
          <strong>{{ mod.displayName }}</strong>
          <button class="edit-btn" @click.stop="emit('edit', mod)" title="Edit schema">&#9998;</button>
        </span>
        <span v-if="mod.description" class="module-desc">{{ mod.description }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.module-selector h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
}
.module-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.module-item {
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 4px;
  margin-bottom: 4px;
}
.module-item:hover {
  background: var(--bg-hover);
}
.module-name-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.module-item strong {
  font-size: 14px;
}
.edit-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-muted);
  padding: 0 2px;
}
.edit-btn:hover {
  color: var(--accent-blue);
}
.module-desc {
  font-size: 12px;
  color: var(--text-secondary);
}
.empty-state {
  font-size: 13px;
  color: var(--text-muted);
  padding: 12px;
}
</style>
