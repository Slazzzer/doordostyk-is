<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import api from '../api.js'
import { useAuthStore } from '../stores/auth.js'
import * as XLSX from 'xlsx'
import jsPDF from 'jspdf'
import autoTable from 'jspdf-autotable'
import PaginationBar from '../components/PaginationBar.vue'
import { applyRobotoFont } from '../utils/jspdfRoboto.js'

const auth = useAuthStore()
const tab = ref('sales')
const page = ref(1)
const pageSize = 10

const fromDate = ref('')
const toDate = ref('')
const categoryId = ref('')
const supplierId = ref('')

const categories = ref([])
const suppliers = ref([])

const salesRows = ref([])
const salesTotalQty = ref(0)
const salesTotalAmount = ref(0)

const receiptsRows = ref([])
const receiptsTotalQty = ref(0)
const receiptsTotalAmount = ref(0)

const loading = ref(false)
const exportFormat = ref('csv')
const canSeeSales = computed(() => auth.isSeller || auth.isAdmin)
const canSeeReceipts = computed(() => auth.isStoreman || auth.isAdmin)
const pagedSalesRows = computed(() => salesRows.value.slice((page.value - 1) * pageSize, page.value * pageSize))
const pagedReceiptsRows = computed(() => receiptsRows.value.slice((page.value - 1) * pageSize, page.value * pageSize))

const fmtDate = d => {
  if (!d) return ''
  const s = String(d).slice(0, 10)
  if (/^\d{4}-\d{2}-\d{2}$/.test(s)) {
    const [y, m, day] = s.split('-')
    return `${day}.${m}.${y}`
  }
  return new Date(d).toLocaleDateString('ru-RU')
}
const fmt = n => new Intl.NumberFormat('ru-RU').format(n) + ' ₽'

async function loadReport() {
  loading.value = true
  try {
    if (tab.value === 'sales') {
      const r = await api.get('/seller/reports/sales', {
        params: { from: fromDate.value || undefined, to: toDate.value || undefined, category_id: categoryId.value || undefined }
      })
      salesRows.value = r.data.rows || []
      salesTotalQty.value = r.data.total_qty || 0
      salesTotalAmount.value = r.data.total_amount || 0
    } else {
      const r = await api.get('/storeman/reports/receipts', {
        params: { from: fromDate.value || undefined, to: toDate.value || undefined,
                  supplier_id: supplierId.value || undefined, category_id: categoryId.value || undefined }
      })
      receiptsRows.value = r.data.rows || []
      receiptsTotalQty.value = r.data.total_qty || 0
      receiptsTotalAmount.value = r.data.total_amount || 0
    }
  } finally { loading.value = false }
}

function exportCSV() {
  let header, rows
  if (tab.value === 'sales') {
    header = ['№', 'Дата', 'Категория', 'Товар', 'Размеры', 'Кол-во', 'Цена', 'Сумма', 'Продавец']
    rows = salesRows.value.map(r => [r.sale_id, fmtDate(r.sale_date), r.category_name, r.product_name,
      r.product_dimensions||'', r.sale_quantity, r.sale_price, r.sale_amount, r.seller_name])
  } else {
    header = ['№', 'Дата', 'Поставщик', 'Категория', 'Товар', 'Кол-во', 'Цена', 'Сумма']
    rows = receiptsRows.value.map(r => [r.receipt_id, fmtDate(r.receipt_date), r.supplier_name, r.category_name,
      r.product_name, r.receipt_quantity, r.receipt_purchase_price, r.receipt_amount])
  }
  const csv = [header, ...rows].map(r => r.map(c => `"${String(c).replace(/"/g, '""')}"`).join(';')).join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `report-${tab.value}-${new Date().toISOString().slice(0,10)}.csv`
  a.click()
}

function exportExcel() {
  let header, rows
  if (tab.value === 'sales') {
    header = ['№', 'Дата', 'Категория', 'Товар', 'Размеры', 'Кол-во', 'Цена', 'Сумма', 'Продавец']
    rows = salesRows.value.map(r => [r.sale_id, fmtDate(r.sale_date), r.category_name, r.product_name,
      r.product_dimensions || '', r.sale_quantity, r.sale_price, r.sale_amount, r.seller_name])
  } else {
    header = ['№', 'Дата', 'Поставщик', 'Категория', 'Товар', 'Кол-во', 'Цена', 'Сумма']
    rows = receiptsRows.value.map(r => [r.receipt_id, fmtDate(r.receipt_date), r.supplier_name, r.category_name,
      r.product_name, r.receipt_quantity, r.receipt_purchase_price, r.receipt_amount])
  }
  const ws = XLSX.utils.aoa_to_sheet([header, ...rows])
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, tab.value === 'sales' ? 'Продажи' : 'Поступления')
  XLSX.writeFile(wb, `report-${tab.value}-${new Date().toISOString().slice(0, 10)}.xlsx`)
}

async function exportPDF() {
  const doc = new jsPDF('l', 'pt', 'a4')
  await applyRobotoFont(doc)
  const isSales = tab.value === 'sales'
  const title = isSales ? 'Отчёт по продажам' : 'Отчёт по поступлениям'
  const subtitle = `Дата формирования: ${new Date().toLocaleString('ru-RU')}`
  const fromText = fromDate.value ? fmtDate(fromDate.value) : 'начало'
  const toText = toDate.value ? fmtDate(toDate.value) : 'текущая дата'
  doc.setFontSize(16)
  doc.text(title, 40, 36)
  doc.setFontSize(10)
  doc.text(subtitle, 40, 54)
  doc.text(`Период: ${fromText} — ${toText}`, 40, 68)

  const head = isSales
    ? [['№', 'Дата', 'Категория', 'Товар', 'Кол-во', 'Цена', 'Сумма', 'Продавец']]
    : [['№', 'Дата', 'Поставщик', 'Категория', 'Товар', 'Кол-во', 'Цена', 'Сумма']]
  const body = isSales
    ? salesRows.value.map(r => [r.sale_id, fmtDate(r.sale_date), r.category_name, r.product_name, r.sale_quantity, fmt(r.sale_price), fmt(r.sale_amount), r.seller_name])
    : receiptsRows.value.map(r => [r.receipt_id, fmtDate(r.receipt_date), r.supplier_name, r.category_name, r.product_name, r.receipt_quantity, fmt(r.receipt_purchase_price), fmt(r.receipt_amount)])

  autoTable(doc, {
    startY: 82,
    head,
    body,
    styles: { font: 'Roboto', fontSize: 9, cellPadding: 4 },
    headStyles: { font: 'Roboto', fillColor: [0, 137, 123], textColor: [255, 255, 255], halign: 'center' },
    bodyStyles: { font: 'Roboto', textColor: [43, 43, 43] },
    alternateRowStyles: { fillColor: [245, 250, 249] },
    margin: { left: 30, right: 30 }
  })

  const totalText = isSales
    ? `Итого: ${salesTotalQty.value} шт., ${fmt(salesTotalAmount.value)}`
    : `Итого: ${receiptsTotalQty.value} шт., ${fmt(receiptsTotalAmount.value)}`
  doc.text(totalText, 40, doc.internal.pageSize.getHeight() - 24)
  doc.save(`report-${tab.value}-${new Date().toISOString().slice(0, 10)}.pdf`)
}

function exportReport() {
  if (exportFormat.value === 'csv') return exportCSV()
  if (exportFormat.value === 'excel') return exportExcel()
  if (exportFormat.value === 'pdf') return exportPDF()
}

onMounted(async () => {
  if (!canSeeSales.value && canSeeReceipts.value) tab.value = 'receipts'
  const cats = await api.get('/catalog/categories')
  categories.value = cats.data
  if (auth.isStoreman || auth.isAdmin) {
    try {
      const sup = await api.get('/suppliers')
      suppliers.value = sup.data
    } catch (e) {}
  }
  await loadReport()
})

let reportTimer = null
watch([tab, fromDate, toDate, categoryId, supplierId], () => {
  page.value = 1
  clearTimeout(reportTimer)
  reportTimer = setTimeout(loadReport, 250)
})
</script>

<template>
  <h1>Отчёты</h1>

  <div class="card">
    <div class="row report-tabs" style="margin-bottom:12px">
      <button :class="['secondary', tab==='sales' && 'active-tab']" @click="tab='sales'"
              v-if="auth.isSeller || auth.isAdmin">Продажи за период</button>
      <button :class="['secondary', tab==='receipts' && 'active-tab']" @click="tab='receipts'"
              v-if="auth.isStoreman || auth.isAdmin">Поступления / поставщики</button>
    </div>
    <div class="report-toolbar">
      <div class="report-fields">
        <div class="report-field"><label>Дата с</label><input type="date" v-model="fromDate" /></div>
        <div class="report-field"><label>Дата по</label><input type="date" v-model="toDate" /></div>
        <div class="report-field">
          <label>Категория</label>
          <select v-model="categoryId">
            <option value="">Все</option>
            <option v-for="c in categories" :key="c.category_id" :value="c.category_id">{{ c.category_name }}</option>
          </select>
        </div>
        <div class="report-field" v-if="tab === 'receipts'">
          <label>Поставщик</label>
          <select v-model="supplierId">
            <option value="">Все</option>
            <option v-for="s in suppliers" :key="s.supplier_id" :value="s.supplier_id">{{ s.organization_name }}</option>
          </select>
        </div>
        <div class="report-field">
          <label>Формат экспорта</label>
          <select v-model="exportFormat">
            <option value="csv">CSV</option>
            <option value="excel">Excel</option>
            <option value="pdf">PDF</option>
          </select>
        </div>
      </div>
      <div class="report-actions">
        <button class="secondary" @click="exportReport">Скачать</button>
      </div>
    </div>
  </div>

  <!-- SALES -->
  <template v-if="tab === 'sales'">
    <div class="stat-grid mb-4">
      <div class="stat-card"><div class="stat-label">Строк отчёта</div><div class="stat-value">{{ salesRows.length }}</div></div>
      <div class="stat-card"><div class="stat-label">Продано, шт.</div><div class="stat-value">{{ salesTotalQty }}</div></div>
      <div class="stat-card"><div class="stat-label">Общая сумма</div><div class="stat-value">{{ fmt(salesTotalAmount) }}</div></div>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr>
          <th>№</th><th>Дата</th><th>Категория</th><th>Товар</th><th>Кол-во</th><th>Цена</th><th>Сумма</th><th>Продавец</th>
        </tr></thead>
        <tbody>
          <tr v-for="r in pagedSalesRows" :key="r.sale_id">
            <td>#{{ r.sale_id }}</td>
            <td>{{ fmtDate(r.sale_date) }}</td>
            <td>{{ r.category_name }}</td>
            <td>{{ r.product_name }}</td>
            <td>{{ r.sale_quantity }}</td>
            <td>{{ fmt(r.sale_price) }}</td>
            <td><strong>{{ fmt(r.sale_amount) }}</strong></td>
            <td>{{ r.seller_name }}</td>
          </tr>
          <tr v-if="!loading && salesRows.length === 0"><td colspan="8" class="empty">Нет данных за период</td></tr>
        </tbody>
      </table>
      <PaginationBar v-model:page="page" :page-size="pageSize" :total="salesRows.length" />
    </div>
  </template>

  <!-- RECEIPTS -->
  <template v-if="tab === 'receipts'">
    <div class="stat-grid mb-4">
      <div class="stat-card"><div class="stat-label">Строк отчёта</div><div class="stat-value">{{ receiptsRows.length }}</div></div>
      <div class="stat-card"><div class="stat-label">Принято, шт.</div><div class="stat-value">{{ receiptsTotalQty }}</div></div>
      <div class="stat-card"><div class="stat-label">Сумма закупок</div><div class="stat-value">{{ fmt(receiptsTotalAmount) }}</div></div>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr>
          <th>№</th><th>Дата</th><th>Поставщик</th><th>Категория</th><th>Товар</th><th>Кол-во</th><th>Цена</th><th>Сумма</th>
        </tr></thead>
        <tbody>
          <tr v-for="r in pagedReceiptsRows" :key="r.receipt_id">
            <td>#{{ r.receipt_id }}</td>
            <td>{{ fmtDate(r.receipt_date) }}</td>
            <td>{{ r.supplier_name }}</td>
            <td>{{ r.category_name }}</td>
            <td>{{ r.product_name }}</td>
            <td>{{ r.receipt_quantity }}</td>
            <td>{{ fmt(r.receipt_purchase_price) }}</td>
            <td><strong>{{ fmt(r.receipt_amount) }}</strong></td>
          </tr>
          <tr v-if="!loading && receiptsRows.length === 0"><td colspan="8" class="empty">Нет данных за период</td></tr>
        </tbody>
      </table>
      <PaginationBar v-model:page="page" :page-size="pageSize" :total="receiptsRows.length" />
    </div>
  </template>
</template>

<style scoped>
.active-tab { background: var(--primary); color: white; }
.report-tabs { user-select: none; }
.report-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: flex-end;
}
.report-fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
  min-width: 0;
}
.report-field {
  min-width: 0;
}
.report-field label { display: block; }
.report-field input,
.report-field select { width: 100%; box-sizing: border-box; }
.report-actions {
  display: flex;
  flex-wrap: nowrap;
  gap: 8px;
  flex-shrink: 0;
  align-items: flex-start;
  padding-top: 24px;
}
.report-actions button { min-height: 38px; white-space: nowrap; }
@media (max-width: 820px) {
  .report-toolbar { grid-template-columns: 1fr; }
  .report-actions { justify-content: flex-end; }
}
</style>
