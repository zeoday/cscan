<template>
  <div class="dirscan-view">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item :label="$t('dirscan.target')">
          <el-input v-model="searchForm.authority" :placeholder="$t('dirscan.targetPlaceholder')" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('dirscan.path')">
          <el-input v-model="searchForm.path" :placeholder="$t('dirscan.pathPlaceholder')" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('dirscan.statusCode')">
          <el-select v-model="searchForm.statusCode" :placeholder="$t('common.all')" clearable style="width: 120px">
            <el-option label="200" :value="200" />
            <el-option label="301" :value="301" />
            <el-option label="302" :value="302" />
            <el-option label="403" :value="403" />
            <el-option label="404" :value="404" />
            <el-option label="500" :value="500" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
          <el-button type="danger" plain @click="handleClear">{{ $t('dirscan.clearData') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 统计信息 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ stat.total }}</div>
          <div class="stat-label">{{ $t('dirscan.total') }}</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card status-2xx">
          <div class="stat-value">{{ stat.status_2xx || 0 }}</div>
          <div class="stat-label">2xx</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card status-3xx">
          <div class="stat-value">{{ stat.status_3xx || 0 }}</div>
          <div class="stat-label">3xx</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card status-4xx">
          <div class="stat-value">{{ stat.status_4xx || 0 }}</div>
          <div class="stat-label">4xx</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card status-5xx">
          <div class="stat-value">{{ stat.status_5xx || 0 }}</div>
          <div class="stat-label">5xx</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 按目标分组的折叠面板 -->
    <el-card class="collapse-card" v-loading="loading">
      <div class="collapse-header">
        <span class="total-info">{{ $t('dirscan.totalTargets', { targets: Object.keys(groupedData).length, records: pagination.total }) }}</span>
        <div class="collapse-actions">
          <el-button size="small" @click="expandAll">{{ $t('dirscan.expandAll') }}</el-button>
          <el-button size="small" @click="collapseAll">{{ $t('dirscan.collapseAll') }}</el-button>
          <el-dropdown style="margin-left: 10px" @command="handleExport">
            <el-button type="success" size="small">
              {{ $t('common.export') }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="all-url">{{ $t('dirscan.exportAllUrl') }}</el-dropdown-item>
                <el-dropdown-item command="csv">{{ $t('dirscan.exportCsv') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <el-collapse v-model="activeNames" class="target-collapse">
        <el-collapse-item v-for="(items, authority) in groupedData" :key="authority" :name="authority">
          <template #title>
            <div class="collapse-title">
              <span class="target-name">{{ authority }}</span>
              <el-badge :value="items.length" :max="999" type="primary" style="margin-left: 10px" />
            </div>
          </template>
          <el-table :data="items" stripe size="small" @sort-change="handleSortChange">
            <el-table-column prop="url" label="URL" min-width="300" show-overflow-tooltip>
              <template #default="{ row }">
                <a :href="row.url" target="_blank" rel="noopener" class="url-link">{{ row.url }}</a>
              </template>
            </el-table-column>
            <el-table-column prop="path" :label="$t('dirscan.path')" min-width="120" show-overflow-tooltip />
            <el-table-column prop="statusCode" :label="$t('dirscan.statusCode')" width="100" sortable="custom">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.statusCode)" size="small">{{ row.statusCode }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="contentLength" :label="$t('dirscan.size')" width="100" sortable="custom">
              <template #default="{ row }">{{ formatSize(row.contentLength) }}</template>
            </el-table-column>
            <el-table-column prop="title" :label="$t('dirscan.title')" min-width="120" show-overflow-tooltip />
            <el-table-column prop="contentType" :label="$t('dirscan.contentType')" min-width="120" show-overflow-tooltip />
            <el-table-column prop="redirectUrl" :label="$t('dirscan.redirectUrl')" min-width="150" show-overflow-tooltip />
            <el-table-column prop="createTime" :label="$t('dirscan.discoveryTime')" width="150" />
            <el-table-column :label="$t('common.operation')" width="80" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" link size="small" @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>

      <el-empty v-if="Object.keys(groupedData).length === 0 && !loading" :description="$t('common.noData')" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const emit = defineEmits(['data-changed'])

const loading = ref(false)
const tableData = ref([])
const activeNames = ref([])

const searchForm = reactive({ authority: '', path: '', statusCode: null })
const sortForm = reactive({ sortField: '', sortOrder: '' })
const stat = reactive({ total: 0, status_2xx: 0, status_3xx: 0, status_4xx: 0, status_5xx: 0 })
const pagination = reactive({ page: 1, pageSize: 1000, total: 0 })

// 按目标分组数据
const groupedData = computed(() => {
  const groups = {}
  for (const item of tableData.value) {
    const key = item.authority || 'unknown'
    if (!groups[key]) groups[key] = []
    groups[key].push(item)
  }
  return groups
})

function handleWorkspaceChanged() { loadData(); loadStat() }

onMounted(() => {
  loadData(); loadStat()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})
onUnmounted(() => { window.removeEventListener('workspace-changed', handleWorkspaceChanged) })

async function loadData() {
  loading.value = true
  try {
    const params = { page: 1, pageSize: pagination.pageSize }
    if (searchForm.authority) params.authority = searchForm.authority
    if (searchForm.path) params.path = searchForm.path
    if (searchForm.statusCode != null) params.statusCode = searchForm.statusCode
    if (sortForm.sortField) params.sortField = sortForm.sortField
    if (sortForm.sortOrder) params.sortOrder = sortForm.sortOrder
    
    const res = await request.post('/dirscan/result/list', params)
    if (res.code === 0) { 
      tableData.value = res.list || []
      pagination.total = res.total || 0
      // 默认展开第一个目标
      const keys = Object.keys(groupedData.value)
      if (keys.length > 0 && activeNames.value.length === 0) {
        activeNames.value = [keys[0]]
      }
    }
  } catch (e) {
    console.error('[DirScan] loadData error:', e)
  } finally { 
    loading.value = false 
  }
}

async function loadStat() {
  try {
    const res = await request.post('/dirscan/result/stat', {})
    if (res.code === 0 && res.stat) {
      stat.total = res.stat.total || 0
      stat.status_2xx = res.stat.status_2xx || 0
      stat.status_3xx = res.stat.status_3xx || 0
      stat.status_4xx = res.stat.status_4xx || 0
      stat.status_5xx = res.stat.status_5xx || 0
    }
  } catch (e) { console.error(e) }
}

function handleSearch() { loadData() }
function handleReset() {
  Object.assign(searchForm, { authority: '', path: '', statusCode: null })
  Object.assign(sortForm, { sortField: '', sortOrder: '' })
  handleSearch()
}

// 处理排序变化
function handleSortChange({ prop, order }) {
  if (order) {
    sortForm.sortField = prop
    sortForm.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  } else {
    sortForm.sortField = ''
    sortForm.sortOrder = ''
  }
  loadData()
}

function expandAll() { activeNames.value = Object.keys(groupedData.value) }
function collapseAll() { activeNames.value = [] }

function getStatusType(code) {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'warning'
  if (code >= 400 && code < 500) return 'danger'
  if (code >= 500) return 'danger'
  return 'info'
}

function formatSize(bytes) {
  if (!bytes || bytes < 0) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

async function handleDelete(row) {
  await ElMessageBox.confirm(t('dirscan.confirmDelete'), t('common.tip'), { type: 'warning' })
  const res = await request.post('/dirscan/result/delete', { id: row.id })
  if (res.code === 0) { ElMessage.success(t('common.deleteSuccess')); loadData(); loadStat() }
}

async function handleClear() {
  await ElMessageBox.confirm(t('dirscan.confirmClear'), t('common.warning'), { type: 'error', confirmButtonText: t('dirscan.confirmClearBtn'), cancelButtonText: t('common.cancel') })
  const res = await request.post('/dirscan/result/clear', {})
  if (res.code === 0) { ElMessage.success(res.msg || t('dirscan.clearSuccess')); loadData(); loadStat(); emit('data-changed') }
  else { ElMessage.error(res.msg || t('dirscan.clearFailed')) }
}

async function handleExport(command) {
  if (tableData.value.length === 0) {
    ElMessage.warning(t('dirscan.noDataToExport'))
    return
  }
  
  if (command === 'csv') {
    // CSV导出所有字段
    const headers = ['URL', 'Path', 'StatusCode', 'ContentLength', 'ContentType', 'Title', 'RedirectUrl', 'Host', 'Port', 'Authority', 'CreateTime']
    const csvRows = [headers.join(',')]
    
    for (const row of tableData.value) {
      const values = [
        escapeCsvField(row.url || ''),
        escapeCsvField(row.path || ''),
        row.statusCode || '',
        row.contentLength || 0,
        escapeCsvField(row.contentType || ''),
        escapeCsvField(row.title || ''),
        escapeCsvField(row.redirectUrl || ''),
        escapeCsvField(row.host || ''),
        row.port || '',
        escapeCsvField(row.authority || ''),
        escapeCsvField(row.createTime || '')
      ]
      csvRows.push(values.join(','))
    }
    
    // 添加BOM以支持Excel正确识别UTF-8
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvRows.join('\n')], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `dirscan_results_${new Date().toISOString().slice(0, 10)}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success(t('dirscan.exportSuccess', { count: tableData.value.length }))
    return
  }
  
  // 原有的URL导出逻辑
  const seen = new Set()
  const exportData = []
  for (const row of tableData.value) {
    if (row.url && !seen.has(row.url)) {
      seen.add(row.url)
      exportData.push(row.url)
    }
  }
  
  if (exportData.length === 0) {
    ElMessage.warning(t('dirscan.noDataToExport'))
    return
  }
  
  const blob = new Blob([exportData.join('\n')], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'dirscan_urls_all.txt'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success(t('dirscan.exportSuccess', { count: exportData.length }))
}

// CSV字段转义
function escapeCsvField(field) {
  if (field == null) return ''
  const str = String(field)
  if (str.includes(',') || str.includes('"') || str.includes('\n') || str.includes('\r')) {
    return '"' + str.replace(/"/g, '""') + '"'
  }
  return str
}

function refresh() { loadData(); loadStat() }

defineExpose({ refresh })
</script>

<style scoped>
.dirscan-view {
  .search-card { margin-bottom: 16px; }
  .stat-row {
    margin-bottom: 16px;
    .stat-card {
      text-align: center;
      .stat-value { font-size: 24px; font-weight: 600; color: var(--el-color-primary); }
      .stat-label { color: var(--el-text-color-secondary); margin-top: 8px; font-size: 13px; }
      &.status-2xx .stat-value { color: #67c23a; }
      &.status-3xx .stat-value { color: #e6a23c; }
      &.status-4xx .stat-value { color: #f56c6c; }
      &.status-5xx .stat-value { color: #909399; }
    }
  }
  .collapse-card {
    .collapse-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      .total-info { color: var(--el-text-color-secondary); font-size: 14px; }
    }
  }
  .target-collapse {
    .collapse-title {
      display: flex;
      align-items: center;
      .target-name {
        font-weight: 500;
        color: var(--el-color-primary);
      }
    }
  }
  .url-link {
    color: var(--el-color-primary);
    text-decoration: none;
    &:hover { text-decoration: underline; }
  }
}
</style>

