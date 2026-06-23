<template>
  <div class="ecg-wave" :style="{ background: config.bg_color || 'rgba(0,0,0,0.3)' }">
    <canvas ref="canvas" class="ecg-canvas"></canvas>
    <div class="ecg-value" :style="{ fontSize: (config.font_size || 36) + 'px', color: config.font_color || '#fff' }">
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
      default: () => ({ line_color: '#00FF00', line_width: 2, bg_color: 'rgba(0,0,0,0.3)', grid_color: 'rgba(255,255,255,0.1)', history_secs: 10 })
    }
  },
  data() {
    return {
      history: [],
      maxPoints: 200,
    }
  },
  computed: {
    displayValue() { return this.heartRate || '--' }
  },
  watch: {
    heartRate(val) {
      if (val) {
        this.history.push({ time: Date.now(), value: val })
        if (this.history.length > this.maxPoints) {
          this.history.shift()
        }
        this.drawCanvas()
      }
    }
  },
  mounted() {
    this.drawCanvas()
  },
  methods: {
    drawCanvas() {
      const canvas = this.$refs.canvas
      if (!canvas) return
      const ctx = canvas.getContext('2d')
      const w = canvas.width = canvas.offsetWidth * (window.devicePixelRatio || 1)
      const h = canvas.height = canvas.offsetHeight * (window.devicePixelRatio || 1)
      ctx.scale(window.devicePixelRatio || 1, window.devicePixelRatio || 1)
      const cw = canvas.offsetWidth
      const ch = canvas.offsetHeight

      ctx.clearRect(0, 0, cw, ch)

      // Grid lines
      const gridColor = this.config.grid_color || 'rgba(255,255,255,0.1)'
      ctx.strokeStyle = gridColor
      ctx.lineWidth = 0.5
      for (let i = 0; i < 5; i++) {
        const y = (ch / 5) * i
        ctx.beginPath()
        ctx.moveTo(0, y)
        ctx.lineTo(cw, y)
        ctx.stroke()
      }

      // ECG line
      if (this.history.length < 2) return

      const lineColor = this.config.line_color || '#00FF00'
      const lineWidth = this.config.line_width || 2

      // Calculate value range
      let minVal = 40, maxVal = 200
      if (this.history.length > 0) {
        const vals = this.history.map(h => h.value)
        minVal = Math.max(30, Math.min(...vals) - 10)
        maxVal = Math.min(220, Math.max(...vals) + 10)
      }

      const valRange = maxVal - minVal || 1
      const now = Date.now()
      const historySecs = this.config.history_secs || 10
      const timeWindow = historySecs * 1000

      ctx.strokeStyle = lineColor
      ctx.lineWidth = lineWidth
      ctx.lineJoin = 'round'
      ctx.beginPath()

      let firstPoint = true
      for (const pt of this.history) {
        const x = cw - ((now - pt.time) / timeWindow) * cw
        if (x < -10) continue
        const y = ch - ((pt.value - minVal) / valRange) * ch
        if (firstPoint) {
          ctx.moveTo(x, y)
          firstPoint = false
        } else {
          ctx.lineTo(x, y)
        }
      }
      ctx.stroke()

      // Glow effect
      ctx.shadowBlur = 6
      ctx.shadowColor = lineColor
      ctx.stroke()
      ctx.shadowBlur = 0
    }
  }
}
</script>

<style scoped>
.ecg-wave {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ecg-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}
.ecg-value {
  position: relative;
  z-index: 1;
  font-weight: 700;
  text-shadow: 0 0 10px rgba(0,0,0,0.7);
  background: rgba(0,0,0,0.3);
  padding: 4px 12px;
  border-radius: 4px;
}
</style>
