<script setup>
// 四维倾向占比条：dimensions 为后端 DimensionStat 数组
defineProps({
  dimensions: { type: Array, required: true },
  compact: { type: Boolean, default: false }
})
</script>

<template>
  <div class="dims" :class="{ compact }">
    <div v-for="d in dimensions" :key="d.pair" class="dim-row">
      <div class="dim-pole" :class="{ active: d.dominant === d.pole_a }">
        <b>{{ d.pole_a }}</b><span>{{ d.label_a }}</span>
      </div>
      <div class="dim-track-wrap">
        <div class="dim-percents">
          <span :class="{ on: d.dominant === d.pole_a }">{{ d.percent_a }}%</span>
          <span class="pair-name">{{ d.pair }}</span>
          <span :class="{ on: d.dominant === d.pole_b }">{{ d.percent_b }}%</span>
        </div>
        <div class="dim-track">
          <div class="dim-fill dim-fill-a" :style="{ width: d.percent_a + '%' }"></div>
          <div class="dim-mid"></div>
          <div class="dim-fill dim-fill-b" :style="{ width: d.percent_b + '%' }"></div>
        </div>
      </div>
      <div class="dim-pole right" :class="{ active: d.dominant === d.pole_b }">
        <b>{{ d.pole_b }}</b><span>{{ d.label_b }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dims { display: flex; flex-direction: column; gap: 16px; }
.dims.compact { gap: 10px; }
.dim-row { display: flex; align-items: center; gap: 12px; }
.dim-pole {
  width: 76px; display: flex; flex-direction: column; align-items: center;
  color: var(--ink-faint); line-height: 1.3;
}
.dim-pole.right { text-align: center; }
.dim-pole b { font-size: 19px; font-weight: 800; }
.dim-pole span { font-size: 12px; }
.dim-pole.active { color: var(--primary); }
.dim-pole.active b { font-size: 22px; }
.dim-track-wrap { flex: 1; }
.dim-percents {
  display: flex; align-items: center; justify-content: space-between;
  font-size: 12.5px; color: var(--ink-faint); font-weight: 700; margin-bottom: 4px;
}
.dim-percents .on { color: var(--primary); font-size: 14px; }
.dim-percents .pair-name { letter-spacing: 1px; font-weight: 600; opacity: .7; }
.dim-track {
  height: 10px; border-radius: 999px; background: #efedfa; overflow: hidden;
  display: flex; position: relative;
}
.dim-fill-a { background: linear-gradient(90deg, #8a78ff, var(--primary)); border-radius: 999px 0 0 999px; transition: width .5s ease; }
.dim-fill-b { background: linear-gradient(90deg, #2ec5a6, var(--accent)); border-radius: 0 999px 999px 0; transition: width .5s ease; }
.dim-mid { width: 2px; background: #fff; }
.compact .dim-pole { width: 58px; }
.compact .dim-pole b { font-size: 16px; }
.compact .dim-track { height: 8px; }
</style>
