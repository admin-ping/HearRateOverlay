<template>
  <div class="speed-gauge" :style="{ background: config.bg_color || 'transparent' }">
    <svg viewBox="0 0 200 140" class="gauge-svg">
      <!-- Gauge arc background -->
      <defs>
        <linearGradient id="gaugeGrad" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" :style="{ stopColor: config.gauge_color || '#FF4444', stopOpacity: 0.3 }" />
          <stop offset="50%" :style="{ stopColor: config.gauge_color || '#FF4444', stopOpacity: 0.6 }" />
          <stop offset="100%" :style="{ stopColor: config.gauge_color || '#FF4444', stopOpacity: 1 }" />
        </linearGradient>
      </defs>
      <!-- Arc -->
      <path :d="arcPath" fill="none" stroke="url(#gaugeGrad)" :stroke-width="arcWidth" stroke-linecap="round" />
      <!-- Tick marks -->
      <line
        v-for="tick in ticks" :key="tick.angle"
        :x1="tick.x1" :y1="tick.y1" :x2="tick.x2" :y2="tick.y2"
        :stroke="tick.major ? 'rgba(255,255,255,0.4)' : 'rgba(255,255,255,0.15)'"
        :stroke-width="tick.major ? 2 : 1"
      />
      <!-- Needle -->
      <line
        :x1="cx" :y1="cy"
        :x2="needleX" :y2="needleY"
        :stroke="config.needle_color || '#FFFFFF'"
        stroke-width="2.5"
        stroke-linecap="round"
        class="needle"
      />
      <!-- Center dot -->
      <circle :cx="cx" :cy="cy" r="5" :fill="config.needle_color || '#FFFFFF'" />
    </svg>
    <div class="gauge-value" :style="{ fontSize: (config.font_size || 36) + 'px', color: config.font_color || '#fff' }">
      {{ displayValue }}
    </div>
  </div>
</template>

<script>
export default {
  props: {
    heartRate: { type: Number, default: 0 },
    config: {
      type: Object,
      default: () => ({ gauge_color: '#FF4444', needle_color: '#FFFFFF', font_size: 36, font_color: '#fff', max_value: 220 })
    }
  },
  computed: {
    displayValue() { return this.heartRate || '--' },
    cx() { return 100 },
    cy() { return 110 },
    radius() { return 80 },
    arcWidth() { return 14 },
    maxVal() { return this.config.max_value || 220 },
    valueAngle() {
      const pct = Math.min(1, Math.max(0, (this.heartRate || 0) / this.maxVal))
      return Math.PI + pct * Math.PI // 180° to 360° (bottom half)
    },
    arcPath() {
      const r = this.radius
      const startX = this.cx - r
      const endX = this.cx + r
      const y = this.cy
      return `M ${startX} ${y} A ${r} ${r} 0 0 1 ${endX} ${y}`
    },
    needleX() {
      return this.cx + this.radius * Math.cos(this.valueAngle)
    },
    needleY() {
      return this.cy + this.radius * Math.sin(this.valueAngle)
    },
    ticks() {
      const result = []
      for (let i = 0; i <= 10; i++) {
        const angle = Math.PI + (i / 10) * Math.PI
        const inner = this.radius - this.arcWidth
        const outer = this.radius + 4
        result.push({
          angle,
          x1: this.cx + inner * Math.cos(angle),
          y1: this.cy + inner * Math.sin(angle),
          x2: this.cx + outer * Math.cos(angle),
          y2: this.cy + outer * Math.sin(angle),
          major: i % 2 === 0,
        })
      }
      return result
    }
  }
}
</script>

<style scoped>
.speed-gauge {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.gauge-svg {
  position: absolute;
  width: 100%;
  height: 100%;
}
.needle {
  transition: all 0.3s ease;
}
.gauge-value {
  position: relative;
  z-index: 1;
  font-weight: 700;
  margin-top: 30px;
  text-shadow: 0 0 10px rgba(0,0,0,0.5);
}
</style>
