<script setup>
import { onMounted, ref, computed } from 'vue'
import api from '../api'
import { useUserStore } from '../stores/user'
import TypeBadge from '../components/TypeBadge.vue'
import DimensionBars from '../components/DimensionBars.vue'
import RequestCard from '../components/RequestCard.vue'

const store = useUserStore()

const myResults = ref([])
const myRequests = ref([])
const myInterps = ref([])
const txns = ref([])
const tab = ref('requests')

const latestResult = computed(() => myResults.value[0] || null)

const txnMeta = {
  register: { label: '注册赠送', cls: 'in', icon: '🎁' },
  freeze: { label: '发布冻结', cls: 'out', icon: '🔒' },
  append_freeze: { label: '追加冻结', cls: 'out', icon: '🔒' },
  settle_in: { label: '解读收入', cls: 'in', icon: '💰' },
  settle_out: { label: '采纳结算', cls: 'neutral', icon: '✅' },
  refund: { label: '关闭退回', cls: 'in', icon: '↩️' }
}

async function load() {
  await store.fetchMe()
  const [r1, r2, r3, r4] = await Promise.all([
    api.get('/my/results'),
    api.get('/my/requests', { params: { mine: 1 } }),
    api.get('/my/interpretations'),
    api.get('/my/transactions')
  ])
  myResults.value = r1.results
  myRequests.value = r2.list
  myInterps.value = r3.list
  txns.value = r4.transactions
}
onMounted(load)

function fmtTime(t) {
  return new Date(t).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <main class="page">
    <div class="container profile-wrap">
      <!-- 顶部：档案 + 钱包 -->
      <div class="top-grid">
        <section class="card profile-card">
          <div class="avatar">{{ store.user?.nickname?.[0] || '?' }}</div>
          <div class="who">
            <h2>{{ store.user?.nickname }} <small>@{{ store.user?.username }}</small></h2>
            <p class="joined">加入于 {{ store.user ? new Date(store.user.created_at).toLocaleDateString() : '' }}</p>
            <div v-if="store.user?.mbti_type" class="type-line">
              <TypeBadge :code="store.user.mbti_type" size="md" />
              <span>当前人格类型</span>
            </div>
            <RouterLink v-else to="/assess" class="btn btn-primary btn-sm">还没测评，去测一下 →</RouterLink>
          </div>
        </section>

        <section class="card wallet-card">
          <h3>我的测评币</h3>
          <div class="wallet-rows">
            <div class="wallet">
              <span class="w-label">🪙 可用余额</span>
              <b class="w-num gold">{{ store.coins }}</b>
            </div>
            <div class="wallet">
              <span class="w-label">❄ 悬赏冻结中</span>
              <b class="w-num muted">{{ store.frozen }}</b>
            </div>
          </div>
          <p class="wallet-tip">
            注册赠送 200 币；发布求解读时冻结悬赏，采纳后结算给解读人，关闭则退回。
          </p>
        </section>
      </div>

      <!-- 最近一次四维占比 -->
      <section v-if="latestResult" class="card block">
        <div class="section-title">
          <h2>人格类型档案</h2>
          <span class="sub">最近一次测评 · {{ fmtTime(latestResult.created_at) }}</span>
        </div>
        <DimensionBars :dimensions="latestResult.dimensions" />
      </section>

      <!-- 记录区 tabs -->
      <section class="card block">
        <div class="rec-tabs">
          <button :class="{ on: tab === 'requests' }" @click="tab = 'requests'">
            📣 我的求解读 <b>{{ myRequests.length }}</b>
          </button>
          <button :class="{ on: tab === 'interps' }" @click="tab = 'interps'">
            💬 我的解读 <b>{{ myInterps.length }}</b>
          </button>
          <button :class="{ on: tab === 'results' }" @click="tab = 'results'">
            🧭 测评历史 <b>{{ myResults.length }}</b>
          </button>
          <button :class="{ on: tab === 'txns' }" @click="tab = 'txns'">
            🧾 测评币流水 <b>{{ txns.length }}</b>
          </button>
        </div>

        <!-- 我的求解读 -->
        <div v-if="tab === 'requests'" class="tab-pane">
          <div v-if="!myRequests.length" class="empty"><span class="emoji">📭</span>还没有发布过求解读</div>
          <div v-else class="card-list">
            <RequestCard v-for="r in myRequests" :key="r.id" :req="r" @changed="load" />
          </div>
        </div>

        <!-- 我的解读 -->
        <div v-if="tab === 'interps'" class="tab-pane">
          <div v-if="!myInterps.length" class="empty"><span class="emoji">🌗</span>还没有帮别人解读过</div>
          <div v-else class="card-list">
            <RequestCard v-for="r in myInterps" :key="r.id" :req="r" @changed="load" />
          </div>
        </div>

        <!-- 测评历史 -->
        <div v-if="tab === 'results'" class="tab-pane">
          <div v-if="!myResults.length" class="empty"><span class="emji">🧭</span>暂无测评记录</div>
          <div v-else class="result-history">
            <RouterLink v-for="r in myResults" :key="r.id" :to="`/result/${r.id}`" class="history-item">
              <TypeBadge :code="r.type_code" size="md" />
              <span class="h-time">{{ fmtTime(r.created_at) }}</span>
              <span class="h-mini">
                <span v-for="d in r.dimensions" :key="d.pair" class="mini-dim">
                  {{ d.dominant }} {{ d.dominant === d.pole_a ? d.percent_a : d.percent_b }}%
                </span>
              </span>
              <span class="h-go">查看 →</span>
            </RouterLink>
          </div>
        </div>

        <!-- 测评币流水 -->
        <div v-if="tab === 'txns'" class="tab-pane">
          <div v-if="!txns.length" class="empty"><span class="emoji">🧾</span>暂无流水</div>
          <table v-else class="txn-table">
            <thead>
              <tr><th>时间</th><th>类型</th><th>说明</th><th class="num">变动</th><th class="num">变动后余额</th></tr>
            </thead>
            <tbody>
              <tr v-for="t in txns" :key="t.id">
                <td class="time">{{ fmtTime(t.created_at) }}</td>
                <td>
                  <span class="tag tag-gray">{{ (txnMeta[t.type] || { icon: '•', label: t.type }).icon }} {{ (txnMeta[t.type] || { label: t.type }).label }}</span>
                </td>
                <td class="remark">{{ t.remark }}</td>
                <td class="num">
                  <b :class="t.change_amount > 0 ? 'plus' : t.change_amount < 0 ? 'minus' : 'zero'">
                    {{ t.change_amount > 0 ? '+' : '' }}{{ t.change_amount }}
                  </b>
                </td>
                <td class="num">{{ t.balance_after }}</td>
              </tr>
            </tbody>
          </table>
          <p class="table-note">
            注：发布/追加时币从「可用」转入「冻结」（计负向流水）；采纳结算时冻结直接核销给对方，
            本人可用余额不再变化，故记一条 0 变动的结算流水。
          </p>
        </div>
      </section>
    </div>
  </main>
</template>

<style scoped>
.profile-wrap { max-width: 960px; }
.top-grid { display: grid; grid-template-columns: 1.3fr 1fr; gap: 18px; margin-bottom: 18px; }
.profile-card { display: flex; gap: 18px; align-items: center; padding: 24px 28px; }
.avatar {
  width: 72px; height: 72px; border-radius: 22px; flex: none;
  display: grid; place-items: center;
  font-size: 30px; font-weight: 900; color: #fff;
  background: linear-gradient(135deg, var(--primary), #9b8cff);
  box-shadow: 0 10px 24px rgba(108,92,231,.3);
}
.who h2 { font-size: 22px; font-weight: 900; }
.who h2 small { font-size: 13px; color: var(--ink-faint); font-weight: 500; margin-left: 6px; }
.joined { font-size: 13px; color: var(--ink-faint); margin: 3px 0 10px; }
.type-line { display: flex; align-items: center; gap: 10px; font-size: 14px; color: var(--ink-soft); font-weight: 600; }

.wallet-card { padding: 24px 26px; }
.wallet-card h3 { font-size: 16px; margin-bottom: 14px; }
.wallet-rows { display: flex; gap: 24px; }
.wallet { display: flex; flex-direction: column; gap: 4px; }
.w-label { font-size: 13px; color: var(--ink-faint); }
.w-num { font-size: 30px; font-weight: 900; line-height: 1.1; }
.w-num.gold { color: #b9760b; }
.w-num.muted { color: var(--ink-soft); }
.wallet-tip { font-size: 12.5px; color: var(--ink-faint); margin-top: 14px; line-height: 1.7; }

.block { padding: 24px 28px; margin-bottom: 18px; }

.rec-tabs { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 20px; }
.rec-tabs button {
  border: 1.5px solid var(--border); background: #fff; border-radius: 999px;
  padding: 8px 18px; font-size: 14px; font-weight: 700; color: var(--ink-soft);
}
.rec-tabs button b {
  display: inline-grid; place-items: center; min-width: 20px; height: 20px;
  border-radius: 999px; background: #eeecfa; color: var(--primary-deep);
  font-size: 12px; margin-left: 5px; padding: 0 5px;
}
.rec-tabs button.on { background: var(--primary); color: #fff; border-color: var(--primary); }
.rec-tabs button.on b { background: rgba(255,255,255,.25); color: #fff; }

.card-list { display: flex; flex-direction: column; gap: 16px; }

.result-history { display: flex; flex-direction: column; gap: 10px; }
.history-item {
  display: flex; align-items: center; gap: 16px;
  border: 1.5px solid var(--border); border-radius: 14px; padding: 13px 18px;
  transition: .15s;
}
.history-item:hover { border-color: var(--primary); background: #faf9ff; }
.h-time { color: var(--ink-faint); font-size: 13px; width: 110px; }
.h-mini { flex: 1; display: flex; gap: 8px; flex-wrap: wrap; }
.mini-dim {
  font-size: 12px; font-weight: 700; color: var(--primary-deep);
  background: var(--primary-soft); padding: 2px 9px; border-radius: 999px;
}
.h-go { color: var(--ink-faint); font-size: 13px; font-weight: 700; }

.txn-table { width: 100%; border-collapse: collapse; font-size: 14px; }
.txn-table th {
  text-align: left; color: var(--ink-faint); font-size: 12.5px; font-weight: 700;
  padding: 8px 10px; border-bottom: 1.5px solid var(--border);
}
.txn-table td { padding: 12px 10px; border-bottom: 1px solid #f1effa; }
.txn-table .num { text-align: right; font-variant-numeric: tabular-nums; }
.txn-table .time { color: var(--ink-faint); font-size: 13px; white-space: nowrap; }
.txn-table .remark { color: var(--ink-soft); font-size: 13px; }
.plus { color: #00865f; }
.minus { color: var(--danger); }
.zero { color: var(--ink-faint); font-weight: 600; }
.table-note { font-size: 12.5px; color: var(--ink-faint); margin-top: 14px; line-height: 1.7; }

@media (max-width: 800px) {
  .top-grid { grid-template-columns: 1fr; }
  .txn-table .remark { display: none; }
}
</style>
