<script lang="ts" setup>
import { useI18n } from 'vue-i18n'
import { ref, computed, onUnmounted } from 'vue'

import { useBool, useMessage } from '@/hooks'
import { DraggableOptions } from '@/constant'
import { useAppSettingsStore, type Menu } from '@/stores'
import type { KernelConnectionsWS } from '@/api/kernel.schema'
import { formatBytes, formatRelativeTime } from '@/utils'
import { getKernelConnectionsWS, deleteConnection } from '@/api/kernel'

import type { Column } from '@/components/Table/index.vue'

type TrafficCacheType = { up: number; down: number }
const TrafficCache: Record<string, TrafficCacheType> = {}

const appSettingsStore = useAppSettingsStore()

const columns = computed(() =>
  (
    [
      {
        title: 'home.connections.host',
        key: 'metadata.host',
        hidden: appSettingsStore.app.connections.visibility['metadata.host'],
        sort: (a, b) => b.metadata.host.localeCompare(a.metadata.host),
        customRender: ({ value, record }) => {
          return (value || record.metadata.destinationIP) + ':' + record.metadata.destinationPort
        }
      },
      {
        title: 'home.connections.inbound',
        align: 'center',
        key: 'metadata.type',
        hidden: appSettingsStore.app.connections.visibility['metadata.type'],
        sort: (a, b) => b.metadata.type.localeCompare(a.metadata.type),
        customRender: ({ value }) => value
      },
      {
        title: 'home.connections.rule',
        align: 'center',
        key: 'rule',
        hidden: appSettingsStore.app.connections.visibility['rule'],
        sort: (a, b) => b.rule.localeCompare(a.rule),
        customRender: ({ value, record }) => {
          return value + (record.rulePayload ? '::' + record.rulePayload : '')
        }
      },
      {
        title: 'home.connections.chains',
        key: 'chains',
        hidden: appSettingsStore.app.connections.visibility['chains'],
        sort: (a, b) => b.chains[0].localeCompare(a.chains[0]),
        customRender: ({ value }) => value[0]
      },
      {
        title: 'home.connections.uploadSpeed',
        align: 'center',
        key: 'up',
        minWidth: '90px',
        hidden: appSettingsStore.app.connections.visibility['up'],
        sort: (a, b) => b.upload - b.up - (a.upload - a.up),
        customRender: ({ value, record }) => formatBytes((record.upload - value) / 1000) + '/s'
      },
      {
        title: 'home.connections.downSpeed',
        align: 'center',
        key: 'down',
        minWidth: '90px',
        hidden: appSettingsStore.app.connections.visibility['down'],
        sort: (a, b) => b.download - b.down - (a.download - a.down),
        customRender: ({ value, record }) => formatBytes((record.download - value) / 1000) + '/s'
      },
      {
        title: 'home.connections.upload',
        align: 'center',
        key: 'upload',
        hidden: appSettingsStore.app.connections.visibility['upload'],
        sort: (a, b) => b.upload - a.upload,
        customRender: ({ value }) => formatBytes(value)
      },
      {
        title: 'home.connections.download',
        align: 'center',
        key: 'download',
        hidden: appSettingsStore.app.connections.visibility['download'],
        sort: (a, b) => b.download - a.download,
        customRender: ({ value }) => formatBytes(value)
      },
      {
        title: 'home.connections.time',
        align: 'center',
        key: 'start',
        hidden: appSettingsStore.app.connections.visibility['start'],
        sort: (a, b) => new Date(a.start).getTime() - new Date(b.start).getTime(),
        customRender: ({ value }) => formatRelativeTime(value)
      }
    ] as Column[]
  ).sort(
    (a, b) =>
      appSettingsStore.app.connections.order.indexOf(a.key) -
      appSettingsStore.app.connections.order.indexOf(b.key)
  )
)

const columnTitleMap = computed(() => {
  const map: Record<string, string | undefined> = {}
  appSettingsStore.app.connections.order.forEach(
    (field) => (map[field] = columns.value.find((column) => column.key === field)?.title)
  )
  return map
})

const menu: Menu[] = [
  {
    label: 'common.details',
    handler: (record: Record<string, any>) => {
      details.value = JSON.stringify(record, null, 2)
      toggleDetails()
    }
  },
  {
    label: 'home.connections.close',
    handler: async (record: Record<string, any>) => {
      if (!isActive.value) return
      try {
        await deleteConnection(record.id)
      } catch (error: any) {
        console.log(error)
        message.error(error)
      }
    }
  }
  // ...[
  //   ['home.connections.addToDirect', 'direct'],
  //   ['home.connections.addToProxy', 'proxy'],
  //   ['home.connections.addToReject', 'block']
  // ].map(([label, ruleset]) => {
  //   return {
  //     label,
  //     handler: async (record: Record<string, any>) => {
  //       try {
  //         const behavior = record.metadata.host ? 'domain' : 'ip_cidr'
  //         const payload = record.metadata.host || record.metadata.destinationIP + '/32'
  //         await addToRuleSet(ruleset as any, behavior + ',' + payload)
  //         await ignoredError(updateProvidersRules, ruleset)
  //         message.success('common.success')
  //       } catch (error: any) {
  //         message.error(error)
  //         console.log(error)
  //       }
  //     }
  //   }
  // })
]

const details = ref()
const isActive = ref(true)
const keywords = ref('')
const dataSource = ref<(KernelConnectionsWS['connections'][0] & TrafficCacheType)[]>([])
const disconnectedData = ref<KernelConnectionsWS['connections']>([])
const [showDetails, toggleDetails] = useBool(false)
const [showSettings, toggleSettings] = useBool(false)
const [isPause, togglePause] = useBool(false)
const { message } = useMessage()
const { t } = useI18n()

const onConnections = (data: KernelConnectionsWS) => {
  if (isPause.value) return
  const connections = data.connections || []

  dataSource.value.forEach((connection) => {
    // Record Disconnected Connections
    const exist = connections.some((v) => v.id === connection.id)
    !exist && disconnectedData.value.push(connection)
  })

  dataSource.value = connections.map((connection) => {
    // Record Previous Traffic Information
    const result = { ...connection, up: 0, down: 0 }
    const cache = TrafficCache[connection.id]
    result.up = cache?.up || connection.upload
    result.down = cache?.down || connection.download
    TrafficCache[connection.id] = {
      down: connection.download,
      up: connection.upload
    }
    return result
  })
}

const filteredConnections = computed(() => {
  if (!keywords.value) return isActive.value ? dataSource.value : disconnectedData.value
  return (isActive.value ? dataSource.value : disconnectedData.value).filter((connection) =>
    connection.metadata.host.includes(keywords.value)
  )
})

const handleCloseAll = async () => {
  if (!isActive.value) return
  await Promise.all(dataSource.value.map((connection) => deleteConnection(connection.id)))
}

const disconnect = getKernelConnectionsWS(onConnections)

onUnmounted(disconnect)
</script>

<template>
  <div class="connections">
    <div class="tools">
      <Radio
        v-model="isActive"
        :options="[
          { label: 'home.connections.active', value: true },
          { label: 'home.connections.closed', value: false }
        ]"
        size="small"
      />
      <Input
        v-model="keywords"
        size="small"
        placeholder="Search"
        style="margin-left: 8px; flex: 1"
      />
      <Button @click="togglePause" size="small" type="text" style="margin-left: 8px">
        <Icon :icon="isPause ? 'play' : 'pause'" fill="var(--color)" />
      </Button>
      <Button @click="handleCloseAll" v-tips="'home.connections.closeAll'" size="small" type="text">
        <Icon icon="close" fill="var(--color)" />
      </Button>
      <Button @click="toggleSettings" size="small" type="text">
        <Icon icon="settings" />
      </Button>
    </div>
    <Table
      :columns="columns"
      :menu="menu"
      :data-source="filteredConnections"
      sort="start"
      style="margin-top: 8px"
    />
  </div>

  <Modal
    v-model:open="showDetails"
    :submit="false"
    cancel-text="common.close"
    max-height="90"
    max-width="90"
    mask-closable
  >
    <CodeViewer v-model="details" />
  </Modal>

  <Modal
    v-model:open="showSettings"
    :submit="false"
    cancel-text="common.close"
    mask-closable
    title="home.connections.sort"
  >
    <div v-draggable="[appSettingsStore.app.connections.order, DraggableOptions]">
      <Card
        v-for="column in appSettingsStore.app.connections.order"
        :key="column"
        class="field-item"
      >
        <span class="name">{{ t(columnTitleMap[column] || column) }}</span>
        <Switch v-model="appSettingsStore.app.connections.visibility[column]" class="action" />
      </Card>
    </div>
  </Modal>
</template>

<style lang="less" scoped>
.connections {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.tools {
  display: flex;
  align-items: center;
}

.field-item {
  display: flex;
  align-items: center;
  padding: 0 8px;
  margin-bottom: 2px;
  .name {
    font-weight: bold;
  }
  .action {
    margin-left: auto;
  }
}
</style>
