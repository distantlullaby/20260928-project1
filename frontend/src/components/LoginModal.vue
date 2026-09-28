<script setup>
import { ref } from 'vue'
import { auth } from '../store/auth'
import { toast } from '../composables/toast'

const emit = defineEmits(['closed', 'success'])
const mode = ref('login') // login | register
const username = ref('')
const password = ref('')
const nickname = ref('')
const loading = ref(false)

async function submit() {
  if (!username.value.trim() || !password.value) {
    toast('请输入用户名和密码', 'error')
    return
  }
  loading.value = true
  try {
    if (mode.value === 'login') {
      const u = await auth.login(username.value.trim(), password.value)
      toast(`欢迎回来，${u.nickname}`)
    } else {
      const u = await auth.register(username.value.trim(), password.value, nickname.value.trim())
      toast(`注册成功，已赠送 ${u.coinBalance} 测评币 🎉`)
    }
    emit('success')
    emit('closed')
  } catch (e) {
    toast(e.message, 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="modal-mask" @click.self="emit('closed')">
    <div class="modal">
      <button class="modal-close" @click="emit('closed')">×</button>
      <h3>{{ mode === 'login' ? '登录' : '注册' }}</h3>
      <p class="sub">
        {{ mode === 'login' ? '登录后发布求解读、帮他人解读赚测评币' : '注册即赠送 100 测评币，开启人格探索之旅' }}
      </p>

      <div class="field">
        <label>用户名</label>
        <input class="input" v-model="username" placeholder="3-20 个字符" maxlength="20"
          @keyup.enter="submit" />
      </div>
      <div v-if="mode === 'register'" class="field">
        <label>昵称（选填）</label>
        <input class="input" v-model="nickname" placeholder="默认与用户名相同" maxlength="20" />
      </div>
      <div class="field">
        <label>密码</label>
        <input class="input" type="password" v-model="password" placeholder="至少 6 位"
          @keyup.enter="submit" />
      </div>

      <button class="btn btn-primary btn-block" :disabled="loading" @click="submit">
        {{ loading ? '请稍候…' : (mode === 'login' ? '登录' : '注册并领取测评币') }}
      </button>
      <p class="switch">
        {{ mode === 'login' ? '还没有账号？' : '已有账号？' }}
        <a @click="mode = mode === 'login' ? 'register' : 'login'">
          {{ mode === 'login' ? '立即注册' : '去登录' }}
        </a>
      </p>
    </div>
  </div>
</template>

<style scoped>
.switch {
  margin-top: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-dim);
}
.switch a { color: var(--primary); cursor: pointer; font-weight: 600; }
</style>
