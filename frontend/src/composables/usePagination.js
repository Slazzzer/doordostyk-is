import { computed, ref, watch } from 'vue'

/** Пагинация клиентского списка (slice). */
export function usePagination(itemsRef, pageSize = 10) {
  const page = ref(1)
  const total = computed(() => itemsRef.value?.length ?? 0)
  const paged = computed(() => {
    const list = itemsRef.value ?? []
    const start = (page.value - 1) * pageSize
    return list.slice(start, start + pageSize)
  })
  watch(total, () => {
    const maxPage = Math.max(1, Math.ceil(total.value / pageSize))
    if (page.value > maxPage) page.value = maxPage
  })
  function resetPage() {
    page.value = 1
  }
  return { page, pageSize, total, paged, resetPage }
}
