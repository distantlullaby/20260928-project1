<template>
  <div class="toast-wrap">
    <div v-for="t in toastState.items" :key="t.id" class="toast" :class="`toast-${t.type}`">
      {{ t.message }}
    </div>
  </div>
</template>

<script setup>
import { toastState } from '../composables/toast'
</script>
