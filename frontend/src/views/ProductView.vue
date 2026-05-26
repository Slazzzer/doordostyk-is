<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api.js'
import { useAuthStore } from '../stores/auth.js'
import { useFlash } from '../composables/useFlash.js'
import { parsePositiveInt, blockNonDigitKey } from '../composables/validators.js'
import { getProductIcon } from '../composables/productIcons.js'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { showFlash } = useFlash()

const product = ref(null)
const qty = ref(1)
const loading = ref(false)

const stockBalance = computed(() => product.value?.balance ?? 0)
const unitPrice = computed(() => product.value?.product_retail_price ?? 0)
const lineTotal = computed(() => (unitPrice.value || 0) * qty.value)
const overStock = computed(() => qty.value > stockBalance.value)
const canOrder = computed(() => auth.isCustomer && stockBalance.value > 0 && !overStock.value)

async function load() {
  const r = await api.get(`/catalog/products/${route.params.id}`)
  product.value = r.data
  qty.value = 1
}

function onQtyInput(e) {
  const max = stockBalance.value
  if (max <= 0) {
    qty.value = 0
    e.target.value = '0'
    return
  }
  qty.value = parsePositiveInt(e.target.value, 1, Math.min(max, 99999))
  e.target.value = String(qty.value)
}

async function order() {
  if (!canOrder.value || qty.value > stockBalance.value) {
    showFlash('Укажите количество в пределах остатка на складе', 'warn')
    return
  }
  loading.value = true
  try {
    const r = await api.post('/customer/orders', {
      product_id: product.value.product_id,
      quantity: qty.value
    })
    showFlash(`Заказ № ${r.data.order_id} создан и принят в обработку`, 'success')
    await load()
  } catch (e) {
    showFlash(e.response?.data?.error || 'Не удалось создать заказ', 'error')
  } finally {
    loading.value = false
  }
}

const fmt = n => n == null ? '-' : new Intl.NumberFormat('ru-RU').format(n) + ' ₽'

watch(stockBalance, (b) => {
  if (qty.value > b && b > 0) qty.value = b
})

onMounted(load)
</script>

<template>
  <button class="secondary small" @click="router.back()">← Назад</button>
  <div v-if="product" class="card mt-3">
    <div class="row" style="gap:24px; align-items:flex-start">
      <div style="flex:0 0 120px; text-align:center">
        <img :src="getProductIcon(product)" alt="" class="product-detail-icon" />
      </div>
      <div style="flex:1">
        <div class="product-meta">{{ product.category_name }}</div>
        <h1 style="margin:4px 0 8px">{{ product.product_name }}</h1>
        <p v-if="product.product_dimensions" class="text-muted">Размеры: {{ product.product_dimensions }} мм</p>
        <p v-if="product.product_description">{{ product.product_description }}</p>
        <div class="product-price" style="font-size:28px; margin:16px 0">{{ fmt(product.product_retail_price) }}</div>
        <p class="text-muted" style="font-size:13px">На складе: <strong>{{ stockBalance }}</strong> шт.</p>

        <div v-if="auth.isCustomer" style="margin-top:16px">
          <div class="row" style="align-items:flex-end">
            <div style="max-width:140px">
              <label>Количество</label>
              <input
                type="text"
                inputmode="numeric"
                :value="qty"
                maxlength="5"
                @keydown="blockNonDigitKey"
                @input="onQtyInput"
              />
            </div>
            <button @click="order" :disabled="loading || !canOrder">
              {{ loading ? 'Оформление…' : 'Оформить заказ' }}
            </button>
          </div>

          <div class="order-summary">
            <div>Цена за единицу: <strong>{{ fmt(unitPrice) }}</strong></div>
            <div style="margin-top:6px">Сумма заказа: <strong>{{ fmt(lineTotal) }}</strong></div>
          </div>

          <div v-if="overStock" class="alert warn mt-3">
            Запрошено {{ qty }} шт., на складе только {{ stockBalance }} шт. Доступный остаток: {{ stockBalance }} шт.
          </div>
          <div v-else-if="stockBalance === 0" class="alert warn mt-3">
            Товар временно отсутствует на складе.
          </div>
        </div>
        <div v-else-if="!auth.isAuth" class="alert info mt-3">
          Чтобы оформить заказ, <router-link to="/login">войдите как клиент</router-link> или <router-link to="/register">зарегистрируйтесь</router-link>.
        </div>
      </div>
    </div>
  </div>
</template>
