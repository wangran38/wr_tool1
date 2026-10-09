<script setup>
import { onMounted, ref } from 'vue'
import DropZone from '@/components/DropZone.vue'
import Toolbar from '@/components/Toolbar.vue'
import DiffViewer from '@/components/DiffViewer.vue'
import LicenseModal from '@/components/LicenseModal.vue'
import {
  CompareFiles,
  ChooseExportPath,
  ExportReport,
  LicenseStatus,
  PickDocument,
} from '@wailsjs/go/main/App'
import { OnFileDrop } from '@wailsjs/runtime/runtime'

const fileA = ref('')
const fileB = ref('')
const options = ref({
  ignoreWhitespace: true,
  ignoreLineBreaks: true,
  ignoreCase: false,
  ignorePunctuation: false,
  ignoreNumbering: true,
})
const result = ref(null)
const busy = ref(false)
const error = ref('')
const machineCode = ref('')
const activated = ref(false)
const trialDays = ref(0)
const expired = ref(false)
const showLicense = ref(false)

onMounted(async () => {
  await refreshLicense()
  // zero width/height registers the whole window as the drop target
  OnFileDrop((paths) => {
    if (!fileA.value) fileA.value = paths[0]
    else if (!fileB.value) fileB.value = paths[0]
    result.value = null
  }, 0, 0, 0, 0)
})

async function refreshLicense() {
  try {
    const st = await LicenseStatus()
    machineCode.value = st.machineCode
    activated.value = st.activated
    trialDays.value = st.trialDays
    expired.value = st.expired
  } catch (e) {
    error.value = String(e)
  }
}

async function onActivated() {
  showLicense.value = false
  await refreshLicense()
}

async function pick(target) {
  const path = await PickDocument()
  if (!path) return
  if (target === 'a') fileA.value = path
  else fileB.value = path
  result.value = null
}

async function run() {
  busy.value = true
  error.value = ''
  try {
    result.value = await CompareFiles(fileA.value, fileB.value, options.value)
  } catch (e) {
    error.value = String(e)
    await refreshLicense()
    if (expired.value) showLicense.value = true
  } finally {
    busy.value = false
  }
}

async function exportReport(format) {
  if (!result.value) return
  error.value = ''
  try {
    const out = await ChooseExportPath(result.value.newFile, format)
    if (!out) return
    await ExportReport(result.value, out, format)
  } catch (e) {
    error.value = String(e)
  }
}
</script>

<template>
  <div class="flex h-screen flex-col bg-gray-50 text-sm">
    <header class="flex items-center justify-between border-b border-gray-200 bg-white px-4 py-2">
      <div class="text-base font-semibold">合同比对工具</div>
      <div class="flex items-center gap-3">
        <span :class="activated ? 'text-green-600' : expired ? 'text-red-600' : 'text-amber-600'">
          {{ activated ? '已激活' : expired ? '试用已到期' : `试用中·剩 ${trialDays} 天` }}
        </span>
        <button class="rounded border border-gray-300 px-2 py-1 hover:bg-gray-100" @click="showLicense = true">
          激活
        </button>
      </div>
    </header>

    <section class="grid grid-cols-2 gap-3 border-b border-gray-200 bg-white p-3">
      <DropZone v-model="fileA" label="原文（旧版）" @browse="pick('a')" />
      <DropZone v-model="fileB" label="对比文（新版）" @browse="pick('b')" />
    </section>

    <Toolbar v-model="options" :disabled="!fileA || !fileB || busy" @compare="run" @export="exportReport" />

    <p v-if="error" class="bg-red-50 px-4 py-2 text-red-700">{{ error }}</p>

    <main class="min-h-0 flex-1 overflow-hidden">
      <DiffViewer v-if="result" :result="result" />
      <div v-else class="flex h-full items-center justify-center text-gray-400">
        {{ busy ? '比对中…' : '拖入或选择两个合同文件后开始比对' }}
      </div>
    </main>

    <LicenseModal
      v-if="showLicense"
      :machine-code="machineCode"
      :trial-days="trialDays"
      :activated="activated"
      :expired="expired"
      @close="showLicense = false"
      @activated="onActivated"
    />
  </div>
</template>
