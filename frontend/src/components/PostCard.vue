<script setup>
import { ref, computed } from 'vue'
import { auth } from '../store/auth'
import { api } from '../api'
import { toast } from '../composables/toast'

const props = defineProps({
  post: { type: Object, required: true },
})
const emit = defineEmits(['changed'])

const expanded = ref(false)
const flipped = ref(false) // false=正面（我的类型） true=背面（他人解读）

const showRespondForm = ref(false)
const responseText = ref('')
const submitting = ref(false)
const appending = ref(false)
const appendAmount = ref(10)
const acceptingId = ref(null)

const assessment = computed(() => props.post.assessment || {})
const profile = computed(() => props.post.assessment?.profile || {})
const isOwner = computed(() => auth.user && auth.user.id === props.post.userId)
const myResponse = computed(() =>
  props.post.responses.find(r => auth.user && r.userId === auth.user.id))
const canRespond = computed(() =>
  auth.isLoggedIn && !isOwner.value && props.post.status === 'open' && !myResponse.value)

function timeText(t) {
  if (!t) return ''
  const d = new Date(t)
  const diff = Date.now() - d.getTime()
  const h = Math.floor(diff / 3600000)
  if (h < 1) return Math.max(1, Math.floor(diff / 60000)) + ' 分钟前'
  if (h < 24) return h + ' 小时前'
  return Math.floor(h / 24) + ' 天前'
}

async function submitResponse() {
  if (responseText.value.trim().length < 5) {
    toast('解读内容至少 5 个字', 'error')
    return
  }
  submitting.value = true
  try {
    await api.post(`/posts/${props.post.id}/responses`, { content: responseText.value.trim() })
    toast('解读已提交，等待发起人采纳')
    showRespondForm.value = false
    responseText.value = ''
    emit('changed')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    submitting.value = false
  }
}

async function appendBounty() {
  if (!appendAmount.value || appendAmount.value <= 0) {
    toast('请输入正确的追加金额', 'error')
    return
  }
  appending.value = true
  try {
    const d = await api.post(`/posts/${props.post.id}/append`, { amount: Number(appendAmount.value) })
    auth.setBalance(d.coinBalance)
    toast(`已追加 ${appendAmount.value} 币，新悬赏 ${props.post.bounty + Number(appendAmount.value)}`)
    appendAmount.value = 10
    emit('changed')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    appending.value = false
  }
}

async function accept(rid) {
  acceptingId.value = rid
  try {
    await api.post(`/posts/${props.post.id}/responses/${rid}/accept`)
    toast('已采纳，悬赏币已结算给对方 🪙')
    emit('changed')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    acceptingId.value = null
  }
}
</script>

<template>
  <div class="post-card" :class="{ expanded }">
    <!-- 卡片头部：始终可见 -->
    <div class="card-head" @click="expanded = !expanded">
      <div class="author">
        <div class="avatar">{{ post.author.nickname?.[0] || '?' }}</div>
        <div>
          <div class="name">{{ post.author.nickname }}</div>
          <div class="meta">
            <span class="type-badge">{{ assessment.typeCode || '----' }}</span>
            <span>{{ timeText(post.createdAt) }}</span>
          </div>
        </div>
      </div>
      <div class="head-right">
        <span class="tag tag-coin">🪙 {{ post.bounty }}</span>
        <span class="tag" :class="post.status === 'open' ? 'tag-open' : 'tag-settled'">
          {{ post.status === 'open' ? '征集中' : '已采纳' }}
        </span>
        <span class="caret">{{ expanded ? '收起 ▲' : '展开双面卡片 ▼' }}</span>
      </div>
    </div>

    <div v-if="post.note" class="post-note">📝 {{ post.note }}</div>

    <!-- 展开后的双面卡片 -->
    <div v-if="expanded" class="card-body">
      <div class="flip-tabs">
        <button :class="{ active: !flipped }" @click="flipped = false">
          🔮 测评结果 · 类型画像
        </button>
        <button :class="{ active: flipped }" @click="flipped = true">
          💬 他人解读 · 洞察视角
          <span v-if="post.responses.length" class="count">{{ post.responses.length }}</span>
        </button>
      </div>

      <transition name="fade" mode="out-in">
        <!-- ===== 正面：测评结果 + 类型画像 ===== -->
        <div v-if="!flipped" key="front" class="face face-front">
          <div class="profile-top">
            <div class="big-type">{{ profile.code }}</div>
            <div>
              <div class="type-name">{{ profile.name }} · {{ profile.title }}</div>
              <p class="summary">{{ profile.summary }}</p>
            </div>
          </div>

          <div class="bars">
            <div v-for="d in assessment.dimensions" :key="d.positive" class="bar-row">
              <span class="pole" :class="{ weak: d.dominant === d.negative }">{{ d.positive }} {{ d.posPercent }}%</span>
              <div class="bar-track">
                <div class="bar-pos" :style="{ width: d.posPercent + '%' }"></div>
                <div class="bar-neg" :style="{ width: d.negPercent + '%' }"></div>
              </div>
              <span class="pole" :class="{ weak: d.dominant === d.positive }">{{ d.negPercent }}% {{ d.negative }}</span>
            </div>
          </div>

          <div class="trait-grid">
            <div class="trait-block">
              <h4>✨ 性格特质</h4>
              <div class="chips"><span v-for="t in profile.traits" :key="t">{{ t }}</span></div>
            </div>
            <div class="trait-block">
              <h4>💪 优势</h4>
              <ul><li v-for="s in profile.strengths" :key="s">{{ s }}</li></ul>
            </div>
            <div class="trait-block">
              <h4>🌫 盲区</h4>
              <ul><li v-for="b in profile.blind" :key="b">{{ b }}</li></ul>
            </div>
          </div>
        </div>

        <!-- ===== 背面：他人解读 ===== -->
        <div v-else key="back" class="face face-back">
          <div v-if="post.responses.length === 0" class="empty">
            <span class="big">🪶</span>
            还没有人解读，来做 TA 的第一面镜子吧
          </div>

          <div v-for="r in post.responses" :key="r.id" class="response"
            :class="{ accepted: post.acceptedResponseId === r.id }">
            <div class="response-head">
              <div class="avatar sm">{{ r.resolver.nickname?.[0] || '?' }}</div>
              <strong>{{ r.resolver.nickname }}</strong>
              <span class="resp-time">{{ timeText(r.createdAt) }}</span>
              <span v-if="post.acceptedResponseId === r.id" class="tag tag-settled">⭐ 已采纳</span>
            </div>
            <p class="response-content">{{ r.content }}</p>
            <div v-if="isOwner && post.status === 'open'" class="response-actions">
              <button class="btn btn-gold btn-sm"
                :disabled="acceptingId === r.id"
                @click="accept(r.id)">
                {{ acceptingId === r.id ? '结算中…' : '✅ 采纳并结算 🪙' + post.bounty }}
              </button>
            </div>
          </div>

          <!-- 帮 TA 解读 -->
          <div v-if="canRespond" class="respond-box">
            <button v-if="!showRespondForm" class="btn btn-primary btn-block"
              @click="showRespondForm = true">🙌 帮 TA 解读</button>
            <template v-else>
              <textarea class="textarea" v-model="responseText"
                placeholder="说说你从这份结果里看到了什么：TA 可能没意识到的优势、盲区，或一个新的视角…（5-1000 字）"
                maxlength="1000"></textarea>
              <div class="row-end">
                <button class="btn btn-ghost btn-sm" @click="showRespondForm = false">取消</button>
                <button class="btn btn-primary btn-sm" :disabled="submitting" @click="submitResponse">
                  {{ submitting ? '提交中…' : '提交解读' }}
                </button>
              </div>
            </template>
          </div>
          <p v-else-if="!auth.isLoggedIn && post.status === 'open'" class="hint">登录后即可帮 TA 解读</p>
          <p v-else-if="isOwner && post.status === 'open'" class="hint">这是你发布的需求 · 可追加悬赏吸引更多解读</p>
          <p v-else-if="myResponse && post.status === 'open'" class="hint">你已提交解读，等待发起人采纳中…</p>

          <!-- 发帖人：追加悬赏 -->
          <div v-if="isOwner && post.status === 'open'" class="append-box">
            <input class="input append-input" type="number" min="1" v-model.number="appendAmount" />
            <button class="btn btn-ghost btn-sm" :disabled="appending" @click="appendBounty">
              {{ appending ? '处理中…' : '➕ 追加悬赏' }}
            </button>
            <span class="hint-inline">可用余额 {{ auth.user?.coinBalance }} 币</span>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<style scoped>
.post-card {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  margin-bottom: 16px;
  overflow: hidden;
  transition: border-color 0.2s, transform 0.2s;
}
.post-card:hover { border-color: #3d4288; }
.post-card.expanded { border-color: var(--primary-soft); box-shadow: var(--shadow); }

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 20px;
  cursor: pointer;
  user-select: none;
}
.author { display: flex; align-items: center; gap: 12px; }
.avatar {
  width: 42px; height: 42px;
  border-radius: 50%;
  display: grid; place-items: center;
  font-weight: 700; font-size: 17px;
  background: linear-gradient(135deg, var(--primary), var(--cyan));
  color: #fff;
  flex-shrink: 0;
}
.avatar.sm { width: 30px; height: 30px; font-size: 13px; }
.name { font-weight: 700; font-size: 15px; }
.meta { display: flex; align-items: center; gap: 8px; margin-top: 3px; font-size: 12px; color: var(--text-faint); }
.type-badge {
  background: var(--primary-soft);
  color: #d7d0ff;
  font-weight: 700;
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 6px;
  letter-spacing: 0.5px;
}
.head-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.caret { font-size: 12px; color: var(--text-faint); margin-left: 4px; }

.post-note {
  padding: 0 20px 14px;
  color: var(--text-dim);
  font-size: 13px;
}

.card-body { padding: 0 20px 20px; }

.flip-tabs {
  display: flex;
  gap: 8px;
  background: var(--bg-soft);
  border-radius: 10px;
  padding: 4px;
  margin-bottom: 16px;
}
.flip-tabs button {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--text-dim);
  font-size: 13px;
  font-weight: 600;
  padding: 9px;
  border-radius: 8px;
  position: relative;
  transition: all 0.2s;
}
.flip-tabs button.active {
  background: linear-gradient(135deg, var(--primary), #9a86ff);
  color: #fff;
  box-shadow: 0 4px 14px var(--primary-glow);
}
.count {
  display: inline-grid;
  place-items: center;
  background: rgba(255,255,255,0.25);
  border-radius: 999px;
  min-width: 18px;
  height: 18px;
  font-size: 11px;
  margin-left: 4px;
}

.face { animation: face-in 0.3s ease; }
@keyframes face-in {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s, transform 0.15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(6px); }

/* 正面 */
.profile-top { display: flex; gap: 18px; align-items: flex-start; margin-bottom: 18px; }
.big-type {
  font-size: 42px;
  font-weight: 800;
  letter-spacing: 2px;
  background: linear-gradient(135deg, #9a86ff, var(--cyan));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1.1;
}
.type-name { font-size: 16px; font-weight: 700; margin-bottom: 6px; }
.summary { color: var(--text-dim); font-size: 13px; line-height: 1.7; }

.bars { display: flex; flex-direction: column; gap: 10px; margin-bottom: 18px; }
.bar-row { display: flex; align-items: center; gap: 10px; }
.pole {
  width: 62px;
  font-size: 12px;
  font-weight: 700;
  color: var(--text);
  text-align: center;
}
.pole.weak { color: var(--text-faint); font-weight: 500; }
.bar-track {
  flex: 1;
  display: flex;
  height: 10px;
  border-radius: 999px;
  overflow: hidden;
  background: var(--bg-soft);
}
.bar-neg { background: linear-gradient(90deg, #45498a, var(--cyan)); }
.bar-pos { background: linear-gradient(90deg, var(--primary), #9a86ff); }

.trait-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.trait-block {
  background: var(--bg-soft);
  border-radius: 10px;
  padding: 13px 15px;
}
.trait-block:first-child { grid-column: 1 / -1; }
.trait-block h4 { font-size: 13px; margin-bottom: 9px; }
.trait-block ul { list-style: none; }
.trait-block li {
  font-size: 12.5px;
  color: var(--text-dim);
  padding: 3px 0 3px 16px;
  position: relative;
}
.trait-block li::before {
  content: '·';
  position: absolute;
  left: 4px;
  color: var(--primary);
  font-weight: 800;
}
.chips { display: flex; flex-wrap: wrap; gap: 8px; }
.chips span {
  background: var(--card);
  border: 1px solid var(--line);
  color: var(--text);
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 999px;
}

/* 背面 */
.response {
  background: var(--bg-soft);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 14px 16px;
  margin-bottom: 12px;
}
.response.accepted { border-color: rgba(245, 181, 68, 0.5); background: rgba(245, 181, 68, 0.06); }
.response-head { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.response-head strong { font-size: 13.5px; }
.resp-time { font-size: 12px; color: var(--text-faint); }
.response-content { font-size: 14px; line-height: 1.75; color: var(--text); }
.response-actions { margin-top: 10px; display: flex; justify-content: flex-end; }

.respond-box { margin-top: 14px; }
.row-end { display: flex; justify-content: flex-end; gap: 8px; margin-top: 10px; }
.hint { text-align: center; color: var(--text-faint); font-size: 13px; margin-top: 12px; }

.append-box {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed var(--line);
}
.append-input { width: 110px; padding: 7px 10px; }
.hint-inline { font-size: 12px; color: var(--text-faint); }

@media (max-width: 640px) {
  .card-head { flex-direction: column; align-items: flex-start; }
  .trait-grid { grid-template-columns: 1fr; }
  .profile-top { flex-direction: column; gap: 8px; }
}
</style>
