<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { auth } from '../store/auth'
import { toast } from '../composables/toast'

const router = useRouter()

const me = ref(null)
const overview = ref(null)
const transactions = ref([])
const loading = ref(true)
const tab = ref('posts')

const DIM_NAMES = { E: '外向', I: '内向', S: '实感', N: '直觉', T: '思考', F: '情感', J: '判断', P: '感知' }

onMounted(load)

async function load() {
  loading.value = true
  try {
    const [m, o, t] = await Promise.all([
      api.get('/me'),
      api.get('/me/overview'),
      api.get('/me/transactions'),
    ])
    me.value = m
    overview.value = o
    transactions.value = t.transactions
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    loading.value = false
  }
}

const TX_LABEL = {
  register_gift: '注册赠送',
  post_freeze: '发帖冻结悬赏',
  bounty_append: '追加悬赏冻结',
  settle_income: '解读被采纳收入',
}
</script>

<template>
  <div class="container profile">
    <div v-if="loading" class="empty">加载中…</div>
    <template v-else-if="me">
      <!-- 顶部用户卡 -->
      <div class="user-card">
        <div class="uc-left">
          <div class="uc-avatar">{{ me.user.nickname?.[0] || '?' }}</div>
          <div>
            <div class="uc-name">{{ me.user.nickname }}
              <span class="uc-username">@{{ me.user.username }}</span>
            </div>
            <div class="uc-since">加入于 {{ new Date(me.user.createdAt).toLocaleDateString('zh-CN') }}</div>
          </div>
        </div>
        <div class="uc-coin">
          <div class="coin-num">🪙 {{ me.user.coinBalance }}</div>
          <div class="coin-label">测评币余额</div>
        </div>
      </div>

      <!-- 统计 -->
      <div class="stats">
        <div class="stat">
          <div class="stat-num">{{ me.stats.assessmentCount }}</div>
          <div class="stat-label">测评次数</div>
        </div>
        <div class="stat">
          <div class="stat-num">{{ me.stats.postCount }}</div>
          <div class="stat-label">发布求解读</div>
        </div>
        <div class="stat">
          <div class="stat-num">{{ me.stats.settledCount }}</div>
          <div class="stat-label">已采纳结算</div>
        </div>
        <div class="stat">
          <div class="stat-num gold">🪙 {{ me.stats.earnedTotal }}</div>
          <div class="stat-label">解读累计收入</div>
        </div>
      </div>

      <!-- 人格类型档案 -->
      <div class="panel archive">
        <h3>🧬 人格类型档案</h3>
        <div v-if="me.latestAssessment" class="archive-body">
          <div class="arc-type">
            <div class="arc-code">{{ me.latestAssessment.typeCode }}</div>
            <div class="arc-name">{{ me.latestAssessment.profile.name }}</div>
            <div class="arc-title">{{ me.latestAssessment.profile.title }}</div>
            <button class="btn btn-ghost btn-sm" @click="router.push('/quiz')">重新测评</button>
          </div>
          <div class="arc-dims">
            <div v-for="d in me.latestAssessment.dimensions" :key="d.positive" class="arc-dim">
              <div class="arc-labels">
                <span :class="{ on: d.dominant === d.positive }">
                  {{ d.positive }} {{ DIM_NAMES[d.positive] }} · {{ d.posPercent }}%
                </span>
                <span :class="{ on: d.dominant === d.negative }">
                  {{ d.negPercent }}% · {{ DIM_NAMES[d.negative] }} {{ d.negative }}
                </span>
              </div>
              <div class="arc-track">
                <div class="arc-pos" :style="{ width: d.posPercent + '%' }"></div>
                <div class="arc-neg" :style="{ width: d.negPercent + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty">
          <span class="big">🧭</span>
          还没有测评记录
          <div style="margin-top:14px">
            <button class="btn btn-primary btn-sm" @click="router.push('/quiz')">去测评</button>
          </div>
        </div>
      </div>

      <!-- 记录区 -->
      <div class="panel records">
        <div class="rec-tabs">
          <button :class="{ active: tab === 'posts' }" @click="tab = 'posts'">
            📮 我发布的 ({{ overview.myPosts.length }})
          </button>
          <button :class="{ active: tab === 'responses' }" @click="tab = 'responses'">
            🙌 我解读的 ({{ overview.myResponses.length }})
          </button>
          <button :class="{ active: tab === 'coins' }" @click="tab = 'coins'">
            🧾 测评币流水 ({{ transactions.length }})
          </button>
        </div>

        <!-- 我发布的 -->
        <div v-if="tab === 'posts'">
          <div v-if="overview.myPosts.length === 0" class="empty">
            <span class="big">📭</span>还没有发布过求解读
          </div>
          <div v-for="p in overview.myPosts" :key="p.id" class="rec-row">
            <span class="tag tag-coin">🪙 {{ p.bounty }}</span>
            <span class="tag" :class="p.status === 'open' ? 'tag-open' : 'tag-settled'">
              {{ p.status === 'open' ? '征集中' : '已采纳' }}
            </span>
            <span class="rec-mid">{{ p.responseCount }} 条解读</span>
            <span class="rec-time">{{ new Date(p.createdAt).toLocaleString('zh-CN') }}</span>
            <button class="btn btn-ghost btn-sm" @click="router.push('/')">去查看</button>
          </div>
        </div>

        <!-- 我解读的 -->
        <div v-if="tab === 'responses'">
          <div v-if="overview.myResponses.length === 0" class="empty">
            <span class="big">🪶</span>还没有帮别人解读过，去广场看看吧
          </div>
          <div v-for="r in overview.myResponses" :key="r.responseId" class="rec-row resp-row">
            <span v-if="r.earned" class="tag tag-settled">⭐ 已采纳 +{{ r.bounty }} 币</span>
            <span v-else-if="r.postStatus === 'open'" class="tag tag-open">待采纳</span>
            <span v-else class="tag">已结束</span>
            <div class="resp-info">
              <div class="resp-to">给 {{ r.author.nickname }} 的解读</div>
              <div class="resp-text">{{ r.content }}</div>
            </div>
            <span class="rec-time">{{ new Date(r.createdAt).toLocaleDateString('zh-CN') }}</span>
          </div>
        </div>

        <!-- 流水 -->
        <div v-if="tab === 'coins'">
          <div v-if="transactions.length === 0" class="empty">暂无流水</div>
          <table v-else class="ledger">
            <thead>
              <tr><th>类型</th><th>变动</th><th>变动后余额</th><th>备注</th><th>时间</th></tr>
            </thead>
            <tbody>
              <tr v-for="t in transactions" :key="t.id">
                <td>{{ TX_LABEL[t.type] || t.type }}</td>
                <td :class="t.change >= 0 ? 'plus' : 'minus'">
                  {{ t.change >= 0 ? '+' : '' }}{{ t.change }}
                </td>
                <td>🪙 {{ t.balanceAfter }}</td>
                <td class="remark">{{ t.remark }}</td>
                <td class="rec-time">{{ new Date(t.createdAt).toLocaleString('zh-CN') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.profile { padding-top: 28px; max-width: 920px; }

.user-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, rgba(124,108,255,0.2), var(--card));
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 26px 30px;
  margin-bottom: 16px;
}
.uc-left { display: flex; align-items: center; gap: 16px; }
.uc-avatar {
  width: 58px; height: 58px;
  border-radius: 50%;
  display: grid; place-items: center;
  font-size: 24px; font-weight: 800;
  background: linear-gradient(135deg, var(--primary), var(--cyan));
}
.uc-name { font-size: 20px; font-weight: 800; }
.uc-username { font-size: 13px; color: var(--text-faint); font-weight: 400; margin-left: 8px; }
.uc-since { font-size: 12.5px; color: var(--text-dim); margin-top: 4px; }
.uc-coin { text-align: right; }
.coin-num { font-size: 28px; font-weight: 800; color: var(--gold); }
.coin-label { font-size: 12px; color: var(--text-dim); margin-top: 2px; }

.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}
.stat {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 18px;
  text-align: center;
}
.stat-num { font-size: 24px; font-weight: 800; }
.stat-num.gold { color: var(--gold); font-size: 20px; }
.stat-label { font-size: 12.5px; color: var(--text-dim); margin-top: 5px; }

.panel {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 24px 26px;
  margin-bottom: 16px;
}
.panel h3 { margin-bottom: 18px; font-size: 16px; }

.archive-body { display: flex; gap: 30px; }
.arc-type {
  width: 168px;
  flex-shrink: 0;
  text-align: center;
  border-right: 1px solid var(--line);
  padding-right: 24px;
}
.arc-code {
  font-size: 40px;
  font-weight: 800;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #b3a6ff, var(--cyan));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.arc-name { font-size: 16px; font-weight: 700; margin: 6px 0 2px; }
.arc-title { font-size: 12.5px; color: var(--text-dim); margin-bottom: 14px; }
.arc-dims { flex: 1; display: flex; flex-direction: column; justify-content: center; gap: 18px; }
.arc-labels {
  display: flex;
  justify-content: space-between;
  font-size: 12.5px;
  color: var(--text-dim);
  margin-bottom: 6px;
}
.arc-labels .on { color: var(--text); font-weight: 700; }
.arc-track { display: flex; height: 10px; border-radius: 999px; overflow: hidden; background: var(--bg-soft); }
.arc-neg { background: linear-gradient(90deg, #3a3f7a, var(--cyan)); }
.arc-pos { background: linear-gradient(90deg, var(--primary), #b3a6ff); }

.rec-tabs { display: flex; gap: 6px; border-bottom: 1px solid var(--line); margin-bottom: 16px; }
.rec-tabs button {
  border: none;
  background: transparent;
  color: var(--text-dim);
  font-size: 13.5px;
  font-weight: 600;
  padding: 10px 14px;
  position: relative;
}
.rec-tabs button.active { color: var(--text); }
.rec-tabs button.active::after {
  content: '';
  position: absolute;
  left: 10px; right: 10px; bottom: -1px;
  height: 2px;
  background: var(--primary);
  border-radius: 2px;
}

.rec-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 4px;
  border-bottom: 1px solid rgba(46, 50, 104, 0.5);
  font-size: 13.5px;
}
.rec-mid { color: var(--text-dim); }
.rec-time { color: var(--text-faint); font-size: 12px; margin-left: auto; }
.resp-row { align-items: flex-start; }
.resp-info { flex: 1; }
.resp-to { font-size: 12.5px; color: var(--text-dim); margin-bottom: 3px; }
.resp-text {
  font-size: 13px;
  color: var(--text);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.ledger { width: 100%; border-collapse: collapse; font-size: 13px; }
.ledger th {
  text-align: left;
  color: var(--text-faint);
  font-weight: 600;
  font-size: 12px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--line);
}
.ledger td { padding: 11px 10px; border-bottom: 1px solid rgba(46, 50, 104, 0.4); }
.ledger .plus { color: var(--green); font-weight: 700; }
.ledger .minus { color: var(--red); }
.ledger .remark { color: var(--text-dim); }

@media (max-width: 680px) {
  .stats { grid-template-columns: repeat(2, 1fr); }
  .archive-body { flex-direction: column; }
  .arc-type { width: 100%; border-right: none; padding-right: 0; border-bottom: 1px solid var(--line); padding-bottom: 18px; }
  .user-card { flex-direction: column; gap: 18px; align-items: flex-start; }
  .uc-coin { text-align: left; }
  .rec-time { margin-left: 0; }
}
</style>
