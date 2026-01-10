<template>
  <div class="dirscan-view">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="目标">
          <el-input v-model="searchForm.authority" placeholder="IP:端口" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="searchForm.path" placeholder="路径关键词" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态码">
          <el-select v-model="searchForm.statusCode" placeholder="全部" clearable style="width: 120px">
            <el-option label="200" :value="200" />
            <el-option label="301" :value="301" />
            <el-option label="302" :value="302" />
            <el-option label="403" :value="403" />
            <el-option label="404" :value="404" />
            <el-option label="500" :value="500" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="danger" plain @click="handleClear">清空数据</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 统计信息 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="4">
        <el-card class="stat-card">
          <div class="stat-value">{{ stat.total }}</div>
          <div class="stat-label">总数</div>
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
        <span class="total-info">共 {{ Object.keys(groupedData).length }} 个目标，{{ pagination.total }} 条记录</span>
        <div class="collapse-actions">
          <el-button size="small" @click="expandAll">全部展开</el-button>
          <el-button size="small" @click="collapseAll">全部收起</el-button>
          <el-dropdown style="margin-left: 10px" @command="handleExport">
            <el-button type="success" size="small">
              导出<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="all-url">导出全部URL</el-dropdown-item>
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
          <el-table :data="items" stripe size="small" max-height="400">
            <el-table-column prop="url" label="URL" min-width="300" show-overflow-tooltip>
              <template #default="{ row }">
                <a :href="row.url" target="_blank" rel="noopener" class="url-link">{{ row.url }}</a>
              </template>
            </el-table-column>
            <el-table-column prop="path" label="路径" min-width="120" show-overflow-tooltip />
            <el-table-column prop="statusCode" label="状态码" width="90">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.statusCode)" size="small">{{ row.statusCode }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="contentLength" label="大小" width="90">
              <template #default="{ row }">{{ formatSize(row.contentLength) }}</template>
            </el-table-column>
            <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />
            <el-table-column prop="createTime" label="发现时间" width="150" />
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>

      <el-empty v-if="Object.keys(groupedData).length === 0 && !loading" description="暂无数据" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import request from '@/api/request'

const emit = defineEmits(['data-changed'])

const loading = ref(false)
const tableData = ref([])
const activeNames = ref([])

const searchForm = reactive({ authority: '', path: '', statusCode: null })
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
  handleSearch()
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
  await ElMessageBox.confirm('确定删除该记录吗？', '提示', { type: 'warning' })
  const res = await request.post('/dirscan/result/delete', { id: row.id })
  if (res.code === 0) { ElMessage.success('删除成功'); loadData(); loadStat() }
}

async function handleClear() {
  await ElMessageBox.confirm('确定清空所有目录扫描数据吗？此操作不可恢复！', '警告', { type: 'error', confirmButtonText: '确定清空', cancelButtonText: '取消' })
  const res = await request.post('/dirscan/result/clear', {})
  if (res.code === 0) { ElMessage.success(res.msg || '清空成功'); loadData(); loadStat(); emit('data-changed') }
  else { ElMessage.error(res.msg || '清空失败') }
}

async function handleExport(command) {
  if (tableData.value.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }
  
  const seen = new Set()
  const exportData = []
  for (const row of tableData.value) {
    if (row.url && !seen.has(row.url)) {
      seen.add(row.url)
      exportData.push(row.url)
    }
  }
  
  if (exportData.length === 0) {
    ElMessage.warning('没有可导出的数据')
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
  
  ElMessage.success(`成功导出 ${exportData.length} 条数据`)
}

function refresh() { loadData(); loadStat() }

defineExpose({ refresh })
</script>

<style lang="scss" scoped>
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
