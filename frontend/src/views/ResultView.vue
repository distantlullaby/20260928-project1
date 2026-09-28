<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import TypeBadge from '../components/TypeBadge.vue'
import DimensionBars from '../components/DimensionBars.vue'
import { useUserStore } from '../stores/user'
import { toast } from '../composables/toast'

const route = useRoute()
const router = useRouter()
const store = useUserStore()

const result = ref(null)
const profile = ref(null)
const loading = ref(true)

const splitList = (s) => (s || '').split(/[、,]/).filter(Boolean)
const strengths = computed(() => splitList(profile.value?.strengths))
const weaknesses = computed(() => splitList(profile.value?.weaknesses))
const careers = computed(() => splitList(profile.value?.careers))
const traits = computed(() => splitList(profile.value?.traits))

onMounted(async () => {
  try {
    const data = await api.get(`/results/${route.params.id}`)
    result.value = data
    profile.value = data.profile
  } catch (e) {
    toast(e.message, 'err')
  } finally {
    loading.value = false
  }
})

const showPublish = ref(false)
const form = ref({ title: '', content: '', reward: 20 })
const publishing = ref(false)
async function publish() {
  publishing.value = true
  try {
    const r = await api.post('/requests', {
      assessment_result_id: Number(route.params.id),
      title: form.value.title,
      content: form.value.content,
      reward: Number(form.value.reward)
    })
    toast('已发布，悬赏币已冻结')
    router.push('/')
  } catch (e) {
    toast(e.message, 'err')
  } finally {
    publishing.value = false
  }
}
</script>

<template>
  <main class="page">
    <div class="container result-wrap">
      <div v-if="loading" class="empty"><span class="emoji">⏳</span>正在生成你的人格报告…</div>

      <template v-else-if="result">
        <div v-if="route.query.fresh" class="form-ok celebrate">🎉 测评完成！这是属于你的性格画像</div>

        <!-- 类型头卡 -->
        <section class="hero-card card" :style="{ '--type-color': profile?.color || '#6c5ce7' }">
          <TypeBadge :code="result.type_code" size="lg" />
          <div class="hero-info">
            <h1>
              {{ result.type_code }}
              <span class="nick">{{ profile?.nickname }}</span>
            </h1>
            <p>
              <span class="tag tag-purple">{{ profile?.group }}</span>
              <span class="tag tag-gray">{{ profile?.group_key }} 型家族</span>
            </p>
            <p class="desc">{{ profile?.description }}</p>
          </div>
          <div class="hero-cta">
            <button class="btn btn-gold" @click="showPublish = true">📣 发布求解读</button>
            <RouterLink class="btn btn-outline" to="/assess">重新测一次</RouterLink>
          </div>
        </section>

        <!-- 四维占比 -->
        <section class="card block">
          <div class="section-title"><h2>四维倾向占比</h2><span class="sub">占比越悬殊，倾向越明显</span></div>
          <DimensionBars :dimensions="result.dimensions" />
        </section>

        <!-- 画像网格 -->
        <div class="grid-2">
          <section class="card block">
            <div class="section-title"><h2>✨ 性格关键词</h2></div>
            <div class="chips">
              <span v-for="t in traits" :key="t" class="chip purple">{{ t }}</span>
            </div>
            <div class="section-title" style="margin-top:22px"><h2>💪 优势天赋</h2></div>
            <ul class="line-list good">
              <li v-for="s in strengths" :key="s">{{ s }}</li>
            </ul>
            <div class="section-title" style="margin-top:22px"><h2>⚠️ 需要留意</h2></div>
            <ul class="line-list warn">
              <li v-for="w in weaknesses" :key="w">{{ w }}</li>
            </ul>
          </section>

          <section class="card block">
            <div class="section-title"><h2>🧭 可能适合的方向</h2></div>
            <div class="chips">
              <span v-for="c in careers" :key="c" class="chip gold">{{ c }}</span>
            </div>
            <div class="insight-box">
              <h3>💡 下一步：借他人视角看见盲区</h3>
              <p>
                标准化测评描述的是「类型中的你」，而真实的你还活在具体的关系与职场里。
                发布一条求解读悬赏，让了解 MBTI 的朋友结合你的真实困惑给出观察——
                类型画像负责定位，他人解读负责立体。
              </p>
              <button class="btn btn-primary btn-block" @click="showPublish = true">用 {{ form.reward }} 币发布求解读</button>
            </div>
          </section>
        </div>
      </template>

      <div v-else class="card empty"><span class="emoji">🔍</span>未找到该测评结果</div>
    </div>

    <!-- 发布弹层 -->
    <div v-if="showPublish" class="modal-mask" @click.self="showPublish = false">
      <div class="modal">
        <h3>📣 发布求解读</h3>
        <p v-if="!store.isLoggedIn" class="form-error">请先登录后再发布</p>
        <template v-else>
          <div class="field">
            <label>标题（你想被解读的困惑）</label>
            <input v-model="form.title" class="input" maxlength="100" placeholder="例如：这个类型在职场如何发挥优势？" />
          </div>
          <div class="field">
            <label>补充描述（选填）</label>
            <textarea v-model="form.content" class="textarea" placeholder="说说你的具体情境"></textarea>
          </div>
          <div class="field">
            <label>悬赏测评币</label>
            <input v-model.number="form.reward" type="number" min="1" :max="store.coins" class="input" />
            <p class="field-hint">当前可用 {{ store.coins }} 币。发布即冻结，采纳后结算给解读人；无人解读可关闭退回。</p>
          </div>
          <div class="modal-actions">
            <button class="btn btn-outline" @click="showPublish = false">取消</button>
            <button class="btn btn-primary" :disabled="publishing || !form.title || form.reward > store.coins || form.reward < 1" @click="publish">
              {{ publishing ? '发布中…' : `确认发布并冻结 ${form.reward || 0} 币` }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </main>
</template>

<style scoped>
.result-wrap { max-width: 960px; }
.celebrate { text-align: center; font-size: 15px; margin-bottom: 16px; }
.hero-card {
  display: flex; gap: 22px; align-items: flex-start;
  padding: 30px 32px; margin-bottom: 20px;
  border-top: 4px solid var(--type-color);
}
.hero-info { flex: 1; }
.hero-info h1 { font-size: 34px; font-weight: 900; display: flex; align-items: baseline; gap: 12px; flex-wrap: wrap; }
.hero-info .nick { font-size: 19px; color: var(--ink-soft); font-weight: 700; }
.hero-info p { margin-top: 8px; }
.hero-info p.desc { color: var(--ink-soft); font-size: 15px; margin-top: 14px; line-height: 1.8; }
.hero-cta { display: flex; flex-direction: column; gap: 9px; flex: none; }
.block { padding: 26px 30px; margin-bottom: 20px; }
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
.chips { display: flex; flex-wrap: wrap; gap: 8px; }
.chip { padding: 6px 15px; border-radius: 999px; font-size: 13.5px; font-weight: 600; }
.chip.purple { background: var(--primary-soft); color: var(--primary-deep); }
.chip.gold { background: var(--gold-soft); color: #b9760b; }
.line-list { list-style: none; display: flex; flex-direction: column; gap: 9px; }
.line-list li { position: relative; padding-left: 26px; font-size: 14.5px; color: var(--ink-soft); line-height: 1.6; }
.line-list li::before { position: absolute; left: 0; font-weight: 800; }
.line-list.good li::before { content: '✓'; color: var(--accent); }
.line-list.warn li::before { content: '!'; color: var(--gold); }
.insight-box {
  margin-top: 22px; border-radius: 14px; padding: 18px 20px;
  background: linear-gradient(135deg, var(--primary-soft), #e8faf5);
}
.insight-box h3 { font-size: 15.5px; margin-bottom: 8px; }
.insight-box p { font-size: 13.5px; color: var(--ink-soft); margin-bottom: 14px; line-height: 1.8; }

@media (max-width: 800px) {
  .hero-card { flex-direction: column; }
  .hero-cta { flex-direction: row; width: 100%; }
  .grid-2 { grid-template-columns: 1fr; }
}
</style>
