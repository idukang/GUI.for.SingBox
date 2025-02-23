<script setup lang="ts">
import { ref } from 'vue'

import { APP_TITLE, APP_VERSION, APP_CHANNEL, exitApp } from '@/utils'
import { type Menu, useAppSettingsStore, useKernelApiStore, useEnvStore } from '@/stores'
import {
  WindowFullscreen,
  WindowIsFullscreen,
  WindowUnfullscreen,
  WindowSetAlwaysOnTop,
  WindowHide,
  WindowSetSize,
  WindowCenter,
  WindowReloadApp
} from '@/utils/bridge'

const isPinned = ref(false)
const isFullScreen = ref(false)

const appSettingsStore = useAppSettingsStore()
const kernelApiStore = useKernelApiStore()
const envStore = useEnvStore()

const toggleFullScreen = async () => {
  const isFull = await WindowIsFullscreen()
  isFull ? WindowUnfullscreen() : WindowFullscreen()
  isFullScreen.value = !isFull
}

const pinWindow = () => {
  isPinned.value = !isPinned.value
  WindowSetAlwaysOnTop(isPinned.value)
}

const closeWindow = async () => {
  if (appSettingsStore.app.exitOnClose) {
    exitApp()
  } else {
    WindowHide()
  }
}

const menus: Menu[] = [
  {
    label: 'titlebar.resetSize',
    handler: async () => {
      WindowUnfullscreen()
      await WindowSetSize(800, 540)
      WindowCenter()
      isFullScreen.value = false
    }
  },
  {
    label: 'titlebar.reload',
    handler: WindowReloadApp
  }
]
</script>

<template>
  <div
    v-if="envStore.env.os === 'windows'"
    @dblclick="toggleFullScreen"
    class="titlebar"
    style="--wails-draggable: drag"
  >
    <img class="logo" draggable="false" src="@/assets/logo.png" />
    <div
      :style="{
        color: appSettingsStore.app.kernel.running ? 'var(--primary-color)' : 'var(--color)'
      }"
      class="appname"
    >
      {{ APP_TITLE }} {{ APP_VERSION }} {{ APP_CHANNEL === 'Stable' ? '' : APP_CHANNEL }}
    </div>
    <Button v-if="kernelApiStore.loading" loading type="text" size="small" />
    <div v-menu="menus" class="menus"></div>
    <div class="action" style="--wails-draggable: disabled">
      <Button @click.stop="pinWindow" type="text">
        <Icon :icon="isPinned ? 'pinFill' : 'pin'" />
      </Button>
      <Button @click.stop="WindowHide" type="text">
        <Icon icon="minimize" />
      </Button>
      <Button @click.stop="toggleFullScreen" type="text">
        <Icon :icon="isFullScreen ? 'maximize2' : 'maximize'" />
      </Button>
      <Button @click.stop="closeWindow" type="text">
        <Icon icon="close" />
      </Button>
    </div>
  </div>
  <div v-else class="placeholder" style="--wails-draggable: drag">
    <div
      :style="{
        color: appSettingsStore.app.kernel.running ? 'var(--primary-color)' : 'var(--color)'
      }"
      v-menu="menus"
      class="appname"
    >
      {{ APP_TITLE }} {{ APP_VERSION }} {{ APP_CHANNEL === 'Stable' ? '' : APP_CHANNEL }}
    </div>
    <Button v-if="kernelApiStore.loading" loading type="text" size="small" />
  </div>
</template>

<style lang="less" scoped>
.titlebar {
  user-select: none;
  display: flex;
  padding: 4px 12px;
  align-items: center;
}
.logo {
  width: 24px;
  height: 24px;
  user-select: none;
}
.appname {
  font-size: 14px;
  margin-left: 8px;
  margin-top: 2px;
  font-weight: bold;
}

.menus {
  flex: 1;
  height: 100%;
}
.action {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  &-btn {
    width: 32px;
    height: 32px;
    line-height: 32px;
    text-align: center;
    border-radius: 4px;
    &:hover {
      background-color: var(--hover-bg-color);
    }
  }
}
.placeholder {
  user-select: none;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  .appname {
    font-size: 12px;
  }
}
</style>
