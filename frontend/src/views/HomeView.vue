<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import RequestCard from '../components/RequestCard.vue'
import { useUserStore } from '../stores/user'
import { toast } from '../composables/toast'

const router = useRouter()
const store = useUserStore()

const requests = ref([])
const loading = ref(true)
const tab = ref('open')
const page = ref(1)
const total = ref(0)
const pageSize = 10

const tabs = [
  { key: 'open', label: '悬赏中' },
  { key: 'settled', label: '已采纳' },
  { key: 'all', label: '全部' }
]

async function load() {
  loading.value = true
  try {
    const data = await api.get('/requests', {
      params: { status: tab.value, page: page.value, page_size: pageSize }
    })
    requests.value = data.list
    total.value = data.total
  } finally {
    loading.value = false
  }
}

function switchTab(k) {
  tab.value = k
  page.value = 1
  load()
}

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

onMounted(load)

// ---- 发布求解读 ----
const showPublish = ref(false)
const myResults = ref([])
const form = ref({ assessment_result_id: '', title: '', content: '', reward: 20 })
const publishing = ref(false)

async function openPublish() {
  if (!store.isLoggedIn) {
    toast('请先登录后发布求解读', 'info')
    router.push({ name: 'login' })
    return
  }
  await store.fetchMe()
  const { results } = await api.get('/my/results')
  myResults.value = results
  if (!results.length) {
    if (confirm('发布求解读前需要先完成一次 MBTI 测评，现在去测评吗？')) {
      router.push('/assess')
    }
    return
  }
  form.value = { assessment_result_id: results[0].id, title: '', content: '', reward: 20 }
  showPublish.value = true
}

async function publish() {
  publishing.value = true
  try {
    await api.post('/requests', {
      ...form.value,
      assessment_result_id: Number(form.value.assessment_result_id),
      reward: Number(form.value.reward)
    })
    toast('已发布，悬赏币已冻结')
    showPublish.value = false
    tab.value = 'open'
    page.value = 1
    await load()
    store.fetchMe()
  } catch (e) {
    toast(e.message, 'err')
  } finally {
    publishing.value = false
  }
}

function onChanged() {
  load()
  store.fetchMe()
}

function startAssess() {
  router.push('/assess')
}
</script>

<template>
  <main class="page">
    <div class="container">
      <!-- Hero -->
      <section class="hero card">
        <div class="hero-text">
          <span class="tag tag-purple">标准化 32 题 · 四维八极倾向分析</span>
          <h1>测出你的人格类型，<br />再借他人的眼睛<span class="grad">读懂自己</span></h1>
          <p>
            每个人都有看不见的自己。完成 MBTI 测评获得专属画像，发布「求解读」悬赏，
            让不同视角的洞察，补上自我认知的最后一块拼图。
          </p>
          <div class="hero-actions">
            <button class="btn btn-primary" @click="startAssess">🧭 开始免费测评</button>
            <button class="btn btn-gold" @click="openPublish">📣 发布求解读</button>
          </div>
          <div class="hero-flow">
            <span>① 注册赠 <b>200</b> 测评币</span>
            <i>→</i>
            <span>② 发帖冻结悬赏</span>
            <i>→</i>
            <span>③ 他人帮解读</span>
            <i>→</i>
            <span>④ 采纳自动结算</span>
          </div>
        </div>
        <div class="hero-art" aria-hidden="true">
          <div class="orbit orbit-1"><span>INTJ</span></div>
          <div class="orbit orbit-2"><span>ENFP</span></div>
          <div class="orbit orbit-3"><span>ISFJ</span></div>
          <div class="mirror">🪞</div>
        </div>
      </section>

      <!-- 求解读列表 -->
      <section style="margin-top: 34px;">
        <div class="section-title">
          <h2>求解读大厅</h2>
          <span class="sub">点击卡片可翻面查看「测评结果+画像」与「他人解读+洞察视角」</span>
        </div>

        <div class="tabs">
          <button v-for="t in tabs" :key="t.key"
                  class="tab" :class="{ on: tab === t.key }"
                  @click="switchTab(t.key)">{{ t.label }}</button>
        </div>

        <div v-if="loading" class="empty"><span class="emoji">⏳</span>加载中…</div>
        <div v-else-if="!requests.length" class="card empty">
          <span class="emoji">🌫️</span>
          这里还没有内容，成为第一个发布求解读的人吧
        </div>
        <div v-else class="req-list">
          <RequestCard v-for="r in requests" :key="r.id" :req="r" @changed="onChanged" />
        </div>

        <div v-if="totalPages > 1" class="pager">
          <button class="btn btn-outline btn-sm" :disabled="page === 1" @click="page--; load()">上一页</button>
          <span>{{ page }} / {{ totalPages }}</span>
          <button class="btn btn-outline btn-sm" :disabled="page === totalPages" @click="page++; load()">下一页</button>
        </div>
      </section>
    </div>

    <!-- 发布求解读弹层 -->
    <div v-if="showPublish" class="modal-mask" @click.self="showPublish = false">
      <div class="modal">
        <h3>📣 发布求解读</h3>
        <div class="field">
          <label>选择测评结果</label>
          <select v-model="form.assessment_result_id" class="input">
            <option v-for="r in myResults" :key="r.id" :value="r.id">
              {{ r.type_code }} · {{ new Date(r.created_at).toLocaleString() }}
            </option>
          </select>
        </div>
        <div class="field">
          <label>标题（你想被解读的困惑）</label>
          <input v-model="form.title" class="input" maxlength="100" placeholder="例如：INFP 在职场总被说太敏感，该改变吗？" />
        </div>
        <div class="field">
          <label>补充描述（选填）</label>
          <textarea v-model="form.content" class="textarea"
            placeholder="说说你的具体情境、困扰，能帮助他人给出更有针对性的洞察"></textarea>
        </div>
        <div class="field">
          <label>悬赏测评币（发布即冻结，采纳后结算给解读人）</label>
          <input v-model.number="form.reward" type="number" min="1" :max="store.coins" class="input" />
          <div class="quick-amounts">
            <button v-for="n in [10, 20, 50, 100]" :key="n" class="qa" @click="form.reward = n">{{ n }} 币</button>
          </div>
          <p class="field-hint">当前可用 {{ store.coins }} 币，冻结中 {{ store.frozen }} 币。无人解读时可关闭并全额退回。</p>
        </div>
        <p v-if="form.reward > store.coins" class="form-error">可用测评币不足，当前只有 {{ store.coins }} 币</p>
        <div class="modal-actions">
          <button class="btn btn-outline" @click="showPublish = false">取消</button>
          <button class="btn btn-primary" :disabled="publishing || !form.title || form.reward > store.coins || form.reward < 1" @click="publish">
            {{ publishing ? '发布中…' : `确认发布并冻结 ${form.reward || 0} 币` }}
          </button>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.hero {
  display: flex; align-items: center; gap: 30px;
  padding: 40px 44px; overflow: hidden;
  background: linear-gradient(120deg, #ffffff 40%, #f3efff 100%);
}
.hero-text { flex: 1.4; }
.hero-text h1 { font-size: 32px; line-height: 1.35; margin: 14px 0 12px; font-weight: 900; }
.grad { background: linear-gradient(135deg, var(--primary), #00b8a9); -webkit-background-clip: text; background-clip: text; color: transparent; }
.hero-text p { color: var(--ink-soft); font-size: 15px; max-width: 520px; }
.hero-actions { display: flex; gap: 12px; margin: 22px 0 18px; }
.hero-flow {
  display: flex; flex-wrap: wrap; align-items: center; gap: 10px;
  font-size: 13px; color: var(--ink-faint);
}
.hero-flow b { color: #b9760b; }
.hero-flow i { color: var(--primary); font-style: normal; }

.hero-art { position: relative; flex: 1; height: 230px; }
.mirror {
  position: absolute; left: 50%; top: 50%; transform: translate(-50%,-50%);
  font-size: 84px; filter: drop-shadow(0 12px 24px rgba(108,92,231,.3));
  animation: floaty 3.4s ease-in-out infinite;
}
@keyframes floaty { 50% { transform: translate(-50%,-58%) } }
.orbit {
  position: absolute; border-radius: 999px; padding: 6px 16px;
  font-weight: 800; font-size: 14px; color: #fff;
  box-shadow: 0 8px 20px rgba(0,0,0,.12);
}
.orbit-1 { background: linear-gradient(135deg,#6366f1,#818cf8); left: 4%; top: 18%; animation: floaty 4s .2s ease-in-out infinite; }
.orbit-2 { background: linear-gradient(135deg,#10b981,#34d399); right: 0; top: 10%; animation: floaty 3s .6s ease-in-out infinite; }
.orbit-3 { background: linear-gradient(135deg,#f59e0b,#fbbf24); left: 12%; bottom: 6%; animation: floaty 3.8s 1s ease-in-out infinite; }

.tabs { display: flex; gap: 8px; margin-bottom: 18px; }
.tab {
  border: none; background: #fff; border-radius: 999px;
  padding: 8px 20px; font-size: 14px; font-weight: 700; color: var(--ink-soft);
  border: 1.5px solid var(--border); transition: .15s;
}
.tab.on { background: var(--primary); color: #fff; border-color: var(--primary); box-shadow: 0 6px 16px rgba(108,92,231,.3); }
.req-list { display: flex; flex-direction: column; gap: 18px; }
.pager { display: flex; align-items: center; justify-content: center; gap: 16px; margin-top: 24px; color: var(--ink-faint); font-size: 14px; }
.quick-amounts { display: flex; gap: 8px; margin-top: 10px; }
.qa {
  border: 1.5px solid var(--border); background: #fff; border-radius: 999px;
  padding: 4px 16px; font-size: 13px; color: var(--ink-soft);
}
.qa:hover { border-color: var(--primary); color: var(--primary); }

@media (max-width: 820px) {
  .hero { flex-direction: column; padding: 30px 24px; }
  .hero-art { width: 100%; height: 170px; }
  .hero-text h1 { font-size: 25px; }
}
</style>
