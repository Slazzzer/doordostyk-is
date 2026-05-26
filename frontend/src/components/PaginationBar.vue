<script setup>
const props = defineProps({
  page: { type: Number, required: true },
  pageSize: { type: Number, required: true },
  total: { type: Number, required: true }
})

const emit = defineEmits(['update:page'])

function setPage(value) {
  const maxPage = Math.max(1, Math.ceil(props.total / props.pageSize))
  emit('update:page', Math.min(Math.max(1, value), maxPage))
}
</script>

<template>
  <div v-if="total > pageSize" class="pagination-bar">
    <button class="small secondary" :disabled="page <= 1" @click="setPage(page - 1)">Назад</button>
    <span>Страница {{ page }} из {{ Math.ceil(total / pageSize) }}</span>
    <button class="small secondary" :disabled="page >= Math.ceil(total / pageSize)" @click="setPage(page + 1)">Вперёд</button>
  </div>
</template>
