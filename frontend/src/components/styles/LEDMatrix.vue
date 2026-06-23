<template>
  <div class="led-matrix" :style="{ background: config.bg_color || 'transparent' }">
    <div class="led-display" :style="displayStyle">
      <span
        v-for="(row, ri) in ledGrid"
        :key="ri"
        class="led-row"
      >
        <span
          v-for="(led, ci) in row"
          :key="ci"
          class="led-dot"
          :style="{
            width: dotSize + 'px',
            height: dotSize + 'px',
            backgroundColor: led ? ledOnColor : ledOffColor,
            boxShadow: led ? `0 0 ${dotSize}px ${ledOnColor}` : 'none',
            borderRadius: dotSize * 0.3 + 'px',
          }"
        />
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
      default: () => ({ led_color_on: '#FF0000', led_color_off: '#330000', dot_size: 8, font_size: 64, bg_color: 'transparent' })
    }
  },
  data() {
    return {
      // 5x7 font for digits (simplified)
      digitPatterns: {
        '0': ['111', '101', '101', '101', '111'],
        '1': ['010', '110', '010', '010', '111'],
        '2': ['111', '001', '111', '100', '111'],
        '3': ['111', '001', '111', '001', '111'],
        '4': ['101', '101', '111', '001', '001'],
        '5': ['111', '100', '111', '001', '111'],
        '6': ['111', '100', '111', '101', '111'],
        '7': ['111', '001', '010', '010', '010'],
        '8': ['111', '101', '111', '101', '111'],
        '9': ['111', '101', '111', '001', '111'],
        '-': ['000', '000', '111', '000', '000'],
      }
    }
  },
  computed: {
    dotSize() { return this.config.dot_size || 8 },
    ledOnColor() { return this.config.led_color_on || '#FF0000' },
    ledOffColor() { return this.config.led_color_off || '#330000' },
    displayValue() {
      const val = this.heartRate
      if (!val) return '--'
      return String(val).padStart(3, ' ')
    },
    ledGrid() {
      const chars = this.displayValue.split('')
      const rows = []
      for (let row = 0; row < 5; row++) {
        const line = []
        for (const ch of chars) {
          const pattern = this.digitPatterns[ch] || this.digitPatterns['-']
          const patRow = pattern[row] || '000'
          for (const px of patRow) {
            line.push(px === '1')
          }
          // Gap between digits
          line.push(false)
        }
        rows.push(line)
      }
      return rows
    },
    displayStyle() {
      const gap = Math.max(1, this.dotSize * 0.3)
      return {
        gap: gap + 'px',
      }
    }
  }
}
</script>

<style scoped>
.led-matrix {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.led-display {
  display: flex;
  flex-direction: column;
}
.led-row {
  display: flex;
  gap: inherit;
}
.led-dot {
  transition: background-color 0.15s, box-shadow 0.15s;
}
</style>
