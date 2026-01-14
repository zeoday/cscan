<template>
  <div class="settings-page">
    <!-- 在线API配置 -->
    <el-card v-if="activeTab === 'onlineapi'">
      <template #header>
        <div class="card-header">
          <span>{{ $t('navigation.onlineApiConfig') }}</span>
        </div>
      </template>
      <el-alert type="info" :closable="false" style="margin-bottom: 20px">
        <template #title>{{ $t('settings.onlineApiTip') }}</template>
      </el-alert>
      
      <el-tabs v-model="apiConfigTab" type="card">
        <el-tab-pane label="Fofa" name="fofa">
          <el-form label-width="100px" style="max-width: 500px; margin-top: 20px">
            <el-form-item :label="$t('settings.apiVersion')">
              <el-radio-group v-model="apiConfigs.fofa.version">
                <el-radio value="v1">v1 (fofa.info)</el-radio>
                <el-radio value="v5">v5 (v5.fofa.info)</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('settings.apiKey')">
              <el-input v-model="apiConfigs.fofa.key" placeholder="Fofa API Key" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveApiConfig('fofa')">{{ $t('common.save') }}</el-button>
              <el-button type="success" @click="openApiUrl(apiConfigs.fofa.version === 'v5' ? 'https://v5.fofa.info/userInfo' : 'https://fofa.info/userInfo')">{{ $t('settings.applyApi') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="Hunter" name="hunter">
          <el-form label-width="100px" style="max-width: 500px; margin-top: 20px">
            <el-form-item :label="$t('settings.apiKey')">
              <el-input v-model="apiConfigs.hunter.key" placeholder="Hunter API Key" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveApiConfig('hunter')">{{ $t('common.save') }}</el-button>
              <el-button type="success" @click="openApiUrl('https://hunter.qianxin.com/home/myInfo')">{{ $t('settings.applyApi') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="Quake" name="quake">
          <el-form label-width="100px" style="max-width: 500px; margin-top: 20px">
            <el-form-item :label="$t('settings.apiKey')">
              <el-input v-model="apiConfigs.quake.key" placeholder="Quake API Key" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveApiConfig('quake')">{{ $t('common.save') }}</el-button>
              <el-button type="success" @click="openApiUrl('https://quake.360.net/quake/#/personal?tab=message')">{{ $t('settings.applyApi') }}</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Subfinder数据源配置 -->
    <el-card v-else-if="activeTab === 'subfinder'">
      <template #header>
        <div class="card-header">
          <span>{{ $t('navigation.subdomainConfig') }}</span>
        </div>
      </template>
      <el-alert type="info" :closable="false" style="margin-bottom: 20px">
        <template #title>{{ $t('settings.subfinderTip') }}</template>
      </el-alert>
      
      <el-table :data="subfinderProviders" v-loading="subfinderLoading" max-height="500" stripe>
        <el-table-column prop="name" :label="$t('settings.dataSource')" width="130" />
        <el-table-column prop="description" :label="$t('common.description')" width="180" show-overflow-tooltip />
        <el-table-column prop="keyFormat" :label="$t('settings.keyFormat')" width="140" />
        <el-table-column :label="$t('settings.apiKeyColumn')" min-width="200">
          <template #default="{ row }">
            <el-input v-model="row.keyInput" :placeholder="row.maskedKey || row.keyFormat" size="small" clearable />
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.status')" width="70">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" size="small" />
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.operation')" width="140">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="saveSubfinderProvider(row)">{{ $t('common.save') }}</el-button>
            <el-button type="success" link size="small" @click="openApiUrl(row.url)">{{ $t('settings.applyApi') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 工作空间 -->
    <el-card v-else-if="activeTab === 'workspace'">
      <template #header>
        <div class="card-header">
          <span>{{ $t('navigation.workspaceManagement') }}</span>
          <el-button type="primary" size="small" @click="showWorkspaceDialog()">
            <el-icon><Plus /></el-icon>{{ $t('workspace.newWorkspace') }}
          </el-button>
        </div>
      </template>
      <el-table :data="workspaceList" v-loading="workspaceLoading" stripe max-height="500">
        <el-table-column prop="name" :label="$t('common.name')" min-width="150" />
        <el-table-column prop="description" :label="$t('common.description')" min-width="250" />
        <el-table-column prop="status" :label="$t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enable' ? 'success' : 'danger'">
              {{ row.status === 'enable' ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" :label="$t('common.createTime')" width="160" />
        <el-table-column :label="$t('common.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showWorkspaceDialog(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteWorkspace(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 组织管理 -->
    <el-card v-else-if="activeTab === 'organization'">
      <template #header>
        <div class="card-header">
          <span>{{ $t('navigation.organizationManagement') }}</span>
          <el-button type="primary" size="small" @click="showOrgDialog()">
            <el-icon><Plus /></el-icon>{{ $t('organization.newOrganization') }}
          </el-button>
        </div>
      </template>
      <el-table :data="orgList" v-loading="orgLoading" stripe max-height="500">
        <el-table-column prop="name" :label="$t('organization.organizationName')" min-width="150" />
        <el-table-column prop="description" :label="$t('common.description')" min-width="250" />
        <el-table-column prop="status" :label="$t('common.status')" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.status" active-value="enable" inactive-value="disable" @change="handleOrgStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="createTime" :label="$t('common.createTime')" width="160" />
        <el-table-column :label="$t('common.operation')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showOrgDialog(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteOrg(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 通知配置 -->
    <el-card v-else-if="activeTab === 'notify'">
      <template #header>
        <div class="card-header">
          <span>{{ $t('navigation.notifyConfig') }}</span>
          <el-button type="primary" size="small" @click="showNotifyDialog()">
            <el-icon><Plus /></el-icon>{{ $t('settings.addNotifyChannel') }}
          </el-button>
        </div>
      </template>
      <el-alert type="info" :closable="false" style="margin-bottom: 20px">
        <template #title>{{ $t('settings.notifyTip') }}</template>
      </el-alert>
      
      <el-table :data="notifyConfigList" v-loading="notifyLoading" stripe max-height="500">
        <el-table-column prop="name" :label="$t('common.name')" min-width="120" />
        <el-table-column :label="$t('settings.channel')" width="140">
          <template #default="{ row }">
            <el-tag>{{ getProviderName(row.provider) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="$t('common.status')" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.status" active-value="enable" inactive-value="disable" @change="handleNotifyStatusChange(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="updateTime" :label="$t('common.updateTime')" width="160" />
        <el-table-column :label="$t('common.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showNotifyDialog(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="success" link size="small" @click="handleTestNotify(row)" :loading="row.testing">{{ $t('settings.test') }}</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteNotify(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 用户管理 -->
    <el-card v-else-if="activeTab === 'user'">
      <template #header>
        <div class="card-header">
          <span>{{ $t('navigation.userManagement') }}</span>
          <el-button type="primary" size="small" @click="showUserDialog()">
            <el-icon><Plus /></el-icon>{{ $t('user.newUser') }}
          </el-button>
        </div>
      </template>
      <el-table :data="userList" v-loading="userLoading" stripe max-height="500">
        <el-table-column prop="username" :label="$t('user.userName')" min-width="150" />
        <el-table-column prop="status" :label="$t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'enable' ? 'success' : 'danger'">
              {{ row.status === 'enable' ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('common.operation')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showUserDialog(row)">{{ $t('common.edit') }}</el-button>
            <el-button type="warning" link size="small" @click="showResetPasswordDialog(row)">{{ $t('user.resetPassword') }}</el-button>
            <el-button type="danger" link size="small" @click="handleDeleteUser(row)">{{ $t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 工作空间对话框 -->
    <el-dialog v-model="workspaceDialogVisible" :title="workspaceForm.id ? $t('workspace.editWorkspace') : $t('workspace.newWorkspace')" width="500px">
      <el-form ref="workspaceFormRef" :model="workspaceForm" :rules="workspaceRules" label-width="80px">
        <el-form-item :label="$t('common.name')" prop="name">
          <el-input v-model="workspaceForm.name" :placeholder="$t('workspace.pleaseEnterName')" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="workspaceForm.description" type="textarea" :rows="3" :placeholder="$t('workspace.pleaseEnterDescription')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="workspaceDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="workspaceSubmitting" @click="handleWorkspaceSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 用户对话框 -->
    <el-dialog v-model="userDialogVisible" :title="userForm.id ? $t('user.editUser') : $t('user.newUser')" width="500px">
      <el-form ref="userFormRef" :model="userForm" :rules="userRules" label-width="80px">
        <el-form-item :label="$t('user.userName')" prop="username">
          <el-input v-model="userForm.username" :placeholder="$t('user.pleaseEnterUsername')" />
        </el-form-item>
        <el-form-item v-if="!userForm.id" :label="$t('user.password')" prop="password">
          <el-input v-model="userForm.password" type="password" :placeholder="$t('user.pleaseEnterPassword')" />
        </el-form-item>
        <el-form-item :label="$t('common.status')" prop="status">
          <el-select v-model="userForm.status" :placeholder="$t('user.pleaseSelectStatus')">
            <el-option :label="$t('common.enabled')" value="enable" />
            <el-option :label="$t('common.disabled')" value="disable" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="userSubmitting" @click="handleUserSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog v-model="resetPasswordVisible" :title="$t('user.resetPassword')" width="400px">
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="80px">
        <el-form-item :label="$t('user.newPassword')" prop="newPassword">
          <el-input v-model="resetForm.newPassword" type="password" :placeholder="$t('user.pleaseEnterNewPassword')" />
        </el-form-item>
        <el-form-item :label="$t('user.confirmPassword')" prop="confirmPassword">
          <el-input v-model="resetForm.confirmPassword" type="password" :placeholder="$t('user.pleaseConfirmPassword')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPasswordVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="resetting" @click="handleResetPassword">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 组织对话框 -->
    <el-dialog v-model="orgDialogVisible" :title="orgForm.id ? $t('organization.editOrganization') : $t('organization.newOrganization')" width="500px">
      <el-form ref="orgFormRef" :model="orgForm" :rules="orgRules" label-width="80px">
        <el-form-item :label="$t('common.name')" prop="name">
          <el-input v-model="orgForm.name" :placeholder="$t('organization.pleaseEnterOrgName')" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model="orgForm.description" type="textarea" :rows="3" :placeholder="$t('organization.pleaseEnterDescription')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orgDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="orgSubmitting" @click="handleOrgSubmit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 通知配置对话框 -->
    <el-dialog v-model="notifyDialogVisible" :title="notifyForm.id ? $t('settings.editNotifyConfig') : $t('settings.addNotifyChannelTitle')" width="700px">
      <el-form ref="notifyFormRef" :model="notifyForm" :rules="notifyRules" label-width="120px">
        <el-form-item :label="$t('settings.channelType')" prop="provider">
          <el-select v-model="notifyForm.provider" :placeholder="$t('settings.selectNotifyChannel')" @change="handleProviderChange" :disabled="!!notifyForm.id">
            <el-option v-for="p in notifyProviders" :key="p.id" :label="p.name" :value="p.id">
              <span>{{ p.name }}</span>
              <span class="option-desc">{{ p.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('settings.configName')" prop="name">
          <el-input v-model="notifyForm.name" :placeholder="$t('settings.enterConfigName')" />
        </el-form-item>
        
        <!-- 动态配置字段 -->
        <template v-if="currentProviderFields.length > 0">
          <el-divider content-position="left">{{ $t('settings.channelConfig') }}</el-divider>
          <el-form-item 
            v-for="field in currentProviderFields" 
            :key="field.name" 
            :label="field.label"
            :prop="'configData.' + field.name"
            :rules="field.required ? [{ required: true, message: t('settings.pleaseEnterInput') + field.label, trigger: 'blur' }] : []"
          >
            <el-input 
              v-if="field.type === 'text'" 
              v-model="notifyForm.configData[field.name]" 
              :placeholder="field.placeholder" 
            />
            <el-input 
              v-else-if="field.type === 'password'" 
              v-model="notifyForm.configData[field.name]" 
              :placeholder="field.placeholder" 
              show-password 
            />
            <el-input 
              v-else-if="field.type === 'textarea'" 
              v-model="notifyForm.configData[field.name]" 
              type="textarea" 
              :rows="3" 
              :placeholder="field.placeholder" 
            />
            <el-input-number 
              v-else-if="field.type === 'number'" 
              v-model="notifyForm.configData[field.name]" 
              :placeholder="field.placeholder" 
              controls-position="right"
            />
            <el-switch 
              v-else-if="field.type === 'switch'" 
              v-model="notifyForm.configData[field.name]" 
            />
            <el-select 
              v-else-if="field.type === 'select'" 
              v-model="notifyForm.configData[field.name]" 
              :placeholder="field.placeholder || $t('common.pleaseSelect')"
              clearable
            >
              <el-option v-for="opt in field.options" :key="opt" :label="opt || $t('common.default')" :value="opt" />
            </el-select>
          </el-form-item>
        </template>
        
        <el-divider content-position="left">{{ $t('settings.messageTemplate') }}</el-divider>
        <el-form-item :label="$t('settings.customTemplate')">
          <el-input 
            v-model="notifyForm.messageTemplate" 
            type="textarea" 
            :rows="4" 
            :placeholder="$t('settings.templatePlaceholder')" 
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="notifyDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="success" @click="handleTestNotifyForm" :loading="notifyTesting">{{ $t('settings.test') }}</el-button>
        <el-button type="primary" :loading="notifySubmitting" @click="handleNotifySubmit">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>


<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import request from '@/api/request'
import { getSubfinderProviderList, getSubfinderProviderInfo, saveSubfinderProvider as saveSubfinderProviderApi } from '@/api/subfinder'
import { getUserList, createUser, updateUser, deleteUser, resetUserPassword } from '@/api/auth'
import { getNotifyProviders, getNotifyConfigList, saveNotifyConfig, deleteNotifyConfig, testNotifyConfig } from '@/api/notify'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()

// 有效的tab名称
const validTabs = ['onlineapi', 'subfinder', 'workspace', 'organization', 'notify', 'user']

// 从URL获取当前tab
const activeTab = computed(() => {
  const tab = route.query.tab
  return validTabs.includes(tab) ? tab : 'onlineapi'
})
const apiConfigTab = ref('fofa')
const subfinderLoading = ref(false)
const subfinderProviders = ref([])

const apiConfigs = reactive({
  fofa: { key: '', secret: '', version: 'v1' },
  hunter: { key: '', secret: '' },
  quake: { key: '', secret: '' }
})

// 工作空间相关
const workspaceLoading = ref(false)
const workspaceList = ref([])
const workspaceDialogVisible = ref(false)
const workspaceSubmitting = ref(false)
const workspaceFormRef = ref()
const workspaceForm = reactive({ id: '', name: '', description: '' })
const workspaceRules = computed(() => ({
  name: [{ required: true, message: t('workspace.pleaseEnterName'), trigger: 'blur' }]
}))

// 用户管理相关
const userLoading = ref(false)
const userList = ref([])
const userDialogVisible = ref(false)
const userSubmitting = ref(false)
const userFormRef = ref()
const userForm = ref({ id: '', username: '', password: '', status: 'enable' })
const userRules = computed(() => ({
  username: [{ required: true, message: t('user.pleaseEnterUsername'), trigger: 'blur' }],
  password: [{ required: true, message: t('user.pleaseEnterPassword'), trigger: 'blur' }],
  status: [{ required: true, message: t('user.pleaseSelectStatus'), trigger: 'change' }]
}))

// 重置密码相关
const resetPasswordVisible = ref(false)
const resetting = ref(false)
const resetFormRef = ref()
const resetForm = ref({ id: '', newPassword: '', confirmPassword: '' })
const resetRules = computed(() => ({
  newPassword: [{ required: true, message: t('user.pleaseEnterNewPassword'), trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: t('user.pleaseConfirmPassword'), trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== resetForm.value.newPassword) {
          callback(new Error(t('user.passwordMismatch')))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}))

// 组织管理相关
const orgLoading = ref(false)
const orgList = ref([])
const orgDialogVisible = ref(false)
const orgSubmitting = ref(false)
const orgFormRef = ref()
const orgForm = reactive({ id: '', name: '', description: '' })
const orgRules = computed(() => ({
  name: [{ required: true, message: t('organization.pleaseEnterOrgName'), trigger: 'blur' }]
}))

// 通知配置相关
const notifyLoading = ref(false)
const notifyConfigList = ref([])
const notifyProviders = ref([])
const notifyDialogVisible = ref(false)
const notifySubmitting = ref(false)
const notifyTesting = ref(false)
const notifyFormRef = ref()
const notifyForm = ref({ 
  id: '', 
  name: '', 
  provider: '', 
  configData: {}, 
  messageTemplate: '', 
  status: 'enable'
})
const notifyRules = computed(() => ({
  provider: [{ required: true, message: t('settings.selectNotifyChannel'), trigger: 'change' }],
  name: [{ required: true, message: t('settings.enterConfigName'), trigger: 'blur' }]
}))
const currentProviderFields = ref([])

onMounted(() => {
  // 根据当前tab加载对应数据
  loadDataByTab(activeTab.value)
})

// 监听tab变化，加载对应数据
watch(activeTab, (val) => {
  loadDataByTab(val)
})

function loadDataByTab(tab) {
  if (tab === 'onlineapi') {
    loadApiConfigs()
  } else if (tab === 'subfinder') {
    loadSubfinderProviders()
  } else if (tab === 'workspace') {
    loadWorkspaceList()
  } else if (tab === 'user') {
    loadUserList()
  } else if (tab === 'organization') {
    loadOrgList()
  } else if (tab === 'notify') {
    loadNotifyProviders()
    loadNotifyConfigList()
  }
}

// 在线API配置
async function loadApiConfigs() {
  const res = await request.post('/onlineapi/config/list', {})
  if (res.code === 0 && res.list) {
    res.list.forEach(item => {
      if (apiConfigs[item.platform]) {
        apiConfigs[item.platform].key = item.key
        apiConfigs[item.platform].secret = item.secret
        if (item.platform === 'fofa' && item.version) {
          apiConfigs.fofa.version = item.version
        }
      }
    })
  }
}

async function saveApiConfig(platform) {
  const config = apiConfigs[platform]
  const data = {
    platform,
    key: config.key,
    secret: config.secret
  }
  // Fofa需要传递版本信?
  if (platform === 'fofa') {
    data.version = config.version
  }
  const res = await request.post('/onlineapi/config/save', data)
  if (res.code === 0) {
    ElMessage.success(t('common.operationSuccess'))
  } else {
    ElMessage.error(res.msg || t('common.operationFailed'))
  }
}

// Subfinder配置
async function loadSubfinderProviders() {
  subfinderLoading.value = true
  try {
    const infoRes = await getSubfinderProviderInfo()
    if (infoRes.code !== 0) {
      ElMessage.error(infoRes.msg || '获取数据源信息失败')
      return
    }

    const listRes = await getSubfinderProviderList()
    const configuredMap = {}
    if (listRes.code === 0 && listRes.list) {
      listRes.list.forEach(item => {
        configuredMap[item.provider] = item
      })
    }

    subfinderProviders.value = infoRes.list.map(info => {
      const configured = configuredMap[info.provider]
      return {
        ...info,
        keyInput: '',
        enabled: configured ? configured.status === 'enable' : false,
        configured: !!configured,
        maskedKey: configured && configured.keys?.length > 0 ? configured.keys[0] : ''
      }
    })
  } finally {
    subfinderLoading.value = false
  }
}

async function saveSubfinderProvider(row) {
  if (!row.keyInput && !row.configured) {
    ElMessage.warning(t('settings.pleaseEnterInput') + t('settings.apiKey'))
    return
  }

  const data = {
    provider: row.provider,
    keys: row.keyInput ? [row.keyInput] : [],
    status: row.enabled ? 'enable' : 'disable',
    description: row.description
  }

  const res = await saveSubfinderProviderApi(data)
  if (res.code === 0) {
    ElMessage.success(t('common.operationSuccess'))
    row.configured = true
    row.keyInput = ''
    await loadSubfinderProviders()
  } else {
    ElMessage.error(res.msg || t('common.operationFailed'))
  }
}

function openApiUrl(url) {
  window.open(url, '_blank')
}

// 工作空间管理
async function loadWorkspaceList() {
  workspaceLoading.value = true
  try {
    const res = await request.post('/workspace/list', { page: 1, pageSize: 100 })
    if (res.code === 0) workspaceList.value = res.list || []
  } finally {
    workspaceLoading.value = false
  }
}

function showWorkspaceDialog(row = null) {
  if (row) {
    Object.assign(workspaceForm, { id: row.id, name: row.name, description: row.description })
  } else {
    Object.assign(workspaceForm, { id: '', name: '', description: '' })
  }
  workspaceDialogVisible.value = true
}

async function handleWorkspaceSubmit() {
  await workspaceFormRef.value.validate()
  workspaceSubmitting.value = true
  try {
    const res = await request.post('/workspace/save', workspaceForm)
    if (res.code === 0) {
      ElMessage.success(workspaceForm.id ? t('common.updateSuccess') : t('common.createSuccess'))
      workspaceDialogVisible.value = false
      loadWorkspaceList()
    } else {
      ElMessage.error(res.msg)
    }
  } finally {
    workspaceSubmitting.value = false
  }
}

async function handleDeleteWorkspace(row) {
  await ElMessageBox.confirm(t('workspace.confirmDeleteWorkspace'), t('common.tip'), { type: 'warning' })
  const res = await request.post('/workspace/delete', { id: row.id })
  if (res.code === 0) {
    ElMessage.success(t('common.deleteSuccess'))
    loadWorkspaceList()
  }
}

// 用户管理
async function loadUserList() {
  userLoading.value = true
  try {
    const res = await getUserList({ page: 1, pageSize: 100 })
    if (res.code === 0) userList.value = res.list || []
  } finally {
    userLoading.value = false
  }
}

function showUserDialog(row = null) {
  if (row) {
    userForm.value = { ...row, password: '' }
  } else {
    userForm.value = { id: '', username: '', password: '', status: 'enable' }
  }
  userDialogVisible.value = true
}

async function handleUserSubmit() {
  if (!userFormRef.value) return
  try {
    await userFormRef.value.validate()
    userSubmitting.value = true
    const api = userForm.value.id ? updateUser : createUser
    const res = await api(userForm.value)
    if (res.code === 0) {
      ElMessage.success(res.msg || t('common.operationSuccess'))
      userDialogVisible.value = false
      loadUserList()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    userSubmitting.value = false
  }
}

async function handleDeleteUser(row) {
  try {
    await ElMessageBox.confirm(t('user.confirmDeleteUser'), t('common.tip'), { type: 'warning' })
    const res = await deleteUser({ id: row.id })
    if (res.code === 0) {
      ElMessage.success(res.msg || t('common.deleteSuccess'))
      loadUserList()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {}
}

function showResetPasswordDialog(row) {
  resetForm.value = { id: row.id, newPassword: '', confirmPassword: '' }
  resetPasswordVisible.value = true
}

async function handleResetPassword() {
  if (!resetFormRef.value) return
  try {
    await resetFormRef.value.validate()
    resetting.value = true
    const res = await resetUserPassword({
      id: resetForm.value.id,
      newPassword: resetForm.value.newPassword
    })
    if (res.code === 0) {
      ElMessage.success(res.msg || t('user.passwordResetSuccess'))
      resetPasswordVisible.value = false
    } else {
      ElMessage.error(res.msg || t('user.passwordResetFailed'))
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    resetting.value = false
  }
}

// 组织管理
async function loadOrgList() {
  orgLoading.value = true
  try {
    const res = await request.post('/organization/list', { page: 1, pageSize: 100 })
    const data = res.data || res
    if (data.code === 0) orgList.value = data.list || []
  } finally {
    orgLoading.value = false
  }
}

function showOrgDialog(row = null) {
  if (row) {
    Object.assign(orgForm, { id: row.id, name: row.name, description: row.description })
  } else {
    Object.assign(orgForm, { id: '', name: '', description: '' })
  }
  orgDialogVisible.value = true
}

async function handleOrgSubmit() {
  await orgFormRef.value.validate()
  orgSubmitting.value = true
  try {
    const res = await request.post('/organization/save', orgForm)
    const data = res.data || res
    if (data.code === 0) {
      ElMessage.success(orgForm.id ? t('common.updateSuccess') : t('common.createSuccess'))
      orgDialogVisible.value = false
      loadOrgList()
    } else {
      ElMessage.error(data.msg)
    }
  } finally {
    orgSubmitting.value = false
  }
}

async function handleDeleteOrg(row) {
  await ElMessageBox.confirm(t('organization.confirmDeleteOrg'), t('common.tip'), { type: 'warning' })
  const res = await request.post('/organization/delete', { id: row.id })
  const data = res.data || res
  if (data.code === 0) {
    ElMessage.success(t('common.deleteSuccess'))
    loadOrgList()
  }
}

async function handleOrgStatusChange(row) {
  const res = await request.post('/organization/updateStatus', {
    id: row.id,
    status: row.status
  })
  const data = res.data || res
  if (data.code === 0) {
    ElMessage.success(t('common.statusUpdateSuccess'))
  } else {
    row.status = row.status === 'enable' ? 'disable' : 'enable'
    ElMessage.error(data.msg || t('common.statusUpdateFailed'))
  }
}

// 通知配置管理
async function loadNotifyProviders() {
  try {
    const res = await getNotifyProviders()
    if (res.code === 0) {
      notifyProviders.value = res.list || []
    }
  } catch (error) {
    console.error('加载通知提供者失败', error)
  }
}

async function loadNotifyConfigList() {
  notifyLoading.value = true
  try {
    const res = await getNotifyConfigList()
    if (res.code === 0) {
      notifyConfigList.value = (res.list || []).map(item => ({ ...item, testing: false }))
    }
  } finally {
    notifyLoading.value = false
  }
}

function getProviderName(providerId) {
  const provider = notifyProviders.value.find(p => p.id === providerId)
  return provider ? provider.name : providerId
}

function handleProviderChange(providerId) {
  const provider = notifyProviders.value.find(p => p.id === providerId)
  currentProviderFields.value = provider ? provider.configFields || [] : []
  // 重置配置数据
  notifyForm.value.configData = {}
}

function showNotifyDialog(row = null) {
  if (row) {
    // 编辑模式
    let configData = {}
    try {
      configData = JSON.parse(row.config || '{}')
    } catch (e) {
      configData = {}
    }
    notifyForm.value = {
      id: row.id,
      name: row.name,
      provider: row.provider,
      configData: configData,
      messageTemplate: row.messageTemplate || '',
      status: row.status
    }
    // 加载对应provider的字段
    const provider = notifyProviders.value.find(p => p.id === row.provider)
    currentProviderFields.value = provider ? provider.configFields || [] : []
  } else {
    // 新增模式
    notifyForm.value = { 
      id: '', 
      name: '', 
      provider: '', 
      configData: {}, 
      messageTemplate: '', 
      status: 'enable'
    }
    currentProviderFields.value = []
  }
  notifyDialogVisible.value = true
}

async function handleNotifySubmit() {
  if (!notifyFormRef.value) return
  try {
    await notifyFormRef.value.validate()
    notifySubmitting.value = true
    
    const data = {
      id: notifyForm.value.id,
      name: notifyForm.value.name,
      provider: notifyForm.value.provider,
      config: JSON.stringify(notifyForm.value.configData),
      messageTemplate: notifyForm.value.messageTemplate,
      status: notifyForm.value.status
    }
    
    const res = await saveNotifyConfig(data)
    if (res.code === 0) {
      ElMessage.success(res.msg || t('common.operationSuccess'))
      notifyDialogVisible.value = false
      loadNotifyConfigList()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    notifySubmitting.value = false
  }
}

async function handleNotifyStatusChange(row) {
  const data = {
    id: row.id,
    name: row.name,
    provider: row.provider,
    config: row.config,
    messageTemplate: row.messageTemplate,
    status: row.status
  }
  const res = await saveNotifyConfig(data)
  if (res.code === 0) {
    ElMessage.success(t('common.statusUpdateSuccess'))
  } else {
    row.status = row.status === 'enable' ? 'disable' : 'enable'
    ElMessage.error(res.msg || t('common.statusUpdateFailed'))
  }
}

async function handleTestNotify(row) {
  row.testing = true
  try {
    const res = await testNotifyConfig({
      provider: row.provider,
      config: row.config,
      messageTemplate: row.messageTemplate
    })
    if (res.code === 0) {
      ElMessage.success(res.msg || t('settings.testSuccess'))
    } else {
      ElMessage.error(res.msg || t('settings.testFailed'))
    }
  } finally {
    row.testing = false
  }
}

async function handleTestNotifyForm() {
  if (!notifyForm.value.provider) {
    ElMessage.warning(t('settings.selectChannelFirst'))
    return
  }
  notifyTesting.value = true
  try {
    const res = await testNotifyConfig({
      provider: notifyForm.value.provider,
      config: JSON.stringify(notifyForm.value.configData),
      messageTemplate: notifyForm.value.messageTemplate
    })
    if (res.code === 0) {
      ElMessage.success(res.msg || t('settings.testSuccess'))
    } else {
      ElMessage.error(res.msg || t('settings.testFailed'))
    }
  } finally {
    notifyTesting.value = false
  }
}

async function handleDeleteNotify(row) {
  try {
    await ElMessageBox.confirm(t('settings.confirmDeleteNotify'), t('common.tip'), { type: 'warning' })
    const res = await deleteNotifyConfig(row.id)
    if (res.code === 0) {
      ElMessage.success(res.msg || t('common.deleteSuccess'))
      loadNotifyConfigList()
    } else {
      ElMessage.error(res.msg || t('common.operationFailed'))
    }
  } catch (error) {}
}
</script>

<style scoped>
.settings-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 16px;
    font-weight: 500;
  }

  .option-desc {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    margin-left: 8px;
  }
}
</style>
