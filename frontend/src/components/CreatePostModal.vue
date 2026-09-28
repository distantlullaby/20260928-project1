<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../api'
import { auth } from '../store/auth'
import { toast } from '../composables/toast'

const emit = defineEmits(['closed', 'created'])

const latest = ref(null)
const bounty = ref(20)
const note = ref('')
const loading = ref(false)
const sending = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const d = await api.get('/quiz/latest')
    latest.value = d.result
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    loading.value = false
  }
})

async function submit() {
  if (!latest.value) return
  if (!bounty.value || bounty.value <= 0) {
    toast('请输入悬赏币数量', 'error')
    return
  }
  if (bounty.value > auth.user.coinBalance) {
    toast('余额不足，无法冻结悬赏', 'error')
    return
  }
  sending.value = true
  try {
    const d = await api.post('/posts', {
      assessmentResultId: latest.value.resultId,
      bounty: Number(bounty.value),
      note: note.value.trim(),
    })
    auth.setBalance(auth.user.coinBalance - Number(bounty.value))
    toast(`发布成功，已冻结 ${bounty.value} 测评币`)
    emit('created', d.post)
    emit('closed')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="modal-mask" @click.self="emit('closed')">
    <div class="modal">
      <button class="modal-close" @click="emit('closed')">×</button>
      <h3>发布「求解读」</h3>
      <p class="sub">冻结悬赏币作为答谢，采纳心仪解读后自动结算给对方</p>

      <div v-if="loading" class="loading">加载测评结果中…</div>
      <template v-else-if="latest">
        <div class="latest-box">
          <div class="lt-type">{{ latest.typeCode }}</div>
          <div>
            <div class="lt-name">{{ latest.profile.name }} · {{ latest.profile.title }}</div>
            <div class="lt-time">测评于 {{ new Date(latest.createdAt).toLocaleString('zh-CN') }}</div>
          </div>
        </div>

        <div class="field">
          <label>悬赏测评币（冻结）</label>
          <input class="input" type="number" min="1" v-model.number="bounty" />
          <div class="balance-tip">
            当前余额 <strong>🪙 {{ auth.user?.coinBalance }}</strong>
            <span v-if="bounty > auth.user?.coinBalance" class="warn">· 余额不足</span>
          </div>
        </div>
        <div class="field">
          <label>给解读人的话（选填）</label>
          <textarea class="textarea" v-model="note" maxlength="500"
            placeholder="例如：最近在职业选择上很纠结，想听听大家从我的类型里看到了什么…"></textarea>
        </div>

        <button class="btn btn-gold btn-block" :disabled="sending" @click="submit">
          {{ sending ? '发布中…' : `确认发布 · 冻结 🪙${bounty || 0}` }}
        </button>
      </template>
      <template v-else>
        <div class="empty">
          <span class="big">🧭</span>
          你还没有测评结果，请先完成一次 MBTI 测评
        </div>
        <button class="btn btn-primary btn-block" @click="emit('closed')">我知道了</button>
      </template>
    </div>
  </div>
</template>

<style scoped>
.loading { text-align: center; color: var(--text-dim); padding: 40px 0; }
.latest-box {
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--bg-soft);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 14px 16px;
  margin-bottom: 18px;
}
.lt-type {
  font-size: 26px;
  font-weight: 800;
  color: #9a86ff;
  letter-spacing: 1px;
}
.lt-name { font-weight: 700; font-size: 14px; }
.lt-time { font-size: 12px; color: var(--text-faint); margin-top: 3px; }
.balance-tip { font-size: 12.5px; color: var(--text-dim); margin-top: 7px; }
.balance-tip strong { color: var(--gold); }
.balance-tip .warn { color: var(--red); }
</style>
