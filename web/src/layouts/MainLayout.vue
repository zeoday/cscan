<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '250px'" class="aside">
      <div class="logo">
        <img src="/logo.png" alt="logo" />
        <span v-show="!isCollapse">CSCAN</span>
      </div>

      <div class="menu-wrapper">
        <el-menu :default-active="$route.path" :default-openeds="defaultOpeneds" :collapse="isCollapse" router
          :unique-opened="false">
          <!-- 主控台分组 -->
          <div v-show="!isCollapse" class="menu-group-title">{{ $t('navigation.groupDashboard') }}</div>
          <el-menu-item index="/dashboard">
            <el-icon>
              <Odometer />
            </el-icon>
            <template #title>{{ $t('navigation.dashboard') }}</template>
          </el-menu-item>
          <el-menu-item index="/asset-management">
            <el-icon>
              <Monitor />
            </el-icon>
            <template #title>{{ $t('navigation.assetManagement') }}</template>
          </el-menu-item>

          <!-- 扫描分组 -->
          <div v-show="!isCollapse" class="menu-group-title">{{ $t('navigation.groupScan') }}</div>
            <el-menu-item index="/task">
              <el-icon>
                <List />
              </el-icon>
              <template #title>{{ $t('navigation.taskManagement') }}</template>
            </el-menu-item>
            <el-menu-item index="/cron-task">
              <el-icon>
                <Timer />
              </el-icon>
              <template #title>{{ $t('navigation.cronTask') }}</template>
            </el-menu-item>
            <el-menu-item index="/poc">
              <el-icon>
                <Aim />
              </el-icon>
              <template #title>{{ $t('navigation.pocManagement') }}</template>
            </el-menu-item>
            <el-menu-item index="/fingerprint">
              <el-icon>
                <Stamp />
              </el-icon>
              <template #title>{{ $t('navigation.fingerprintManagement') }}</template>
            </el-menu-item>
                        <el-menu-item index="/blacklist">
              <el-icon>
                <CircleClose />
              </el-icon>
              <template #title>{{ $t('navigation.blacklist') }}</template>
            </el-menu-item>
            <el-menu-item index="/settings?tab=subfinder">
              <el-icon>
                <Search />
              </el-icon>
              <template #title>{{ $t('navigation.subdomainConfig') }}</template>
          </el-menu-item>

          <!-- 工具分组 -->
          <div v-show="!isCollapse" class="menu-group-title">{{ $t('navigation.groupTools') }}</div>
          <el-menu-item index="/online-search">
            <el-icon>
              <Search />
            </el-icon>
            <template #title>{{ $t('navigation.onlineSearch') }}</template>

          </el-menu-item>
            <el-menu-item index="/settings?tab=onlineapi">
              <el-icon>
                <Key />
              </el-icon>
              <template #title>{{ $t('navigation.onlineApiConfig') }}</template>
            </el-menu-item>
          <!-- 系统管理分组 -->
          <div v-show="!isCollapse" class="menu-group-title">{{ $t('navigation.groupSystem') }}</div>
            <el-menu-item index="/worker">
              <el-icon>
                <Connection />
              </el-icon>
              <template #title>{{ $t('navigation.workerNodes') }}</template>
            </el-menu-item>
            <el-menu-item index="/worker-logs">
              <el-icon>
                <Document />
              </el-icon>
              <template #title>{{ $t('navigation.workerLogs') }}</template>
            </el-menu-item>
            <el-menu-item index="/settings?tab=notify">
              <el-icon>
                <Bell />
              </el-icon>
              <template #title>{{ $t('navigation.notifyConfig') }}</template>
            </el-menu-item>
            <el-menu-item index="/high-risk-filter">
              <el-icon>
                <Warning />
              </el-icon>
              <template #title>{{ $t('navigation.highRiskFilter') }}</template>
            </el-menu-item>
            <el-menu-item index="/settings?tab=workspace">
              <el-icon>
                <Folder />
              </el-icon>
              <template #title>{{ $t('navigation.workspaceManagement') }}</template>
            </el-menu-item>
            <el-menu-item index="/settings?tab=organization">
              <el-icon>
                <OfficeBuilding />
              </el-icon>
              <template #title>{{ $t('navigation.organizationManagement') }}</template>
            </el-menu-item>
            <el-menu-item index="/settings?tab=user">
              <el-icon>
                <User />
              </el-icon>
              <template #title>{{ $t('navigation.userManagement') }}</template>
            </el-menu-item>

        </el-menu>
      </div>

    </el-aside>

    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <!-- 工作空间选择 -->
          <el-select v-model="workspaceStore.currentWorkspaceId" :placeholder="$t('common.allWorkspaces')"
            style="width: 160px; margin-right: 16px;" @change="handleWorkspaceChange">
            <el-option :label="$t('common.allWorkspaces')" value="all" />
            <el-option v-for="ws in workspaceStore.workspaces" :key="ws.id" :label="ws.name" :value="ws.id" />
          </el-select>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">{{ $t('common.home') }}</el-breadcrumb-item>
            <el-breadcrumb-item>{{ $route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <!-- 语言切换 -->
          <LanguageSwitcher />
          <!-- 主题切换 -->
          <ThemeSwitcher />
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" icon="User" />
              <span class="username">{{ userStore.username }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">{{ $t('auth.logout') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import { useWorkspaceStore } from '@/stores/workspace'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import { Setting, Monitor, List, Search, Aim, Odometer, Stamp, Connection, Fold, Expand, Key, Folder, OfficeBuilding, Bell, User, Document, CircleClose, Warning, Timer } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()
const workspaceStore = useWorkspaceStore()
const isCollapse = ref(false)
const defaultOpeneds = ref(['scan-group', 'system-group'])

onMounted(() => {
  workspaceStore.loadWorkspaces()
})

function handleWorkspaceChange(val) {
  workspaceStore.setCurrentWorkspace(val)
  // 触发页面刷新数据
  window.dispatchEvent(new CustomEvent('workspace-changed', { detail: val }))
}

function handleCommand(command) {
  if (command === 'logout') {
    userStore.logout()
    router.push('/login')
  }
}
</script>

<style lang="scss" scoped>
.layout-container {
  height: 100vh;
}

.aside {
  background: hsl(var(--sidebar));
  color: hsl(var(--sidebar-foreground));
  transition: width 0.3s, background 0.3s;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-right: 1px solid hsl(var(--sidebar-border));
  display: flex;
  flex-direction: column;

  .logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: hsl(var(--sidebar-foreground));
    font-size: 18px;
    font-weight: 600;
    letter-spacing: 2px;
    border-bottom: 1px solid hsl(var(--sidebar-border));
    flex-shrink: 0;

    img {
      width: 36px;
      height: 36px;
      margin-right: 10px;
      border-radius: 6px;
      background: transparent;
    }
  }

  .menu-wrapper {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;

    &::-webkit-scrollbar {
      width: 4px;
    }

    &::-webkit-scrollbar-thumb {
      background: hsl(var(--sidebar-border));
      border-radius: 2px;
    }
  }

  .menu-group-title {
    padding: 16px 20px 8px;
    font-size: 11px;
    font-weight: 500;
    color: hsl(var(--sidebar-foreground) / 0.5);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .el-menu {
    border-right: none;
    background: transparent !important;

    .el-menu-item {
      margin: 2px 8px;
      border-radius: 8px;
      height: 40px;
      line-height: 40px;
      transition: all 0.3s;
      color: hsl(var(--sidebar-foreground));

      &:hover {
        background: hsl(var(--sidebar-accent)) !important;
        color: hsl(var(--sidebar-accent-foreground)) !important;
      }

      &.is-active {
        background: var(--el-color-primary) !important;
        color: var(--el-color-primary-contrast, #fff) !important;
        box-shadow: 0 2px 8px var(--el-color-primary-light-5);
      }
    }

    .el-sub-menu {
      .el-sub-menu__title {
        margin: 2px 8px;
        border-radius: 8px;
        height: 40px;
        line-height: 40px;
        color: hsl(var(--sidebar-foreground));

        &:hover {
          background: hsl(var(--sidebar-accent)) !important;
          color: hsl(var(--sidebar-accent-foreground)) !important;
        }
      }

      &.is-opened>.el-sub-menu__title {
        color: hsl(var(--sidebar-foreground));
      }

      .el-menu {
        background: transparent !important;

        .el-menu-item {
          padding-left: 50px !important;
          min-width: auto;
          height: 36px;
          line-height: 36px;
          font-size: 13px;
        }
      }
    }

    // 收起状态下的样式
    &.el-menu--collapse {
      .el-menu-item {
        margin: 2px 8px;
        padding: 0 !important;
        justify-content: center;
        
        .el-icon {
          margin-right: 0;
        }
      }

      .el-sub-menu {
        .el-sub-menu__title {
          margin: 2px 8px;
          padding: 0 12px !important;
          
          .el-icon {
            margin-right: 0;
          }
        }
      }
    }
  }

}

.header {
  background: hsl(var(--background));
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  border-bottom: 1px solid hsl(var(--border));
  transition: background 0.3s;

  .header-left {
    display: flex;
    align-items: center;

    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
      margin-right: 20px;
      color: hsl(var(--muted-foreground));
      transition: color 0.3s;

      &:hover {
        color: hsl(var(--primary));
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .theme-switch {
      width: 36px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 8px;
      cursor: pointer;
      color: hsl(var(--muted-foreground));
      transition: all 0.3s;

      &:hover {
        background: hsl(var(--accent));
        color: hsl(var(--primary));
      }

      .el-icon {
        font-size: 18px;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 8px;
      transition: background 0.3s;

      &:hover {
        background: hsl(var(--accent));
      }

      .username {
        margin-left: 8px;
        color: hsl(var(--foreground));
      }
    }
  }
}

.main {
  background: hsl(var(--background));
  padding: 20px 200px;
  overflow-y: auto;
  transition: background 0.3s;
}
</style>
