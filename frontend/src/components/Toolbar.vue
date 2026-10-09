<script setup>
const props = defineProps({
  modelValue: { type: Object, required: true },
  disabled: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue', 'compare', 'export'])
const filters = [
  { key: 'ignoreWhitespace', label: '忽略空格' },
  { key: 'ignoreLineBreaks', label: '忽略换行' },
  { key: 'ignoreCase', label: '忽略大小写' },
  { key: 'ignorePunctuation', label: '忽略标点' },
  { key: 'ignoreNumbering', label: '忽略条款编号' },
]

function toggle(key, value) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-4 border-b border-gray-200 bg-white px-4 py-2">
    <label v-for="f in filters" :key="f.key" class="flex items-center gap-1.5">
      <input
        type="checkbox"
        :checked="modelValue[f.key]"
        :disabled="disabled"
        @change="toggle(f.key, $event.target.checked)"
      />
      <span>{{ f.label }}</span>
    </label>

    <div class="ml-auto flex gap-2">
      <button
        class="rounded bg-blue-600 px-3 py-1.5 text-white hover:bg-blue-700 disabled:bg-gray-300"
        :disabled="disabled"
        @click="emit('compare')"
      >
        开始比对
      </button>
      <button
        class="rounded border border-gray-300 px-3 py-1.5 hover:bg-gray-100 disabled:text-gray-300"
        :disabled="disabled"
        @click="emit('export', 'html')"
      >
        导出 HTML
      </button>
      <button
        class="rounded border border-gray-300 px-3 py-1.5 hover:bg-gray-100 disabled:text-gray-300"
        :disabled="disabled"
        @click="emit('export', 'pdf')"
      >
        导出 PDF
      </button>
    </div>
  </div>
</template>
