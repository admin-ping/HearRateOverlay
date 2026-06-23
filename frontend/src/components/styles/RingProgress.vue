<template>
  <div class="ring-progress" :style="{ background: config.bg_color || 'transparent' }">
    <svg viewBox="0 0 200 200" class="ring-svg">
      <!-- Background ring -->
      <circle
        cx="100" cy="100" :r="radius"
        fill="none"
        :stroke="bgRingColor"
        :stroke-width="ringWidth"
      />
      <!-- Progress ring -->
      <circle
        cx="100" cy="100" :r="radius"
        fill="none"
        :stroke="config.ring_color || '#00FF88'"
        :stroke-width="ringWidth"
        :stroke-dasharray="circumference"
        :stroke-dashoffset="dashOffset"
        stroke-linecap="round"
        class="progress-ring"
      />
    </svg>
    <div class="ring-center">
      <span class="ring-value" :style="{ fontSize: (config.font_size || 48) + 'px', color: config.font_color || '#fff' }">
        {{ displayValue }}
      </span>
      <span v-if="config.show_percent" class="ring-pct" :style="{ color: config.font_color || '#fff' }">
        {{ percent }}%
      </span>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    heartRate: { type: Number, default: 0 },
    config: {
      type: Object,
      default: () => ({ ring_color: '#00FF88', ring_width: 8, font_size: 48, font_color: '#fff', show_percent: true })
    }
  },
  computed: {
    displayValue() { return this.heartRate || '--' },
    radius() { return 85 },
    ringWidth() { return this.config.ring_width || 8 },
    circumference() { return 2 * Math.PI * this.radius },
    maxHR() { return this.config.max_value || 220 },
    percent() {
      if (!this.heartRate) return 0
      return Math.min(100, Math.round((this.heartRate / this.maxHR) * 100))
    },
    dashOffset() {
      return this.circumference * (1 - this.percent / 100)
    },
    bgRingColor() {
      return 'rgba(255,255,255,0.1)'
    }
  }
}
</script>

<style scoped>
.ring-progress {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ring-svg {
  position: absolute;
  width: 100%;
  height: 100%;
}
.progress-ring {
  transition: stroke-dashoffset 0.5s ease;
}
.ring-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  z-index: 1;
}
.ring-value {
  font-weight: 700;
  line-height: 1;
  text-shadow: 0 0 10px rgba(0,0,0,0.5);
}
.ring-pct {
  font-size: 14px;
  opacity: 0.7;
  margin-top: 2px;
}
</style>
