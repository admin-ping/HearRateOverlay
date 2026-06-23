<template>
  <!-- ============================================================ -->
  <!-- OVERLAY MODE: separate process, shows only the style + BPM   -->
  <!-- ============================================================ -->
  <div v-if="isOverlayMode" class="overlay-root" @contextmenu.prevent="closeOverlay">
    <component
      :is="overlayComponent"
      :heart-rate="bpm"
      :config="styleCfg"
      class="style-comp"
    />
    <div v-if="!connected" class="overlay-placeholder">
      <span>--</span>
    </div>
    <button class="overlay-close no-drag" @click="closeOverlay" title="关闭悬浮窗">✕</button>
  </div>

  <!-- ============================================================ -->
  <!-- SETTINGS MODE: main control panel                           -->
  <!-- ============================================================ -->
  <div v-else class="settings-root">
    <div class="settings-header">
      <h2>❤️ HeartRateOverlay</h2>
      <div class="header-actions no-drag">
        <span class="status-badge" :class="connectionStatus">{{ connectionStatus }}</span>
        <button class="btn primary" @click="launchOverlay">
          🚀 打开悬浮窗
        </button>
        <button class="btn icon" @click="minimizeWindow" title="最小化">─</button>
        <button class="btn icon" @click="toggleMaximize" title="最大化">☐</button>
        <button class="btn icon danger" @click="closeWindow" title="关闭">✕</button>
      </div>
    </div>

    <div class="settings-body no-drag">
      <div class="settings-col">
        <div class="card">
          <h3>蓝牙设备</h3>
          <div class="card-row">
            <button @click="scanDevices" :disabled="scanning">
              {{ scanning ? '⏳ 扫描中...' : '🔍 扫描设备' }}
            </button>
            <button v-if="scanning" @click="stopScan" class="btn-danger btn-sm">停止</button>
            <button @click="disconnect" :disabled="!connected" class="btn-danger">断开</button>
          </div>
          <select v-if="devices.length" v-model="selectedDevice" class="full-width">
            <option value="">-- 选择设备 --</option>
            <option v-for="d in devices" :key="d.address" :value="d.address">
              {{ d.name || d.address }} (RSSI: {{ d.rssi }})
            </option>
          </select>
          <button v-if="selectedDevice" @click="connectToSelected" class="btn full-width mt-4">连接</button>
          <div v-if="!devices.length && !scanning" class="hint">点击「扫描设备」搜索附近的心率设备</div>
        </div>

        <div class="card">
          <h3>实时统计</h3>
          <div class="stats-grid">
            <div class="stat-item"><span class="stat-label">当前</span><span class="stat-value">{{ stats.current_bpm || '--' }}</span></div>
            <div class="stat-item"><span class="stat-label">10秒均</span><span class="stat-value">{{ stats.recent_avg ? stats.recent_avg.toFixed(0) : '--' }}</span></div>
            <div class="stat-item"><span class="stat-label">平均</span><span class="stat-value">{{ stats.avg_hr || '--' }}</span></div>
            <div class="stat-item"><span class="stat-label">最大</span><span class="stat-value">{{ stats.max_hr || '--' }}</span></div>
            <div class="stat-item"><span class="stat-label">最小</span><span class="stat-value">{{ stats.min_hr || '--' }}</span></div>
            <div class="stat-item"><span class="stat-label">时长</span><span class="stat-value small">{{ stats.duration || '--' }}</span></div>
          </div>
          <div v-if="stats.zones" class="zone-bar">
            <span v-for="(pct, zone) in stats.zones" :key="zone" class="zone-tag">{{ zone }} {{ pct.toFixed(0) }}%</span>
          </div>
          <button @click="resetSession" class="btn-sm mt-4">重置会话</button>
        </div>
      </div>

      <div class="settings-col">
        <div class="card">
          <h3>样式 ({{ styles.length }})</h3>
          <div class="style-grid">
            <button v-for="s in styles" :key="s.name" :class="{ active: currentStyle === s.name }" @click="switchStyle(s.name)" class="style-btn">{{ s.name }}</button>
          </div>
        </div>
        <div class="card">
          <h3>场景</h3>
          <div class="scene-list">
            <button v-for="s in scenes" :key="s.name" :class="{ active: currentScene === s.name }" @click="switchScene(s.name)" class="scene-btn">{{ s.name }}</button>
          </div>
          <div class="card-row mt-4">
            <input v-model="newSceneName" placeholder="新场景名" />
            <button @click="createScene" class="btn-sm">创建</button>
          </div>
        </div>
        <div class="card">
          <h3>设置</h3>
          <div class="setting-row"><label>最大心率</label><input v-model.number="config.max_hr" type="number" @change="updateMaxHR" /></div>
          <div class="setting-row"><label>语言</label><select v-model="config.language" @change="updateLang"><option value="zh">中文</option><option value="en">English</option></select></div>
          <div class="setting-row"><label>开机启动</label><input type="checkbox" v-model="config.auto_start" @change="updateConfigValue" /></div>
          <div class="setting-row"><label>检查更新</label><input type="checkbox" v-model="config.check_update" @change="updateConfigValue" /></div>
          <button @click="checkUpdate" class="btn-sm mt-4">检查更新</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import MinimalistNumber from './components/styles/MinimalistNumber.vue'
import RingProgress from './components/styles/RingProgress.vue'
import SpeedGauge from './components/styles/SpeedGauge.vue'
import ECGWave from './components/styles/ECGWave.vue'
import LEDMatrix from './components/styles/LEDMatrix.vue'
import NeonGlow from './components/styles/NeonGlow.vue'

const styleComponentMap = { MinimalistNumber, RingProgress, SpeedGauge, ECGWave, LEDMatrix, NeonGlow }

export default {
  components: styleComponentMap,
  setup() {
    // Determine mode by checking which Go backend is available
    const isOverlayMode = ref(!!window.go?.main?.OverlayApp)

    // Overlay state
    const bpm = ref(0)
    const connected = ref(false)
    const styleCfg = ref({})
    const overlayComponentName = ref('')

    // Overlay component: use component name from shared state directly
    const overlayComponent = computed(() => {
      if (!overlayComponentName.value) return null
      return styleComponentMap[overlayComponentName.value] || null
    })

    // Settings state
    const connectionStatus = ref('idle')
    const scanning = ref(false)
    const devices = ref([])
    const selectedDevice = ref('')
    const styles = ref([])
    const currentStyle = ref('')
    const scenes = ref([])
    const currentScene = ref('')
    const newSceneName = ref('')
    const stats = ref({})
    const config = ref({ max_hr: 185, language: 'zh', auto_start: false, check_update: true })

    let statsInterval = null

    const currentStyleComponent = computed(() => {
      if (!currentStyle.value) return null
      const s = styles.value.find(x => x.name === currentStyle.value)
      return s?.component || null
    })

    // ========== OVERLAY METHODS ==========
    async function closeOverlay() {
      try { await window.go.main.OverlayApp.CloseOverlay() } catch (e) { /* */ }
    }

    async function overlayPoll() {
      if (!isOverlayMode.value) return
      try { await window.go.main.OverlayApp.PollState() } catch (e) { /* */ }
    }

    // ========== SETTINGS METHODS ==========
    async function loadData() {
      try {
        styles.value = await window.go.main.App.GetStyles()
        currentStyle.value = await window.go.main.App.GetCurrentStyle()
        if (!currentStyle.value && styles.value.length > 0) currentStyle.value = styles.value[0].name
        scenes.value = await window.go.main.App.GetScenes()
        config.value = await window.go.main.App.GetConfig()
        connectionStatus.value = await window.go.main.App.GetConnectionStatus()
        connected.value = connectionStatus.value === 'subscribed' || connectionStatus.value === 'connected'
        updateStyleCfg()
        updateStats()
      } catch (e) { console.error(e) }
    }

    function updateStyleCfg() {
      const s = styles.value.find(x => x.name === currentStyle.value)
      if (s) styleCfg.value = s.default || {}
    }

    async function updateStats() { try { stats.value = await window.go.main.App.GetSessionStats() } catch (e) { /* */ } }
    async function scanDevices() { scanning.value = true; try { await window.go.main.App.StartScan() } catch (e) { console.error(e) } }
    async function stopScan() { try { await window.go.main.App.StopScan() } catch (e) { console.error(e) }; scanning.value = false }
    async function connectToSelected() { if (!selectedDevice.value) return; try { await window.go.main.App.Connect(selectedDevice.value) } catch (e) { console.error(e) } }
    async function disconnect() { try { await window.go.main.App.Disconnect() } catch (e) { console.error(e) } }
    async function switchStyle(name) { try { await window.go.main.App.SwitchStyle(name); currentStyle.value = name; updateStyleCfg() } catch (e) { console.error(e) } }
    async function switchScene(name) { try { await window.go.main.App.SwitchScene(name); currentScene.value = name } catch (e) { console.error(e) } }
    async function createScene() {
      if (!newSceneName.value) return
      try { await window.go.main.App.SaveScene({ name: newSceneName.value, style: currentStyle.value, font_size: 80, window: { x: 100, y: 100, width: 300, height: 200, always_on_top: true, opacity: 1.0 } }); scenes.value = await window.go.main.App.GetScenes(); newSceneName.value = '' } catch (e) { console.error(e) }
    }
    async function updateMaxHR() { try { await window.go.main.App.UpdateConfig(config.value) } catch (e) { console.error(e) } }
    async function updateConfigValue() { try { await window.go.main.App.UpdateConfig(config.value) } catch (e) { console.error(e) } }
    async function updateLang() { try { await window.go.main.App.SetLanguage(config.value.language); await window.go.main.App.UpdateConfig(config.value) } catch (e) { console.error(e) } }
    async function resetSession() { try { await window.go.main.App.ResetSession(); updateStats() } catch (e) { console.error(e) } }
    async function launchOverlay() { try { await window.go.main.App.LaunchOverlay() } catch (e) { console.error(e) } }
    function minimizeWindow() { window.runtime?.WindowMinimise() }
    function toggleMaximize() { window.runtime?.WindowToggleMaximise() }
    function closeWindow() { window.runtime?.Quit() }

    async function checkUpdate() {
      try { const r = await window.go.main.App.CheckForUpdate(); if (r.has_update) alert(`新版本: ${r.version}\n${r.url}`); else alert('已是最新版本') } catch (e) { alert('检查失败') }
    }

    onMounted(() => {
      if (isOverlayMode.value) {
        overlayPoll()
        if (window.runtime?.EventsOn) {
          window.runtime.EventsOn('state-update', (data) => {
            bpm.value = data.bpm
            connected.value = data.connected
            if (data.styleCfg) styleCfg.value = data.styleCfg
            if (data.styleName) currentStyle.value = data.styleName
            if (data.component) overlayComponentName.value = data.component
          })
        }
      } else {
        loadData()
        if (window.runtime?.EventsOn) {
          window.runtime.EventsOn('heart-rate-update', (data) => { bpm.value = data.bpm })
          window.runtime.EventsOn('connection-status', (s) => {
            connectionStatus.value = typeof s === 'string' ? s : 'unknown'
            connected.value = s === 'subscribed' || s === 'connected'
          })
          window.runtime.EventsOn('style-changed', (def) => { currentStyle.value = def.name; styleCfg.value = def.default || {} })
          window.runtime.EventsOn('scene-changed', (scene) => { currentScene.value = scene.name; if (scene.style) { currentStyle.value = scene.style; updateStyleCfg() } })
          window.runtime.EventsOn('scan-update', (list) => { devices.value = list || [] })
        }
        statsInterval = setInterval(updateStats, 2000)
      }
    })

    onUnmounted(() => { if (statsInterval) clearInterval(statsInterval) })

    return {
      isOverlayMode, bpm, connected, styleCfg, overlayComponentName, overlayComponent,
      connectionStatus, scanning, devices, selectedDevice, styles, currentStyle, currentStyleComponent, scenes, currentScene, newSceneName, stats, config,
      closeOverlay, scanDevices, stopScan, connectToSelected, disconnect, switchStyle, switchScene, createScene,
      updateMaxHR, updateConfigValue, updateLang, resetSession, launchOverlay, minimizeWindow, toggleMaximize, closeWindow, checkUpdate,
    }
  }
}
</script>

<style scoped>
/* ===== OVERLAY MODE ===== */
.overlay-root {
  width: 100%; height: 100%;
  display: flex; align-items: center; justify-content: center;
  background: transparent !important;
}
.style-comp { width: 100%; height: 100%; z-index: 1; }
.overlay-placeholder { position: absolute; z-index: 0; font-size: 72px; color: rgba(255,255,255,0.3); }
.overlay-close {
  position: absolute; top: 4px; right: 4px; z-index: 10;
  width: 22px; height: 22px; padding: 0; font-size: 12px; line-height: 1;
  background: rgba(0,0,0,0.4); border: none; color: #fff; border-radius: 3px; cursor: pointer;
}
.overlay-close:hover { background: rgba(200,0,0,0.7); }

/* ===== SETTINGS MODE ===== */
.settings-root { width: 100%; height: 100%; background: #1e1e28; color: #ddd; display: flex; flex-direction: column; overflow: hidden; font-size: 13px; }
.settings-header { display: flex; align-items: center; justify-content: space-between; padding: 12px 16px; background: #252530; border-bottom: 1px solid #333; }
.settings-header h2 { margin: 0; font-size: 18px; color: #fff; }
.header-actions { display: flex; align-items: center; gap: 10px; }
.status-badge { padding: 4px 10px; border-radius: 12px; font-size: 11px; text-transform: uppercase; background: #444; color: #aaa; }
.status-badge.subscribed, .status-badge.connected { background: #1a5; color: #fff; }
.settings-body { flex: 1; display: flex; gap: 12px; padding: 12px; overflow-y: auto; }
.settings-col { flex: 1; display: flex; flex-direction: column; gap: 12px; min-width: 0; }
.card { background: #2a2a36; border-radius: 8px; padding: 12px; border: 1px solid #333; }
.card h3 { margin: 0 0 8px 0; font-size: 13px; color: #8af; text-transform: uppercase; letter-spacing: 1px; }
.card-row { display: flex; gap: 6px; align-items: center; }
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 6px; }
.stat-item { text-align: center; }
.stat-label { font-size: 10px; color: #888; display: block; }
.stat-value { font-size: 22px; font-weight: 700; color: #fff; }
.stat-value.small { font-size: 13px; }
.zone-bar { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 6px; }
.zone-tag { font-size: 10px; background: #333; padding: 2px 6px; border-radius: 4px; color: #aaa; }
.style-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 4px; }
.style-btn { padding: 6px 4px; font-size: 11px; background: #333; border: 1px solid #444; color: #ccc; border-radius: 4px; cursor: pointer; }
.style-btn:hover { background: #444; }
.style-btn.active { background: #1a6; border-color: #2a8; color: #fff; }
.scene-list { display: flex; flex-wrap: wrap; gap: 4px; }
.scene-btn { padding: 4px 10px; font-size: 11px; background: #333; border: 1px solid #444; color: #ccc; border-radius: 4px; cursor: pointer; }
.scene-btn:hover { background: #444; }
.scene-btn.active { background: #26a; border-color: #48c; color: #fff; }
.setting-row { display: flex; align-items: center; justify-content: space-between; padding: 4px 0; }
.setting-row label { color: #aaa; font-size: 12px; }
.setting-row input[type="number"] { width: 70px; }
.setting-row select { width: 100px; }
.hint { font-size: 11px; color: #666; margin-top: 6px; }

button { background: #3a3a5c; border: 1px solid #555; color: #ddd; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-size: 12px; }
button:hover { background: #4a4a7c; }
button:disabled { opacity: 0.4; cursor: default; }
button.btn { padding: 8px 16px; font-size: 13px; }
button.btn.primary { background: #1a6; border-color: #2a8; color: #fff; }
button.btn.primary:hover { background: #2a8; }
button.btn.icon { width: 28px; height: 28px; padding: 0; font-size: 14px; line-height: 1; background: transparent; border: 1px solid #444; display: inline-flex; align-items: center; justify-content: center; }
button.btn.icon:hover { background: #444; }
button.btn.icon.danger:hover { background: #c44; border-color: #c44; }
button.btn-danger { background: #a22; border-color: #c44; }
button.btn-sm { padding: 3px 8px; font-size: 11px; }
button.full-width { width: 100%; }
button.active { background: #1a6; border-color: #2a8; color: #fff; }

select, input[type="text"], input[type="number"] { background: #333; border: 1px solid #555; color: #ddd; padding: 5px 8px; border-radius: 4px; font-size: 12px; }
select.full-width { width: 100%; }
.mt-4 { margin-top: 8px; }
</style>
