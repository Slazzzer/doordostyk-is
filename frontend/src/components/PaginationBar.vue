<script setup>
import { computed } from 'vue'

const props = defineProps({
  page: { type: Number, required: true },
  pageSize: { type: Number, required: true },
  total: { type: Number, required: true }
})

const emit = defineEmits(['update:page'])

const maxPage = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

function setPage(value) {
  emit('update:page', Math.min(Math.max(1, value), maxPage.value))
}
</script>

<template>
  <div v-if="total > pageSize" class="pagination-bar">
    <button class="small secondary" :disabled="page <= 1" title="На первую страницу" @click="setPage(1)">В начало</button>
    <button class="small secondary" :disabled="page <= 1" @click="setPage(page - 1)">Назад</button>
    <span class="pagination-info">Страница {{ page }} из {{ maxPage }}</span>
    <button class="small secondary" :disabled="page >= maxPage" @click="setPage(page + 1)">Вперёд</button>
    <button class="small secondary" :disabled="page >= maxPage" title="На последнюю страницу" @click="setPage(maxPage)">В конец</button>
  </div>
</template>
