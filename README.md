# ❤️ HeartRateOverlay

开源、免费、直连蓝牙心率设备的桌面悬浮窗。超低延迟，丰富样式，OBS 直播友好。

[中文](#中文) | [English](#english)

---

## 中文

### 简介

HeartRateOverlay 是一款桌面端心率实时显示工具，通过 PC 蓝牙直接接收心率手环/胸带的广播数据，以透明悬浮窗展示。可作为 [Pulsoid](https://pulsoid.net) 的开源替代品 — 无需手机中转，延迟 < 100ms，完全本地运行，不收集任何数据。

### 特性

- **🎯 直连蓝牙** — PC 蓝牙直接读取 BLE 心率设备（0x180D 服务），无需手机
- **🪟 透明悬浮窗** — 无边框、始终置顶、全透明背景，OBS 窗口采集完美支持
- **🎨 6 种内置样式** — 极简数字、环形进度、速度表盘、心电图曲线、LED 点阵、霓虹
- **📐 样式可扩展** — 通过 `styles/` 目录下的 YAML + Vue 组件添加新样式
- **🗂️ 场景管理** — 多场景独立保存位置/大小/样式/置顶设置，一键切换
- **📊 数据统计** — 会话记录、实时平均/最大/最小、5 区心率分布（热身/燃脂/有氧/无氧/极限）
- **💾 本地存储** — SQLite 数据库 + YAML 配置文件，所有数据留在本地
- **🔔 系统托盘** — 托盘图标快速操作：设置 / 悬浮窗 / 退出
- **🌐 中英双语** — i18n 框架，初始提供中文和英文
- **🔄 自动重连** — BLE 断连后指数退避自动重连（1s ~ 30s）
- **🔍 更新检查** — 启动时通过 GitHub API 检查新版本

### 系统要求

| 平台 | 要求 |
|------|------|
| Windows | 10/11，蓝牙 4.0+ 适配器 |
| macOS | 11+，CoreBluetooth |
| Linux | Ubuntu 20.04+，BlueZ 5.50+ |

### 快速开始

从 [Releases](https://github.com/admin-ping/HearRateOverlay/releases) 下载最新版本。

**Windows**：解压 `HearRateOverlay.zip`，运行 `HearRateOverlay.exe`

**从源码构建**：
```bash
# 前置依赖
# Go 1.21+, Node.js 18+, Wails CLI

# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 构建
git clone https://github.com/admin-ping/HearRateOverlay.git
cd HearRateOverlay
wails build
```

### 使用说明

1. 启动应用 → 打开**设置窗口**（800×600）
2. 点击「🔍 扫描设备」→ 选择你的心率设备 → 点击「连接」
3. 连接成功后点击「🚀 打开悬浮窗」→ 独立透明悬浮窗弹出
4. 悬浮窗可拖动、始终置顶，在 OBS 中添加**窗口采集**并勾选「允许透明度」
5. 右键悬浮窗或点击 ✕ 关闭；从系统托盘可随时重新打开

### 添加新样式

1. 在 `styles/` 下新建文件夹，如 `styles/my-style/`
2. 创建 `style.yaml`：
```yaml
name: "我的样式"
version: 1
component: "MyStyle"
default:
  font_size: 60
  font_color: "#FF8800"
  bg_color: "transparent"
```
3. 在 `frontend/src/components/styles/` 下创建 Vue 组件 `MyStyle.vue`，接收 `heartRate` 和 `config` props
4. 重启应用即可在样式中选择

### 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.21+ |
| GUI 框架 | Wails v2 |
| 前端 | Vue 3 + TypeScript + Vite |
| 蓝牙 | tinygo.org/x/bluetooth |
| 数据库 | modernc.org/sqlite（纯 Go，无 CGO） |
| 配置 | YAML (gopkg.in/yaml.v3) |
| 系统托盘 | github.com/getlantern/systray |

### 项目结构

```
HearRateOverlay/
├── main.go                      # 入口（设置模式 / --overlay 悬浮窗模式）
├── app.go                       # 设置窗口逻辑 + Wails 绑定
├── overlay_app.go               # 悬浮窗进程逻辑
├── internal/
│   ├── ble/                     # 蓝牙管理（扫描/连接/解析/重连）
│   ├── config/                  # 配置 + 场景管理（YAML）
│   ├── db/                      # SQLite 持久层
│   ├── i18n/                    # 中英双语
│   ├── state/                   # 双进程共享状态
│   ├── stats/                   # 会话统计
│   ├── style/                   # 样式引擎
│   ├── tray/                    # 系统托盘
│   └── update/                  # GitHub 更新检查
├── frontend/src/
│   ├── App.vue                  # 主界面（设置 / 悬浮窗双模式）
│   └── components/styles/       # 6 种样式组件
├── styles/                      # 样式 YAML 定义
└── build/                       # 构建配置
```

### 常见问题

**Q: 扫描不到设备？**
- 确保电脑蓝牙已开启
- 确保心率设备处于广播模式（部分手环需手动开启）
- 尝试靠近设备

**Q: 悬浮窗在 OBS 中是黑底/毛玻璃？**
- OBS 窗口采集 → 勾选「允许透明度」
- 如仍有问题，尝试 `wails build -debug` 模式排查

**Q: macOS/Linux 支持？**
- 项目已预留跨平台代码，但主要测试在 Windows
- macOS/Linux 用户可通过 `wails build` 自行构建

### 许可证

MIT License

---

## English

### Overview

HeartRateOverlay is a desktop heart rate monitor that connects directly to BLE heart rate devices via PC Bluetooth and displays your BPM in a transparent overlay window. An open-source alternative to Pulsoid — no phone required, < 100ms latency, fully local.

### Features

- BLE direct connection (0x180D Heart Rate Service)
- Transparent frameless always-on-top overlay, OBS capture ready
- 6 built-in visual styles, extensible via YAML + Vue components
- Scene management with independent position/size/style/opacity
- Session recording with real-time stats and 5-zone heart rate distribution
- SQLite local storage, no data collection
- System tray quick actions
- Chinese/English i18n
- Auto-reconnect with exponential backoff
- GitHub update checker

### Quick Start

Download from [Releases](https://github.com/admin-ping/HearRateOverlay/releases) or build from source:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
git clone https://github.com/admin-ping/HearRateOverlay.git
cd HearRateOverlay
wails build
```

### License

MIT
