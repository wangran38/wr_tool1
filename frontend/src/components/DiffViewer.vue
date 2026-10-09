<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  result: { type: Object, required: true },
})

const tab = ref('text')
const left = ref(null)
const right = ref(null)
let syncing = false

const stats = computed(() => props.result.text.stats)

const rowClass = (kind) => ({
  insert: 'diff-row-ins',
  delete: 'diff-row-del',
  replace: 'diff-row-rep',
}[kind] || '')

const segClass = (kind) => (kind === 'insert' ? 'seg-ins' : kind === 'delete' ? 'seg-del' : '')

function cellClass(row, col) {
  const change = (row.cells || []).find((c) => c.col === col)
  return change ? rowClass(change.kind) : ''
}

function sync(from, to) {
  if (syncing || !from || !to) return
  syncing = true
  to.scrollTop = from.scrollTop
  requestAnimationFrame(() => { syncing = false })
}
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="flex items-center gap-4 border-b border-gray-200 bg-white px-4 py-1.5 text-xs">
      <button :class="tab === 'text' && 'font-semibold text-blue-600'" @click="tab = 'text'">正文差异</button>
      <button
        v-if="result.tables && result.tables.length"
        :class="tab === 'table' && 'font-semibold text-blue-600'"
        @click="tab = 'table'"
      >
        表格差异（{{ result.tables.length }}）
      </button>
      <span class="ml-auto text-gray-500">
        一致 {{ stats.equal }} · 新增 {{ stats.insert }} · 删除 {{ stats.delete }} · 修改 {{ stats.replace }} ·
        相似度 {{ (stats.similarity * 100).toFixed(1) }}%
      </span>
    </div>

    <div v-if="tab === 'text'" class="grid min-h-0 flex-1 grid-cols-2">
      <div ref="left" class="overflow-auto border-r border-gray-200" @scroll="sync(left, right)">
        <table class="w-full table-fixed border-collapse">
          <tbody>
            <tr v-for="(p, i) in result.text.pairs" :key="i" :class="rowClass(p.kind)">
              <td class="w-1/12 border-b border-gray-100 px-2 text-right text-xs text-gray-400">{{ i + 1 }}</td>
              <td class="whitespace-pre-wrap break-all border-b border-gray-100 px-2">
                <template v-if="p.oldSegs && p.oldSegs.length">
                  <span v-for="(s, si) in p.oldSegs" :key="si" :class="segClass(s.kind)">{{ s.text }}</span>
                </template>
                <template v-else>{{ p.old || '（无）' }}</template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div ref="right" class="overflow-auto" @scroll="sync(right, left)">
        <table class="w-full table-fixed border-collapse">
          <tbody>
            <tr v-for="(p, i) in result.text.pairs" :key="i" :class="rowClass(p.kind)">
              <td class="w-1/12 border-b border-gray-100 px-2 text-right text-xs text-gray-400">{{ i + 1 }}</td>
              <td class="whitespace-pre-wrap break-all border-b border-gray-100 px-2">
                <template v-if="p.newSegs && p.newSegs.length">
                  <span v-for="(s, si) in p.newSegs" :key="si" :class="segClass(s.kind)">{{ s.text }}</span>
                </template>
                <template v-else>{{ p.new || '（无）' }}</template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-else class="min-h-0 flex-1 overflow-auto p-4">
      <section v-for="t in result.tables" :key="t.index" class="mb-6">
        <h3 class="mb-2 font-semibold">
          表格 {{ t.index + 1 }}
          <span class="ml-2 text-xs font-normal text-gray-500">{{ t.rows.length }} 处行级差异</span>
        </h3>
        <table class="w-full border-collapse text-xs">
          <thead>
            <tr class="bg-gray-50 text-left">
              <th class="w-1/2 border border-gray-200 px-2 py-1">原表格行</th>
              <th class="w-1/2 border border-gray-200 px-2 py-1">新表格行</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(r, ri) in t.rows" :key="ri" :class="rowClass(r.kind)">
              <td class="border border-gray-200 px-2 py-1">
                <span
                  v-for="(cell, ci) in r.old || []"
                  :key="ci"
                  class="mr-1 inline-block border-l border-gray-300 pl-1"
                  :class="cellClass(r, ci)"
                >{{ cell }}</span>
                <span v-if="r.kind === 'insert'" class="text-gray-400">（新增行）</span>
              </td>
              <td class="border border-gray-200 px-2 py-1">
                <span
                  v-for="(cell, ci) in r.new || []"
                  :key="ci"
                  class="mr-1 inline-block border-l border-gray-300 pl-1"
                  :class="cellClass(r, ci)"
                >{{ cell }}</span>
                <span v-if="r.kind === 'delete'" class="text-gray-400">（已删除行）</span>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
      <p v-if="!result.tables || !result.tables.length" class="text-gray-400">未检测到表格</p>
    </div>
  </div>
</template>
