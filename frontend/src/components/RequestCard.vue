<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import TypeBadge from './TypeBadge.vue'
import DimensionBars from './DimensionBars.vue'
import { useUserStore } from '../stores/user'
import { toast } from '../composables/toast'
import api from '../api'

const props = defineProps({
  req: { type: Object, required: true }
})
const emit = defineEmits(['changed'])

const router = useRouter()
const store = useUserStore()
const flipped = ref(false)
const detail = ref(null)
const loadingDetail = ref(false)

// 列表接口已带来正面数据；解读列表按需展开时加载
async function loadDetail() {
  loadingDetail.value = true
  try {
    const { request } = await api.get(`/requests/${props.req.id}`)
    detail.value = request
  } catch (e) {
    toast(e.message, 'err')
  } finally {
    loadingDetail.value = false
  }
}
async function flip() {
  flipped.value = !flipped.value
  if (flipped.value && !detail.value) await loadDetail()
}

const profile = computed(() => props.req.profile)
const assessment = computed(() => props.req.assessment)
const interpretations = computed(() => detail.value?.interpretations || [])
const accepted = computed(() => interpretations.value.find((i) => i.accepted))

const isOwner = computed(() => store.user && store.user.id === props.req.user_id)
const statusText = computed(() => ({
  open: '悬赏中', settled: '已采纳结算', closed: '已关闭退款'
}[props.req.status]))
const statusTag = computed(() => ({
  open: 'tag-gold', settled: 'tag-green', closed: 'tag-gray'
}[props.req.status]))

const traitList = computed(() => (profile.value?.traits || '').split(/[、,]/).filter(Boolean))

// ---- 帮 TA 解读 ----
const showInterpretBox = ref(false)
const interpText = ref('')
const submitting = ref(false)
async function submitInterpretation() {
  if (!store.isLoggedIn) {
    toast('请先登录后再解读', 'info')
    router.push({ name: 'login', query: { redirect: `/requests/${props.req.id}` } })
    return
  }
  submitting.value = true
  try {
    await api.post(`/requests/${props.req.id}/interpretations`, { content: interpText.value })
    toast('解读已提交，等待发起人采纳')
    interpText.value = ''
    showInterpretBox.value = false
    detail.value = null
    await loadDetail()
    flipped.value = true
    emit('changed')
  } catch (e) {
    toast(e.message, 'err')
  } finally {
    submitting.value = false
  }
}

// ---- 采纳 ----
const accepting = ref(false)
async function accept(item) {
  if (!confirm(`确认采纳 ${item.user.nickname} 的解读？${props.req.reward} 枚测评币将立即结算给对方，不可撤销。`)) return
  accepting.value = true
  try {
    await api.post(`/requests/${props.req.id}/interpretations/${item.id}/accept`)
    toast(`已采纳，${props.req.reward} 枚测评币已结算`)
    detail.value = null
    emit('changed')
    await loadDetail()
    flipped.value = true
  } catch (e) {
    toast(e.message, 'err')
  } finally {
    accepting.value = false
  }
}

// ---- 追加悬赏 / 关闭 ----
const showAppend = ref(false)
const appendAmount = ref(10)
async function appendReward() {
  try {
    await api.post(`/requests/${props.req.id}/append`, { amount: Number(appendAmount.value) })
    toast(`已追加 ${appendAmount.value} 币并冻结`)
    showAppend.value = false
    emit('changed')
  } catch (e) { toast(e.message, 'err') }
}
async function closeRequest() {
  if (!confirm('确认关闭这条求解读？冻结的悬赏币将全额退回。')) return
  try {
    await api.post(`/requests/${props.req.id}/close`)
    toast('已关闭，悬赏币已退回')
    emit('changed')
  } catch (e) { toast(e.message, 'err') }
}

const hasMyInterpretation = computed(() =>
  interpretations.value.some((i) => i.user?.id === store.user?.id)
)
const canInterpret = computed(() =>
  store.isLoggedIn && !isOwner.value && props.req.status === 'open' && !hasMyInterpretation.value
)
</script>

<template>
  <div class="flip-card" :class="{ flipped }">
    <div class="flip-inner">
      <!-- ========== 正面：测评结果 + 类型画像 ========== -->
      <div class="face face-front card">
        <div class="face-head">
          <div class="head-main">
            <span class="tag" :class="statusTag">{{ statusText }}</span>
            <h3 class="req-title">{{ req.title }}</h3>
            <div class="req-meta">
              <span class="who">🙋 {{ req.user?.nickname }}</span>
              <span class="reward">🪙 悬赏 {{ req.reward }} 币</span>
              <span class="count">💬 {{ req.interpretation_count ?? 0 }} 条解读</span>
            </div>
          </div>
          <TypeBadge :code="assessment?.type_code" size="lg" />
        </div>

        <p v-if="req.content" class="req-content">{{ req.content }}</p>

        <div v-if="assessment" class="front-grid">
          <div class="profile-box">
            <div class="profile-name">
              <b>{{ profile?.nickname || '——' }}</b>
              <span class="tag tag-purple">{{ profile?.group }}</span>
            </div>
            <p class="profile-desc">{{ profile?.description }}</p>
            <div v-if="traitList.length" class="trait-row">
              <span v-for="t in traitList" :key="t" class="trait-chip">{{ t }}</span>
            </div>
          </div>
          <DimensionBars :dimensions="assessment.dimensions" compact />
        </div>

        <div class="face-actions">
          <button v-if="isOwner && req.status === 'open'" class="btn btn-outline btn-sm" @click="showAppend = true">＋ 追加悬赏</button>
          <button v-if="isOwner && req.status === 'open'" class="btn btn-danger btn-sm" @click="closeRequest">关闭并退款</button>
          <span class="spacer"></span>
          <button class="btn btn-ghost btn-sm" @click="flip">
            🔄 翻面看他人解读
          </button>
        </div>
      </div>

      <!-- ========== 背面：他人解读 + 洞察视角 ========== -->
      <div class="face face-back card">
        <div class="back-head">
          <div>
            <span class="tag tag-purple">他人解读 · 洞察视角</span>
            <h3 class="req-title small">{{ req.title }}</h3>
          </div>
          <button class="btn btn-outline btn-sm" @click="flip">↩ 返回画像</button>
        </div>

        <div v-if="loadingDetail" class="back-loading">正在加载解读…</div>

        <div v-else class="interp-list">
          <div v-if="!interpretations.length" class="interp-empty">
            <span>🌗</span>
            <p>还没有人解读，你的视角可能正是 TA 需要的</p>
          </div>

          <div v-for="item in interpretations" :key="item.id"
               class="interp-item" :class="{ accepted: item.accepted }">
            <div class="interp-meta">
              <span class="interp-who">👤 {{ item.user?.nickname }}</span>
              <span v-if="item.accepted" class="tag tag-green">✓ 已采纳 · 获 {{ req.reward }} 币</span>
            </div>
            <p class="interp-text">{{ item.content }}</p>
            <button v-if="isOwner && req.status === 'open' && !accepting"
                    class="btn btn-gold btn-sm" @click="accept(item)">
              ✅ 确认采纳并结算 {{ req.reward }} 币
            </button>
          </div>
        </div>

        <div v-if="canInterpret && !loadingDetail" class="interpret-entry">
          <button v-if="!showInterpretBox" class="btn btn-primary btn-block" @click="showInterpretBox = true">
            ✍️ 帮 TA 解读（被采纳可获 {{ req.reward }} 币）
          </button>
          <div v-else class="interpret-form">
            <textarea v-model="interpText" class="textarea" rows="4"
              placeholder="结合 TA 的性格画像说说你的观察与建议，至少 10 个字…"></textarea>
            <div class="form-actions">
              <button class="btn btn-outline btn-sm" @click="showInterpretBox = false">取消</button>
              <button class="btn btn-primary btn-sm" :disabled="submitting || interpText.length < 10" @click="submitInterpretation">
                {{ submitting ? '提交中…' : '提交解读' }}
              </button>
            </div>
          </div>
        </div>
        <p v-else-if="!store.isLoggedIn && req.status === 'open'" class="login-hint">
          <RouterLink :to="{ name: 'login' }">登录</RouterLink> 后即可帮 TA 解读
        </p>
        <p v-else-if="isOwner" class="login-hint">这是你发布的求解读，等待他人提交视角后可在此采纳</p>
        <p v-else-if="hasMyInterpretation && req.status === 'open'" class="login-hint">
          ✍️ 你已提交解读，等待发起人采纳
        </p>
      </div>
    </div>

    <!-- 追加悬赏弹层 -->
    <div v-if="showAppend" class="modal-mask" @click.self="showAppend = false">
      <div class="modal">
        <h3>＋ 追加悬赏币</h3>
        <p class="field-hint" style="margin-bottom:12px">
          追加的测评币会立即从可用余额冻结，与原悬赏合并。当前可用 {{ store.coins }} 币。
        </p>
        <input v-model.number="appendAmount" type="number" min="1" class="input" />
        <div class="quick-amounts">
          <button v-for="n in [5, 10, 20, 50]" :key="n" class="qa" @click="appendAmount = n">{{ n }}</button>
        </div>
        <div class="modal-actions">
          <button class="btn btn-outline" @click="showAppend = false">取消</button>
          <button class="btn btn-primary" @click="appendReward">确认追加</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.flip-card { perspective: 2000px; }
.flip-inner {
  position: relative;
  transition: transform .55s cubic-bezier(.4,.2,.2,1);
  transform-style: preserve-3d;
}
.flip-card.flipped .flip-inner { transform: rotateY(180deg); }
.face {
  padding: 22px 24px;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}
.face-back {
  position: absolute; inset: 0;
  transform: rotateY(180deg);
  overflow-y: auto;
  max-height: 560px;
}

.face-head { display: flex; gap: 14px; justify-content: space-between; align-items: flex-start; }
.head-main { flex: 1; min-width: 0; }
.req-title { font-size: 17.5px; font-weight: 800; margin: 8px 0 6px; line-height: 1.5; }
.req-title.small { font-size: 15.5px; }
.req-meta { display: flex; flex-wrap: wrap; gap: 14px; font-size: 13px; color: var(--ink-faint); }
.req-meta .reward { color: #b9760b; font-weight: 700; }
.req-content {
  margin: 12px 0 14px; padding: 12px 14px;
  background: #faf9ff; border-left: 3px solid var(--primary);
  border-radius: 0 10px 10px 0; font-size: 14px; color: var(--ink-soft);
}
.front-grid {
  display: grid; grid-template-columns: 1fr 1.15fr; gap: 20px;
  border-top: 1px dashed var(--border); padding-top: 14px;
}
.profile-name { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; }
.profile-name b { font-size: 17px; }
.profile-desc { font-size: 13.5px; color: var(--ink-soft); margin-bottom: 10px; }
.trait-row { display: flex; flex-wrap: wrap; gap: 6px; }
.trait-chip {
  font-size: 12px; padding: 3px 10px; border-radius: 999px;
  background: var(--primary-soft); color: var(--primary-deep); font-weight: 600;
}
.face-actions { display: flex; align-items: center; gap: 8px; margin-top: 16px; }
.face-actions .spacer { flex: 1; }

.back-head { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 12px; }
.back-loading { padding: 40px; text-align: center; color: var(--ink-faint); }
.interp-list { display: flex; flex-direction: column; gap: 12px; }
.interp-empty { text-align: center; padding: 26px 10px; color: var(--ink-faint); }
.interp-empty span { font-size: 34px; display: block; margin-bottom: 6px; }
.interp-empty p { font-size: 14px; }
.interp-item {
  border: 1px solid var(--border); border-radius: 14px; padding: 14px 16px;
  background: #fcfcff; transition: .2s;
}
.interp-item.accepted {
  border-color: #7fd8bf; background: linear-gradient(180deg, #f3fffb, #fff);
  box-shadow: 0 6px 18px rgba(0,184,148,.12);
}
.interp-meta { display: flex; justify-content: space-between; align-items: center; margin-bottom: 7px; }
.interp-who { font-size: 13px; font-weight: 700; color: var(--ink-soft); }
.interp-text { font-size: 14px; color: var(--ink); line-height: 1.75; margin-bottom: 10px; }
.interpret-entry { margin-top: 14px; }
.interpret-form .form-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }
.login-hint { text-align: center; font-size: 13px; color: var(--ink-faint); margin-top: 12px; }
.login-hint a { color: var(--primary); font-weight: 700; }
.quick-amounts { display: flex; gap: 8px; margin-top: 10px; }
.qa {
  border: 1.5px solid var(--border); background: #fff; border-radius: 999px;
  padding: 4px 16px; font-size: 13px; color: var(--ink-soft);
}
.qa:hover { border-color: var(--primary); color: var(--primary); }

@media (max-width: 720px) {
  .front-grid { grid-template-columns: 1fr; }
  .face { padding: 18px; }
}
</style>
