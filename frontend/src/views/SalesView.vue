<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import api from '../api.js'
import { useFlash } from '../composables/useFlash.js'
import { parseMoneyLimit, parsePositiveInt, blockNonDigitKey, blockNonDigitDecimalKey } from '../composables/validators.js'
import { mapApiError } from '../composables/apiErrors.js'
import PaginationBar from '../components/PaginationBar.vue'

const { showFlash } = useFlash()
const sales = ref([])
const products = ref([])
const form = ref({ product_id: '', quantity: 1, price: '' })
const page = ref(1)
const pageSize = 10
const selectedProduct = computed(() => products.value.find(p => p.product_id === Number(form.value.product_id)))
const selectedBalance = computed(() => selectedProduct.value?.balance ?? 0)
const qtyMax = computed(() => Math.max(1, Math.min(99999, selectedBalance.value || 0)))
const pagedSales = computed(() => sales.value.slice((page.value - 1) * pageSize, page.value * pageSize))

async function load() {
  const r = await api.get('/seller/sales')
  sales.value = r.data
}

async function loadProducts() {
  const r = await api.get('/catalog/products')
  products.value = r.data
}

async function submit() {
  try {
    if (!selectedProduct.value) {
      showFlash('Выберите товар', 'error')
      return
    }
    if (selectedBalance.value < Number(form.value.quantity)) {
      showFlash(`Недостаточно товара на складе: осталось ${selectedBalance.value} шт.`, 'error')
      return
    }
    const payload = {
      product_id: Number(form.value.product_id),
      quantity: Number(form.value.quantity),
      price: form.value.price ? Number(form.value.price) : 0
    }
    const r = await api.post('/seller/sales', payload)
    showFlash(`Продажа № ${r.data.sale_id} оформлена`, 'success')
    form.value = { product_id: '', quantity: 1, price: '' }
    await Promise.all([load(), loadProducts()])
  } catch (e) { showFlash(mapApiError(e, 'Ошибка продажи'), 'error') }
}

function onQtyInput(e) {
  form.value.quantity = parsePositiveInt(e.target.value, 1, qtyMax.value)
  e.target.value = String(form.value.quantity)
}

function onPriceInput(e) {
  form.value.price = parseMoneyLimit(e.target.value)
}

function stockText(p) {
  return `${p.product_name} - ${p.balance ?? 0} шт. на складе`
}

const fmtDate = d => new Date(d).toLocaleDateString('ru-RU')
const fmt = n => new Intl.NumberFormat('ru-RU').format(n) + ' ₽'

onMounted(async () => { await Promise.all([load(), loadProducts()]) })
watch(sales, () => { page.value = 1 })
</script>

<template>
  <h1>Продажи</h1>

  <div class="card">
    <h3>Новая продажа («с полки»)</h3>
    <form @submit.prevent="submit">
      <div class="row">
        <div class="col" style="flex:2">
          <label>Товар</label>
          <select v-model="form.product_id" required>
            <option value="">Выберите товар</option>
            <option v-for="p in products" :key="p.product_id" :value="p.product_id">
              {{ stockText(p) }}
            </option>
          </select>
          <p v-if="form.product_id" class="text-muted" style="font-size:12px; margin-top:6px">
            На складе: <strong>{{ selectedBalance }}</strong> шт.
          </p>
        </div>
        <div class="col">
          <label>Количество (до 99 999)</label>
          <input type="text" inputmode="numeric" :value="form.quantity" required maxlength="5" @keydown="blockNonDigitKey" @input="onQtyInput" />
        </div>
        <div class="col">
          <label>Цена (пусто: из карточки)</label>
          <input type="text" inputmode="decimal" :value="form.price" maxlength="10" @keydown="blockNonDigitDecimalKey" @input="onPriceInput" />
        </div>
        <div class="form-action">
          <button type="submit">Оформить продажу</button>
        </div>
      </div>
    </form>
  </div>

  <h2>История продаж</h2>
  <div class="card" style="padding:0; overflow:hidden">
    <table>
      <thead>
        <tr><th>№</th><th>Дата</th><th>Товар</th><th>Кол-во</th><th>Цена</th><th>Сумма</th><th>Заказ</th></tr>
      </thead>
      <tbody>
        <tr v-for="s in pagedSales" :key="s.sale_id">
          <td>#{{ s.sale_id }}</td>
          <td>{{ fmtDate(s.sale_date) }}</td>
          <td>{{ s.product_name }}</td>
          <td>{{ s.sale_quantity }}</td>
          <td>{{ fmt(s.sale_price) }}</td>
          <td><strong>{{ fmt(s.sale_amount) }}</strong></td>
          <td>{{ s.order_id ? '№ ' + s.order_id : '-' }}</td>
        </tr>
        <tr v-if="sales.length === 0"><td colspan="7" class="empty">Продаж пока нет</td></tr>
      </tbody>
    </table>
    <PaginationBar v-model:page="page" :page-size="pageSize" :total="sales.length" />
  </div>
</template>
