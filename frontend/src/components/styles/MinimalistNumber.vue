<template>
  <div class="minimalist" :style="containerStyle">
    <span class="number" :style="numberStyle">{{ displayValue }}</span>
    <span v-if="config.show_unit" class="unit" :style="unitStyle">BPM</span>
  </div>
</template>

<script>
export default {
  props: {
    heartRate: { type: Number, default: 0 },
    config: {
      type: Object,
      default: () => ({ font_size: 80, font_color: '#FFFFFF', show_unit: true, animation: 'pulse' })
    }
  },
  computed: {
    displayValue() {
      return this.heartRate || '--'
    },
    containerStyle() {
      return {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'transparent',
      }
    },
    numberStyle() {
      return {
        fontSize: (this.config.font_size || 80) + 'px',
        color: this.config.font_color || '#FFFFFF',
        fontWeight: '700',
        fontFamily: '"Segoe UI", system-ui, sans-serif',
        lineHeight: 1,
        textShadow: this.config.animation === 'glow'
          ? `0 0 20px ${this.config.font_color || '#FFFFFF'}`
          : '0 2px 8px rgba(0,0,0,0.5)',
        animation: this.config.animation === 'pulse' ? 'pulse 1.5s ease-in-out infinite' : 'none',
      }
    },
    unitStyle() {
      return {
        fontSize: Math.floor((this.config.font_size || 80) * 0.25) + 'px',
        color: this.config.font_color || '#FFFFFF',
        opacity: 0.7,
        marginTop: '4px',
        letterSpacing: '4px',
      }
    }
  }
}
</script>

<style scoped>
.minimalist {
  width: 100%;
  height: 100%;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.03); }
}
</style>
