<script lang="ts" setup>
import {ref} from 'vue'
import {SelectImageFile, ProcessImage} from '../../wailsjs/go/main/App'

const props = defineProps<{
  images: string[]
}>()

const emit = defineEmits<{
  'update:images': [filenames: string[]]
}>()

const error = ref<string | null>(null)
const processing = ref(false)

async function onAddImage() {
  error.value = null
  try {
    const path = await SelectImageFile()
    if (!path) return // user cancelled

    processing.value = true
    const result = await ProcessImage(path)
    emit('update:images', [...props.images, result.filename])
  } catch (e: any) {
    error.value = e?.message ?? String(e)
  } finally {
    processing.value = false
  }
}

function removeImage(index: number) {
  const updated = [...props.images]
  updated.splice(index, 1)
  emit('update:images', updated)
}
</script>

<template>
  <div class="image-attach">
    <label class="field-label">Images</label>

    <div v-if="images.length > 0" class="image-previews">
      <div v-for="(filename, idx) in images" :key="filename" class="image-preview">
        <img :src="'/thumbnails/' + filename" alt="Attached image" />
        <button type="button" class="remove-btn" @click="removeImage(idx)">x</button>
      </div>
    </div>

    <button type="button" class="btn btn-secondary" :disabled="processing" @click="onAddImage">
      {{ processing ? 'Processing...' : 'Add Image' }}
    </button>

    <div v-if="error" class="field-error">{{ error }}</div>
  </div>
</template>

<style scoped>
.image-attach {
  margin-bottom: 12px;
}
.field-label {
  display: block;
  font-weight: 600;
  margin-bottom: 4px;
  font-size: 14px;
}
.image-previews {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
}
.image-preview {
  position: relative;
  width: 80px;
  height: 80px;
}
.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 4px;
}
.remove-btn {
  position: absolute;
  top: -4px;
  right: -4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: none;
  background: #e53e3e;
  color: white;
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.btn-secondary {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  background: #e2e8f0;
  color: #333;
}
.btn-secondary:hover {
  background: #cbd5e0;
}
.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.field-error {
  color: #e53e3e;
  font-size: 12px;
  margin-top: 4px;
}
</style>
