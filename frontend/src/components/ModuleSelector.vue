<script lang="ts" setup>
import {main} from '../../wailsjs/go/models'

defineProps<{
  modules: main.ModuleSchema[]
}>()

const emit = defineEmits<{
  select: [module: main.ModuleSchema]
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
        <strong>{{ mod.displayName }}</strong>
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
  background: rgba(0,0,0,0.05);
}
.module-item strong {
  display: block;
  font-size: 14px;
}
.module-desc {
  font-size: 12px;
  color: #666;
}
.empty-state {
  font-size: 13px;
  color: #888;
  padding: 12px;
}
</style>
