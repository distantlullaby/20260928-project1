import { createRouter, createWebHistory } from 'vue-router'
import { auth } from '../store/auth'

const routes = [
  { path: '/', name: 'home', component: () => import('../views/HomeView.vue') },
  { path: '/quiz', name: 'quiz', component: () => import('../views/QuizView.vue'), meta: { requiresAuth: true } },
  { path: '/me', name: 'me', component: () => import('../views/ProfileView.vue'), meta: { requiresAuth: true } },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return { name: 'home', query: { login: '1', redirect: to.fullPath } }
  }
})
