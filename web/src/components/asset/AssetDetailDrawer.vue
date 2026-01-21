<template>
  <el-drawer
    :model-value="visible"
    :title="drawerTitle"
    size="60%"
    direction="rtl"
    @update:model-value="handleClose"
  >
    <div v-if="asset" class="asset-detail">
      <!-- 顶部截图和基本信息 -->
      <div class="detail-header">
        <div 
          class="detail-screenshot"
          @mouseenter="handlePreviewShow"
          @mouseleave="handlePreviewHide"
        >
          <img 
            v-if="asset.screenshot"
            :src="formatScreenshotUrl(asset.screenshot)"
            :alt="asset.title"
            class="detail-screenshot-img"
          />
          <div v-else class="detail-screenshot-placeholder">
            {{ t('asset.noScreenshot') }}
          </div>
        </div>
        <div class="detail-basic-info">
          <div class="info-row">
            <span class="info-label">URL:</span>
            <a :href="assetUrl" target="_blank" class="info-value link">
              {{ assetUrl }}
            </a>
          </div>
          <div class="info-row">
            <span class="info-label">{{ t('asset.ip') }}:</span>
            <span class="info-value">{{ asset.ip || '-' }}</span>
          </div>
          <div v-if="asset.status && asset.status !== '0'" class="info-row">
            <span class="info-label">{{ t('asset.statusCode') }}:</span>
            <el-tag :type="getStatusType(asset.status)" size="small">
              {{ asset.status }}
            </el-tag>
          </div>
          <div v-if="asset.asn" class="info-row">
            <span class="info-label">ASN:</span>
            <span class="info-value">{{ asset.asn }}</span>
          </div>
          <div v-if="asset.title" class="info-row">
            <span class="info-label">{{ t('asset.title') }}:</span>
            <span class="info-value">{{ asset.title }}</span>
          </div>
        </div>
      </div>
      
      <!-- 标签页 -->
      <el-tabs v-model="activeTab" class="detail-tabs">
        <!-- Overview 标签页 -->
        <el-tab-pane :label="t('asset.assetDetail.overview')" name="overview">
          <div class="tab-content">
            <div class="section">
              <h4 class="section-title">{{ t('asset.assetDetail.networkInfo') }}</h4>
              <div class="info-grid">
                <div class="info-item">
                  <span class="item-label">{{ t('asset.assetDetail.host') }}:</span>
                  <span class="item-value">{{ asset.host || asset.name }}</span>
                </div>
                <div v-if="asset.port && asset.port !== 0" class="info-item">
                  <span class="item-label">{{ t('asset.assetDetail.port') }}:</span>
                  <span class="item-value">{{ asset.port }}</span>
                </div>
                <div class="info-item">
                  <span class="item-label">{{ t('asset.assetDetail.service') }}:</span>
                  <span class="item-value">{{ asset.service || '-' }}</span>
                </div>
                <div v-if="asset.cname" class="info-item">
                  <span class="item-label">{{ t('asset.assetDetail.cname') }}:</span>
                  <span class="item-value">{{ asset.cname }}</span>
                </div>
                <div v-if="asset.iconHash" class="info-item">
                  <span class="item-label">{{ t('asset.assetDetail.iconHash') }}:</span>
                  <div class="icon-hash-display">
                    <img 
                      v-if="asset.iconHashBytes"
                      :src="'data:image/x-icon;base64,' + asset.iconHashBytes"
                      class="favicon-large"
                      @error="(e) => e.target.style.display = 'none'"
                    />
                    <span class="item-value">{{ asset.iconHash }}</span>
                  </div>
                </div>
              </div>
            </div>
            
            <div class="section">
              <h4 class="section-title">{{ t('asset.assetDetail.httpResponse') }}</h4>
              <div class="code-block">
                <pre>{{ asset.httpHeader || t('asset.assetDetail.noHttpData') }}</pre>
              </div>
            </div>
            
            <div v-if="asset.httpBody" class="section">
              <h4 class="section-title">{{ t('asset.assetDetail.httpBody') }}</h4>
              <div class="code-block">
                <pre>{{ asset.httpBody.substring(0, 1000) }}{{ asset.httpBody.length > 1000 ? '...' : '' }}</pre>
              </div>
            </div>
          </div>
        </el-tab-pane>
        
        <!-- Exposures 标签页 -->
        <el-tab-pane name="exposures">
          <template #label>
            <span>{{ t('asset.assetDetail.exposures') }} <el-badge :value="exposuresCount" class="tab-badge" /></span>
          </template>
          <div class="tab-content">
            <!-- 端口服务 -->
            <div class="section">
              <h4 class="section-title">{{ t('asset.assetDetail.portServices') }}</h4>
              <div class="exposure-item">
                <div class="exposure-header">
                  <el-tag size="small">{{ asset.port }}</el-tag>
                  <span class="exposure-service">{{ asset.service || t('asset.assetDetail.unknown') }}</span>
                </div>
                <div class="exposure-details">
                  <div class="detail-item">
                    <span class="detail-label">{{ t('asset.assetDetail.protocol') }}:</span>
                    <span class="detail-value">{{ asset.port === 443 ? 'HTTPS' : 'HTTP' }}</span>
                  </div>
                  <div v-if="asset.banner" class="detail-item">
                    <span class="detail-label">{{ t('asset.assetDetail.banner') }}:</span>
                    <div class="code-block small">
                      <pre>{{ asset.banner }}</pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 目录扫描结果 -->
            <div class="section">
              <div class="section-header">
                <h4 class="section-title">
                  {{ t('asset.assetDetail.dirScanResults') }}
                  <el-badge :value="filteredDirScanResults.length" class="count-badge" />
                </h4>
                <el-input
                  v-if="dirScanResults.length > 0"
                  v-model="dirScanSearch"
                  :placeholder="t('asset.assetDetail.searchDirScan')"
                  clearable
                  size="small"
                  class="section-search"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </div>
              <div v-if="filteredDirScanResults.length > 0" class="dir-scan-list">
                <div v-for="(dir, index) in filteredDirScanResults" :key="index" class="dir-scan-item">
                  <div class="dir-scan-header">
                    <a :href="dir.url" target="_blank" class="dir-url">{{ dir.path }}</a>
                    <el-tag :type="getStatusType(dir.status)" size="small">{{ dir.status }}</el-tag>
                  </div>
                  <div class="dir-scan-meta">
                    <span class="meta-item">
                      <el-icon><Document /></el-icon>
                      {{ dir.contentLength || 0 }} bytes
                    </span>
                    <span class="meta-item">
                      <el-icon><Clock /></el-icon>
                      {{ dir.responseTime || 0 }}ms
                    </span>
                    <span v-if="dir.title" class="meta-item">
                      <el-icon><Document /></el-icon>
                      {{ dir.title }}
                    </span>
                  </div>
                </div>
              </div>
              <div v-else-if="dirScanResults.length > 0 && dirScanSearch" class="empty-state">
                {{ t('asset.assetDetail.noSearchResults') }}
              </div>
              <div v-else class="empty-state">
                {{ t('asset.assetDetail.noDirScanResults') }}
              </div>
            </div>
            
            <!-- 漏洞扫描结果 -->
            <div class="section">
              <div class="section-header">
                <h4 class="section-title">
                  {{ t('asset.assetDetail.vulnScanResults') }}
                  <el-badge :value="filteredVulnScanResults.length" class="count-badge" />
                </h4>
                <el-input
                  v-if="vulnScanResults.length > 0"
                  v-model="vulnScanSearch"
                  :placeholder="t('asset.assetDetail.searchVuln')"
                  clearable
                  size="small"
                  class="section-search"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </div>
              <div v-if="filteredVulnScanResults.length > 0" class="vuln-scan-list">
                <div v-for="(vuln, index) in filteredVulnScanResults" :key="index" class="vuln-scan-item">
                  <div class="vuln-scan-header">
                    <div class="vuln-title-row">
                      <el-tag :type="getVulnSeverityType(vuln.severity)" size="small" class="severity-tag">
                        {{ vuln.severity }}
                      </el-tag>
                      <span class="vuln-name">{{ vuln.name }}</span>
                    </div>
                    <span class="vuln-id">{{ vuln.id }}</span>
                  </div>
                  <div v-if="vuln.description" class="vuln-description">
                    {{ vuln.description }}
                  </div>
                  <div class="vuln-meta">
                    <span v-if="vuln.cvss" class="meta-item">
                      <el-icon><Warning /></el-icon>
                      CVSS: {{ vuln.cvss }}
                    </span>
                    <span v-if="vuln.cve" class="meta-item">
                      <el-icon><Document /></el-icon>
                      {{ vuln.cve }}
                    </span>
                    <span class="meta-item">
                      <el-icon><Clock /></el-icon>
                      {{ vuln.discoveredAt }}
                    </span>
                  </div>
                  <div v-if="vuln.matchedUrl" class="vuln-matched-url">
                    <span class="matched-label">{{ t('asset.assetDetail.matchedUrl') }}:</span>
                    <a :href="vuln.matchedUrl" target="_blank" class="matched-url">{{ vuln.matchedUrl }}</a>
                  </div>
                </div>
              </div>
              <div v-else-if="vulnScanResults.length > 0 && vulnScanSearch" class="empty-state">
                {{ t('asset.assetDetail.noSearchResults') }}
              </div>
              <div v-else class="empty-state">
                {{ t('asset.assetDetail.noVulnScanResults') }}
              </div>
            </div>
          </div>
        </el-tab-pane>
        
        <!-- Technologies 标签页 -->
        <el-tab-pane name="technologies">
          <template #label>
            <span>{{ t('asset.assetDetail.technologies') }} <el-badge :value="technologies.length" class="tab-badge" /></span>
          </template>
          <div class="tab-content">
            <!-- 技术栈搜索框 -->
            <div v-if="technologies.length > 0" class="section-header standalone">
              <el-input
                v-model="techSearch"
                :placeholder="t('asset.assetDetail.searchTech')"
                clearable
                size="small"
                class="section-search full-width"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
            </div>
            <div v-if="filteredTechnologies.length > 0" class="tech-list-detail">
              <div v-for="(tech, index) in filteredTechnologies" :key="index" class="tech-item-detail">
                <div class="tech-icon">
                  <el-icon><Box /></el-icon>
                </div>
                <div class="tech-info">
                  <div class="tech-name">{{ tech }}</div>
                  <div class="tech-category">{{ t('asset.assetDetail.techCategory') }}</div>
                </div>
              </div>
            </div>
            <div v-else-if="technologies.length > 0 && techSearch" class="empty-state">
              {{ t('asset.assetDetail.noSearchResults') }}
            </div>
            <div v-else class="empty-state">
              {{ t('asset.assetDetail.noTechDetected') }}
            </div>
          </div>
        </el-tab-pane>
        
        <!-- Changelogs 标签页 -->
        <el-tab-pane name="changelogs">
          <template #label>
            <span>{{ t('asset.assetDetail.changelogs') }} <el-badge :value="changelogs.length" class="tab-badge" /></span>
          </template>
          <div class="tab-content">
            <div v-if="changelogs.length > 0" class="changelog-list">
              <div v-for="(log, index) in changelogs" :key="index" class="changelog-item">
                <div class="changelog-header">
                  <div class="changelog-time-info">
                    <el-icon class="time-icon"><Clock /></el-icon>
                    <span class="changelog-time">{{ log.time }}</span>
                  </div>
                  <el-tag size="small" type="info">{{ log.taskId }}</el-tag>
                </div>
                <div class="changelog-changes">
                  <div v-for="(change, idx) in log.changes" :key="idx" class="change-item">
                    <div class="change-field-name">
                      <el-icon class="field-icon"><Edit /></el-icon>
                      <span class="field-label">{{ translateFieldName(change.field) }}</span>
                    </div>
                    <div class="change-values">
                      <div class="change-value-box old-value">
                        <div class="value-label">{{ t('asset.assetDetail.oldValue') }}</div>
                        <div class="value-content">{{ change.oldValue || '-' }}</div>
                      </div>
                      <el-icon class="change-arrow"><Right /></el-icon>
                      <div class="change-value-box new-value">
                        <div class="value-label">{{ t('asset.assetDetail.newValue') }}</div>
                        <div class="value-content">{{ change.newValue || '-' }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty-state">
              {{ t('asset.assetDetail.noChangeHistory') }}
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Box,
  Document,
  Clock,
  Warning,
  Edit,
  Right,
  Search
} from '@element-plus/icons-vue'
import { formatScreenshotUrl } from '@/utils/screenshot'

const { t } = useI18n()

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  asset: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'preview-show', 'preview-hide'])

const activeTab = ref('overview')

// 搜索关键词
const dirScanSearch = ref('')
const vulnScanSearch = ref('')
const techSearch = ref('')

// Computed properties
const drawerTitle = computed(() => {
  if (!props.asset) return ''
  const host = props.asset.host || props.asset.name
  const port = props.asset.port && props.asset.port !== 0 ? `:${props.asset.port}` : ''
  return `${host}${port}`
})

const assetUrl = computed(() => {
  if (!props.asset) return ''
  if (props.asset.url) return props.asset.url
  const host = props.asset.host || props.asset.name
  const port = props.asset.port
  if (port && port !== 0) {
    return `${port === 443 ? 'https' : 'http'}://${host}:${port}`
  }
  return `http://${host}`
})

const technologies = computed(() => {
  if (!props.asset || !props.asset.technologies) return []
  // Handle both string array and object array formats
  return props.asset.technologies.map(tech => 
    typeof tech === 'string' ? tech : tech.name
  )
})

const dirScanResults = computed(() => {
  return props.asset?.dirScanResults || []
})

const vulnScanResults = computed(() => {
  return props.asset?.vulnScanResults || []
})

const changelogs = computed(() => {
  return props.asset?.changelogs || []
})

// 过滤后的目录扫描结果
const filteredDirScanResults = computed(() => {
  if (!dirScanSearch.value) return dirScanResults.value
  const query = dirScanSearch.value.toLowerCase()
  return dirScanResults.value.filter(dir => 
    (dir.path && dir.path.toLowerCase().includes(query)) ||
    (dir.url && dir.url.toLowerCase().includes(query)) ||
    (dir.title && dir.title.toLowerCase().includes(query)) ||
    (dir.status && String(dir.status).includes(query))
  )
})

// 过滤后的漏洞扫描结果
const filteredVulnScanResults = computed(() => {
  if (!vulnScanSearch.value) return vulnScanResults.value
  const query = vulnScanSearch.value.toLowerCase()
  return vulnScanResults.value.filter(vuln => 
    (vuln.name && vuln.name.toLowerCase().includes(query)) ||
    (vuln.severity && vuln.severity.toLowerCase().includes(query)) ||
    (vuln.cve && vuln.cve.toLowerCase().includes(query)) ||
    (vuln.description && vuln.description.toLowerCase().includes(query)) ||
    (vuln.matchedUrl && vuln.matchedUrl.toLowerCase().includes(query))
  )
})

// 过滤后的技术栈
const filteredTechnologies = computed(() => {
  if (!techSearch.value) return technologies.value
  const query = techSearch.value.toLowerCase()
  return technologies.value.filter(tech => 
    tech.toLowerCase().includes(query)
  )
})

const exposuresCount = computed(() => {
  return dirScanResults.value.length + vulnScanResults.value.length
})

// Methods
const handleClose = (value) => {
  emit('update:visible', value)
}

const handlePreviewShow = (event) => {
  emit('preview-show', props.asset, event)
}

const handlePreviewHide = () => {
  emit('preview-hide')
}

const getStatusType = (status) => {
  const code = parseInt(status)
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'warning'
  if (code >= 400 && code < 500) return 'danger'
  if (code >= 500) return 'danger'
  return 'info'
}

const getVulnSeverityType = (severity) => {
  const level = String(severity || '').toLowerCase()
  if (level === 'critical' || level === 'high') return 'danger'
  if (level === 'medium') return 'warning'
  if (level === 'low') return 'info'
  return 'info'
}

const translateFieldName = (field) => {
  const fieldMap = {
    'title': t('asset.assetDetail.title'),
    'status': t('asset.assetDetail.statusCode'),
    'technologies': t('asset.assetDetail.technologies'),
    'httpHeader': t('asset.assetDetail.httpHeader'),
    'httpBody': t('asset.assetDetail.httpBody'),
    'screenshot': t('asset.assetDetail.screenshot'),
    'ip': t('asset.ip'),
    'port': t('asset.assetDetail.port'),
    'service': t('asset.assetDetail.service')
  }
  return fieldMap[field] || field
}

// Watch for visibility changes to reset tab and search
watch(() => props.visible, (newVal) => {
  if (newVal) {
    activeTab.value = 'overview'
    // 重置搜索关键词
    dirScanSearch.value = ''
    vulnScanSearch.value = ''
    techSearch.value = ''
  }
})
</script>

<style scoped lang="scss">
.asset-detail {
  padding: 0;
}

.detail-header {
  display: flex;
  gap: 20px;
  margin-bottom: 24px;
  padding: 16px;
  background: hsl(var(--muted) / 0.5);
  border-radius: 8px;
  border: 1px solid hsl(var(--border));
}

.detail-screenshot {
  flex-shrink: 0;
  width: 300px;
  height: 200px;
  border-radius: 8px;
  overflow: hidden;
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.detail-screenshot-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.detail-screenshot-placeholder {
  color: hsl(var(--muted-foreground));
  font-size: 14px;
}

.detail-basic-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-label {
  font-weight: 500;
  color: hsl(var(--muted-foreground));
  min-width: 80px;
}

.info-value {
  color: hsl(var(--foreground));
  
  &.link {
    color: hsl(var(--primary));
    text-decoration: none;
    
    &:hover {
      text-decoration: underline;
    }
  }
}

.detail-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 20px;
  }
}

.tab-content {
  padding: 0 16px;
}

.section {
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  
  &.standalone {
    margin-bottom: 16px;
  }
  
  .section-title {
    margin-bottom: 0;
  }
}

.section-search {
  width: 200px;
  flex-shrink: 0;
  
  &.full-width {
    width: 100%;
  }
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: hsl(var(--foreground));
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.count-badge {
  :deep(.el-badge__content) {
    background-color: hsl(var(--muted-foreground));
  }
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-label {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.item-value {
  font-size: 14px;
  color: hsl(var(--foreground));
}

.icon-hash-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.favicon-large {
  width: 24px;
  height: 24px;
}

.code-block {
  background: hsl(var(--muted) / 0.5);
  border: 1px solid hsl(var(--border));
  border-radius: 4px;
  padding: 12px;
  overflow-x: auto;
  
  &.small {
    padding: 8px;
  }
  
  pre {
    margin: 0;
    font-family: 'Courier New', monospace;
    font-size: 12px;
    line-height: 1.5;
    color: hsl(var(--foreground));
    white-space: pre-wrap;
    word-break: break-all;
  }
}

.exposure-item {
  background: hsl(var(--muted) / 0.5);
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.exposure-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.exposure-service {
  font-weight: 500;
  color: hsl(var(--foreground));
}

.exposure-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item {
  display: flex;
  gap: 8px;
}

.detail-label {
  font-weight: 500;
  color: hsl(var(--muted-foreground));
  min-width: 80px;
}

.detail-value {
  color: hsl(var(--foreground));
}

.dir-scan-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dir-scan-item {
  background: hsl(var(--muted) / 0.5);
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 12px;
}

.dir-scan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.dir-url {
  color: hsl(var(--primary));
  text-decoration: none;
  font-weight: 500;
  
  &:hover {
    text-decoration: underline;
  }
}

.dir-scan-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.vuln-scan-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.vuln-scan-item {
  background: hsl(var(--muted) / 0.5);
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 16px;
}

.vuln-scan-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.vuln-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.severity-tag {
  flex-shrink: 0;
}

.vuln-name {
  font-weight: 600;
  color: hsl(var(--foreground));
  font-size: 14px;
}

.vuln-id {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  flex-shrink: 0;
}

.vuln-description {
  margin-bottom: 12px;
  color: hsl(var(--muted-foreground));
  font-size: 13px;
  line-height: 1.6;
}

.vuln-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  margin-bottom: 8px;
}

.vuln-matched-url {
  display: flex;
  gap: 8px;
  font-size: 12px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid hsl(var(--border));
}

.matched-label {
  color: hsl(var(--muted-foreground));
  flex-shrink: 0;
}

.matched-url {
  color: hsl(var(--primary));
  text-decoration: none;
  word-break: break-all;
  
  &:hover {
    text-decoration: underline;
  }
}

.tech-list-detail {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 12px;
}

.tech-item-detail {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: hsl(var(--muted) / 0.5);
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
}

.tech-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: hsl(var(--primary));
  color: hsl(var(--primary-foreground));
  border-radius: 8px;
  flex-shrink: 0;
}

.tech-info {
  flex: 1;
  min-width: 0;
}

.tech-name {
  font-weight: 500;
  color: hsl(var(--foreground));
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tech-category {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  margin-top: 2px;
}

.changelog-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.changelog-item {
  background: hsl(var(--muted) / 0.5);
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 16px;
}

.changelog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.changelog-time-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-icon {
  color: hsl(var(--muted-foreground));
}

.changelog-time {
  font-weight: 500;
  color: hsl(var(--foreground));
}

.changelog-changes {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.change-item {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  padding: 12px;
}

.change-field-name {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.field-icon {
  color: hsl(var(--primary));
}

.field-label {
  font-weight: 600;
  color: hsl(var(--foreground));
}

.change-values {
  display: flex;
  align-items: center;
  gap: 12px;
}

.change-value-box {
  flex: 1;
  padding: 12px;
  border-radius: 6px;
  
  &.old-value {
    background: hsl(var(--destructive) / 0.1);
    border: 1px solid hsl(var(--destructive) / 0.3);
  }
  
  &.new-value {
    background: hsl(var(--primary) / 0.1);
    border: 1px solid hsl(var(--primary) / 0.3);
  }
}

.value-label {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  margin-bottom: 4px;
}

.value-content {
  font-size: 13px;
  color: hsl(var(--foreground));
  word-break: break-all;
}

.change-arrow {
  color: hsl(var(--muted-foreground));
  flex-shrink: 0;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: hsl(var(--muted-foreground));
  font-size: 14px;
}

.tab-badge {
  :deep(.el-badge__content) {
    background-color: hsl(var(--muted-foreground));
  }
}

// Element Plus Drawer 深色主题覆盖
:deep(.el-drawer) {
  background: hsl(var(--background));
  
  .el-drawer__header {
    color: hsl(var(--foreground));
    border-bottom: 1px solid hsl(var(--border));
    margin-bottom: 0;
    padding-bottom: 16px;
  }
  
  .el-drawer__title {
    color: hsl(var(--foreground));
  }
  
  .el-drawer__body {
    background: hsl(var(--background));
  }
}

// Element Plus Tabs 深色主题覆盖
:deep(.el-tabs) {
  .el-tabs__nav-wrap::after {
    background-color: hsl(var(--border));
  }
  
  .el-tabs__item {
    color: hsl(var(--muted-foreground));
    
    &.is-active {
      color: hsl(var(--primary));
    }
    
    &:hover {
      color: hsl(var(--foreground));
    }
  }
  
  .el-tabs__active-bar {
    background-color: hsl(var(--primary));
  }
}
</style>
