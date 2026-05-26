<script setup>
import { ref, onMounted } from 'vue'
import api from '../api.js'

const orders = ref([])
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const r = await api.get('/customer/orders/my')
    orders.value = r.data
  } finally { loading.value = false }
}
const fmtDate = d => new Date(d).toLocaleDateString('ru-RU')
const statusClass = s => ({ 'новый':'new', 'выполнен':'done', 'отклонён':'rejected' }[s] || 'new')
onMounted(load)
</script>

<template>
  <h1>Мои заказы</h1>
  <div v-if="loading" class="empty">Загрузка…</div>
  <div v-else-if="orders.length === 0" class="empty">У вас пока нет заказов. <router-link to="/catalog">Перейти в каталог</router-link></div>
  <div v-else class="card" style="padding:0; overflow:hidden">
    <table>
      <thead>
        <tr>
          <th>№</th><th>Дата</th><th>Товар</th><th>Кол-во</th><th>Статус</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in orders" :key="o.order_id">
          <td>#{{ o.order_id }}</td>
          <td>{{ fmtDate(o.order_date) }}</td>
          <td>{{ o.product_name }}</td>
          <td>{{ o.order_quantity }}</td>
          <td><span :class="['badge', statusClass(o.order_status)]">{{ o.order_status }}</span></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
