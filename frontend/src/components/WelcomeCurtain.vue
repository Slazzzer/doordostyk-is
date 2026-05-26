<script setup>
import { ref, onMounted } from 'vue'

const emit = defineEmits(['done'])
const phase = ref('idle')

onMounted(() => {
  const showMs = 2600
  const openDurationMs = 1300
  requestAnimationFrame(() => { phase.value = 'show' })
  setTimeout(() => { phase.value = 'opening' }, showMs)
  setTimeout(() => {
    phase.value = 'done'
    emit('done')
  }, showMs + openDurationMs + 300)
})
</script>

<template>
  <div class="welcome-curtain" :class="phase" aria-hidden="true">
    <div class="curtain-panel left" />
    <div class="curtain-panel right" />
    <p class="welcome-text">
      Добро пожаловать в<br />
      <strong>Дверной Достык</strong>
    </p>
  </div>
</template>

<style scoped>
.welcome-curtain {
  position: fixed;
  inset: 0;
  z-index: 10000;
  pointer-events: none;
}
.curtain-panel {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 50%;
  background: linear-gradient(160deg, #004d40 0%, #00695c 45%, #003d33 100%);
  box-shadow: inset 0 0 80px rgba(0, 0, 0, 0.35);
  transition: transform 1.3s cubic-bezier(0.65, 0, 0.35, 1);
}
.curtain-panel.left {
  left: 0;
  border-right: 3px solid rgba(255, 193, 7, 0.35);
}
.curtain-panel.right {
  right: 0;
  border-left: 3px solid rgba(255, 193, 7, 0.35);
}
.welcome-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 2;
  margin: 0;
  text-align: center;
  color: #fff;
  font-size: clamp(1.25rem, 4vw, 2rem);
  line-height: 1.45;
  text-shadow: 0 2px 16px rgba(0, 0, 0, 0.45);
  opacity: 0;
  transition: opacity 0.5s ease;
}
.welcome-text strong {
  color: #ffd54f;
  font-weight: 700;
  font-size: 1.15em;
}
.show .welcome-text {
  opacity: 1;
}
.opening .curtain-panel.left,
.done .curtain-panel.left {
  transform: translateX(-100%);
}
.opening .curtain-panel.right,
.done .curtain-panel.right {
  transform: translateX(100%);
}
.opening .welcome-text,
.done .welcome-text {
  opacity: 0;
}
.done.welcome-curtain {
  visibility: hidden;
}
</style>
