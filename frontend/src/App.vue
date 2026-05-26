<script setup>
import { useAuthStore } from './stores/auth.js'
import { useToastStore } from './stores/toast.js'
import { useRouter } from 'vue-router'
import { computed } from 'vue'
import ToastContainer from './components/ToastContainer.vue'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()

function logout() {
  auth.logout()
  toast.push('Вы вышли из системы', 'info')
  router.push('/catalog')
}

const roleLabel = computed(() => {
  if (!auth.isAuth) return 'Гость'
  if (auth.isCustomer) return 'Клиент'
  return { administrator: 'Администратор', seller: 'Продавец', storekeeper: 'Кладовщик' }[auth.user?.role] || ''
})
</script>

<template>
  <div class="layout">
    <header class="topbar">
      <router-link to="/" class="brand">
        <span class="brand-icon">🚪</span>
        <span class="brand-text">Дверной Достык</span>
      </router-link>
      <nav class="nav">
        <router-link to="/catalog">Каталог</router-link>
        <router-link v-if="auth.isCustomer" to="/my-orders">Мои заказы</router-link>
        <router-link v-if="auth.isSeller" to="/orders">Заказы</router-link>
        <router-link v-if="auth.isSeller" to="/sales">Продажи</router-link>
        <router-link v-if="auth.isStoreman" to="/receipts">Поступления</router-link>
        <router-link v-if="auth.isStoreman || auth.isSeller" to="/stock">Остатки</router-link>
        <router-link v-if="auth.isSeller || auth.isStoreman" to="/reports">Отчёты</router-link>
        <router-link v-if="auth.isAdmin" to="/admin">Администрирование</router-link>
        <router-link v-if="auth.isAdmin" to="/dashboard">Дашборд</router-link>
      </nav>
      <div class="user-box">
        <template v-if="auth.isAuth">
          <span class="user-name">{{ auth.displayName }}</span>
          <span class="badge new" style="margin-right:8px">{{ roleLabel }}</span>
          <button class="small secondary" @click="logout">Выйти</button>
        </template>
        <template v-else>
          <span class="badge new" style="margin-right:8px">Гость</span>
          <router-link to="/login"><button class="small">Войти</button></router-link>
          <router-link to="/register" style="margin-left:6px"><button class="small secondary">Регистрация</button></router-link>
        </template>
      </div>
    </header>

    <main class="content">
      <router-view />
    </main>

    <footer class="footer">
      <span>ИС «Дверной Достык» · курсовой проект, ЛГТУ ·</span>
      <a
        class="footer-repo"
        href="https://github.com/Slazzzer/doordostyk-is"
        target="_blank"
        rel="noopener noreferrer"
      >github.com/Slazzzer/doordostyk-is</a>
    </footer>

    <ToastContainer />
  </div>
</template>

<style>
.layout { min-height: 100vh; display: flex; flex-direction: column; }
.topbar {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 55%, #004d40 100%);
  color: white;
  padding: 10px 24px;
  display: flex;
  align-items: center;
  gap: 24px;
  box-shadow: 0 2px 8px rgba(0, 77, 64, 0.25);
  border-bottom: 3px solid var(--accent);
}
.brand { display: flex; align-items: center; gap: 8px; color: white; font-weight: 700; font-size: 17px; }
.brand:hover { text-decoration: none; }
.brand-icon { font-size: 24px; }
.brand-text { user-select: none; }
.nav { display: flex; gap: 4px; flex: 1; flex-wrap: wrap; }
.nav a {
  color: rgba(255,255,255,0.85);
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  transition: background 0.15s, color 0.15s;
}
.nav a:hover { background: rgba(255,255,255,0.15); color: white; text-decoration: none; }
.nav a.router-link-active { background: var(--accent); color: var(--primary-dark); font-weight: 600; }
.user-box { display: flex; align-items: center; gap: 6px; font-size: 13px; }
.user-name { color: rgba(255,255,255,0.95); margin-right: 4px; font-weight: 500; }
.topbar button { user-select: none; }

.content { flex: 1; max-width: 1280px; width: 100%; margin: 0 auto; padding: 24px; }

.footer {
  text-align: center;
  padding: 12px;
  color: var(--text-muted);
  font-size: 12px;
  border-top: 1px solid var(--border);
  background: white;
}
.footer-repo {
  color: var(--primary);
  font-weight: 500;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.footer-repo:hover { color: var(--primary-dark); }
</style>
