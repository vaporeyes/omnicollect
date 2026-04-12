<!-- ABOUTME: Global toast notification overlay. -->
<!-- ABOUTME: Renders sliding toasts in the bottom-right corner with auto-dismiss. -->
<script lang="ts" setup>
import {useToastStore} from '../stores/toastStore'

const toastStore = useToastStore()
</script>

<template>
  <div class="toast-container" aria-live="polite">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toastStore.toasts"
        :key="toast.id"
        :class="['toast', 'toast-' + toast.type]"
        @click="toastStore.dismiss(toast.id)"
      >
        <span class="toast-icon" v-if="toast.type === 'success'">&#10003;</span>
        <span class="toast-icon" v-else-if="toast.type === 'error'">&#10007;</span>
        <span class="toast-icon" v-else>&#9432;</span>
        <span class="toast-message">{{ toast.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column-reverse;
  gap: 8px;
  pointer-events: none;
  max-width: 380px;
}
.toast {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: var(--radius-md, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  pointer-events: auto;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
}
.toast-icon {
  flex-shrink: 0;
  font-size: 15px;
  line-height: 1;
}
.toast-message {
  flex: 1;
  line-height: 1.4;
}

.toast-success {
  background: var(--success-bg, rgba(34, 154, 22, 0.15));
  color: var(--success-text, #22c55e);
  border-left: 3px solid var(--success-text, #22c55e);
}
.toast-error {
  background: var(--error-bg, rgba(239, 68, 68, 0.15));
  color: var(--error-text, #ef4444);
  border-left: 3px solid var(--error-border, #ef4444);
}
.toast-info {
  background: var(--bg-secondary, rgba(100, 116, 139, 0.15));
  color: var(--text-primary, #e2e8f0);
  border-left: 3px solid var(--accent-blue, #3b82f6);
}

/* Slide-in from right animation */
.toast-enter-active {
  transition: transform 0.3s ease-out, opacity 0.3s ease-out;
}
.toast-leave-active {
  transition: transform 0.2s ease-in, opacity 0.2s ease-in;
}
.toast-enter-from {
  transform: translateX(100%);
  opacity: 0;
}
.toast-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
