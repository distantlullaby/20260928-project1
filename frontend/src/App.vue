<script setup>
import { onMounted } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useUserStore } from './stores/user'
import ToastHost from './components/ToastHost.vue'

const store = useUserStore()
const router = useRouter()

onMounted(() => {
  store.fetchMe()
})

function logout() {
  store.logout()
  router.push('/')
}
</script>

<template>
  <header class="nav">
    <div class="container nav-inner">
      <RouterLink to="/" class="logo">
        <span class="logo-badge">🪞</span>
        <span>镜我<small>&nbsp;MBTI 解读社区</small></span>
      </RouterLink>
      <nav class="nav-links">
        <RouterLink to="/">首页</RouterLink>
        <RouterLink to="/assess">开始测评</RouterLink>
        <RouterLink v-if="store.isLoggedIn" to="/profile">个人中心</RouterLink>
      </nav>
      <span class="nav-spacer"></span>
      <template v-if="store.isLoggedIn">
        <RouterLink to="/profile" class="coin-chip" title="可用测评币">
          <span class="dot">🪙</span>{{ store.coins }}
        </RouterLink>
        <span v-if="store.frozen > 0" class="coin-chip frozen" title="悬赏冻结中">
          ❄ {{ store.frozen }}
        </span>
        <button class="btn btn-outline btn-sm" @click="logout">退出</button>
      </template>
      <template v-else>
        <RouterLink to="/login" class="btn btn-primary btn-sm">登录 / 注册</RouterLink>
      </template>
    </div>
  </header>

  <RouterView v-slot="{ Component }">
    <component :is="Component" />
  </RouterView>

  <ToastHost />

  <footer class="footer">
    镜我 MBTI · 借他人的眼睛，看见更立体的自己 ｜ 测评币循环：注册赠送 · 发布冻结 · 采纳结算
  </footer>
</template>
