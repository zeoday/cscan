<template>
  <el-card class="pro-table-wrapper" shadow="never">
    <!-- Toolbar area (for search, filters, actions) -->
    <div class="pro-table-toolbar">
      <div class="toolbar-left">
        <slot name="toolbar-left"></slot>
      </div>
      <div class="toolbar-right">
        <slot name="toolbar-right"></slot>
      </div>
    </div>

    <!-- Main Table -->
    <el-table :data="tableData" v-bind="$attrs">
      <template v-for="(col, index) in columns" :key="index">
        <!-- If column has a slot -->
        <el-table-column v-if="col.slot" v-bind="col">
          <template #default="{ row }">
            <slot :name="col.slot" :row="row"></slot>
          </template>
        </el-table-column>

        <!-- Default rendering -->
        <el-table-column v-else v-bind="col">
          <template #default="{ row }">
            {{ row[col.prop] }}
          </template>
        </el-table-column>
      </template>
    </el-table>

    <!-- Pagination or footer (placeholder for now) -->
    <div class="pro-table-footer">
      <slot name="footer"></slot>
    </div>
  </el-card>
</template>

<script setup>
import { ref } from 'vue'

defineOptions({
  name: 'ProTable',
  inheritAttrs: false
})

const props = defineProps({
  api: {
    type: Function,
    default: null
  },
  columns: {
    type: Array,
    default: () => []
  }
})

// Will hold the actual data fetched from the API (for later stages)
const tableData = ref([])

// Expose anything parent might need (for future)
defineExpose({
  tableData
})
</script>

<style scoped lang="scss">
.pro-table-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;

  :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    padding: 16px;
    height: 100%;
    box-sizing: border-box;
  }
}

.pro-table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .toolbar-left, .toolbar-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.pro-table-footer {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
