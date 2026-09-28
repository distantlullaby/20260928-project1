<script setup>
import { ref } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { auth } from './store/auth'
import { toastState } from './composables/toast'
import LoginModal from './components/LoginModal.vue'

const router = useRouter()
const showLogin = ref(false)

function logout() {
  auth.logout()
  router.push('/')
}
</script>

<template>
  <header class="navbar">
    <div class="nav-inner">
      <RouterLink to="/" class="brand">
        <span class="brand-logo">🪞</span>
        <span class="brand-name">镜像人格</span>
      </RouterLink>

      <nav class="nav-links">
        <RouterLink to="/" class="nav-link" exact-active-class="active">首页广场</RouterLink>
        <RouterLink to="/quiz" class="nav-link" active-class="active">测评</RouterLink>
        <template v-if="auth.isLoggedIn">
          <RouterLink to="/me" class="nav-link" active-class="active">个人中心</RouterLink>
        </template>
      </nav>

      <div class="nav-right">
        <template v-if="auth.isLoggedIn && auth.user">
          <span class="coin-chip" title="测评币余额">🪙 {{ auth.user.coinBalance }}</span>
          <RouterLink to="/me" class="user-chip">
            <span class="mini-avatar">{{ auth.user.nickname?.[0] || '?' }}</span>
            {{ auth.user.nickname }}
          </RouterLink>
          <button class="btn btn-ghost btn-sm" @click="logout">退出</button>
        </template>
        <template v-else>
          <button class="btn btn-primary btn-sm" @click="showLogin = true">登录 / 注册</button>
        </template>
      </div>
    </div>
  </header>

  <RouterView />

  <LoginModal v-if="showLogin" @closed="showLogin = false" />

  <transition name="toast">
    <div v-if="toastState.visible" class="toast" :class="`toast-${toastState.type}`">
      {{ toastState.message }}
    </div>
  </transition>
</template>

<style scoped>
.navbar {
  position: sticky;
  top: 0;
  z-index: 50;
  background: rgba(15, 16, 36, 0.82);
  backdrop-filter: blur(14px);
  border-bottom: 1px solid var(--line);
}
.nav-inner {
  max-width: 1080px;
  margin: 0 auto;
  height: 60px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  gap: 28px;
}
.brand { display: flex; align-items: center; gap: 8px; }
.brand-logo { font-size: 22px; }
.brand-name { font-weight: 800; font-size: 18px; letter-spacing: 1px; }

.nav-links { display: flex; gap: 4px; flex: 1; }
.nav-link {
  padding: 7px 14px;
  border-radius: 8px;
  font-size: 14px;
  color: var(--text-dim);
  font-weight: 600;
  transition: all 0.15s;
}
.nav-link:hover { color: var(--text); background: var(--bg-soft); }
.nav-link.active { color: var(--text); background: var(--primary-soft); }

.nav-right { display: flex; align-items: center; gap: 10px; }
.coin-chip {
  background: var(--gold-soft);
  color: var(--gold);
  font-size: 13px;
  font-weight: 700;
  padding: 6px 12px;
  border-radius: 999px;
}
.user-chip {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 13.5px;
  font-weight: 600;
}
.mini-avatar {
  width: 26px; height: 26px;
  border-radius: 50%;
  display: grid; place-items: center;
  font-size: 12px;
  background: linear-gradient(135deg, var(--primary), var(--cyan));
}

.toast-enter-active, .toast-leave-active { transition: all 0.25s; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translate(-50%, -10px); }

@media (max-width: 640px) {
  .nav-inner { gap: 10px; }
  .brand-name { display: none; }
  .user-chip { display: none; }
}
</style>
