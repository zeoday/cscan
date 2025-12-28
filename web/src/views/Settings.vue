<template>
  <div class="settings-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>系统配置</span>
        </div>
      </template>
      
      <el-tabs v-model="activeTab" tab-position="left" class="settings-tabs">
        <!-- 在线API配置 -->
        <el-tab-pane label="在线API配置" name="onlineapi">
          <div class="tab-content">
            <el-alert type="info" :closable="false" style="margin-bottom: 20px">
              <template #title>配置在线搜索API密钥，用于Fofa、Hunter、Quake等平台的资产搜索</template>
            </el-alert>
            
            <el-tabs v-model="apiConfigTab" type="card">
              <el-tab-pane label="Fofa" name="fofa">
                <el-form label-width="100px" style="max-width: 500px; margin-top: 20px">
                  <el-form-item label="Email">
                    <el-input v-model="apiConfigs.fofa.key" placeholder="Fofa账号邮箱" />
                  </el-form-item>
                  <el-form-item label="API Key">
                    <el-input v-model="apiConfigs.fofa.secret" placeholder="Fofa API Key" show-password />
                  </el-form-item>
                  <el-form-item>
                    <el-button type="primary" @click="saveApiConfig('fofa')">保存</el-button>
                    <el-button type="success" @click="openApiUrl('https://fofa.info/userInfo')">申请API</el-button>
                  </el-form-item>
                </el-form>
              </el-tab-pane>
              <el-tab-pane label="Hunter" name="hunter">
                <el-form label-width="100px" style="max-width: 500px; margin-top: 20px">
                  <el-form-item label="API Key">
                    <el-input v-model="apiConfigs.hunter.key" placeholder="Hunter API Key" show-password />
                  </el-form-item>
                  <el-form-item>
                    <el-button type="primary" @click="saveApiConfig('hunter')">保存</el-button>
                    <el-button type="success" @click="openApiUrl('https://hunter.qianxin.com/home/myInfo')">申请API</el-button>
                  </el-form-item>
                </el-form>
              </el-tab-pane>
              <el-tab-pane label="Quake" name="quake">
                <el-form label-width="100px" style="max-width: 500px; margin-top: 20px">
                  <el-form-item label="API Key">
                    <el-input v-model="apiConfigs.quake.key" placeholder="Quake API Key" show-password />
                  </el-form-item>
                  <el-form-item>
                    <el-button type="primary" @click="saveApiConfig('quake')">保存</el-button>
                    <el-button type="success" @click="openApiUrl('https://quake.360.net/quake/#/personal?tab=message')">申请API</el-button>
                  </el-form-item>
                </el-form>
              </el-tab-pane>
            </el-tabs>
          </div>
        </el-tab-pane>

        <!-- Subfinder数据源配置 -->
        <el-tab-pane label="子域名扫描配置" name="subfinder">
          <div class="tab-content">
            <el-alert type="info" :closable="false" style="margin-bottom: 20px">
              <template #title>Subfinder用于子域名枚举，配置API密钥可以获取更多数据源的结果。点击数据源名称可跳转到官网获取API密钥。</template>
            </el-alert>
            
            <el-table :data="subfinderProviders" v-loading="subfinderLoading" max-height="500" stripe>
              <el-table-column prop="name" label="数据源" width="130">
                <template #default="{ row }">
                  <span>{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="描述" width="180" show-overflow-tooltip />
              <el-table-column prop="keyFormat" label="密钥格式" width="140" />
              <el-table-column label="API密钥" min-width="200">
                <template #default="{ row }">
                  <el-input
                    v-model="row.keyInput"
                    :placeholder="row.maskedKey || row.keyFormat"
                    size="small"
                    clearable
                  />
                </template>
              </el-table-column>
              <el-table-column label="状态" width="70">
                <template #default="{ row }">
                  <el-switch v-model="row.enabled" size="small" />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="140">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="saveSubfinderProvider(row)">保存</el-button>
                  <el-button type="success" link size="small" @click="openApiUrl(row.url)">申请API</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- 工作空间 -->
        <el-tab-pane label="工作空间" name="workspace">
          <div class="tab-content">
            <div class="tab-action-bar">
              <el-button type="primary" @click="showWorkspaceDialog()">
                <el-icon><Plus /></el-icon>新建工作空间
              </el-button>
            </div>
            <el-table :data="workspaceList" v-loading="workspaceLoading" stripe max-height="500">
              <el-table-column prop="name" label="名称" min-width="150" />
              <el-table-column prop="description" label="描述" min-width="250" />
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'enable' ? 'success' : 'danger'">
                    {{ row.status === 'enable' ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="createTime" label="创建时间" width="160" />
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="showWorkspaceDialog(row)">编辑</el-button>
                  <el-button type="danger" link size="small" @click="handleDeleteWorkspace(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- 组织管理 -->
        <el-tab-pane label="组织管理" name="organization">
          <div class="tab-content">
            <div class="tab-action-bar">
              <el-button type="primary" @click="showOrgDialog()">
                <el-icon><Plus /></el-icon>新建组织
              </el-button>
            </div>
            <el-table :data="orgList" v-loading="orgLoading" stripe max-height="500">
              <el-table-column prop="name" label="组织名称" min-width="150" />
              <el-table-column prop="description" label="描述" min-width="250" />
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-switch
                    v-model="row.status"
                    active-value="enable"
                    inactive-value="disable"
                    @change="handleOrgStatusChange(row)"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="createTime" label="创建时间" width="160" />
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="showOrgDialog(row)">编辑</el-button>
                  <el-button type="danger" link size="small" @click="handleDeleteOrg(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>


        <!-- 用户管理 -->
        <el-tab-pane label="用户管理" name="user">
          <div class="tab-content">
            <div class="tab-action-bar">
              <el-button type="primary" @click="showUserDialog()">
                <el-icon><Plus /></el-icon>新建用户
              </el-button>
            </div>
            <el-table :data="userList" v-loading="userLoading" stripe max-height="500">
              <el-table-column prop="username" label="用户名" min-width="150" />
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'enable' ? 'success' : 'danger'">
                    {{ row.status === 'enable' ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="showUserDialog(row)">编辑</el-button>
                  <el-button type="warning" link size="small" @click="showResetPasswordDialog(row)">重置密码</el-button>
                  <el-button type="danger" link size="small" @click="handleDeleteUser(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 工作空间对话框 -->
    <el-dialog v-model="workspaceDialogVisible" :title="workspaceForm.id ? '编辑工作空间' : '新建工作空间'" width="500px">
      <el-form ref="workspaceFormRef" :model="workspaceForm" :rules="workspaceRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="workspaceForm.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="workspaceForm.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="workspaceDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="workspaceSubmitting" @click="handleWorkspaceSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 用户对话框 -->
    <el-dialog v-model="userDialogVisible" :title="userForm.id ? '编辑用户' : '新建用户'" width="500px">
      <el-form ref="userFormRef" :model="userForm" :rules="userRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item v-if="!userForm.id" label="密码" prop="password">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="userForm.status" placeholder="请选择状态">
            <el-option label="启用" value="enable" />
            <el-option label="禁用" value="disable" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="userSubmitting" @click="handleUserSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog v-model="resetPasswordVisible" title="重置密码" width="400px">
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="80px">
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="resetForm.newPassword" type="password" placeholder="请输入新密码" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="resetForm.confirmPassword" type="password" placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPasswordVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetting" @click="handleResetPassword">确定</el-button>
      </template>
    </el-dialog>

    <!-- 组织对话框 -->
    <el-dialog v-model="orgDialogVisible" :title="orgForm.id ? '编辑组织' : '新建组织'" width="500px">
      <el-form ref="orgFormRef" :model="orgForm" :rules="orgRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="orgForm.name" placeholder="请输入组织名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="orgForm.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orgDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="orgSubmitting" @click="handleOrgSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>


<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/api/request'
import { getSubfinderProviderList, getSubfinderProviderInfo, saveSubfinderProvider as saveSubfinderProviderApi } from '@/api/subfinder'
import { getUserList, createUser, updateUser, deleteUser, resetUserPassword } from '@/api/auth'

const route = useRoute()
const activeTab = ref('onlineapi')
const apiConfigTab = ref('fofa')
const subfinderLoading = ref(false)
const subfinderProviders = ref([])

const apiConfigs = reactive({
  fofa: { key: '', secret: '' },
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
const workspaceRules = { name: [{ required: true, message: '请输入名称', trigger: 'blur' }] }

// 用户管理相关
const userLoading = ref(false)
const userList = ref([])
const userDialogVisible = ref(false)
const userSubmitting = ref(false)
const userFormRef = ref()
const userForm = ref({ id: '', username: '', password: '', status: 'enable' })
const userRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

// 重置密码相关
const resetPasswordVisible = ref(false)
const resetting = ref(false)
const resetFormRef = ref()
const resetForm = ref({ id: '', newPassword: '', confirmPassword: '' })
const resetRules = {
  newPassword: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== resetForm.value.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 组织管理相关
const orgLoading = ref(false)
const orgList = ref([])
const orgDialogVisible = ref(false)
const orgSubmitting = ref(false)
const orgFormRef = ref()
const orgForm = reactive({ id: '', name: '', description: '' })
const orgRules = { name: [{ required: true, message: '请输入组织名称', trigger: 'blur' }] }

onMounted(() => {
  // 处理URL中的tab参数
  if (route.query.tab) {
    activeTab.value = route.query.tab
  }
  loadApiConfigs()
  loadSubfinderProviders()
})

// 监听tab切换，按需加载数据
watch(activeTab, (val) => {
  if (val === 'workspace' && workspaceList.value.length === 0) {
    loadWorkspaceList()
  } else if (val === 'user' && userList.value.length === 0) {
    loadUserList()
  } else if (val === 'organization' && orgList.value.length === 0) {
    loadOrgList()
  }
})

// 在线API配置
async function loadApiConfigs() {
  const res = await request.post('/onlineapi/config/list', {})
  if (res.code === 0 && res.list) {
    res.list.forEach(item => {
      if (apiConfigs[item.platform]) {
        apiConfigs[item.platform].key = item.key
        apiConfigs[item.platform].secret = item.secret
      }
    })
  }
}

async function saveApiConfig(platform) {
  const config = apiConfigs[platform]
  const res = await request.post('/onlineapi/config/save', {
    platform,
    key: config.key,
    secret: config.secret
  })
  if (res.code === 0) {
    ElMessage.success('保存成功')
  } else {
    ElMessage.error(res.msg || '保存失败')
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
    ElMessage.warning('请输入API密钥')
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
    ElMessage.success('保存成功')
    row.configured = true
    row.keyInput = ''
    await loadSubfinderProviders()
  } else {
    ElMessage.error(res.msg || '保存失败')
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
      ElMessage.success(workspaceForm.id ? '更新成功' : '创建成功')
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
  await ElMessageBox.confirm('确定删除该工作空间吗？', '提示', { type: 'warning' })
  const res = await request.post('/workspace/delete', { id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
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
      ElMessage.success(res.msg || '操作成功')
      userDialogVisible.value = false
      loadUserList()
    } else {
      ElMessage.error(res.msg || '操作失败')
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    userSubmitting.value = false
  }
}

async function handleDeleteUser(row) {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗？', '提示', { type: 'warning' })
    const res = await deleteUser({ id: row.id })
    if (res.code === 0) {
      ElMessage.success(res.msg || '删除成功')
      loadUserList()
    } else {
      ElMessage.error(res.msg || '删除失败')
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
      ElMessage.success(res.msg || '密码重置成功')
      resetPasswordVisible.value = false
    } else {
      ElMessage.error(res.msg || '密码重置失败')
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
      ElMessage.success(orgForm.id ? '更新成功' : '创建成功')
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
  await ElMessageBox.confirm('确定删除该组织吗？', '提示', { type: 'warning' })
  const res = await request.post('/organization/delete', { id: row.id })
  const data = res.data || res
  if (data.code === 0) {
    ElMessage.success('删除成功')
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
    ElMessage.success('状态更新成功')
  } else {
    row.status = row.status === 'enable' ? 'disable' : 'enable'
    ElMessage.error(data.msg || '状态更新失败')
  }
}
</script>

<style lang="scss" scoped>
.settings-page {
  .card-header {
    font-size: 16px;
    font-weight: 500;
  }

  .settings-tabs {
    min-height: 500px;

    :deep(.el-tabs__item) {
      height: 50px;
      line-height: 50px;
    }
  }

  .tab-content {
    padding: 0 20px;
  }

  .tab-action-bar {
    margin-bottom: 16px;
  }
}
</style>
