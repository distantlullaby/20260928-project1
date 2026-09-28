<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { auth } from '../store/auth'
import { toast } from '../composables/toast'
import PostCard from '../components/PostCard.vue'
import CreatePostModal from '../components/CreatePostModal.vue'
import LoginModal from '../components/LoginModal.vue'

const route = useRoute()
const router = useRouter()

const posts = ref([])
const loading = ref(false)
const tab = ref('all') // all | open | settled
const showCreate = ref(false)
const showLogin = ref(false)

onMounted(async () => {
  if (auth.isLoggedIn) {
    try { await auth.fetchMe() } catch { auth.logout() }
  }
  await loadPosts()
  if (route.query.publish === '1') {
    openCreate()
    router.replace({ query: {} })
  }
  if (route.query.login === '1') showLogin.value = true
})

async function loadPosts() {
  loading.value = true
  try {
    const qs = tab.value === 'all' ? '' : `?status=${tab.value}`
    const d = await api.get(`/posts${qs}`)
    posts.value = d.posts
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    loading.value = false
  }
}

function switchTab(t) {
  tab.value = t
  loadPosts()
}

function openCreate() {
  if (!auth.isLoggedIn) {
    showLogin.value = true
    return
  }
  showCreate.value = true
}

function onCreated() {
  tab.value = 'all'
  loadPosts()
}
</script>

<template>
  <div>
    <!-- Hero -->
    <section class="hero">
      <div class="container hero-inner">
        <div class="hero-text">
          <h1>看见自己，<br />也看见<span class="grad">别人眼中的你</span></h1>
          <p>
            完成标准化 MBTI 测评定位人格类型，发布「求解读」悬赏，
            让他人的洞察成为你的另一面镜子。
          </p>
          <div class="hero-actions">
            <button class="btn btn-primary" @click="router.push('/quiz')">🧭 开始测评</button>
            <button class="btn btn-gold" @click="openCreate">🪙 发布求解读</button>
          </div>
          <div class="hero-hint">注册即赠 100 测评币 · 采纳解读后悬赏币自动结算</div>
        </div>
        <div class="hero-cards">
          <div class="float-card c1">
            <div class="fc-type">INFJ</div>
            <div class="fc-label">提倡者</div>
            <div class="fc-bar"><i style="width:72%"></i></div>
          </div>
          <div class="float-card c2">
            <div class="fc-emoji">💬</div>
            <div class="fc-line">“你的直觉很强，<br/>只是常常忽略自己的需要”</div>
          </div>
        </div>
      </div>
    </section>

    <div class="container">
      <!-- 求解读信息流 -->
      <div class="feed-head">
        <h2>求解读广场</h2>
        <div class="tabs">
          <button :class="{ active: tab === 'all' }" @click="switchTab('all')">全部</button>
          <button :class="{ active: tab === 'open' }" @click="switchTab('open')">征集中</button>
          <button :class="{ active: tab === 'settled' }" @click="switchTab('settled')">已采纳</button>
        </div>
      </div>

      <div v-if="loading" class="empty">加载中…</div>
      <div v-else-if="posts.length === 0" class="empty">
        <span class="big">🪐</span>
        这里还空空如也，完成测评后发布第一个「求解读」吧
      </div>
      <PostCard v-else v-for="p in posts" :key="p.id" :post="p" @changed="loadPosts" />
    </div>

    <CreatePostModal v-if="showCreate" @closed="showCreate = false" @created="onCreated" />
    <LoginModal v-if="showLogin" @closed="showLogin = false" @success="showCreate = true" />
  </div>
</template>

<style scoped>
.hero {
  padding: 54px 0 44px;
}
.hero-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 40px;
}
.hero-text { max-width: 600px; }
.hero-text h1 {
  font-size: 40px;
  line-height: 1.3;
  margin-bottom: 16px;
  font-weight: 800;
}
.grad {
  background: linear-gradient(135deg, #b3a6ff, var(--cyan));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.hero-text p {
  color: var(--text-dim);
  font-size: 15px;
  line-height: 1.8;
  margin-bottom: 26px;
}
.hero-actions { display: flex; gap: 12px; margin-bottom: 14px; }
.hero-hint { font-size: 12.5px; color: var(--text-faint); }

.hero-cards {
  position: relative;
  width: 280px;
  height: 240px;
  flex-shrink: 0;
}
.float-card {
  position: absolute;
  border-radius: 18px;
  border: 1px solid var(--line);
  box-shadow: var(--shadow);
  animation: floaty 5s ease-in-out infinite;
}
.c1 {
  width: 190px;
  background: linear-gradient(160deg, rgba(124,108,255,0.25), var(--card));
  padding: 22px;
  top: 0; left: 0;
}
.fc-type { font-size: 34px; font-weight: 800; color: #b3a6ff; letter-spacing: 2px; }
.fc-label { color: var(--text-dim); font-size: 13px; margin: 4px 0 16px; }
.fc-bar { height: 8px; background: var(--bg-soft); border-radius: 99px; overflow: hidden; }
.fc-bar i { display: block; height: 100%; background: linear-gradient(90deg, var(--primary), var(--cyan)); }
.c2 {
  width: 200px;
  background: var(--card);
  padding: 20px;
  right: 0; bottom: 0;
  animation-delay: 1.2s;
}
.fc-emoji { font-size: 26px; margin-bottom: 8px; }
.fc-line { font-size: 13px; color: var(--text-dim); line-height: 1.7; }
@keyframes floaty {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-12px); }
}

.feed-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 10px 0 18px;
}
.feed-head h2 { font-size: 21px; }
.tabs { display: flex; gap: 6px; background: var(--bg-soft); padding: 4px; border-radius: 10px; }
.tabs button {
  border: none;
  background: transparent;
  color: var(--text-dim);
  font-size: 13px;
  font-weight: 600;
  padding: 7px 16px;
  border-radius: 8px;
  transition: all 0.15s;
}
.tabs button.active {
  background: var(--primary);
  color: #fff;
  box-shadow: 0 4px 12px var(--primary-glow);
}

@media (max-width: 760px) {
  .hero-inner { flex-direction: column; align-items: flex-start; }
  .hero-cards { display: none; }
  .hero-text h1 { font-size: 30px; }
}
</style>
