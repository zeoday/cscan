<template>
  <div class="cron-task-page">
    <!-- 操作栏 -->
    <el-card class="action-card">
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>{{ $t('cronTask.newCronTask') }}
      </el-button>
      <el-button @click="loadData">
        <el-icon><Refresh /></el-icon>{{ $t('common.refresh') }}
      </el-button>
      <el-button 
        type="danger" 
        :disabled="selectedRows.length === 0"
        @click="handleBatchDelete"
      >
        <el-icon><Delete /></el-icon>{{ $t('common.batchDelete') }} {{ selectedRows.length > 0 ? `(${selectedRows.length})` : '' }}
      </el-button>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <el-table 
        :data="tableData" 
        v-loading="loading" 
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" :label="$t('cronTask.cronTaskName')" min-width="140" />
        <el-table-column prop="taskName" :label="$t('cronTask.relatedTask')" min-width="140">
          <template #default="{ row }">
            <span class="task-link" @click="goToTask(row)">{{ row.taskName }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="targetShort" :label="$t('cronTask.scanTarget')" min-width="180" show-overflow-tooltip />
        <el-table-column :label="$t('cronTask.scheduleType')" width="180">
          <template #default="{ row }">
            <div v-if="row.scheduleType === 'cron'">
              <el-tag type="primary" size="small">{{ $t('cronTask.cronExec').split(' ')[0] }}</el-tag>
              <el-tooltip :content="getCronDescription(row.cronSpec)" placement="top">
                <code class="cron-code">{{ row.cronSpec }}</code>
              </el-tooltip>
            </div>
            <div v-else>
              <el-tag type="warning" size="small">{{ $t('cronTask.onceExec').split(' ')[0] }}</el-tag>
              <span class="schedule-time">{{ row.scheduleTime }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="$t('cronTask.status')" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              active-value="enable"
              inactive-value="disable"
              @change="handleToggle(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="nextRunTime" :label="$t('cronTask.nextRunTime')" width="160">
          <template #default="{ row }">
            <span v-if="row.status === 'enable' && row.nextRunTime">{{ row.nextRunTime }}</span>
            <span v-else class="text-muted">{{ row.status === 'disable' ? $t('cronTask.disabled') : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="lastRunTime" :label="$t('cronTask.lastRunTime')" width="160">
          <template #default="{ row }">
            {{ row.lastRunTime || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="runCount" :label="$t('cronTask.runCount')" width="90">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.runCount || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="success" link size="small" @click="handleRunNow(row)">
              <el-icon><VideoPlay /></el-icon>{{ $t('cronTask.runNow') }}
            </el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>{{ $t('common.edit') }}
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>{{ $t('common.delete') }}
            </el-button>
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

    <!-- 新建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? $t('cronTask.editCronTask') : $t('cronTask.newCronTask')" width="750px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item :label="$t('cronTask.cronTaskName')" prop="name">
          <el-input v-model="form.name" :placeholder="$t('cronTask.pleaseEnterName')" />
        </el-form-item>
        
        <el-form-item :label="$t('cronTask.relatedTask')" prop="mainTaskId">
          <el-select 
            v-model="form.mainTaskId" 
            :placeholder="$t('cronTask.pleaseSelectTask')" 
            style="width: 100%" 
            filterable
            @change="onTaskSelect"
          >
            <el-option 
              v-for="task in taskList" 
              :key="task.taskId" 
              :label="task.name" 
              :value="task.taskId"
            >
              <div class="task-option">
                <span class="task-name">{{ task.name }}</span>
                <span class="task-target">{{ truncateTarget(task.target) }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('cronTask.scanTarget')" prop="target">
          <el-input 
            v-model="form.target" 
            type="textarea" 
            :rows="4" 
            :placeholder="$t('cronTask.targetPlaceholder')"
          />
          <div class="form-hint">{{ $t('cronTask.targetHint') }}</div>
        </el-form-item>

        <el-form-item :label="$t('cronTask.scheduleType')" prop="scheduleType">
          <el-radio-group v-model="form.scheduleType">
            <el-radio label="cron">{{ $t('cronTask.cronExec') }}</el-radio>
            <el-radio label="once">{{ $t('cronTask.onceExec') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- Cron表达式 -->
        <el-form-item v-if="form.scheduleType === 'cron'" :label="$t('cronTask.cronExpression')" prop="cronSpec">
          <el-input v-model="form.cronSpec" :placeholder="$t('cronTask.cronPlaceholder')">
            <template #append>
              <el-button @click="validateCron">{{ $t('cronTask.validate') }}</el-button>
            </template>
          </el-input>
          <div class="cron-help">
            <div class="cron-presets">
              <span class="preset-label">{{ $t('cronTask.quickSelect') }}</span>
              <el-tag 
                v-for="preset in cronPresets" 
                :key="preset.value" 
                size="small" 
                class="preset-tag"
                @click="form.cronSpec = preset.value; validateCron()"
              >
                {{ preset.label }}
              </el-tag>
            </div>
            <div v-if="cronValidation.valid" class="cron-next-times">
              <div class="next-label">{{ $t('cronTask.next5Times') }}</div>
              <div v-for="(time, index) in cronValidation.nextTimes" :key="index" class="next-time">
                {{ index + 1 }}. {{ time }}
              </div>
            </div>
            <div v-else-if="cronValidation.error" class="cron-error">
              {{ cronValidation.error }}
            </div>
          </div>
        </el-form-item>

        <!-- 指定时间 -->
        <el-form-item v-if="form.scheduleType === 'once'" :label="$t('cronTask.execTime')" prop="scheduleTime">
          <el-date-picker
            v-model="form.scheduleTimeDate"
            type="datetime"
            :placeholder="$t('common.pleaseSelect')"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            :disabled-date="disabledDate"
            style="width: 100%"
            @change="onScheduleTimeChange"
          />
          <div class="form-hint">{{ $t('cronTask.onceExecHint') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Edit, Delete, VideoPlay } from '@element-plus/icons-vue'
import { 
  getCronTaskList, 
  saveCronTask, 
  toggleCronTask, 
  deleteCronTask,
  batchDeleteCronTask,
  runCronTaskNow,
  validateCronSpec 
} from '@/api/crontask'
import { getTaskList } from '@/api/task'

const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: '',
  name: '',
  scheduleType: 'cron',
  cronSpec: '0 0 2 * * *',
  scheduleTime: '',
  scheduleTimeDate: null,
  mainTaskId: '',
  target: '',
  config: ''
})

const rules = {
  name: [{ required: true, message: t('cronTask.pleaseEnterName'), trigger: 'blur' }],
  mainTaskId: [{ required: true, message: t('cronTask.pleaseSelectTask'), trigger: 'change' }],
  scheduleType: [{ required: true, message: t('common.pleaseSelect'), trigger: 'change' }],
  cronSpec: [{ 
    required: true, 
    validator: (rule, value, callback) => {
      if (form.scheduleType === 'cron' && !value) {
        callback(new Error(t('cronTask.cronValidateError')))
      } else {
        callback()
      }
    },
    trigger: 'blur' 
  }],
  scheduleTime: [{
    required: true,
    validator: (rule, value, callback) => {
      if (form.scheduleType === 'once' && !form.scheduleTimeDate) {
        callback(new Error(t('common.pleaseSelect')))
      } else {
        callback()
      }
    },
    trigger: 'change'
  }]
}

const cronPresets = computed(() => [
  { label: t('cronTask.everyHour'), value: '0 0 * * * *' },
  { label: t('cronTask.everyDay2am'), value: '0 0 2 * * *' },
  { label: t('cronTask.everyMonday'), value: '0 0 3 * * 1' },
  { label: t('cronTask.every6hours'), value: '0 0 */6 * * *' }
])

const cronValidation = reactive({
  valid: false,
  nextTimes: [],
  error: ''
})

const taskList = ref([])

const selectedTask = computed(() => {
  return taskList.value.find(t => t.taskId === form.mainTaskId)
})

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const res = await getCronTaskList({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    console.log('定时任务列表响应:', res)
    if (res.code === 0) {
      tableData.value = res.data?.list || []
      pagination.total = res.data?.total || 0
    } else {
      console.error('加载定时任务失败:', res.msg)
    }
  } catch (error) {
    console.error('加载定时任务失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载任务列表（只加载已创建状态的任务作为模板）
async function loadTaskList() {
  try {
    const res = await getTaskList({ page: 1, pageSize: 500 })
    if (res.code === 0) {
      // 过滤出可用的任务（已创建或已完成的任务）
      // 注意：list直接在res下，不是res.data.list
      taskList.value = (res.list || []).filter(t => 
        ['CREATED', 'SUCCESS', 'FAILURE', 'STOPPED'].includes(t.status)
      )
    }
  } catch (error) {
    console.error('加载任务列表失败:', error)
  }
}

// 截取目标显示
function truncateTarget(target, maxLen = 40) {
  if (!target) return ''
  const firstLine = target.split('\n')[0]
  if (firstLine.length > maxLen) {
    return firstLine.substring(0, maxLen) + '...'
  }
  return firstLine
}

// 显示创建对话框
function showCreateDialog() {
  isEdit.value = false
  Object.assign(form, {
    id: '',
    name: '',
    scheduleType: 'cron',
    cronSpec: '0 0 2 * * *',
    scheduleTime: '',
    scheduleTimeDate: null,
    mainTaskId: '',
    target: '',
    config: ''
  })
  cronValidation.valid = false
  cronValidation.nextTimes = []
  cronValidation.error = ''
  dialogVisible.value = true
  loadTaskList()
}

// 编辑
function handleEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    name: row.name,
    scheduleType: row.scheduleType || 'cron',
    cronSpec: row.cronSpec || '',
    scheduleTime: row.scheduleTime || '',
    scheduleTimeDate: row.scheduleTime || null,
    mainTaskId: row.mainTaskId,
    target: row.target || '',
    config: row.config || ''
  })
  cronValidation.valid = false
  cronValidation.nextTimes = []
  cronValidation.error = ''
  dialogVisible.value = true
  loadTaskList()
  if (form.scheduleType === 'cron' && form.cronSpec) {
    validateCron()
  }
}

// 选择任务时自动填充名称和目标
function onTaskSelect(taskId) {
  const task = taskList.value.find(t => t.taskId === taskId)
  if (task) {
    if (!form.name) {
      form.name = `${t('cronTask.title')}-${task.name}`
    }
    // 新建时自动填充目标
    if (!isEdit.value && !form.target) {
      form.target = task.target || ''
    }
  }
}

// 时间选择变化
function onScheduleTimeChange(val) {
  form.scheduleTime = val
}

// 禁用过去的日期
function disabledDate(time) {
  return time.getTime() < Date.now() - 24 * 60 * 60 * 1000
}

// 验证Cron表达式
async function validateCron() {
  if (!form.cronSpec) {
    cronValidation.valid = false
    cronValidation.error = t('cronTask.cronValidateError')
    return
  }
  try {
    const res = await validateCronSpec({ cronSpec: form.cronSpec })
    if (res.code === 0 && res.data?.valid) {
      cronValidation.valid = true
      cronValidation.nextTimes = res.data.nextTimes || []
      cronValidation.error = ''
    } else {
      cronValidation.valid = false
      cronValidation.nextTimes = []
      cronValidation.error = res.msg || t('cronTask.cronValidateError')
    }
  } catch (error) {
    cronValidation.valid = false
    cronValidation.error = t('cronTask.cronValidateError')
  }
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value.validate()
  } catch {
    return
  }

  // 额外验证
  if (form.scheduleType === 'once' && !form.scheduleTimeDate) {
    ElMessage.error(t('common.pleaseSelect'))
    return
  }

  submitting.value = true
  try {
    // 获取选中任务的workspaceId
    const task = selectedTask.value
    const data = {
      id: form.id,
      name: form.name,
      scheduleType: form.scheduleType,
      cronSpec: form.scheduleType === 'cron' ? form.cronSpec : '',
      scheduleTime: form.scheduleType === 'once' ? form.scheduleTime : '',
      mainTaskId: form.mainTaskId,
      workspaceId: task?.workspaceId || '',
      target: form.target,
      config: form.config
    }
    const res = await saveCronTask(data)
    if (res.code === 0) {
      ElMessage.success(isEdit.value ? t('cronTask.updateSuccess') : t('cronTask.createSuccess'))
      dialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    ElMessage.error(t('common.operationFailed'))
  } finally {
    submitting.value = false
  }
}

// 开关任务
async function handleToggle(row) {
  try {
    const res = await toggleCronTask({ id: row.id, status: row.status })
    if (res.code === 0) {
      ElMessage.success(t('cronTask.statusUpdateSuccess'))
      loadData()
    } else {
      row.status = row.status === 'enable' ? 'disable' : 'enable'
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    row.status = row.status === 'enable' ? 'disable' : 'enable'
    ElMessage.error(t('common.operationFailed'))
  }
}

// 立即执行
async function handleRunNow(row) {
  try {
    await ElMessageBox.confirm(t('cronTask.runNow') + '?', t('common.confirm'), { type: 'warning' })
    const res = await runCronTaskNow({ id: row.id })
    if (res.code === 0) {
      ElMessage.success(t('cronTask.runSuccess'))
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('common.operationFailed'))
    }
  }
}

// 删除
async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(t('cronTask.confirmDelete'), t('common.confirm'), { type: 'warning' })
    const res = await deleteCronTask({ id: row.id })
    if (res.code === 0) {
      ElMessage.success(t('cronTask.deleteSuccess'))
      loadData()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('common.operationFailed'))
    }
  }
}

// 选择变化
function handleSelectionChange(rows) {
  selectedRows.value = rows
}

// 批量删除
async function handleBatchDelete() {
  if (selectedRows.value.length === 0) {
    ElMessage.warning(t('common.pleaseSelect'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('cronTask.confirmBatchDelete', { count: selectedRows.value.length }), 
      t('common.batchDelete'), 
      { type: 'warning' }
    )
    
    const ids = selectedRows.value.map(row => row.id)
    const res = await batchDeleteCronTask({ ids })
    
    if (res.code === 0) {
      ElMessage.success(res.msg || t('cronTask.deleteSuccess'))
      selectedRows.value = []
      loadData()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('common.operationFailed'))
    }
  }
}

// 跳转到任务详情
function goToTask(row) {
  router.push('/task')
}

// 获取Cron描述
function getCronDescription(cronSpec) {
  if (!cronSpec) return ''
  const parts = cronSpec.split(' ')
  if (parts.length !== 6) return cronSpec
  
  const [sec, min, hour, day, month, week] = parts
  let desc = ''
  
  if (week !== '*') {
    const weekNames = [t('common.sunday') || 'Sun', t('common.monday') || 'Mon', t('common.tuesday') || 'Tue', t('common.wednesday') || 'Wed', t('common.thursday') || 'Thu', t('common.friday') || 'Fri', t('common.saturday') || 'Sat']
    desc += `${weekNames[parseInt(week)] || week} `
  }
  if (month !== '*') desc += `${month}M `
  if (day !== '*') desc += `${day}D `
  if (hour !== '*') desc += `${hour}h`
  if (min !== '*') desc += `${min}m`
  if (sec !== '*' && sec !== '0') desc += `${sec}s`
  
  return desc || cronSpec
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.cron-task-page {
  padding: 20px;
}

.action-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
}

.cron-code {
  background: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
  margin-left: 6px;
}

.schedule-time {
  font-size: 12px;
  color: var(--el-text-color-regular);
  margin-left: 6px;
}

.text-muted {
  color: var(--el-text-color-placeholder);
}

.task-link {
  color: var(--el-color-primary);
  cursor: pointer;
}

.task-link:hover {
  text-decoration: underline;
}

.task-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.task-option .task-name {
  flex: 1;
}

.task-option .task-target {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.selected-task-info {
  margin-top: 8px;
}

.cron-help {
  margin-top: 10px;
}

.cron-presets {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.preset-label {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.preset-tag {
  cursor: pointer;
}

.preset-tag:hover {
  background: var(--el-color-primary-light-7);
}

.cron-next-times {
  background: var(--el-fill-color-lighter);
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
}

.next-label {
  color: var(--el-text-color-secondary);
  margin-bottom: 5px;
}

.next-time {
  color: var(--el-text-color-regular);
  line-height: 1.8;
}

.cron-error {
  color: var(--el-color-danger);
  font-size: 12px;
}

.form-hint {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
  margin-top: 5px;
}
</style>
