<script setup>
import { ref, onMounted } from 'vue'
import api from '../api.js'
import { useFlash } from '../composables/useFlash.js'
import { mapApiError } from '../composables/apiErrors.js'
import { usePagination } from '../composables/usePagination.js'
import PaginationBar from '../components/PaginationBar.vue'

const { showFlash } = useFlash()
const orders = ref([])
const loading = ref(true)
const editingId = ref(null)
const editQty = ref(1)
const { page, pageSize, total, paged } = usePagination(orders, 10)

async function load() {
  loading.value = true
  try {
    const r = await api.get('/customer/orders/my')
    orders.value = r.data
  } finally { loading.value = false }
}

function startEdit(o) {
  editingId.value = o.order_id
  editQty.value = o.order_quantity
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(id) {
  const qty = Number(editQty.value)
  if (!qty || qty < 1) {
    showFlash('Укажите количество от 1', 'error')
    return
  }
  try {
    await api.patch(`/customer/orders/${id}`, { quantity: qty })
    showFlash('Заказ обновлён', 'success')
    editingId.value = null
    await load()
  } catch (e) { showFlash(mapApiError(e, 'Ошибка'), 'error') }
}

async function removeOrder(id) {
  if (!confirm(`Удалить заказ № ${id}?`)) return
  try {
    await api.delete(`/customer/orders/${id}`)
    showFlash('Заказ удалён', 'success')
    await load()
  } catch (e) { showFlash(mapApiError(e, 'Ошибка'), 'error') }
}

const fmtDate = d => new Date(d).toLocaleDateString('ru-RU')
const statusClass = s => ({ 'новый':'new', 'выполнен':'done', 'отклонён':'rejected' }[s] || 'new')
onMounted(load)
</script>

<template>
  <h1>Мои заказы</h1>
  <p class="text-muted" style="margin-bottom:12px">Заказы со статусом «новый» можно изменить или отменить.</p>
  <div v-if="loading" class="empty">Загрузка…</div>
  <div v-else-if="orders.length === 0" class="empty">У вас пока нет заказов. <router-link to="/catalog">Перейти в каталог</router-link></div>
  <template v-else>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead>
          <tr>
            <th>№</th><th>Дата</th><th>Товар</th><th>Кол-во</th><th>Статус</th><th>Действия</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in paged" :key="o.order_id">
            <td>#{{ o.order_id }}</td>
            <td>{{ fmtDate(o.order_date) }}</td>
            <td>{{ o.product_name }}</td>
            <td>
              <template v-if="editingId === o.order_id">
                <input type="number" min="1" max="99999" v-model.number="editQty" style="max-width:90px" />
              </template>
              <template v-else>{{ o.order_quantity }}</template>
            </td>
            <td><span :class="['badge', statusClass(o.order_status)]">{{ o.order_status }}</span></td>
            <td>
              <template v-if="o.order_status === 'новый'">
                <template v-if="editingId === o.order_id">
                  <button class="small success" @click="saveEdit(o.order_id)">Сохранить</button>
                  <button class="small secondary" style="margin-left:4px" @click="cancelEdit">Отмена</button>
                </template>
                <template v-else>
                  <button class="small secondary" @click="startEdit(o)">Изменить</button>
                  <button class="small danger" style="margin-left:4px" @click="removeOrder(o.order_id)">Удалить</button>
                </template>
              </template>
              <span v-else class="text-muted">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <PaginationBar v-model:page="page" :page-size="pageSize" :total="total" />
  </template>
</template>
