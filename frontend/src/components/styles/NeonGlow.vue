<template>
  <div class="neon-glow" :style="{ background: config.bg_color || 'transparent' }">
    <span
      class="neon-text"
      :style="neonStyle"
    >{{ displayValue }}</span>
    <span v-if="config.show_unit" class="neon-unit" :style="unitStyle">BPM</span>
  </div>
</template>

<script>
export default {
  props: {
    heartRate: { type: Number, default: 0 },
    config: {
      type: Object,
      default: () => ({ glow_color: '#FF00FF', glow_radius: 20, font_size: 72, font_color: '#fff', show_unit: true })
    }
  },
  computed: {
    displayValue() { return this.heartRate || '--' },
    glowColor() { return this.config.glow_color || '#FF00FF' },
    glowRadius() { return (this.config.glow_radius || 20) + 'px' },
    fontSize() { return (this.config.font_size || 72) + 'px' },
    neonStyle() {
      const c = this.glowColor
      return {
        fontSize: this.fontSize,
        color: this.config.font_color || '#fff',
        fontWeight: '700',
        fontFamily: '"Courier New", monospace',
        textShadow: [
          `0 0 ${this.glowRadius} ${c}`,
          `0 0 ${parseInt(this.glowRadius) * 2}px ${c}`,
          `0 0 ${parseInt(this.glowRadius) * 4}px ${c}`,
          `0 0 2px #fff`,
        ].join(', '),
        animation: 'neonFlicker 3s ease-in-out infinite',
      }
    },
    unitStyle() {
      return {
        fontSize: Math.floor((this.config.font_size || 72) * 0.22) + 'px',
        color: this.glowColor,
        opacity: 0.8,
        marginTop: '8px',
        letterSpacing: '6px',
        textTransform: 'uppercase',
        textShadow: `0 0 10px ${this.glowColor}`,
      }
    }
  }
}
</script>

<style scoped>
.neon-glow {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.neon-text {
  line-height: 1;
}
.neon-unit {
  font-family: '"Courier New", monospace';
}

@keyframes neonFlicker {
  0%, 19%, 21%, 23%, 25%, 54%, 56%, 100% {
    opacity: 1;
  }
  20%, 24%, 55% {
    opacity: 0.85;
  }
}
</style>
