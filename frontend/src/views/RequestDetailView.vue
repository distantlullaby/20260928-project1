<script setup>
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'
import RequestCard from '../components/RequestCard.vue'
import { useUserStore } from '../stores/user'

const route = useRoute()
const store = useUserStore()
const req = ref(null)

async function load() {
  const { request } = await api.get(`/requests/${route.params.id}`)
  req.value = request
}
onMounted(load)
</script>

<template>
  <main class="page">
    <div class="container detail-wrap">
      <div class="back-link">
        <RouterLink to="/" class="btn btn-outline btn-sm">← 返回大厅</RouterLink>
      </div>
      <RequestCard v-if="req" :req="req" @changed="load(); store.fetchMe()" />
      <div v-else class="card empty"><span class="emoji">⏳</span>加载中…</div>
    </div>
  </main>
</template>

<style scoped>
.detail-wrap { max-width: 820px; }
.back-link { margin-bottom: 16px; }
</style>
