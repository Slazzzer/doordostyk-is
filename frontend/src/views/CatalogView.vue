<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import api from '../api.js'
import { getProductIcon } from '../composables/productIcons.js'
import PaginationBar from '../components/PaginationBar.vue'

const categories = ref([])
const products = ref([])
const search = ref('')
const loading = ref(false)
const page = ref(1)
const pageSize = 8
const pagedProducts = computed(() => products.value.slice((page.value - 1) * pageSize, page.value * pageSize))

async function loadProducts() {
  loading.value = true
  try {
    const params = {}
    const q = search.value.trim()
    if (q) params.q = q
    const r = await api.get('/catalog/products', { params })
    products.value = r.data
    page.value = 1
  } finally { loading.value = false }
}

function fmt(n) {
  if (n == null) return '-'
  return new Intl.NumberFormat('ru-RU').format(n) + ' ₽'
}

onMounted(async () => {
  const cats = await api.get('/catalog/categories')
  categories.value = cats.data
  await loadProducts()
})

let t = null
watch(search, () => { clearTimeout(t); t = setTimeout(loadProducts, 300) })
</script>

<template>
  <h1>Каталог дверей</h1>

  <div class="card">
    <label>Поиск по названию или категории</label>
    <div class="search-bar mt-3">
      <input
        v-model="search"
        type="search"
        placeholder="Введите название товара или категории"
        autocomplete="off"
      />
      <button v-if="search" type="button" class="secondary small" @click="search = ''">Сбросить</button>
    </div>
  </div>

  <div v-if="loading" class="empty">Загрузка…</div>
  <div v-else-if="products.length === 0" class="empty">По вашему запросу ничего не найдено</div>
  <div v-else class="grid">
    <template v-for="p in pagedProducts" :key="p.product_id">
      <router-link :to="`/catalog/${p.product_id}`" class="product-card" style="color:inherit; text-decoration:none">
        <img :src="getProductIcon(p)" alt="" class="product-icon" />
        <div class="product-meta">{{ p.category_name }}</div>
        <div class="product-name">{{ p.product_name }}</div>
        <div class="product-meta" v-if="p.product_dimensions">{{ p.product_dimensions }} мм</div>
        <div class="product-meta" v-if="p.product_description">{{ p.product_description }}</div>
        <div class="product-meta">На складе: {{ p.balance ?? 0 }} шт.</div>
        <div class="product-price">{{ fmt(p.product_retail_price) }}</div>
      </router-link>
    </template>
  </div>
  <PaginationBar v-if="!loading" v-model:page="page" :page-size="pageSize" :total="products.length" />
</template>
