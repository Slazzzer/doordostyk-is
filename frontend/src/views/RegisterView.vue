<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import { useToastStore } from '../stores/toast.js'
import { useFlash } from '../composables/useFlash.js'
import { formatPhoneInput, isValidPhone, blockNonDigitKey, isStrongPassword } from '../composables/validators.js'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const { showFlash } = useFlash()

const form = ref({ full_name: '', email: '', phone: '+7', password: '' })
const phoneError = ref('')
const passwordError = ref('')
const loading = ref(false)

function onPhoneInput(e) {
  form.value.phone = formatPhoneInput(e.target.value)
  phoneError.value = ''
}

function onPhoneKeydown(e) {
  const pos = e.target.selectionStart
  if ((e.key === 'Backspace' || e.key === 'Delete') && pos <= 2) {
    e.preventDefault()
    return
  }
  blockNonDigitKey(e)
}

async function submit() {
  phoneError.value = ''
  passwordError.value = ''
  if (!isValidPhone(form.value.phone)) {
    phoneError.value = 'Введите телефон в формате +7XXXXXXXXXX (10 цифр после +7)'
    return
  }
  if (!isStrongPassword(form.value.password)) {
    passwordError.value = 'Пароль: от 8 символов, латиница A-Z/a-z, цифра и спецсимвол'
    return
  }
  if (form.value.full_name.trim().length < 3) {
    showFlash('Укажите ФИО (не менее 3 символов)', 'error')
    return
  }
  loading.value = true
  try {
    await auth.register({
      full_name: form.value.full_name.trim(),
      email: form.value.email.trim(),
      phone: form.value.phone,
      password: form.value.password
    })
    toast.push('Вы зарегистрированы и вошли в систему', 'success')
    router.push('/catalog')
  } catch (e) {
    showFlash(e.response?.data?.error || 'Ошибка регистрации', 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page-wrap">
    <div class="card" style="max-width:480px; width:100%">
      <h2>Регистрация клиента</h2>
      <form @submit.prevent="submit">
        <div class="mb-3">
          <label>ФИО</label>
          <input v-model="form.full_name" required minlength="3" maxlength="150" placeholder="Введите ФИО полностью" />
        </div>
        <div class="mb-3">
          <label>Email</label>
          <input v-model="form.email" type="email" required maxlength="100" autocomplete="email" />
        </div>
        <div class="mb-3">
          <label>Телефон (+7XXXXXXXXXX)</label>
          <input
            :value="form.phone"
            type="tel"
            required
            maxlength="12"
            placeholder="Формат: +7XXXXXXXXXX"
            @input="onPhoneInput"
            @keydown="onPhoneKeydown"
          />
          <span v-if="phoneError" class="field-error">{{ phoneError }}</span>
        </div>
        <div class="mb-3">
          <label>Пароль (8+ символов, A-Z/a-z, цифра, спецсимвол)</label>
          <input v-model="form.password" type="password" minlength="8" maxlength="72" required autocomplete="new-password" />
          <span v-if="passwordError" class="field-error">{{ passwordError }}</span>
        </div>
        <button type="submit" :disabled="loading" style="width:100%">
          {{ loading ? 'Регистрация…' : 'Зарегистрироваться' }}
        </button>
      </form>
      <p class="text-muted mt-3 text-center" style="font-size:12px">
        Уже есть аккаунт? <router-link to="/login">Войти</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.field-error { display: block; font-size: 12px; color: var(--danger); margin-top: 4px; }
</style>
