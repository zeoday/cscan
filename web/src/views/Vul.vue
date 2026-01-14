<template>
  <div class="vul-page">
    <!-- 搜索区域 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item :label="$t('vulnerability.target')">
          <el-input v-model="searchForm.authority" placeholder="IP:Port" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item :label="$t('vulnerability.severity')">
          <el-select v-model="searchForm.severity" :placeholder="$t('common.all')" clearable style="width: 120px">
            <el-option :label="$t('vulnerability.critical')" value="critical" />
            <el-option :label="$t('vulnerability.high')" value="high" />
            <el-option :label="$t('vulnerability.medium')" value="medium" />
            <el-option :label="$t('vulnerability.low')" value="low" />
            <el-option :label="$t('vulnerability.info')" value="info" />
            <el-option :label="$t('vulnerability.unknown')" value="unknown" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('vulnerability.source')">
          <el-select v-model="searchForm.source" :placeholder="$t('common.all')" clearable style="width: 120px">
            <el-option label="Nuclei" value="nuclei" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">{{ $t('common.search') }}</el-button>
          <el-button @click="handleReset">{{ $t('common.reset') }}</el-button>
          <el-button type="danger" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
            {{ $t('common.batchDelete') }} ({{ selectedRows.length }})
          </el-button>
          <el-dropdown style="margin-left: 10px" @command="handleExport">
            <el-button type="success">
              {{ $t('common.export') }}<el-icon class="el-icon--right"><arrow-down /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="selected-target" :disabled="selectedRows.length === 0">{{ $t('vulnerability.exportSelectedTarget') }} ({{ selectedRows.length }})</el-dropdown-item>
                <el-dropdown-item command="selected-url" :disabled="selectedRows.length === 0">{{ $t('vulnerability.exportSelectedUrl') }} ({{ selectedRows.length }})</el-dropdown-item>
                <el-dropdown-item divided command="all-target">{{ $t('vulnerability.exportAllTarget') }}</el-dropdown-item>
                <el-dropdown-item command="all-url">{{ $t('vulnerability.exportAllUrl') }}</el-dropdown-item>
                <el-dropdown-item command="csv">{{ $t('common.exportCsv') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <el-table 
        :data="tableData" 
        v-loading="loading" 
        stripe
        max-height="500"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="authority" :label="$t('vulnerability.target')" min-width="150" />
        <el-table-column prop="url" label="URL" min-width="250" show-overflow-tooltip />
        <el-table-column prop="pocFile" :label="$t('vulnerability.poc')" min-width="200" show-overflow-tooltip />
        <el-table-column prop="severity" :label="$t('vulnerability.severity')" width="100">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" size="small">
              {{ getSeverityLabel(row.severity) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="source" :label="$t('vulnerability.source')" width="100" />
        <el-table-column prop="createTime" :label="$t('vulnerability.discoveryTime')" width="160" />
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
    <el-dialog v-model="detailVisible" :title="$t('vulnerability.vulDetail')" width="800px">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('vulnerability.target')">{{ currentVul.authority }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.severity')">
          <el-tag :type="getSeverityType(currentVul.severity)">
            {{ getSeverityLabel(currentVul.severity) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="URL" :span="2">{{ currentVul.url }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.poc')" :span="2">{{ currentVul.pocFile }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.source')">{{ currentVul.source }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.discoveryTime')">{{ currentVul.createTime }}</el-descriptions-item>
        <!-- 知识库信息 -->
        <el-descriptions-item :label="$t('vulnerability.cveId')" v-if="currentVul.cveId">{{ currentVul.cveId }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.cweId')" v-if="currentVul.cweId">{{ currentVul.cweId }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.cvssScore')" v-if="currentVul.cvssScore">{{ currentVul.cvssScore }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.scanCount')" v-if="currentVul.scanCount">{{ currentVul.scanCount }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.firstSeen')" v-if="currentVul.firstSeenTime">{{ currentVul.firstSeenTime }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.lastSeen')" v-if="currentVul.lastSeenTime">{{ currentVul.lastSeenTime }}</el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.remediation')" :span="2" v-if="currentVul.remediation">
          <pre class="result-pre">{{ currentVul.remediation }}</pre>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.references')" :span="2" v-if="currentVul.references && currentVul.references.length">
          <div v-for="(ref, idx) in currentVul.references" :key="idx">
            <a :href="ref" target="_blank" rel="noopener">{{ ref }}</a>
          </div>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('vulnerability.verifyResult')" :span="2">
          <pre class="result-pre">{{ currentVul.result }}</pre>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 证据链 -->
      <template v-if="currentVul.evidence">
        <el-divider content-position="left">{{ $t('vulnerability.evidence') }}</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item :label="$t('vulnerability.matcherName')" v-if="currentVul.evidence.matcherName">
            {{ currentVul.evidence.matcherName }}
          </el-descriptions-item>
          <el-descriptions-item :label="$t('vulnerability.extractedResults')" v-if="currentVul.evidence.extractedResults && currentVul.evidence.extractedResults.length">
            <el-tag v-for="(item, idx) in currentVul.evidence.extractedResults" :key="idx" style="margin-right: 5px; margin-bottom: 5px;">
              {{ item }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('vulnerability.curlCommand')" v-if="currentVul.evidence.curlCommand">
            <pre class="result-pre">{{ currentVul.evidence.curlCommand }}</pre>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('vulnerability.requestContent')" v-if="currentVul.evidence.request">
            <pre class="result-pre">{{ currentVul.evidence.request }}</pre>
          </el-descriptions-item>
          <el-descriptions-item :label="$t('vulnerability.responseContent')" v-if="currentVul.evidence.response">
            <pre class="result-pre">{{ currentVul.evidence.response }}</pre>
            <el-tag v-if="currentVul.evidence.responseTruncated" type="warning" size="small">{{ $t('vulnerability.responseTruncated') }}</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import request from '@/api/request'
import { useWorkspaceStore } from '@/stores/workspace'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const workspaceStore = useWorkspaceStore()
const loading = ref(false)
const tableData = ref([])
const detailVisible = ref(false)
const currentVul = ref({})
const selectedRows = ref([])

const searchForm = reactive({
  authority: '',
  severity: '',
  source: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 监听工作空间切换
function handleWorkspaceChanged() {
  pagination.page = 1
  loadData()
}

onMounted(() => {
  loadData()
  window.addEventListener('workspace-changed', handleWorkspaceChanged)
})

onUnmounted(() => {
  window.removeEventListener('workspace-changed', handleWorkspaceChanged)
})

async function loadData() {
  loading.value = true
  try {
    const res = await request.post('/vul/list', {
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize,
      workspaceId: workspaceStore.currentWorkspaceId || ''
    })
    if (res.code === 0) {
      tableData.value = res.list || []
      pagination.total = res.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadData()
}

function handleReset() {
  Object.assign(searchForm, { authority: '', severity: '', source: '' })
  handleSearch()
}

function handleSelectionChange(rows) {
  selectedRows.value = rows
}

function getSeverityType(severity) {
  const map = { critical: 'danger', high: 'danger', medium: 'warning', low: 'info', info: '', unknown: 'info' }
  return map[severity] || ''
}

function getSeverityLabel(severity) {
  const map = { 
    critical: t('vulnerability.critical'), 
    high: t('vulnerability.high'), 
    medium: t('vulnerability.medium'), 
    low: t('vulnerability.low'), 
    info: t('vulnerability.info'), 
    unknown: t('vulnerability.unknown') 
  }
  return map[severity] || severity
}

async function showDetail(row) {
  try {
    const res = await request.post('/vul/detail', { id: row.id })
    if (res.code === 0 && res.data) {
      currentVul.value = res.data
    } else {
      currentVul.value = row
    }
  } catch (e) {
    currentVul.value = row
  }
  detailVisible.value = true
}

async function handleDelete(row) {
  await ElMessageBox.confirm(t('vulnerability.confirmDeleteVul'), t('common.tip'), { type: 'warning' })
  const res = await request.post('/vul/delete', { id: row.id })
  if (res.code === 0) {
    ElMessage.success(t('common.deleteSuccess'))
    loadData()
  } else {
    ElMessage.error(res.msg || t('common.operationFailed'))
  }
}

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return
  await ElMessageBox.confirm(t('vulnerability.confirmBatchDeleteVul', { count: selectedRows.value.length }), t('common.tip'), { type: 'warning' })
  const ids = selectedRows.value.map(row => row.id)
  const res = await request.post('/vul/batchDelete', { ids })
  if (res.code === 0) {
    ElMessage.success(res.msg || t('common.deleteSuccess'))
    selectedRows.value = []
    loadData()
  } else {
    ElMessage.error(res.msg || t('common.operationFailed'))
  }
}

// 导出功能
async function handleExport(command) {
  let data = []
  let filename = ''
  
  if (command === 'selected-target' || command === 'selected-url') {
    // 导出选中的
    if (selectedRows.value.length === 0) {
      ElMessage.warning(t('vulnerability.selectToExport'))
      return
    }
    data = selectedRows.value
    filename = command === 'selected-target' ? 'vul_targets_selected.txt' : 'vul_urls_selected.txt'
  } else if (command === 'csv') {
    // CSV导出所有字段
    ElMessage.info(t('vulnerability.fetchingData'))
    try {
      const res = await request.post('/vul/list', {
        ...searchForm,
        page: 1,
        pageSize: 10000,
        workspaceId: workspaceStore.currentWorkspaceId || ''
      })
      if (res.code === 0) {
        data = res.list || []
      } else {
        ElMessage.error(t('vulnerability.fetchFailed'))
        return
      }
    } catch (e) {
      ElMessage.error(t('vulnerability.fetchFailed') + ': ' + e.message)
      return
    }
    
    if (data.length === 0) {
      ElMessage.warning(t('vulnerability.noDataToExport'))
      return
    }
    
    const headers = ['Target', 'URL', 'POC', 'Severity', 'Source', 'Result', 'CreateTime']
    const csvRows = [headers.join(',')]
    
    for (const row of data) {
      const values = [
        escapeCsvField(row.authority || ''),
        escapeCsvField(row.url || ''),
        escapeCsvField(row.pocFile || ''),
        escapeCsvField(row.severity || ''),
        escapeCsvField(row.source || ''),
        escapeCsvField(row.result || ''),
        escapeCsvField(row.createTime || '')
      ]
      csvRows.push(values.join(','))
    }
    
    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csvRows.join('\n')], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `vulnerabilities_${new Date().toISOString().slice(0, 10)}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success(t('vulnerability.exportSuccess', { count: data.length }))
    return
  } else {
    // 导出全部 - 需要获取所有数据
    ElMessage.info(t('vulnerability.fetchingData'))
    try {
      const res = await request.post('/vul/list', {
        ...searchForm,
        page: 1,
        pageSize: 10000, // 获取全部
        workspaceId: workspaceStore.currentWorkspaceId || ''
      })
      if (res.code === 0) {
        data = res.list || []
      } else {
        ElMessage.error(t('vulnerability.fetchFailed'))
        return
      }
    } catch (e) {
      ElMessage.error(t('vulnerability.fetchFailed') + ': ' + e.message)
      return
    }
    filename = command === 'all-target' ? 'vul_targets_all.txt' : 'vul_urls_all.txt'
  }
  
  if (data.length === 0) {
    ElMessage.warning(t('vulnerability.noDataToExport'))
    return
  }
  
  // 根据类型提取数据并去重
  let exportData = []
  const seen = new Set()
  
  if (command.includes('target')) {
    // 导出目标 (authority)
    for (const row of data) {
      if (row.authority && !seen.has(row.authority)) {
        seen.add(row.authority)
        exportData.push(row.authority)
      }
    }
  } else {
    // 导出URL
    for (const row of data) {
      if (row.url && !seen.has(row.url)) {
        seen.add(row.url)
        exportData.push(row.url)
      }
    }
  }
  
  if (exportData.length === 0) {
    ElMessage.warning(t('vulnerability.noDataToExport'))
    return
  }
  
  // 生成文件并下载
  const content = exportData.join('\n')
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success(t('vulnerability.exportSuccess', { count: exportData.length }))
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
</script>

<style lang="scss" scoped>
.vul-page {
  .search-card {
    margin-bottom: 20px;
  }

  .pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }

  .result-pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 300px;
    overflow: auto;
    background: var(--el-fill-color-darker, #1e1e1e) !important;
    color: var(--el-text-color-primary, #d4d4d4) !important;
    padding: 12px;
    border-radius: 6px;
    font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
    font-size: 13px;
    line-height: 1.5;
    border: 1px solid var(--el-border-color);
  }
}

// 证据链代码块样式 - 使用 :deep 穿透 scoped 样式
:deep(.el-descriptions) {
  .el-descriptions__content {
    .result-pre {
      background: var(--el-fill-color-darker, #1e1e1e) !important;
      color: var(--el-text-color-primary, #d4d4d4) !important;
    }
  }
}

// 对话框内的代码块样式
:deep(.el-dialog) {
  .result-pre {
    background: var(--el-fill-color-darker, #1e1e1e) !important;
    color: var(--el-text-color-primary, #d4d4d4) !important;
  }
  
  .el-descriptions__content .result-pre {
    background: var(--el-fill-color-darker, #1e1e1e) !important;
    color: var(--el-text-color-primary, #d4d4d4) !important;
  }
}

// 暗黑模式下的代码块样式
html.dark {
  .result-pre {
    background: var(--el-fill-color-darker, #1e1e1e) !important;
    color: var(--el-text-color-primary) !important;
    border-color: var(--el-border-color);
  }
}
</style>
