<template>
  <div class="site-view">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-form :inline="true" class="search-form">
        <el-form-item :label="$t('site.site')">
          <el-input v-model="searchForm.site" :placeholder="$t('site.sitePlaceholder')" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('site.title')">
          <el-input v-model="searchForm.title" :placeholder="$t('site.titlePlaceholder')" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('site.app')">
          <el-input v-model="searchForm.app" :placeholder="$t('site.appPlaceholder')" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('site.statusCode')">
          <el-input v-model="searchForm.httpStatus" placeholder="200/404..." clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('site.organization')">
          <el-select v-model="searchForm.orgId" :placeholder="$t('common.allOrganizations')" clearable style="width: 140px">
            <el-option :label="$t('common.allOrganizations')" value="" />
            <el-option v-for="org in organizations" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
          <el-button type="danger" plain @click="handleClear">{{ $t('asset.clearData') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 统计信息 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stat.total }}</div>
          <div class="stat-label">{{ $t('site.totalSites') }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stat.httpCount }}</div>
          <div class="stat-label">{{ $t('site.httpSites') }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stat.httpsCount }}</div>
          <div class="stat-label">{{ $t('site.httpsSites') }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-value">{{ stat.newCount }}</div>
          <div class="stat-label">{{ $t('site.newSites') }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <div class="table-header">
        <span class="total-info">{{ $t('site.totalSitesCount', { count: pagination.total }) }}</span>
        <div class="table-actions">
          <el-button type="danger" size="small" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
            {{ $t('common.batchDelete') }} ({{ selectedRows.length }})
          </el-button>
          <el-dropdown style="margin-left: 10px" @command="handleExport">
            <el-button type="success" size="small">
              {{ $t('common.export') }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="selected-site" :disabled="selectedRows.length === 0">{{ $t('site.exportSelectedSites', { count: selectedRows.length }) }}</el-dropdown-item>
                <el-dropdown-item divided command="all-site">{{ $t('site.exportAllSites') }}</el-dropdown-item>
                <el-dropdown-item command="csv">{{ $t('common.exportCsv') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
      
      <el-table :data="tableData" v-loading="loading" stripe max-height="500" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" />
        <el-table-column :label="$t('site.site')" min-width="280">
          <template #default="{ row }">
            <div class="site-cell">
              <a :href="row.site" target="_blank" class="site-link">{{ row.site }}</a>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('site.title')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.title || '-' }}</template>
        </el-table-column>
        <el-table-column :label="$t('site.statusCode')" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.httpStatus)" size="small">{{ row.httpStatus || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('site.fingerprint')" min-width="180">
          <template #default="{ row }">
            <el-tag v-for="app in (row.app || []).slice(0, 3)" :key="app" size="small" type="success" style="margin: 2px">
              {{ app }}
            </el-tag>
            <span v-if="(row.app || []).length > 3" class="more-apps">+{{ (row.app || []).length - 3 }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="$t('site.screenshot')" width="100" align="center">
          <template #default="{ row }">
            <el-image 
              v-if="row.screenshot" 
              :src="getScreenshotUrl(row.screenshot)" 
              :preview-src-list="[getScreenshotUrl(row.screenshot)]" 
              :z-index="9999"
              :preview-teleported="true"
              :hide-on-click-modal="true"
              fit="cover" 
              class="screenshot-img" 
            />
            <span v-else>-</span>
          </template>
        </el-table-column>

        <el-table-column :label="$t('common.updateTime')" width="160">
          <template #default="{ row }">{{ row.updateTime }}</template>
        </el-table-column>
        <el-table-column :label="$t('common.operation')" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showDetail(row)">{{ $t('common.detail') }}</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        class="pagination"
        @size-change="loadData"
        @current-change="loadData"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" :title="$t('site.siteDetail')" width="700px">
      <el-descriptions :column="2" border v-if="currentSite">
        <el-descriptions-item :label="$t('site.siteAddress')" :span="2">
          <a :href="currentSite.site" target="_blank" class="site-link">{{ currentSite.site }}</a>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('site.title')" :span="2">{{ currentSite.title || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('site.statusCode')">
          <el-tag :type="getStatusType(currentSite.httpStatus)" size="small">{{ currentSite.httpStatus || '-' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Server">{{ currentSite.server || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ currentSite.ip || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('site.location')">{{ currentSite.location || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('site.organization')">{{ currentSite.orgName || $t('common.defaultOrganization') }}</el-descriptions-item>
        <el-descriptions-item :label="$t('common.updateTime')">{{ currentSite.updateTime }}</el-descriptions-item>
        <el-descriptions-item :label="$t('site.fingerprint')" :span="2">
          <el-tag v-for="app in (currentSite.app || [])" :key="app" size="small" type="success" style="margin: 2px">
            {{ app }}
          </el-tag>
          <span v-if="!(currentSite.app || []).length">-</span>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('site.screenshot')" :span="2" v-if="currentSite.screenshot">
          <el-image 
            :src="getScreenshotUrl(currentSite.screenshot)" 
            :preview-src-list="[getScreenshotUrl(currentSite.screenshot)]"
            :z-index="9999"
            :preview-teleported="true"
            :hide-on-click-modal="true"
            fit="contain"
            style="max-width: 400px; max-height: 300px; cursor: pointer;"
          />
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">{{ $t('common.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import request from '@/api/request'
import { clearAsset } from '@/api/asset'

const { t } = useI18n()
const emit = defineEmits(['data-changed'])

const loading = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const organizations = ref([])
const detailVisible = ref(false)
const currentSite = ref(null)

const searchForm = reactive({ site: '', title: '', app: '', httpStatus: '', orgId: '' })
const stat = reactive({ total: 0, httpCount: 0, httpsCount: 0, newCount: 0 })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

function handleWorkspaceChanged() { pagination.page = 1; loadData(); loadStat() }

onMounted(() => {
  loadData(); loadStat(); loadOrganizations()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})
onUnmounted(() => { window.removeEventListener('workspace-changed', handleWorkspaceChanged) })

async function loadData() {
  loading.value = true
  try {
    const res = await request.post('/asset/site/list', {
      page: pagination.page, pageSize: pagination.pageSize,
      site: searchForm.site, title: searchForm.title, app: searchForm.app,
      httpStatus: searchForm.httpStatus, orgId: searchForm.orgId
    })
    if (res.code === 0) { tableData.value = res.list || []; pagination.total = res.total || 0 }
  } finally { loading.value = false }
}

async function loadStat() {
  try {
    const res = await request.post('/asset/site/stat', {})
    if (res.code === 0) {
      stat.total = res.total || 0
      stat.httpCount = res.httpCount || 0
      stat.httpsCount = res.httpsCount || 0
      stat.newCount = res.newCount || 0
    }
  } catch (e) { console.error(e) }
}

async function loadOrganizations() {
  try {
    const res = await request.post('/organization/list', { page: 1, pageSize: 100 })
    if (res.code === 0) organizations.value = res.list || []
  } catch (e) { console.error(e) }
}

function handleSearch() { pagination.page = 1; loadData() }
function handleReset() {
  Object.assign(searchForm, { site: '', title: '', app: '', httpStatus: '', orgId: '' })
  handleSearch()
}

function handleSelectionChange(rows) { selectedRows.value = rows }

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(t('site.confirmBatchDelete', { count: selectedRows.value.length }), t('common.tip'), { type: 'warning' })
  const ids = selectedRows.value.map(row => row.id)
  const res = await request.post('/asset/site/batchDelete', { ids })
  if (res.code === 0) { ElMessage.success(t('common.deleteSuccess')); selectedRows.value = []; loadData(); loadStat(); emit('data-changed') }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(t('site.confirmDeleteSite'), t('common.tip'), { type: 'warning' })
  const res = await request.post('/asset/site/batchDelete', { ids: [row.id] })
  if (res.code === 0) { ElMessage.success(t('common.deleteSuccess')); loadData(); loadStat(); emit('data-changed') }
}

function showDetail(row) {
  currentSite.value = row
  detailVisible.value = true
}

async function handleClear() {
  try {
    await ElMessageBox.confirm(t('asset.confirmClearAll'), t('common.warning'), { type: 'error', confirmButtonText: t('asset.confirmClearBtn'), cancelButtonText: t('common.cancel') })
    const res = await clearAsset()
    if (res.code === 0) { 
      ElMessage.success(res.msg || t('asset.clearSuccess'))
      selectedRows.value = []
      loadData()
      loadStat()
      emit('data-changed') 
    } else { 
      ElMessage.error(res.msg || t('asset.clearFailed')) 
    }
  } catch (e) {
    if (e !== 'cancel') {
      console.error('清空资产失败:', e)
      ElMessage.error(t('asset.clearFailed'))
    }
  }
}

// 导出功能
async function handleExport(command) {
  let data = []
  let filename = ''
  
  if (command === 'selected-site') {
    if (selectedRows.value.length === 0) {
      ElMessage.warning(t('site.pleaseSelectSites'))
      return
    }
    data = selectedRows.value
    filename = 'sites_selected.txt'
  } else if (command === 'csv') {
    // CSV导出所有字段
    ElMessage.info(t('asset.gettingAllData'))
    try {
      const res = await request.post('/asset/site/list', {
        ...searchForm, page: 1, pageSize: 10000
      })
      if (res.code === 0) {
        data = res.list || []
      } else {
        ElMessage.error(t('asset.getDataFailed'))
        return
      }
    } catch (e) {
      ElMessage.error(t('asset.getDataFailed'))
      return
    }
    
    if (data.length === 0) {
      ElMessage.warning(t('asset.noDataToExport'))
      return
    }
    
    const headers = ['Site', 'Title', 'StatusCode', 'Server', 'IP', 'Location', 'Apps', 'Organization', 'UpdateTime']
    const csvRows = [headers.join(',')]
    
    for (const row of data) {
      const values = [
        escapeCsvField(row.site || ''),
        escapeCsvField(row.title || ''),
        row.httpStatus || '',
        escapeCsvField(row.server || ''),
        escapeCsvField(row.ip || ''),
        escapeCsvField(row.location || ''),
        escapeCsvField((row.app || []).join(';')),
        escapeCsvField(row.orgName || ''),
        escapeCsvField(row.updateTime || '')
      ]
      csvRows.push(values.join(','))
    }
    
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvRows.join('\n')], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `sites_${new Date().toISOString().slice(0, 10)}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success(t('asset.exportSuccess', { count: data.length }))
    return
  } else {
    ElMessage.info(t('asset.gettingAllData'))
    try {
      const res = await request.post('/asset/site/list', {
        ...searchForm, page: 1, pageSize: 10000
      })
      if (res.code === 0) {
        data = res.list || []
      } else {
        ElMessage.error(t('asset.getDataFailed'))
        return
      }
    } catch (e) {
      ElMessage.error(t('asset.getDataFailed'))
      return
    }
    filename = 'sites_all.txt'
  }
  
  if (data.length === 0) {
    ElMessage.warning(t('asset.noDataToExport'))
    return
  }
  
  const seen = new Set()
  const exportData = []
  for (const row of data) {
    if (row.site && !seen.has(row.site)) {
      seen.add(row.site)
      exportData.push(row.site)
    }
  }
  
  if (exportData.length === 0) {
    ElMessage.warning(t('asset.noDataToExport'))
    return
  }
  
  const blob = new Blob([exportData.join('\n')], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success(t('asset.exportSuccess', { count: exportData.length }))
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

function getStatusType(status) {
  if (!status) return 'info'
  const code = parseInt(status)
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'warning'
  if (code >= 400) return 'danger'
  return 'info'
}

function getScreenshotUrl(screenshot) {
  if (!screenshot) return ''
  if (screenshot.startsWith('data:') || screenshot.startsWith('/9j/') || screenshot.startsWith('iVBOR')) {
    return screenshot.startsWith('data:') ? screenshot : `data:image/png;base64,${screenshot}`
  }
  return `/api/screenshot/${screenshot}`
}

function refresh() { loadData(); loadStat() }

defineExpose({ refresh })
</script>

<style scoped>
.site-view {
  .search-card { margin-bottom: 16px; }
  .stat-row {
    margin-bottom: 16px;
    .stat-card {
      text-align: center;
      .stat-value { font-size: 28px; font-weight: 600; color: var(--el-color-primary); }
      .stat-label { color: var(--el-text-color-secondary); margin-top: 8px; }
    }
  }
  .table-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;
    .total-info { color: var(--el-text-color-secondary); font-size: 14px; }
  }
  .site-cell .site-link { color: #409eff; text-decoration: none; font-family: monospace; &:hover { text-decoration: underline; } }
  .more-apps { color: var(--el-text-color-secondary); font-size: 12px; }
  .screenshot-img { width: 80px; height: 60px; border-radius: 4px; cursor: pointer; }
  .pagination { margin-top: 16px; justify-content: flex-end; }
}
</style>

