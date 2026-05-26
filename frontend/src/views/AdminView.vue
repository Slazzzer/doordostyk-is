<script setup>
import { ref, onMounted } from 'vue'
import api from '../api.js'
import { useFlash } from '../composables/useFlash.js'
import { formatPhoneInput, isValidPhone, blockNonDigitKey, isStrongPassword } from '../composables/validators.js'

const { showFlash } = useFlash()

const tab = ref('users')

// users
const users = ref([])
const userForm = ref({ full_name: '', login: '', role: 'seller', password: '' })

// customers (admin-only)
const customers = ref([])
const customerForm = ref({ full_name: '', email: '', phone: '+7', password: '' })
const customerEdit = ref(null)

// categories
const categories = ref([])
const catForm = ref({ name: '', description: '' })
const catEdit = ref(null)

// products
const products = ref([])
const prodForm = ref({ category_id: '', name: '', description: '', dimensions: '', purchase_price: '', retail_price: '' })
const prodEdit = ref(null)

// suppliers
const suppliers = ref([])
const supForm = ref({ name: '', address: '', phone: '' })
const supEdit = ref(null)

function onSupPhoneInput(e) {
  supForm.value.phone = formatPhoneInput(e.target.value)
}
function onCustomerPhoneInput(e) {
  customerForm.value.phone = formatPhoneInput(e.target.value)
}

function canDeleteUser(u) {
  return !['admin', 'seller', 'storekeeper'].includes(u.user_login)
}

async function loadAll() {
  const [u, cu, c, p, s] = await Promise.all([
    api.get('/admin/users'),
    api.get('/admin/customers'),
    api.get('/catalog/categories'),
    api.get('/catalog/products'),
    api.get('/admin/suppliers')
  ])
  users.value = u.data; customers.value = cu.data; categories.value = c.data; products.value = p.data; suppliers.value = s.data
}

// users
async function createUser() {
  try {
    if (!isStrongPassword(userForm.value.password)) {
      showFlash('Пароль сотрудника: 8+ символов, A-Z/a-z, цифра и спецсимвол', 'error')
      return
    }
    await api.post('/admin/users', userForm.value)
    userForm.value = { full_name: '', login: '', role: 'seller', password: '' }
    await loadAll(); showFlash('Пользователь создан', 'success')
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}
async function delUser(id) {
  if (!confirm('Удалить пользователя?')) return
  try { await api.delete('/admin/users/' + id); await loadAll(); showFlash('Удалено', 'success') }
  catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

// customers
async function saveCustomer() {
  try {
    if (!isValidPhone(customerForm.value.phone)) {
      showFlash('Телефон клиента: формат +7XXXXXXXXXX', 'error')
      return
    }
    if (!isStrongPassword(customerForm.value.password)) {
      showFlash('Пароль клиента: 8+ символов, A-Z/a-z, цифра и спецсимвол', 'error')
      return
    }
    const payload = {
      full_name: customerForm.value.full_name,
      email: customerForm.value.email,
      phone: customerForm.value.phone,
      password: customerForm.value.password
    }
    if (customerEdit.value) await api.patch('/admin/customers/' + customerEdit.value, payload)
    else await api.post('/admin/customers', payload)
    customerForm.value = { full_name: '', email: '', phone: '+7', password: '' }
    customerEdit.value = null
    await loadAll(); showFlash('Клиент сохранён', 'success')
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}
function editCustomer(c) {
  customerEdit.value = c.customer_id
  customerForm.value = {
    full_name: c.customer_full_name,
    email: c.customer_email || '',
    phone: c.customer_phone_number ? formatPhoneInput(c.customer_phone_number) : '+7',
    password: ''
  }
}
async function delCustomer(id) {
  if (!confirm('Удалить клиента?')) return
  try { await api.delete('/admin/customers/' + id); await loadAll(); showFlash('Клиент удалён', 'success') }
  catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

// categories
async function saveCat() {
  try {
    const payload = { name: catForm.value.name, description: catForm.value.description || null }
    if (catEdit.value) await api.patch('/admin/categories/' + catEdit.value, payload)
    else await api.post('/admin/categories', payload)
    catForm.value = { name: '', description: '' }; catEdit.value = null
    await loadAll(); showFlash('Категория сохранена', 'success')
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}
function editCat(c) { catEdit.value = c.category_id; catForm.value = { name: c.category_name, description: c.category_description || '' } }
async function delCat(id) {
  if (!confirm('Удалить категорию?')) return
  try { await api.delete('/admin/categories/' + id); await loadAll(); showFlash('Удалено', 'success') }
  catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

// products
async function saveProd() {
  try {
    const payload = {
      category_id: Number(prodForm.value.category_id),
      name: prodForm.value.name,
      description: prodForm.value.description || null,
      dimensions: prodForm.value.dimensions || null,
      purchase_price: prodForm.value.purchase_price ? Number(prodForm.value.purchase_price) : null,
      retail_price: prodForm.value.retail_price ? Number(prodForm.value.retail_price) : null
    }
    if (prodEdit.value) await api.patch('/admin/products/' + prodEdit.value, payload)
    else await api.post('/admin/products', payload)
    prodForm.value = { category_id: '', name: '', description: '', dimensions: '', purchase_price: '', retail_price: '' }
    prodEdit.value = null
    await loadAll(); showFlash('Товар сохранён', 'success')
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}
function editProd(p) {
  prodEdit.value = p.product_id
  prodForm.value = {
    category_id: p.category_id, name: p.product_name,
    description: p.product_description || '', dimensions: p.product_dimensions || '',
    purchase_price: p.product_purchase_price || '', retail_price: p.product_retail_price || ''
  }
}
async function delProd(id) {
  if (!confirm('Удалить товар?')) return
  try { await api.delete('/admin/products/' + id); await loadAll(); showFlash('Удалено', 'success') }
  catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

// suppliers
async function saveSup() {
  try {
    const phone = supForm.value.phone?.trim()
    if (phone && !isValidPhone(phone)) {
      showFlash('Телефон поставщика: формат +7XXXXXXXXXX', 'error')
      return
    }
    const payload = { name: supForm.value.name, address: supForm.value.address || null, phone: phone || null }
    if (supEdit.value) await api.patch('/admin/suppliers/' + supEdit.value, payload)
    else await api.post('/admin/suppliers', payload)
    supForm.value = { name: '', address: '', phone: '' }; supEdit.value = null
    await loadAll(); showFlash('Поставщик сохранён', 'success')
  } catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}
function editSup(s) {
  supEdit.value = s.supplier_id
  const ph = s.supplier_phone_number || ''
  supForm.value = {
    name: s.organization_name,
    address: s.supplier_address || '',
    phone: ph ? formatPhoneInput(ph) : '+7'
  }
}
async function delSup(id) {
  if (!confirm('Удалить поставщика?')) return
  try { await api.delete('/admin/suppliers/' + id); await loadAll(); showFlash('Удалено', 'success') }
  catch (e) { showFlash(e.response?.data?.error || 'Ошибка', 'error') }
}

onMounted(loadAll)
</script>

<template>
  <h1>Администрирование</h1>

  <div class="card" style="margin-bottom:16px">
    <div class="row">
      <button :class="tab==='users'      ? '' : 'secondary'" @click="tab='users'">Пользователи</button>
      <button :class="tab==='customers'  ? '' : 'secondary'" @click="tab='customers'">Клиенты</button>
      <button :class="tab==='categories' ? '' : 'secondary'" @click="tab='categories'">Категории</button>
      <button :class="tab==='products'   ? '' : 'secondary'" @click="tab='products'">Товары</button>
      <button :class="tab==='suppliers'  ? '' : 'secondary'" @click="tab='suppliers'">Поставщики</button>
    </div>
  </div>

  <!-- USERS -->
  <template v-if="tab === 'users'">
    <div class="card">
      <h3>Создать сотрудника</h3>
      <form @submit.prevent="createUser">
        <div class="row">
          <div class="col"><label>ФИО</label><input v-model="userForm.full_name" required /></div>
          <div class="col"><label>Логин</label><input v-model="userForm.login" required minlength="3" /></div>
          <div class="col"><label>Роль</label>
            <select v-model="userForm.role" required>
              <option value="administrator">Администратор</option>
              <option value="seller">Продавец</option>
              <option value="storekeeper">Кладовщик</option>
            </select>
          </div>
          <div class="col"><label>Пароль</label><input v-model="userForm.password" type="password" required minlength="8" /></div>
          <div class="form-action"><button type="submit">Создать</button></div>
        </div>
      </form>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr><th>ID</th><th>ФИО</th><th>Логин</th><th>Роль</th><th></th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.user_id">
            <td>#{{ u.user_id }}</td><td>{{ u.user_full_name }}</td>
            <td>{{ u.user_login }}</td><td>{{ u.user_role }}</td>
            <td>
              <button
                v-if="canDeleteUser(u)"
                class="small danger"
                @click="delUser(u.user_id)"
              >Удалить</button>
              <span v-else class="text-muted" style="font-size:12px">нельзя удалить</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>

  <!-- CUSTOMERS -->
  <template v-if="tab === 'customers'">
    <div class="card">
      <h3>{{ customerEdit ? `Редактирование клиента #${customerEdit}` : 'Новый клиент' }}</h3>
      <form @submit.prevent="saveCustomer">
        <div class="row">
          <div class="col" style="flex:2"><label>ФИО</label><input v-model="customerForm.full_name" required /></div>
          <div class="col" style="flex:2"><label>Email</label><input v-model="customerForm.email" type="email" required /></div>
          <div class="col"><label>Телефон</label><input :value="customerForm.phone" maxlength="12" placeholder="Формат: +7XXXXXXXXXX" @input="onCustomerPhoneInput" @keydown="blockNonDigitKey" /></div>
          <div class="col"><label>Пароль</label><input v-model="customerForm.password" type="password" required minlength="8" placeholder="Новый пароль клиента" /></div>
          <div class="form-action" style="gap:6px">
            <button type="submit">{{ customerEdit ? 'Сохранить' : 'Создать' }}</button>
            <button v-if="customerEdit" type="button" class="secondary" @click="customerEdit=null; customerForm={full_name:'',email:'',phone:'+7',password:''}">Отмена</button>
          </div>
        </div>
      </form>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr><th>ID</th><th>ФИО</th><th>Email</th><th>Телефон</th><th></th></tr></thead>
        <tbody>
          <tr v-for="c in customers" :key="c.customer_id">
            <td>#{{ c.customer_id }}</td>
            <td>{{ c.customer_full_name }}</td>
            <td>{{ c.customer_email || '-' }}</td>
            <td>{{ c.customer_phone_number || '-' }}</td>
            <td>
              <button class="small secondary" @click="editCustomer(c)">Изменить</button>
              <button class="small danger" style="margin-left:4px" @click="delCustomer(c.customer_id)">Удалить</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>

  <!-- CATEGORIES -->
  <template v-if="tab === 'categories'">
    <div class="card">
      <h3>{{ catEdit ? `Редактирование категории #${catEdit}` : 'Новая категория' }}</h3>
      <form @submit.prevent="saveCat">
        <div class="row">
          <div class="col"><label>Название</label><input v-model="catForm.name" required /></div>
          <div class="col" style="flex:2"><label>Описание</label><input v-model="catForm.description" /></div>
          <div class="form-action" style="gap:6px">
            <button type="submit">{{ catEdit ? 'Сохранить' : 'Создать' }}</button>
            <button v-if="catEdit" type="button" class="secondary" @click="catEdit=null; catForm={name:'',description:''}">Отмена</button>
          </div>
        </div>
      </form>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr><th>ID</th><th>Название</th><th>Описание</th><th></th></tr></thead>
        <tbody>
          <tr v-for="c in categories" :key="c.category_id">
            <td>#{{ c.category_id }}</td><td>{{ c.category_name }}</td><td>{{ c.category_description || '-' }}</td>
            <td>
              <button class="small secondary" @click="editCat(c)">Изменить</button>
              <button v-if="c.can_delete" class="small danger" style="margin-left:4px" @click="delCat(c.category_id)">Удалить</button>
              <span v-else class="text-muted" style="font-size:12px; margin-left:4px">нельзя удалить</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>

  <!-- PRODUCTS -->
  <template v-if="tab === 'products'">
    <div class="card">
      <h3>{{ prodEdit ? `Редактирование товара #${prodEdit}` : 'Новый товар' }}</h3>
      <form @submit.prevent="saveProd">
        <div class="row">
          <div class="col"><label>Категория</label>
            <select v-model="prodForm.category_id" required>
              <option value="">Не выбрано</option>
              <option v-for="c in categories" :key="c.category_id" :value="c.category_id">{{ c.category_name }}</option>
            </select>
          </div>
          <div class="col" style="flex:2"><label>Название</label><input v-model="prodForm.name" required /></div>
          <div class="col"><label>Размеры</label><input v-model="prodForm.dimensions" placeholder="Например: 2050x860 мм" /></div>
        </div>
        <div class="row">
          <div class="col" style="flex:3"><label>Описание</label><input v-model="prodForm.description" /></div>
          <div class="col"><label>Закуп. цена</label><input type="number" min="0" step="0.01" v-model="prodForm.purchase_price" /></div>
          <div class="col"><label>Розн. цена</label><input type="number" min="0" step="0.01" v-model="prodForm.retail_price" /></div>
          <div class="form-action" style="gap:6px">
            <button type="submit">{{ prodEdit ? 'Сохранить' : 'Создать' }}</button>
            <button v-if="prodEdit" type="button" class="secondary"
              @click="prodEdit=null; prodForm={category_id:'',name:'',description:'',dimensions:'',purchase_price:'',retail_price:''}">Отмена</button>
          </div>
        </div>
      </form>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr><th>ID</th><th>Категория</th><th>Название</th><th>Размеры</th><th>Цена</th><th></th></tr></thead>
        <tbody>
          <tr v-for="p in products" :key="p.product_id">
            <td>#{{ p.product_id }}</td><td>{{ p.category_name }}</td>
            <td>{{ p.product_name }}</td><td>{{ p.product_dimensions || '-' }}</td>
            <td>{{ p.product_retail_price || '-' }}</td>
            <td>
              <button class="small secondary" @click="editProd(p)">Изменить</button>
              <button v-if="p.can_delete" class="small danger" style="margin-left:4px" @click="delProd(p.product_id)">Удалить</button>
              <span v-else class="text-muted" style="font-size:12px; margin-left:4px">нельзя удалить</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>

  <!-- SUPPLIERS -->
  <template v-if="tab === 'suppliers'">
    <div class="card">
      <h3>{{ supEdit ? `Редактирование поставщика #${supEdit}` : 'Новый поставщик' }}</h3>
      <form @submit.prevent="saveSup">
        <div class="row">
          <div class="col"><label>Название организации</label><input v-model="supForm.name" required /></div>
          <div class="col"><label>Адрес</label><input v-model="supForm.address" /></div>
          <div class="col"><label>Телефон (+7XXXXXXXXXX)</label>
            <input :value="supForm.phone || '+7'" maxlength="12" placeholder="Формат: +7XXXXXXXXXX" @input="onSupPhoneInput" @keydown="blockNonDigitKey" />
          </div>
          <div class="form-action" style="gap:6px">
            <button type="submit">{{ supEdit ? 'Сохранить' : 'Создать' }}</button>
            <button v-if="supEdit" type="button" class="secondary"
              @click="supEdit=null; supForm={name:'',address:'',phone:''}">Отмена</button>
          </div>
        </div>
      </form>
    </div>
    <div class="card" style="padding:0; overflow:hidden">
      <table>
        <thead><tr><th>ID</th><th>Организация</th><th>Адрес</th><th>Телефон</th><th></th></tr></thead>
        <tbody>
          <tr v-for="s in suppliers" :key="s.supplier_id">
            <td>#{{ s.supplier_id }}</td><td>{{ s.organization_name }}</td>
            <td>{{ s.supplier_address || '-' }}</td><td>{{ s.supplier_phone_number || '-' }}</td>
            <td>
              <button class="small secondary" @click="editSup(s)">Изменить</button>
              <button v-if="s.can_delete" class="small danger" style="margin-left:4px" @click="delSup(s.supplier_id)">Удалить</button>
              <span v-else class="text-muted" style="font-size:12px; margin-left:4px">нельзя удалить</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </template>
</template>
