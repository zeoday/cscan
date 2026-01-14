<template>
  <div class="worker-logs-page">
    <el-card>
      <template #header>
        <div class="log-header">
          <span>{{ $t('worker.workerRunningLogs') }}</span>
          <div class="log-filters">
            <el-input
              v-model="searchKeyword"
              :placeholder="$t('worker.searchLogs')"
              clearable
              size="small"
              style="width: 180px; margin-right: 10px"
              @keyup.enter="searchLogs"
              @clear="searchLogs"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-select 
              v-model="filterWorker" 
              :placeholder="$t('worker.filterWorker')" 
              clearable 
              size="small"
              style="width: 150px; margin-right: 10px"
            >
              <el-option :label="$t('worker.allWorkers')" value="" />
              <el-option 
                v-for="worker in workerList" 
                :key="worker.name" 
                :label="worker.name" 
                :value="worker.name" 
              />
            </el-select>
            <el-select 
              v-model="filterLevel" 
              :placeholder="$t('worker.filterLevel')" 
              clearable 
              size="small"
              style="width: 120px; margin-right: 10px"
            >
              <el-option :label="$t('worker.allLevels')" value="" />
              <el-option label="INFO" value="INFO" />
              <el-option label="WARN" value="WARN" />
              <el-option label="ERROR" value="ERROR" />
              <el-option label="DEBUG" value="DEBUG" />
            </el-select>
            <el-switch v-model="autoScroll" :active-text="$t('worker.autoScrolling')" style="margin-right: 15px" />
            <el-button size="small" @click="clearLogs">{{ $t('worker.clear') }}</el-button>
            <el-button size="small" :type="isConnected ? 'success' : 'danger'" @click="toggleConnection">
              {{ isConnected ? $t('worker.autoRefreshing') : $t('worker.paused') }}
            </el-button>
          </div>
        </div>
      </template>
      <div ref="logContainer" class="log-container">
        <div v-for="(log, index) in filteredLogs" :key="index" class="log-item" :class="'log-' + log.level?.toLowerCase()">
          <span class="log-time">{{ log.timestamp }}</span>
          <span class="log-level">[{{ log.level }}]</span>
          <span class="log-worker">[{{ log.workerName }}]</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
        <div v-if="filteredLogs.length === 0 && logs.length > 0" class="log-empty">{{ $t('worker.noMatchingLogs') }}</div>
        <div v-if="logs.length === 0" class="log-empty">{{ $t('worker.noLogsYet') }}</div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const logs = ref([])
const logContainer = ref(null)
const autoScroll = ref(true)
const isConnected = ref(false)
const filterWorker = ref('')
const filterLevel = ref('')
const searchKeyword = ref('')
const workerList = ref([])
let pollingTimer = null
let logIdSet = new Set()

const filteredLogs = computed(() => {
  const keyword = searchKeyword.value.toLowerCase()
  return logs.value.filter(log => {
    if (filterWorker.value && log.workerName !== filterWorker.value) return false
    if (filterLevel.value && log.level?.toUpperCase() !== filterLevel.value) return false
    if (keyword) {
      const message = (log.message || '').toLowerCase()
      const level = (log.level || '').toLowerCase()
      const workerName = (log.workerName || '').toLowerCase()
      if (!message.includes(keyword) && !level.includes(keyword) && !workerName.includes(keyword)) return false
    }
    return true
  })
})

onMounted(() => {
  loadWorkerList()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})

watch(filteredLogs, () => {
  if (autoScroll.value) {
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }
}, { deep: true })

async function loadWorkerList() {
  try {
    const res = await request.post('/worker/list')
    if (res.code === 0) workerList.value = res.list || []
  } catch (e) {
    console.error('Load worker list error:', e)
  }
}

function startPolling() {
  if (pollingTimer) return
  isConnected.value = true
  fetchLogsHistory()
  pollingTimer = setInterval(fetchLogsHistory, 2000)
}

function stopPolling() {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
  isConnected.value = false
}

async function fetchLogsHistory() {
  try {
    const res = await request.post('/worker/logs/history', { limit: 200 })
    if (res.code === 0 && res.list && res.list.length > 0) {
      for (const log of res.list) {
        const logId = (log.timestamp || '') + (log.message || '')
        if (!logIdSet.has(logId)) {
          logIdSet.add(logId)
          logs.value.push(log)
        }
      }
      if (logs.value.length > 1000) {
        const removed = logs.value.splice(0, logs.value.length - 500)
        removed.forEach(l => logIdSet.delete((l.timestamp || '') + (l.message || '')))
      }
    }
  } catch (e) {
    console.error('[Polling] Fetch logs error:', e)
    isConnected.value = false
  }
}

function toggleConnection() {
  if (isConnected.value) {
    stopPolling()
  } else {
    startPolling()
  }
}

async function clearLogs() {
  try {
    const res = await request.post('/worker/logs/clear')
    if (res.code === 0) {
      logs.value = []
      logIdSet.clear()
    }
  } catch (e) {
    console.error('Clear logs error:', e)
  }
}

function searchLogs() {
  // 前端过滤，无需额外操作
}
</script>

<style lang="scss" scoped>
.worker-logs-page {
  .log-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    
    .log-filters {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 5px;
    }
  }

  .log-container {
    height: 500px;
    overflow-y: auto;
    background: hsl(var(--muted));
    border: 1px solid hsl(var(--border));
    border-radius: 4px;
    padding: 10px;
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 12px;
  }

  .log-item {
    padding: 2px 0;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;

    .log-time { color: hsl(var(--muted-foreground)); margin-right: 10px; }
    .log-level { display: inline-block; width: 60px; margin-right: 8px; font-weight: bold; }
    .log-worker { color: hsl(var(--primary)); margin-right: 8px; }
    .log-message { color: hsl(var(--foreground)); }

    &.log-info .log-level { color: var(--el-color-success); }
    &.log-warn .log-level { color: var(--el-color-warning); }
    &.log-error .log-level { color: var(--el-color-danger); }
    &.log-debug .log-level { color: var(--el-color-primary); }
  }

  .log-empty {
    color: hsl(var(--muted-foreground));
    text-align: center;
    padding: 50px;
  }
}
</style>
