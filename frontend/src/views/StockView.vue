<script setup>
import { ref, onMounted, computed } from 'vue'
import api from '../api.js'
import { useAuthStore } from '../stores/auth.js'

const auth = useAuthStore()
const items = ref([])
const categories = ref([])
const maxBalance = ref('')
const categoryId = ref('')

async function load() {
  const params = {}
  if (maxBalance.value !== '' && maxBalance.value !== null) params.max_balance = maxBalance.value
  if (categoryId.value) params.category_id = categoryId.value
  const path = auth.isStoreman ? '/storeman/stock' : '/seller/stock'
  const r = await api.get(path, { params })
  items.value = r.data
}

const totalBalance = computed(() => items.value.reduce((s, i) => s + i.balance, 0))
const lowCount = computed(() => items.value.filter(i => i.balance < 5).length)

onMounted(async () => {
  const c = await api.get('/catalog/categories')
  categories.value = c.data
  await load()
})

const fmt = n => n == null ? '-' : new Intl.NumberFormat('ru-RU').format(n) + ' ₽'
</script>

<template>
  <h1>Остатки на складе</h1>

  <div class="card">
    <div class="row">
      <div class="col">
        <label>Категория</label>
        <select v-model="categoryId" @change="load">
          <option value="">Все</option>
          <option v-for="c in categories" :key="c.category_id" :value="c.category_id">{{ c.category_name }}</option>
        </select>
      </div>
      <div class="col">
        <label>Максимальный остаток (фильтр дефицита)</label>
        <input type="number" min="0" v-model="maxBalance" @change="load" @input="load" placeholder="Введите порог, например: 5" />
      </div>
    </div>
  </div>

  <div class="stat-grid mb-4">
    <div class="stat-card"><div class="stat-label">Позиций</div><div class="stat-value">{{ items.length }}</div></div>
    <div class="stat-card"><div class="stat-label">Общий остаток, шт.</div><div class="stat-value">{{ totalBalance }}</div></div>
    <div class="stat-card"><div class="stat-label">Позиций с дефицитом (&lt;5)</div><div class="stat-value">{{ lowCount }}</div></div>
  </div>

  <div class="card" style="padding:0; overflow:hidden">
    <table>
      <thead>
        <tr>
          <th>Категория</th><th>Товар</th><th>Размеры</th><th>Цена</th>
          <th>Приход</th><th>Продажи</th><th>Остаток</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="i in items" :key="i.product_id">
          <td>{{ i.category_name }}</td>
          <td>{{ i.product_name }}</td>
          <td>{{ i.product_dimensions || '-' }}</td>
          <td>{{ fmt(i.product_retail_price) }}</td>
          <td>{{ i.received_qty }}</td>
          <td>{{ i.sold_qty }}</td>
          <td>
            <strong :class="i.balance < 5 ? 'low-balance' : ''">{{ i.balance }}</strong>
            <span v-if="i.balance < 5" class="badge low" style="margin-left:6px">дефицит</span>
          </td>
        </tr>
        <tr v-if="items.length === 0"><td colspan="7" class="empty">Нет позиций</td></tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.low-balance { color: var(--danger); }
</style>
