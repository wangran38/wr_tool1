<script setup>
import { ref } from 'vue'
import { VerifyLicense } from '@wailsjs/go/main/App'

const props = defineProps({
  machineCode: { type: String, required: true },
  trialDays: { type: Number, default: 30 },
  activated: { type: Boolean, default: false },
  expired: { type: Boolean, default: false },
})
const emit = defineEmits(['close', 'activated'])

const code = ref('')
const error = ref('')
const busy = ref(false)

async function submit() {
  if (!code.value.trim()) {
    error.value = '请输入激活码'
    return
  }
  busy.value = true
  error.value = ''
  try {
    await VerifyLicense(code.value.trim())
    emit('activated')
  } catch (e) {
    error.value = '激活码无效，请核对机器码后重试'
  } finally {
    busy.value = false
  }
}

async function copy() {
  await navigator.clipboard.writeText(props.machineCode)
}
</script>

<template>
  <div class="fixed inset-0 z-10 flex items-center justify-center bg-black/40" @click.self="emit('close')">
    <div class="w-[420px] rounded-lg bg-white p-5 shadow-lg">
      <h2 class="mb-3 text-base font-semibold">离线激活</h2>

      <p v-if="activated" class="mb-3 rounded bg-green-50 px-3 py-2 text-green-700">
        本机已激活，无需再次输入激活码。
      </p>
      <p v-else class="mb-3 rounded bg-amber-50 px-3 py-2 leading-relaxed text-amber-800">
        <template v-if="expired">试用已到期，比对功能已锁定。</template>
        <template v-else>试用期内可正常使用，剩余 {{ trialDays }} 天。</template>
        请加微信
        <span class="font-semibold select-all">wangran38</span>
        ，把下面的机器码发给对方换取激活码。
      </p>

      <p class="mb-1 text-gray-600">本机机器码</p>
      <div class="mb-3 flex items-center gap-2">
        <code class="flex-1 select-all break-all rounded bg-gray-100 px-2 py-1">{{ machineCode }}</code>
        <button class="rounded border border-gray-300 px-2 py-1 text-xs hover:bg-gray-100" @click="copy">复制</button>
      </div>

      <p class="mb-1 text-gray-600">激活码</p>
      <input
        v-model="code"
        class="mb-3 w-full rounded border border-gray-300 px-2 py-1.5 tracking-wider"
        placeholder="XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"
        @keyup.enter="submit"
      />

      <p v-if="error" class="mb-2 text-red-600">{{ error }}</p>

      <div class="flex justify-end gap-2">
        <button class="rounded border border-gray-300 px-3 py-1.5 hover:bg-gray-100" @click="emit('close')">取消</button>
        <button
          class="rounded bg-blue-600 px-3 py-1.5 text-white hover:bg-blue-700 disabled:bg-gray-300"
          :disabled="busy"
          @click="submit"
        >
          {{ busy ? '校验中…' : '激活' }}
        </button>
      </div>
    </div>
  </div>
</template>
