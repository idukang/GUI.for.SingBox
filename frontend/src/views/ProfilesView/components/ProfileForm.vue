<script setup lang="ts">
import { ref, inject } from 'vue'
import { useI18n } from 'vue-i18n'

import { deepClone, sampleID } from '@/utils'
import { useMessage, useBool } from '@/hooks'
import * as Defaults from '@/constant/profile'
import { type ProfileType, useProfilesStore } from '@/stores'

import GeneralConfig from '@/components/Profile/GeneralConfig.vue'
import AdvancedConfig from '@/components/Profile/AdvancedConfig.vue'
import TunConfig from '@/components/Profile/TunConfig.vue'
import DnsConfig from '@/components/Profile/DnsConfig.vue'
import ProxyGroupsConfig from '@/components/Profile/ProxyGroupsConfig.vue'
import DnsRulesConfig from '@/components/Profile/DnsRulesConfig.vue'
import RulesConfig from '@/components/Profile/RulesConfig.vue'

interface Props {
  id?: string
  step?: number
  isUpdate?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  id: '',
  isUpdate: false,
  step: 0
})

const currentStep = ref(props.step)

const stepItems = [
  { title: 'profile.step.name' },
  { title: 'profile.step.general' },
  { title: 'profile.step.tun' },
  { title: 'profile.step.groups' },
  { title: 'profile.step.rules' },
  { title: 'profile.step.dns' },
  { title: 'profile.step.dnsRules' }
]

const profile = ref<ProfileType>({
  id: sampleID(),
  name: '',
  generalConfig: Defaults.GeneralConfigDefaults,
  advancedConfig: Defaults.AdvancedConfigDefaults(),
  tunConfig: Defaults.TunConfigDefaults,
  dnsConfig: Defaults.DnsConfigDefaults(),
  proxyGroupsConfig: Defaults.ProxyGroupsConfigDefaults(),
  rulesConfig: Defaults.RulesConfigDefaults(),
  dnsRulesConfig: Defaults.DnsRulesConfigDefaults()
})

const { t } = useI18n()
const { message } = useMessage()
const profilesStore = useProfilesStore()
const [showAdvancedSetting, toggleAdvancedSetting] = useBool(false)

const handleCancel = inject('cancel') as any
const handleSubmit = inject('submit') as any

const handleSave = async () => {
  if (props.isUpdate) {
    try {
      await profilesStore.editProfile(props.id, profile.value)
      handleSubmit()
    } catch (error: any) {
      console.error('editProfile: ', error)
      message.error(error)
      return
    }
    return
  }

  try {
    await profilesStore.addProfile(profile.value)
    handleCancel()
  } catch (error: any) {
    console.error('addProfile: ')
    message.error(error)
    return
  }
}

const handlePrevStep = () => {
  currentStep.value--
}

const handleNextStep = async () => {
  currentStep.value++
}

if (props.isUpdate) {
  const p = profilesStore.getProfileById(props.id)
  if (p) {
    profile.value = deepClone(p)
  }
}
</script>

<template>
  <div class="title" style="--wails-draggable: drag">{{ t(stepItems[currentStep].title) }}</div>

  <div class="form">
    <div v-show="currentStep === 0" class="step1">
      <div class="form-item">
        <div class="name">{{ t('profile.name') }} *</div>
        <Input v-model="profile.name" auto-size autofocus style="flex: 1; margin-left: 8px" />
      </div>
    </div>

    <div v-show="currentStep === 1" class="step2">
      <GeneralConfig v-model="profile.generalConfig" />
      <Divider>
        <Button type="text" size="small" @click="toggleAdvancedSetting">
          {{ t('profile.advancedSettings') }}
        </Button>
      </Divider>
      <div v-if="showAdvancedSetting">
        <AdvancedConfig v-model="profile.advancedConfig" :profile="profile" />
      </div>
    </div>

    <div v-show="currentStep === 2" class="step3">
      <TunConfig v-model="profile.tunConfig" />
    </div>

    <div v-show="currentStep === 3" class="step4">
      <ProxyGroupsConfig v-model="profile.proxyGroupsConfig" />
    </div>

    <div v-show="currentStep === 4" class="step5">
      <RulesConfig
        v-model="profile.rulesConfig"
        :proxy-groups="profile.proxyGroupsConfig"
        :profile="profile"
      />
    </div>

    <div v-show="currentStep === 5" class="step6">
      <DnsConfig v-model="profile.dnsConfig" :proxy-groups="profile.proxyGroupsConfig" />
    </div>

    <div v-show="currentStep === 6" class="step7">
      <DnsRulesConfig
        v-model="profile.dnsRulesConfig"
        :dns-config="profile.dnsConfig"
        :proxy-groups="profile.proxyGroupsConfig"
      />
    </div>
  </div>

  <div class="form-action">
    <Button @click="handleCancel" type="text">{{ t('common.cancel') }}</Button>
    <Button @click="handlePrevStep" :disable="currentStep == 0" type="text">
      {{ t('common.prevStep') }}
    </Button>
    <Button
      @click="handleNextStep"
      :disable="!profile.name || currentStep == stepItems.length - 1"
      type="text"
    >
      {{ t('common.nextStep') }}
    </Button>
    <Button @click="handleSave" :disable="!profile.name" type="primary">
      {{ t('common.save') }}
    </Button>
  </div>
</template>

<style lang="less" scoped>
.title {
  font-size: 20px;
  font-weight: bold;
  margin: 0 0 16px 0;
}
.form {
  padding: 0 8px;
  overflow-y: auto;
  max-height: 60vh;
}
</style>
