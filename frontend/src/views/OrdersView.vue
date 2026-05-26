<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '../api.js'
import { useFlash } from '../composables/useFlash.js'
import { usePagination } from '../composables/usePagination.js'
import PaginationBar from '../components/PaginationBar.vue'

const { showFlash } = useFlash()
const orders = ref([])
const status = ref('новый')
const loading = ref(false)
const { page, pageSize, total, paged } = usePagination(orders, 10)

async function load() {
  loading.value = true
  try {
    const r = await api.get('/seller/orders', { params: status.value ? { status: status.value } : {} })
    orders.value = r.data
  } finally { loading.value = false }
}

async function execute(id) {
  if (!confirm(`Выполнить заказ № ${id}?`)) return
  try {
    const r = await api.post(`/seller/orders/${id}/execute`)
    showFlash(`Заказ № ${id} выполнен. Создана продажа № ${r.data.sale_id}.`, 'success')
    await load()
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

async function reject(id) {
  if (!confirm(`Отклонить заказ № ${id}?`)) return
  try {
    await api.post(`/seller/orders/${id}/reject`)
    showFlash(`Заказ № ${id} отклонён`, 'success')
    await load()
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

const fmtDate = d => new Date(d).toLocaleDateString('ru-RU')
const statusClass = s => ({ 'новый':'new', 'выполнен':'done', 'отклонён':'rejected' }[s] || 'new')

onMounted(load)
watch(status, () => { page.value = 1; load() })
</script>

<template>
  <h1>Заказы клиентов</h1>

  <div class="card">
    <label>Фильтр по статусу</label>
    <select v-model="status" style="max-width:240px">
      <option value="">Все</option>
      <option value="новый">новый</option>
      <option value="выполнен">выполнен</option>
      <option value="отклонён">отклонён</option>
    </select>
  </div>

  <div v-if="loading" class="empty">Загрузка…</div>
  <div v-else-if="orders.length === 0" class="empty">Заказы не найдены</div>
  <div v-else class="card" style="padding:0; overflow:hidden">
    <table>
      <thead>
        <tr>
          <th>№</th><th>Дата</th><th>Клиент</th><th>Товар</th><th>Кол-во</th><th>Статус</th><th>Действия</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in paged" :key="o.order_id">
          <td>#{{ o.order_id }}</td>
          <td>{{ fmtDate(o.order_date) }}</td>
          <td>{{ o.customer_name }}</td>
          <td>{{ o.product_name }}</td>
          <td>{{ o.order_quantity }}</td>
          <td><span :class="['badge', statusClass(o.order_status)]">{{ o.order_status }}</span></td>
          <td>
            <template v-if="o.order_status === 'новый'">
              <button class="small success" @click="execute(o.order_id)">Выполнить</button>
              <button class="small danger" style="margin-left:4px" @click="reject(o.order_id)">Отклонить</button>
            </template>
            <template v-else-if="o.sale_id">
              <span class="text-muted">продажа № {{ o.sale_id }}</span>
            </template>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <PaginationBar v-if="orders.length" v-model:page="page" :page-size="pageSize" :total="total" />
</template>
