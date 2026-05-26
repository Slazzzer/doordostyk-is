<script setup>
import { ref, onMounted, computed } from 'vue'
import api from '../api.js'

const data = ref(null)
const loading = ref(true)
const fmt = n => new Intl.NumberFormat('ru-RU').format(n || 0) + ' ₽'

onMounted(async () => {
  try { const r = await api.get('/admin/dashboard'); data.value = r.data }
  finally { loading.value = false }
})

const maxTop = computed(() => Math.max(...(data.value?.top_categories_month?.map(c => c.value) || [1])))
</script>

<template>
  <h1>Дашборд администратора</h1>
  <div v-if="loading" class="empty">Загрузка…</div>
  <div v-else-if="data">
    <div class="stat-grid mb-4">
      <div class="stat-card">
        <div class="stat-label">Выручка сегодня</div>
        <div class="stat-value">{{ fmt(data.sales_today) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Выручка за месяц</div>
        <div class="stat-value">{{ fmt(data.sales_month) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Новых заказов</div>
        <div class="stat-value">{{ data.new_orders }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Позиций с дефицитом</div>
        <div class="stat-value">{{ data.low_stock_count }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Всего товаров в каталоге</div>
        <div class="stat-value">{{ data.products_count }}</div>
      </div>
    </div>

    <div class="card">
      <h3>Топ категорий продаж (текущий месяц)</h3>
      <div v-if="data.top_categories_month?.length">
        <div v-for="c in data.top_categories_month" :key="c.label" class="bar-row">
          <div class="bar-label">{{ c.label }}</div>
          <div class="bar-track">
            <div class="bar-fill" :style="{ width: (c.value / maxTop * 100) + '%' }"></div>
          </div>
          <div class="bar-val">{{ fmt(c.value) }}</div>
        </div>
      </div>
      <div v-else class="empty">Продаж в текущем месяце пока нет</div>
    </div>
  </div>
</template>

<style scoped>
.bar-row { display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid var(--border); }
.bar-row:last-child { border-bottom: 0; }
.bar-label { width: 160px; font-weight: 500; }
.bar-track { flex: 1; height: 22px; background: var(--bg); border-radius: 4px; overflow: hidden; }
.bar-fill { height: 100%; background: linear-gradient(90deg, var(--primary-light), var(--primary)); transition: width 0.3s; }
.bar-val { width: 140px; text-align: right; font-weight: 600; color: var(--primary-dark); }
</style>
