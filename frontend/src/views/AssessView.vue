<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import { useUserStore } from '../stores/user'
import { toast } from '../composables/toast'

const router = useRouter()
const store = useUserStore()

const questions = ref([])
const answers = ref({}) // { questionId: 'A' | 'B' }
const idx = ref(0)
const submitting = ref(false)

onMounted(async () => {
  const { questions: list } = await api.get('/questions')
  questions.value = list
})

const total = computed(() => questions.value.length)
const current = computed(() => questions.value[idx.value])
const progress = computed(() => (Object.keys(answers.value).length / (total.value || 1)) * 100)
const answeredCount = computed(() => Object.keys(answers.value).length)

const dimNames = { EI: '精力方向 E/I', SN: '信息方式 S/N', TF: '决策依据 T/F', JP: '生活方式 J/P' }
const currentDim = computed(() => current.value ? dimNames[current.value.dimension] : '')

function choose(pole) {
  answers.value[current.value.id] = pole
  // 最后一题或已经答完则提交，否则前进到下一道未答题
  setTimeout(() => {
    if (answeredCount.value >= total.value) {
      submit()
    } else {
      next()
    }
  }, 180)
}

function next() {
  if (idx.value < total.value - 1) idx.value++
}
function prev() {
  if (idx.value > 0) idx.value--
}
function jumpTo(i) { idx.value = i }

async function submit() {
  if (!store.isLoggedIn) {
    toast('测评完成，登录后即可保存结果并生成画像', 'info')
    router.push({ name: 'login', query: { redirect: '/assess' } })
    return
  }
  submitting.value = true
  try {
    const payload = {
      answers: questions.value.map((q) => ({ question_id: q.id, pole: answers.value[q.id] }))
    }
    const res = await api.post('/assessments', payload)
    router.push({ name: 'result', params: { id: res.result_id }, query: { fresh: 1 } })
  } catch (e) {
    toast(e.message, 'err')
    submitting.value = false
  }
}
</script>

<template>
  <main class="page">
    <div class="container assess-wrap">
      <div v-if="!questions.length" class="empty"><span class="emoji">⏳</span>题目加载中…</div>

      <template v-else-if="current">
        <div class="assess-head card">
          <div class="head-top">
            <span class="tag tag-purple">{{ currentDim }}</span>
            <span class="counter">第 {{ idx + 1 }} / {{ total }} 题</span>
          </div>
          <div class="progress">
            <div class="progress-bar" :style="{ width: progress + '%' }"></div>
          </div>
          <div class="dots">
            <button v-for="(q, i) in questions" :key="q.id"
                    class="dot" :class="{ done: answers[q.id], active: i === idx, curdim: q.dimension !== current.dimension }"
                    :title="`第${i + 1}题`" @click="jumpTo(i)"></button>
          </div>
        </div>

        <Transition name="q" mode="out-in">
          <div :key="current.id" class="question-card card">
            <h2 class="q-text">{{ current.text }}</h2>
            <div class="options">
              <button class="option" :class="{ picked: answers[current.id] === 'A' }" @click="choose('A')">
                <span class="option-key">A</span>
                <span class="option-text">{{ current.option_a_text }}</span>
                <span class="option-pole">{{ current.option_a_pole }}</span>
              </button>
              <div class="or">— 或 —</div>
              <button class="option" :class="{ picked: answers[current.id] === 'B' }" @click="choose('B')">
                <span class="option-key">B</span>
                <span class="option-text">{{ current.option_b_text }}</span>
                <span class="option-pole">{{ current.option_b_pole }}</span>
              </button>
            </div>
          </div>
        </Transition>

        <div class="assess-nav">
          <button class="btn btn-outline" :disabled="idx === 0" @click="prev">← 上一题</button>
          <span class="hint">凭第一直觉选择最像你的一项，无需深思</span>
          <button class="btn btn-primary" :disabled="idx === total - 1" @click="next">
            下一题 →
          </button>
        </div>
        <p v-if="submitting" class="submitting">正在计算你的人格类型…</p>
      </template>
    </div>
  </main>
</template>

<style scoped>
.assess-wrap { max-width: 760px; }
.assess-head { padding: 18px 24px; margin-bottom: 20px; }
.head-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.counter { font-size: 14px; color: var(--ink-faint); font-weight: 700; }
.progress { height: 8px; background: #eeecfa; border-radius: 999px; overflow: hidden; }
.progress-bar {
  height: 100%; border-radius: 999px;
  background: linear-gradient(90deg, var(--primary), #00b8a9);
  transition: width .3s;
}
.dots { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 12px; }
.dot {
  width: 11px; height: 11px; border-radius: 50%; border: none;
  background: #e3e1f0; transition: .15s; padding: 0;
}
.dot.done { background: #b3a8f5; }
.dot.active { background: var(--primary); transform: scale(1.35); }

.question-card { padding: 34px 36px; }
.q-text { font-size: 22px; font-weight: 800; line-height: 1.55; margin-bottom: 26px; text-align: center; }
.options { display: flex; flex-direction: column; gap: 12px; }
.option {
  display: flex; align-items: center; gap: 14px; text-align: left;
  border: 2px solid var(--border); background: #fcfcff;
  border-radius: 16px; padding: 18px 20px; transition: .18s; position: relative;
}
.option:hover { border-color: #b3a8f5; background: #faf9ff; transform: translateX(4px); }
.option.picked { border-color: var(--primary); background: var(--primary-soft); box-shadow: 0 8px 20px rgba(108,92,231,.15); }
.option-key {
  width: 34px; height: 34px; flex: none; border-radius: 10px;
  display: grid; place-items: center; font-weight: 800; font-size: 16px;
  background: #eeecfa; color: var(--primary-deep);
}
.option.picked .option-key { background: var(--primary); color: #fff; }
.option-text { flex: 1; font-size: 16px; line-height: 1.5; }
.option-pole {
  font-size: 12px; font-weight: 800; color: var(--ink-faint);
  border: 1px solid var(--border); border-radius: 6px; padding: 2px 7px;
}
.option.picked .option-pole { border-color: var(--primary); color: var(--primary); }
.or { text-align: center; color: var(--ink-faint); font-size: 13px; }

.assess-nav { display: flex; align-items: center; gap: 14px; margin-top: 20px; }
.hint { flex: 1; text-align: center; font-size: 13px; color: var(--ink-faint); }
.submitting { text-align: center; color: var(--primary); font-weight: 700; margin-top: 14px; }

.q-enter-active, .q-leave-active { transition: all .22s ease; }
.q-enter-from { opacity: 0; transform: translateX(28px); }
.q-leave-to { opacity: 0; transform: translateX(-28px); }

@media (max-width: 600px) {
  .question-card { padding: 24px 18px; }
  .q-text { font-size: 18px; }
  .option { padding: 14px; }
  .hint { display: none; }
}
</style>
