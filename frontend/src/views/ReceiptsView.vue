<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import api from '../api.js'
import { useFlash } from '../composables/useFlash.js'
import { parseMoneyLimit, parsePositiveInt, blockNonDigitKey, blockNonDigitDecimalKey } from '../composables/validators.js'
import { mapApiError } from '../composables/apiErrors.js'
import PaginationBar from '../components/PaginationBar.vue'

const { showFlash } = useFlash()
const receipts = ref([])
const products = ref([])
const suppliers = ref([])
const form = ref({ supplier_id: '', product_id: '', quantity: 1, purchase_price: '' })
const page = ref(1)
const pageSize = 10
const selectedProduct = computed(() => products.value.find(p => p.product_id === Number(form.value.product_id)))
const selectedBalance = computed(() => selectedProduct.value?.balance ?? 0)
const pagedReceipts = computed(() => receipts.value.slice((page.value - 1) * pageSize, page.value * pageSize))

async function load() {
  const r = await api.get('/storeman/receipts')
  receipts.value = r.data
}

async function submit() {
  try {
    const payload = {
      supplier_id: Number(form.value.supplier_id),
      product_id: Number(form.value.product_id),
      quantity: Number(form.value.quantity),
      purchase_price: Number(form.value.purchase_price)
    }
    const r = await api.post('/storeman/receipts', payload)
    showFlash(`Поступление № ${r.data.receipt_id} оформлено`, 'success')
    form.value = { supplier_id: '', product_id: '', quantity: 1, purchase_price: '' }
    await Promise.all([load(), loadProducts()])
  } catch (e) { showFlash(mapApiError(e, 'Ошибка поступления'), 'error') }
}

function onQtyInput(e) {
  form.value.quantity = parsePositiveInt(e.target.value, 1, 99999)
  e.target.value = String(form.value.quantity)
}

function onPriceInput(e) {
  form.value.purchase_price = parseMoneyLimit(e.target.value)
}

const fmtDate = d => new Date(d).toLocaleDateString('ru-RU')
const fmt = n => new Intl.NumberFormat('ru-RU').format(n) + ' ₽'

async function loadProducts() {
  const r = await api.get('/catalog/products')
  products.value = r.data
}

function stockText(p) {
  return `${p.product_name} - ${p.balance ?? 0} шт. на складе`
}

onMounted(async () => {
  const [r1, r2] = await Promise.all([
    api.get('/catalog/products'),
    api.get('/suppliers')
  ])
  products.value = r1.data
  suppliers.value = r2.data
  await load()
})
watch(receipts, () => { page.value = 1 })
</script>

<template>
  <h1>Поступления на склад</h1>

  <div class="card">
    <h3>Новое поступление</h3>
    <form @submit.prevent="submit">
      <div class="row">
        <div class="col" style="flex:2">
          <label>Поставщик</label>
          <select v-model="form.supplier_id" required>
            <option value="">Выберите поставщика</option>
            <option v-for="s in suppliers" :key="s.supplier_id" :value="s.supplier_id">{{ s.organization_name }}</option>
          </select>
        </div>
        <div class="col" style="flex:2">
          <label>Товар</label>
          <select v-model="form.product_id" required>
            <option value="">Выберите товар</option>
            <option v-for="p in products" :key="p.product_id" :value="p.product_id">
              {{ stockText(p) }}
            </option>
          </select>
          <p v-if="form.product_id" class="text-muted" style="font-size:12px; margin-top:6px">
            На складе сейчас: <strong>{{ selectedBalance }}</strong> шт.
          </p>
        </div>
        <div class="col">
          <label>Количество (до 99 999)</label>
          <input type="text" inputmode="numeric" :value="form.quantity" required maxlength="5" @keydown="blockNonDigitKey" @input="onQtyInput" />
        </div>
        <div class="col">
          <label>Закупочная цена, ₽</label>
          <input type="text" inputmode="decimal" :value="form.purchase_price" required maxlength="10" @keydown="blockNonDigitDecimalKey" @input="onPriceInput" />
        </div>
        <div class="form-action">
          <button type="submit">Принять на склад</button>
        </div>
      </div>
    </form>
  </div>

  <h2>История поступлений</h2>
  <div class="card" style="padding:0; overflow:hidden">
    <table>
      <thead>
        <tr><th>№</th><th>Дата</th><th>Поставщик</th><th>Товар</th><th>Кол-во</th><th>Закуп.</th></tr>
      </thead>
      <tbody>
        <tr v-for="r in pagedReceipts" :key="r.receipt_id">
          <td>#{{ r.receipt_id }}</td>
          <td>{{ fmtDate(r.receipt_date) }}</td>
          <td>{{ r.supplier_name }}</td>
          <td>{{ r.product_name }}</td>
          <td>{{ r.receipt_quantity }}</td>
          <td>{{ fmt(r.receipt_purchase_price) }}</td>
        </tr>
        <tr v-if="receipts.length === 0"><td colspan="6" class="empty">Поступлений нет</td></tr>
      </tbody>
    </table>
    <PaginationBar v-model:page="page" :page-size="pageSize" :total="receipts.length" />
  </div>
</template>
