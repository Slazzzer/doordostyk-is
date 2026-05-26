<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import { useToastStore } from '../stores/toast.js'
import { useFlash } from '../composables/useFlash.js'
import { mapApiError } from '../composables/apiErrors.js'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()
const { showFlash } = useFlash()

const mode = ref('user')
const login = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    if (mode.value === 'user') {
      await auth.loginUser(login.value.trim(), password.value)
      toast.push('Вы вошли в систему', 'success')
      router.push('/orders')
    } else {
      await auth.loginCustomer(email.value.trim(), password.value)
      toast.push('Вы вошли в систему', 'success')
      router.push('/my-orders')
    }
  } catch (e) {
    showFlash(mapApiError(e, 'Ошибка входа'), 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrap">
    <div class="card login-card">
      <h2>Вход в систему</h2>

      <div class="tabs">
        <button :class="['tab', mode==='user' && 'active']" @click="mode='user'" type="button">Сотрудник</button>
        <button :class="['tab', mode==='customer' && 'active']" @click="mode='customer'" type="button">Клиент</button>
      </div>

      <form @submit.prevent="submit">
        <template v-if="mode === 'user'">
          <div class="mb-3">
            <label>Логин</label>
            <input v-model="login" required autocomplete="username" placeholder="Введите логин сотрудника" />
          </div>
        </template>
        <template v-else>
          <div class="mb-3">
            <label>Email</label>
            <input v-model="email" type="email" required autocomplete="email" placeholder="Введите email клиента" />
          </div>
        </template>
        <div class="mb-3">
          <label>Пароль</label>
          <input v-model="password" type="password" required autocomplete="current-password" minlength="8" />
        </div>
        <button type="submit" :disabled="loading" style="width:100%">{{ loading ? 'Вход…' : 'Войти' }}</button>
      </form>

      <p class="text-muted mt-3 text-center" style="font-size:12px">
        Нет аккаунта? <router-link to="/register">Зарегистрироваться</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-wrap { display: flex; align-items: center; justify-content: center; min-height: 60vh; }
.login-card { width: 100%; max-width: 400px; }
.tabs { display: flex; gap: 0; margin-bottom: 20px; border-bottom: 2px solid var(--border); }
.tab {
  flex: 1; padding: 8px; border: none; background: transparent; color: var(--text-muted);
  border-bottom: 2px solid transparent; margin-bottom: -2px; border-radius: 0;
}
.tab.active { color: var(--primary-dark); border-bottom-color: var(--accent); font-weight: 600; }
.tab:hover { background: var(--bg); }
</style>
