<script setup>
import { useToastStore } from '../stores/toast.js'

const toast = useToastStore()
</script>

<template>
  <div class="toast-stack" aria-live="polite">
    <TransitionGroup name="toast">
      <div
        v-for="t in toast.items"
        :key="t.id"
        :class="['toast-item', `toast-${t.type}`]"
      >
        <button class="toast-close" type="button" @click="toast.remove(t.id)" aria-label="Закрыть уведомление">×</button>
        <div class="toast-message">{{ t.message }}</div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  bottom: 16px;
  right: 16px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: min(380px, calc(100vw - 32px));
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  position: relative;
  padding: 12px 40px 12px 14px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.18);
  text-align: left;
}

.toast-message {
  line-height: 1.4;
}

.toast-close {
  position: absolute;
  top: 7px;
  right: 8px;
  width: 24px;
  height: 24px;
  border: 1px solid currentColor;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.5);
  color: inherit;
  font-size: 14px;
  line-height: 1.1;
  font-weight: 700;
  padding: 0;
}

.toast-close:hover {
  background: rgba(255, 255, 255, 0.85);
  transform: scale(1.05);
}

.toast-close:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 1px;
}

.toast-success {
  background: #e8f5e9;
  color: #2e7d32;
  border: 1px solid #a5d6a7;
}

.toast-error {
  background: #ffebee;
  color: #c62828;
  border: 1px solid #ef9a9a;
}

.toast-info {
  background: #e0f7fa;
  color: #0277bd;
  border: 1px solid #80deea;
}

.toast-warn {
  background: #fff8e7;
  color: #bf360c;
  border: 1px solid #ffcc80;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.35s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(24px);
}
</style>
