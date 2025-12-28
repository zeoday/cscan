<template>
  <div class="asset-all-view">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-tabs v-model="activeTab" class="search-tabs">
        <el-tab-pane label="快捷查询" name="quick">
          <div class="quick-search-form">
            <div class="search-row">
              <div class="search-item">
                <label class="search-label">主机</label>
                <el-input v-model="searchForm.host" placeholder="IP/域名" clearable @keyup.enter="handleSearch" />
              </div>
              <div class="search-item">
                <label class="search-label">端口</label>
                <el-input v-model.number="searchForm.port" placeholder="端口号" clearable @keyup.enter="handleSearch" />
              </div>
              <div class="search-item">
                <label class="search-label">服务</label>
                <el-input v-model="searchForm.service" placeholder="http/ssh..." clearable @keyup.enter="handleSearch" />
              </div>
              <div class="search-item">
                <label class="search-label">标题</label>
                <el-input v-model="searchForm.title" placeholder="网页标题" clearable @keyup.enter="handleSearch" />
              </div>
              <div class="search-item">
                <label class="search-label">应用</label>
                <el-input v-model="searchForm.app" placeholder="指纹/应用" clearable @keyup.enter="handleSearch" />
              </div>
              <div class="search-item">
                <label class="search-label">组织</label>
                <el-select v-model="searchForm.orgId" placeholder="全部组织" clearable @change="handleSearch">
                  <el-option label="全部组织" value="" />
                  <el-option v-for="org in organizations" :key="org.id" :label="org.name" :value="org.id" />
                </el-select>
              </div>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="统计信息" name="stat">
          <div class="stat-panel">
            <div class="stat-column">
              <div class="stat-title">Port</div>
              <div v-for="item in stat.topPorts" :key="'port-'+item.name" class="stat-item" @click="quickFilter('port', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column">
              <div class="stat-title">Service</div>
              <div v-for="item in stat.topService" :key="'svc-'+item.name" class="stat-item" @click="quickFilter('service', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column">
              <div class="stat-title">App</div>
              <div v-for="item in stat.topApp" :key="'app-'+item.name" class="stat-item" @click="quickFilter('app', item.name)">
                <span class="stat-count">{{ item.count }}</span>
                <span class="stat-name">{{ item.name }}</span>
              </div>
            </div>
            <div class="stat-column filter-column">
              <el-checkbox v-model="searchForm.onlyUpdated">只看有更新</el-checkbox>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <div class="search-actions">
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <div class="table-header">
        <span class="total-info">共 {{ pagination.total }} 条记录</span>
        <div class="table-actions">
          <el-button type="danger" size="small" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
            批量删除 ({{ selectedRows.length }})
          </el-button>
          <el-button type="danger" size="small" plain @click="handleClear">
            清空数据
          </el-button>
        </div>
      </div>
      <el-table :data="tableData" v-loading="loading" stripe size="small" max-height="500" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="40" />
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column label="资产" min-width="200">
          <template #default="{ row }">
            <div class="asset-cell">
              <a :href="getAssetUrl(row)" target="_blank" class="asset-link">{{ row.authority }}</a>
            </div>
            <div class="org-text">{{ row.orgName || '默认组织' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="IP" width="140">
          <template #default="{ row }">
            <div>{{ row.host }}</div>
            <div v-if="row.location" class="location-text">{{ row.location }}</div>
          </template>
        </el-table-column>
        <el-table-column label="端口/服务" width="120">
          <template #default="{ row }">
            <span class="port-text">{{ row.port }}</span>
            <span v-if="row.service" class="service-text">{{ row.service }}</span>
          </template>
        </el-table-column>
        <el-table-column label="标题" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="指纹" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="app in (row.app || []).slice(0, 3)" :key="app" size="small" type="success" style="margin: 2px">
              {{ getAppName(app) }}
            </el-tag>
            <span v-if="(row.app || []).length > 3" class="more-apps">+{{ (row.app || []).length - 3 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="截图" width="90" align="center">
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
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">
            <div>{{ row.updateTime }}</div>
            <el-tag v-if="row.isNew" type="success" size="small" effect="dark">新</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showDetail(row)">详情</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
    <el-dialog v-model="detailVisible" title="资产详情" width="700px">
      <el-descriptions :column="2" border v-if="currentAsset">
        <el-descriptions-item label="资产地址" :span="2">
          <a :href="getAssetUrl(currentAsset)" target="_blank" class="asset-link">{{ currentAsset.authority }}</a>
        </el-descriptions-item>
        <el-descriptions-item label="主机">{{ currentAsset.host }}</el-descriptions-item>
        <el-descriptions-item label="端口">{{ currentAsset.port }}</el-descriptions-item>
        <el-descriptions-item label="服务">{{ currentAsset.service || '-' }}</el-descriptions-item>
        <el-descriptions-item label="协议">{{ currentAsset.scheme || '-' }}</el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ currentAsset.title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态码">{{ currentAsset.httpStatus || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Server">{{ currentAsset.server || '-' }}</el-descriptions-item>
        <el-descriptions-item label="位置" :span="2">{{ currentAsset.location || '-' }}</el-descriptions-item>
        <el-descriptions-item label="组织">{{ currentAsset.orgName || '默认组织' }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ currentAsset.updateTime }}</el-descriptions-item>
        <el-descriptions-item label="应用指纹" :span="2">
          <el-tag v-for="app in (currentAsset.app || [])" :key="app" size="small" type="success" style="margin: 2px">
            {{ app }}
          </el-tag>
          <span v-if="!(currentAsset.app || []).length">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="截图" :span="2" v-if="currentAsset.screenshot">
          <el-image 
            :src="getScreenshotUrl(currentAsset.screenshot)" 
            :preview-src-list="[getScreenshotUrl(currentAsset.screenshot)]"
            :z-index="9999"
            :preview-teleported="true"
            :hide-on-click-modal="true"
            fit="contain"
            style="max-width: 400px; max-height: 300px; cursor: pointer;"
          />
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAssetList, getAssetStat, batchDeleteAsset, clearAsset } from '@/api/asset'
import { useWorkspaceStore } from '@/stores/workspace'
import request from '@/api/request'

const emit = defineEmits(['data-changed'])

const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const activeTab = ref('quick')
const organizations = ref([])
const detailVisible = ref(false)
const currentAsset = ref(null)

const stat = reactive({ topPorts: [], topService: [], topApp: [] })
const searchForm = reactive({ host: '', port: null, service: '', title: '', app: '', orgId: '', onlyUpdated: false })
const pagination = reactive({ page: 1, pageSize: 50, total: 0 })

function handleWorkspaceChanged() { pagination.page = 1; loadData(); loadStat() }

onMounted(() => {
  loadData(); loadStat(); loadOrganizations()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})
onUnmounted(() => { window.removeEventListener('workspace-changed', handleWorkspaceChanged) })

async function loadData() {
  loading.value = true
  try {
    const res = await getAssetList({
      page: pagination.page, pageSize: pagination.pageSize,
      host: searchForm.host, port: searchForm.port, service: searchForm.service,
      title: searchForm.title, app: searchForm.app, orgId: searchForm.orgId,
      onlyUpdated: searchForm.onlyUpdated
    })
    if (res.code === 0) { tableData.value = res.list || []; pagination.total = res.total }
  } finally { loading.value = false }
}

async function loadStat() {
  const res = await getAssetStat({})
  if (res.code === 0) {
    stat.topPorts = res.topPorts || []
    stat.topService = res.topService || []
    stat.topApp = res.topApp || []
  }
}

async function loadOrganizations() {
  try {
    const res = await request.post('/organization/list', { page: 1, pageSize: 100 })
    if (res.code === 0) organizations.value = res.list || []
  } catch (e) { console.error(e) }
}

function quickFilter(field, value) {
  searchForm[field] = field === 'port' ? parseInt(value, 10) : value
  activeTab.value = 'quick'
  handleSearch()
}

function handleSearch() { pagination.page = 1; loadData() }
function handleReset() {
  Object.assign(searchForm, { host: '', port: null, service: '', title: '', app: '', orgId: '', onlyUpdated: false })
  handleSearch(); loadStat()
}

function handleSelectionChange(rows) { selectedRows.value = rows }

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 条资产吗？`, '提示', { type: 'warning' })
  const ids = selectedRows.value.map(row => row.id)
  const res = await batchDeleteAsset({ ids })
  if (res.code === 0) { ElMessage.success('删除成功'); selectedRows.value = []; loadData(); loadStat(); emit('data-changed') }
}

async function handleDelete(row) {
  await ElMessageBox.confirm('确定删除该资产吗？', '提示', { type: 'warning' })
  const res = await batchDeleteAsset({ ids: [row.id] })
  if (res.code === 0) { ElMessage.success('删除成功'); loadData(); loadStat(); emit('data-changed') }
}

function showDetail(row) {
  currentAsset.value = row
  detailVisible.value = true
}

async function handleClear() {
  await ElMessageBox.confirm('确定清空所有资产数据吗？此操作不可恢复！', '警告', { type: 'error', confirmButtonText: '确定清空', cancelButtonText: '取消' })
  const res = await clearAsset()
  if (res.code === 0) { ElMessage.success(res.msg || '清空成功'); selectedRows.value = []; loadData(); loadStat(); emit('data-changed') }
  else { ElMessage.error(res.msg || '清空失败') }
}

function getAssetUrl(row) {
  const scheme = row.service === 'https' || row.port === 443 ? 'https' : 'http'
  return `${scheme}://${row.host}:${row.port}`
}

function getAppName(app) {
  if (!app) return ''
  const idx = app.indexOf('[')
  return idx > 0 ? app.substring(0, idx) : app
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

<style lang="scss" scoped>
.asset-all-view {
  .search-card { margin-bottom: 15px; }
  .search-tabs { :deep(.el-tabs__header) { margin-bottom: 10px; } }
  .quick-search-form .search-row {
    display: flex; flex-wrap: wrap; gap: 16px;
    .search-item {
      display: flex; flex-direction: column; min-width: 140px; flex: 1; max-width: 180px;
      .search-label { font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 6px; }
    }
  }
  .search-actions { margin-top: 16px; padding-top: 12px; border-top: 1px solid var(--el-border-color-lighter); text-align: right; }
  .stat-panel {
    display: flex; gap: 30px;
    .stat-column {
      min-width: 140px;
      .stat-title { font-weight: bold; margin-bottom: 8px; padding-bottom: 5px; border-bottom: 2px solid #409eff; }
      .stat-item { display: flex; align-items: center; padding: 3px 0; cursor: pointer;
        &:hover { background: var(--el-fill-color); }
        .stat-count { min-width: 30px; padding: 1px 6px; margin-right: 8px; background: #409eff; color: #fff; border-radius: 3px; font-size: 12px; }
        .stat-name { color: #409eff; font-size: 13px; }
      }
    }
    .filter-column { margin-left: auto; }
  }
  .table-card {
    .table-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;
      .total-info { color: var(--el-text-color-secondary); font-size: 14px; }
    }
    .asset-cell .asset-link { color: #409eff; text-decoration: none; &:hover { text-decoration: underline; } }
    .org-text, .location-text { color: var(--el-text-color-secondary); font-size: 12px; }
    .port-text { font-weight: 500; margin-right: 8px; }
    .service-text { color: #67c23a; font-size: 12px; }
    .more-apps { color: var(--el-text-color-secondary); font-size: 12px; }
    .screenshot-img { width: 70px; height: 50px; border-radius: 4px; cursor: pointer; }
    .pagination { margin-top: 15px; justify-content: flex-end; }
  }
}
</style>
