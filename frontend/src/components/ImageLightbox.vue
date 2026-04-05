<script lang="ts" setup>
defineProps<{
  filename: string
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <div v-if="visible" class="lightbox-overlay" @click="emit('close')">
    <div class="lightbox-content" @click.stop>
      <button class="lightbox-close" @click="emit('close')">x</button>
      <img :src="'/originals/' + filename" alt="Full resolution" />
    </div>
  </div>
</template>

<style scoped>
.lightbox-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.lightbox-content {
  position: relative;
  max-width: 90vw;
  max-height: 90vh;
}
.lightbox-content img {
  max-width: 100%;
  max-height: 90vh;
  object-fit: contain;
  border-radius: 4px;
}
.lightbox-close {
  position: absolute;
  top: -12px;
  right: -12px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background: white;
  color: #333;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}
</style>
