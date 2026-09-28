<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { auth } from '../store/auth'
import { toast } from '../composables/toast'

const router = useRouter()

const questions = ref([])
const answers = ref({}) // {题目id: 'A'|'B'}
const current = ref(0)
const submitting = ref(false)
const result = ref(null)

onMounted(async () => {
  try {
    const d = await api.get('/quiz/questions')
    questions.value = d.questions
  } catch (e) {
    toast(e.message, 'error')
  }
})

const q = computed(() => questions.value[current.value])
const progress = computed(() =>
  questions.value.length ? Math.round((Object.keys(answers.value).length / questions.value.length) * 100) : 0)

function choose(option) {
  answers.value[q.value.id] = option
  if (current.value < questions.value.length - 1) {
    setTimeout(() => { current.value++ }, 160)
  }
}
function go(i) { if (i <= Object.keys(answers.value).length) current.value = i }

async function submit() {
  submitting.value = true
  try {
    const d = await api.post('/quiz/submit', { answers: answers.value })
    result.value = d
    auth.fetchMe?.()
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    submitting.value = false
  }
}

function restart() {
  answers.value = {}
  current.value = 0
  result.value = null
}
</script>

<template>
  <div class="container quiz-page">
    <!-- 结果页 -->
    <div v-if="result" class="result-wrap">
      <div class="result-hero">
        <div class="hero-label">你的人格类型是</div>
        <div class="hero-type">{{ result.typeCode }}</div>
        <div class="hero-name">{{ result.profile.name }} · {{ result.profile.title }}</div>
        <p class="hero-summary">{{ result.profile.summary }}</p>
      </div>

      <div class="panel">
        <h3>📊 四维倾向</h3>
        <div v-for="d in result.dimensions" :key="d.positive" class="dim-row">
          <div class="dim-labels">
            <span :class="{ on: d.dominant === d.positive }">
              {{ d.positive }} · {{ dimName(d.positive) }} <b>{{ d.posPercent }}%</b>
            </span>
            <span :class="{ on: d.dominant === d.negative }">
              <b>{{ d.negPercent }}%</b> {{ dimName(d.negative) }} · {{ d.negative }}
            </span>
          </div>
          <div class="dim-track">
            <div class="dim-pos" :style="{ width: d.posPercent + '%' }"></div>
            <div class="dim-neg" :style="{ width: d.negPercent + '%' }"></div>
          </div>
        </div>
      </div>

      <div class="panel">
        <h3>🪞 类型画像</h3>
        <div class="chips">
          <span v-for="t in result.profile.traits" :key="t" class="chip">{{ t }}</span>
        </div>
        <div class="pf-cols">
          <div>
            <h4>💪 优势</h4>
            <ul><li v-for="s in result.profile.strengths" :key="s">{{ s }}</li></ul>
          </div>
          <div>
            <h4>🌫 成长盲区</h4>
            <ul><li v-for="b in result.profile.blind" :key="b">{{ b }}</li></ul>
          </div>
        </div>
      </div>

      <div class="result-actions">
        <button class="btn btn-ghost" @click="restart">🔄 重新测一次</button>
        <button class="btn btn-gold" @click="router.push('/?publish=1')">🪙 发布「求解读」获取他人视角</button>
      </div>
    </div>

    <!-- 答题页 -->
    <template v-else-if="questions.length">
      <div class="quiz-head">
        <h2>🧭 MBTI 人格测评</h2>
        <p>共 {{ questions.length }} 题，凭第一直觉选择更接近你的选项</p>
        <div class="progress">
          <div class="progress-bar" :style="{ width: progress + '%' }"></div>
        </div>
        <div class="progress-text">{{ Object.keys(answers).length }} / {{ questions.length }}</div>
      </div>

      <transition name="q" mode="out-in">
        <div :key="q.id" class="question-card">
          <div class="q-index">第 {{ current + 1 }} 题 / 共 {{ questions.length }} 题</div>
          <div class="q-text">{{ q.text }}</div>
          <div class="options">
            <button class="option" :class="{ picked: answers[q.id] === 'A' }" @click="choose('A')">
              <span class="opt-key">A</span>
              <span>{{ q.optionA }}</span>
            </button>
            <button class="option" :class="{ picked: answers[q.id] === 'B' }" @click="choose('B')">
              <span class="opt-key">B</span>
              <span>{{ q.optionB }}</span>
            </button>
          </div>
        </div>
      </transition>

      <div class="quiz-nav">
        <button class="btn btn-ghost" :disabled="current === 0" @click="current--">上一题</button>
        <div class="dots">
          <span v-for="(item, i) in questions" :key="item.id"
            :class="{ done: answers[item.id], active: i === current }"
            @click="go(i)"></span>
        </div>
        <button v-if="current < questions.length - 1" class="btn btn-primary"
          :disabled="!answers[q.id]" @click="current++">下一题</button>
        <button v-else class="btn btn-gold" :disabled="Object.keys(answers).length < questions.length || submitting"
          @click="submit">
          {{ submitting ? '计算中…' : '✨ 查看测评结果' }}
        </button>
      </div>
    </template>

    <div v-else class="empty">题目加载中…</div>
  </div>
</template>

<script>
const DIM_NAMES = {
  E: '外向', I: '内向', S: '实感', N: '直觉',
  T: '思考', F: '情感', J: '判断', P: '感知',
}
export default {
  methods: { dimName(p) { return DIM_NAMES[p] || p } },
}
</script>

<style scoped>
.quiz-page { padding-top: 30px; max-width: 760px; }
.quiz-head { text-align: center; margin-bottom: 26px; }
.quiz-head h2 { font-size: 26px; margin-bottom: 8px; }
.quiz-head p { color: var(--text-dim); font-size: 14px; }
.progress {
  height: 8px;
  background: var(--bg-soft);
  border-radius: 999px;
  margin: 18px auto 8px;
  max-width: 420px;
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, var(--primary), var(--cyan));
  transition: width 0.3s;
}
.progress-text { font-size: 12px; color: var(--text-faint); }

.question-card {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 34px 30px;
  box-shadow: var(--shadow);
}
.q-index { font-size: 13px; color: var(--primary); font-weight: 700; margin-bottom: 14px; }
.q-text { font-size: 20px; font-weight: 700; line-height: 1.5; margin-bottom: 26px; }
.options { display: flex; flex-direction: column; gap: 14px; }
.option {
  display: flex;
  align-items: center;
  gap: 14px;
  text-align: left;
  background: var(--bg-soft);
  border: 1.5px solid var(--line);
  border-radius: 12px;
  padding: 17px 18px;
  color: var(--text);
  font-size: 15px;
  line-height: 1.5;
  transition: all 0.15s;
}
.option:hover { border-color: var(--primary); background: var(--card-hover); }
.option.picked {
  border-color: var(--primary);
  background: rgba(124, 108, 255, 0.14);
  box-shadow: 0 0 0 3px var(--primary-glow);
}
.opt-key {
  width: 32px; height: 32px;
  flex-shrink: 0;
  display: grid; place-items: center;
  border-radius: 50%;
  background: var(--card);
  border: 1.5px solid var(--line);
  font-weight: 700;
  font-size: 14px;
}
.option.picked .opt-key { background: var(--primary); border-color: var(--primary); color: #fff; }

.q-enter-active, .q-leave-active { transition: all 0.2s; }
.q-enter-from { opacity: 0; transform: translateX(30px); }
.q-leave-to { opacity: 0; transform: translateX(-30px); }

.quiz-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 22px;
}
.dots { display: flex; flex-wrap: wrap; gap: 6px; justify-content: center; flex: 1; }
.dots span {
  width: 9px; height: 9px;
  border-radius: 50%;
  background: var(--line);
  cursor: pointer;
  transition: all 0.15s;
}
.dots span.done { background: var(--primary-soft); }
.dots span.active { background: var(--cyan); transform: scale(1.35); }

/* 结果页 */
.result-wrap { padding-top: 26px; }
.result-hero {
  text-align: center;
  background: linear-gradient(160deg, rgba(124,108,255,0.18), rgba(86,200,255,0.08));
  border: 1px solid var(--line);
  border-radius: 20px;
  padding: 44px 30px;
  margin-bottom: 20px;
}
.hero-label { color: var(--text-dim); font-size: 14px; }
.hero-type {
  font-size: 72px;
  font-weight: 800;
  letter-spacing: 6px;
  line-height: 1.15;
  background: linear-gradient(135deg, #b3a6ff, var(--cyan));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.hero-name { font-size: 18px; font-weight: 700; margin-bottom: 12px; }
.hero-summary { max-width: 560px; margin: 0 auto; color: var(--text-dim); font-size: 14px; line-height: 1.8; }

.panel {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  padding: 24px 26px;
  margin-bottom: 18px;
}
.panel h3 { margin-bottom: 18px; font-size: 16px; }

.dim-row { margin-bottom: 16px; }
.dim-row:last-child { margin-bottom: 0; }
.dim-labels {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 13px;
  color: var(--text-faint);
  margin-bottom: 7px;
}
.dim-labels .on { color: var(--text); font-weight: 700; font-size: 14px; }
.dim-labels b { color: var(--cyan); font-size: 15px; }
.dim-track { display: flex; height: 12px; border-radius: 999px; overflow: hidden; background: var(--bg-soft); }
.dim-neg { background: linear-gradient(90deg, #3a3f7a, var(--cyan)); }
.dim-pos { background: linear-gradient(90deg, var(--primary), #b3a6ff); }

.chips { display: flex; flex-wrap: wrap; gap: 10px; margin-bottom: 20px; }
.chip {
  background: var(--bg-soft);
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 6px 16px;
  font-size: 13px;
}
.pf-cols { display: grid; grid-template-columns: 1fr 1fr; gap: 22px; }
.pf-cols h4 { font-size: 14px; margin-bottom: 10px; }
.pf-cols ul { list-style: none; }
.pf-cols li {
  color: var(--text-dim);
  font-size: 13.5px;
  line-height: 1.6;
  padding: 4px 0 4px 18px;
  position: relative;
}
.pf-cols li::before { content: '·'; position: absolute; left: 6px; color: var(--primary); font-weight: 800; }

.result-actions { display: flex; gap: 12px; justify-content: center; margin-top: 26px; }

@media (max-width: 600px) {
  .pf-cols { grid-template-columns: 1fr; }
  .hero-type { font-size: 52px; }
  .quiz-nav { flex-wrap: wrap; }
  .dots { order: 3; width: 100%; }
}
</style>
