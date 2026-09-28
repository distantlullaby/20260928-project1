<script setup>
import { computed } from 'vue'

const props = defineProps({
  code: { type: String, default: '' },
  size: { type: String, default: 'md' }
})

const colors = {
  NT: ['#6366f1', '#818cf8'], NF: ['#10b981', '#34d399'],
  SJ: ['#f59e0b', '#fbbf24'], SP: ['#f97316', '#fb923c']
}
const groupOf = (code) => {
  if (!code || code.length < 4) return ['#6c5ce7', '#9b8cff']
  const k = code[1] + code[2]
  return colors[k] || colors.NT
}
const gradient = computed(() => {
  const [a, b] = groupOf(props.code)
  return { background: `linear-gradient(135deg, ${a}, ${b})` }
})
</script>

<template>
  <span class="type-badge" :class="size" :style="gradient">{{ code || '----' }}</span>
</template>

<style scoped>
.type-badge {
  display: inline-grid; place-items: center;
  color: #fff; font-weight: 800; letter-spacing: 1px;
  border-radius: 12px; font-family: "Segoe UI", sans-serif;
  box-shadow: 0 6px 14px rgba(99,102,241,.28);
}
.type-badge.sm { width: 52px; height: 26px; font-size: 13px; border-radius: 8px; }
.type-badge.md { width: 64px; height: 32px; font-size: 16px; }
.type-badge.lg { width: 88px; height: 44px; font-size: 23px; border-radius: 14px; }
</style>
