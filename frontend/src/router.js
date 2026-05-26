import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth.js'

import CatalogView from './views/CatalogView.vue'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import MyOrdersView from './views/MyOrdersView.vue'
import OrdersView from './views/OrdersView.vue'
import SalesView from './views/SalesView.vue'
import ReceiptsView from './views/ReceiptsView.vue'
import StockView from './views/StockView.vue'
import ReportsView from './views/ReportsView.vue'
import AdminView from './views/AdminView.vue'
import DashboardView from './views/DashboardView.vue'
import ProductView from './views/ProductView.vue'

const routes = [
  { path: '/', redirect: '/catalog' },
  { path: '/catalog', component: CatalogView, meta: { title: 'Каталог', public: true } },
  { path: '/catalog/:id', component: ProductView, meta: { title: 'Товар', public: true } },
  { path: '/login', component: LoginView, meta: { title: 'Вход', public: true } },
  { path: '/register', component: RegisterView, meta: { title: 'Регистрация', public: true } },

  { path: '/my-orders', component: MyOrdersView, meta: { title: 'Мои заказы', requiresCustomer: true } },

  { path: '/orders', component: OrdersView, meta: { title: 'Заказы', requiresRole: ['seller','administrator'] } },
  { path: '/sales', component: SalesView, meta: { title: 'Продажи', requiresRole: ['seller','administrator'] } },

  { path: '/receipts', component: ReceiptsView, meta: { title: 'Поступления', requiresRole: ['storekeeper','administrator'] } },
  { path: '/stock', component: StockView, meta: { title: 'Остатки', requiresRole: ['storekeeper','seller','administrator'] } },

  { path: '/reports', component: ReportsView, meta: { title: 'Отчёты', requiresRole: ['seller','storekeeper','administrator'] } },
  { path: '/admin', component: AdminView, meta: { title: 'Админ', requiresRole: ['administrator'] } },
  { path: '/dashboard', component: DashboardView, meta: { title: 'Дашборд', requiresRole: ['administrator'] } },

  { path: '/:pathMatch(.*)*', redirect: '/catalog' }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  const page = to.meta.title || 'Достык'
  document.title = page === 'Достык' ? 'Достык' : `${page} · Достык`

  if (to.meta.public) {
    // В этой вкладке нет сессии — показываем гостя (старый localStorage не используем).
    const hasTabSession = typeof sessionStorage !== 'undefined' && sessionStorage.getItem('dd_token')
    if (hasTabSession) {
      if (!auth.token) auth.restore()
    } else {
      auth.enterAsGuest()
    }
  } else {
    const needsAuth = to.meta.requiresCustomer || to.meta.requiresRole
    if (needsAuth && !auth.token) auth.restore()
  }

  if (to.meta.requiresCustomer && !auth.isCustomer) return next('/login')
  if (to.meta.requiresRole) {
    if (!auth.isUser) return next('/login')
    if (!to.meta.requiresRole.includes(auth.user.role)) return next('/catalog')
  }
  next()
})

export default router
