<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { toast } from '../composables/toast'

const router = useRouter()
const route = useRoute()
const store = useUserStore()

const mode = ref('login')
const form = ref({ username: '', password: '', nickname: '' })
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'login') {
      await store.login(form.value.username, form.value.password)
      toast('欢迎回来')
    } else {
      await store.register(form.value.username, form.value.password, form.value.nickname)
      toast('注册成功，已赠送 200 测评币')
    }
    router.push(route.query.redirect || '/')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function fillDemo(name) {
  mode.value = 'login'
  form.value.username = name
  form.value.password = 'demo123'
}
</script>

<template>
  <main class="page">
    <div class="container auth-wrap">
      <div class="auth-card card">
        <div class="auth-brand">
          <div class="brand-emoji">🪞</div>
          <h2>{{ mode === 'login' ? '欢迎回到镜我' : '开启自我探索' }}</h2>
          <p>{{ mode === 'login' ? '登录后继续测评与解读' : '注册即送 200 枚测评币' }}</p>
        </div>

        <div class="mode-switch">
          <button :class="{ on: mode === 'login' }" @click="mode = 'login'">登录</button>
          <button :class="{ on: mode === 'register' }" @click="mode = 'register'">注册</button>
        </div>

        <form @submit.prevent="submit">
          <p v-if="error" class="form-error">{{ error }}</p>
          <div class="field">
            <label>用户名</label>
            <input v-model="form.username" class="input" autocomplete="username"
                   placeholder="3-20 位用户名" />
          </div>
          <div v-if="mode === 'register'" class="field">
            <label>昵称（选填）</label>
            <input v-model="form.nickname" class="input" maxlength="20" placeholder="在社区里如何称呼你" />
          </div>
          <div class="field">
            <label>密码</label>
            <input v-model="form.password" type="password" class="input" autocomplete="current-password"
                   placeholder="至少 6 位" />
          </div>
          <button class="btn btn-primary btn-block" type="submit" :disabled="loading">
            {{ loading ? '请稍候…' : mode === 'login' ? '登录' : '注册并领取 200 币' }}
          </button>
        </form>

        <div class="demo-box">
          <p>演示账号（密码均为 <b>demo123</b>）：</p>
          <div class="demo-accounts">
            <button @click="fillDemo('linxia')">林小夏 · INFP</button>
            <button @click="fillDemo('zhouye')">周野 · ENTJ</button>
            <button @click="fillDemo('chenan')">陈安 · ISFJ</button>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.auth-wrap { max-width: 460px; }
.auth-card { padding: 34px 34px 28px; }
.auth-brand { text-align: center; margin-bottom: 22px; }
.brand-emoji { font-size: 46px; margin-bottom: 8px; }
.auth-brand h2 { font-size: 23px; font-weight: 900; }
.auth-brand p { color: var(--ink-faint); font-size: 14px; margin-top: 4px; }
.mode-switch {
  display: flex; background: #f0eff8; border-radius: 999px; padding: 4px; margin-bottom: 22px;
}
.mode-switch button {
  flex: 1; border: none; background: transparent; border-radius: 999px;
  padding: 9px 0; font-size: 15px; font-weight: 700; color: var(--ink-faint); transition: .15s;
}
.mode-switch button.on { background: #fff; color: var(--primary-deep); box-shadow: 0 4px 12px rgba(0,0,0,.08); }
.demo-box {
  margin-top: 22px; padding-top: 18px; border-top: 1px dashed var(--border);
  text-align: center;
}
.demo-box p { font-size: 13px; color: var(--ink-faint); margin-bottom: 10px; }
.demo-accounts { display: flex; gap: 8px; justify-content: center; flex-wrap: wrap; }
.demo-accounts button {
  border: 1.5px solid var(--border); background: #fff; border-radius: 999px;
  padding: 5px 13px; font-size: 12.5px; color: var(--ink-soft);
}
.demo-accounts button:hover { border-color: var(--primary); color: var(--primary); }
</style>
