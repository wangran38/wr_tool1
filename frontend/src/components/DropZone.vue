<script setup>
import { computed } from 'vue'

const props = defineProps({
  label: { type: String, required: true },
  modelValue: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue', 'browse'])

const fileName = computed(() => {
  if (!props.modelValue) return ''
  const parts = props.modelValue.split(/[\\/]/)
  return parts[parts.length - 1]
})

function clear() {
  emit('update:modelValue', '')
}
</script>

<template>
  <div
    class="flex min-h-[72px] cursor-pointer flex-col justify-center rounded border border-dashed p-3 text-center transition"
    :class="modelValue ? 'border-blue-400 bg-blue-50' : 'border-gray-300 bg-white hover:bg-gray-50'"
    @click="emit('browse')"
  >
    <template v-if="fileName">
      <div class="font-medium text-gray-800">{{ fileName }}</div>
      <div class="mt-1 text-xs text-gray-500">{{ label }}</div>
      <button class="mt-1 text-xs text-red-500 hover:underline" @click.stop="clear">移除</button>
    </template>
    <template v-else>
      <div class="text-gray-500">{{ label }}</div>
      <div class="mt-1 text-xs text-gray-400">点击选择，或直接拖入窗口（.docx / .pdf / .txt）</div>
    </template>
  </div>
</template>
