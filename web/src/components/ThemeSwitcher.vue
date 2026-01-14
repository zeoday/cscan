<template>
  <el-popover
    placement="bottom"
    :width="320"
    trigger="click"
    popper-class="theme-switcher-popover"
  >
    <template #reference>
      <div class="theme-switch-btn" :title="$t('theme.settings')">
        <el-icon v-if="themeStore.isDark">
          <Moon />
        </el-icon>
        <el-icon v-else>
          <Sunny />
        </el-icon>
      </div>
    </template>
    
    <div class="theme-switcher">
      <!-- 主题模式选择 -->
      <div class="theme-section">
        <div class="section-title">{{ $t('theme.mode') }}</div>
        <div class="mode-options">
          <div
            v-for="mode in themeModes"
            :key="mode.value"
            class="mode-option"
            :class="{ active: themeStore.theme === mode.value }"
            @click="themeStore.setTheme(mode.value)"
          >
            <el-icon :size="20">
              <component :is="mode.icon" />
            </el-icon>
            <span>{{ $t(mode.label) }}</span>
          </div>
        </div>
      </div>
      
      <!-- 颜色主题选择 -->
      <div class="theme-section">
        <div class="section-title">{{ $t('theme.colorTheme') }}</div>
        <div class="color-options">
          <div
            v-for="color in colorThemes"
            :key="color.value"
            class="color-option"
            :class="{ active: themeStore.colorTheme === color.value }"
            :style="{ '--theme-color': color.color }"
            :title="$t(color.label)"
            @click="themeStore.setColorTheme(color.value)"
          >
            <div class="color-preview" :style="{ background: color.color }"></div>
            <el-icon v-if="themeStore.colorTheme === color.value" class="check-icon">
              <Check />
            </el-icon>
          </div>
        </div>
      </div>
    </div>
  </el-popover>
</template>

<script setup>
import { computed } from 'vue'
import { useThemeStore, COLOR_THEMES } from '@/stores/theme'
import { Sunny, Moon, Monitor, Check } from '@element-plus/icons-vue'

const themeStore = useThemeStore()

const themeModes = [
  { value: 'light', label: 'theme.light', icon: Sunny },
  { value: 'dark', label: 'theme.dark', icon: Moon },
  { value: 'system', label: 'theme.system', icon: Monitor },
]

const colorThemes = computed(() => COLOR_THEMES)
</script>

<style lang="scss" scoped>
.theme-switch-btn {
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

.theme-switcher {
  .theme-section {
    margin-bottom: 16px;
    
    &:last-child {
      margin-bottom: 0;
    }
  }
  
  .section-title {
    font-size: 13px;
    font-weight: 500;
    color: hsl(var(--muted-foreground));
    margin-bottom: 10px;
  }
  
  .mode-options {
    display: flex;
    gap: 8px;
    
    .mode-option {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 6px;
      padding: 12px 8px;
      border-radius: 8px;
      cursor: pointer;
      border: 2px solid transparent;
      background: hsl(var(--muted));
      color: hsl(var(--muted-foreground));
      transition: all 0.2s;
      
      &:hover {
        background: hsl(var(--accent));
        color: hsl(var(--foreground));
      }
      
      &.active {
        border-color: hsl(var(--primary));
        background: hsl(var(--primary) / 0.1);
        color: hsl(var(--primary));
      }
      
      span {
        font-size: 12px;
      }
    }
  }
  
  .color-options {
    display: grid;
    grid-template-columns: repeat(6, 1fr);
    gap: 8px;
    
    .color-option {
      position: relative;
      width: 100%;
      aspect-ratio: 1;
      border-radius: 8px;
      cursor: pointer;
      border: 2px solid transparent;
      background: hsl(var(--muted));
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.2s;
      
      &:hover {
        transform: scale(1.1);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
      }
      
      &.active {
        border-color: var(--theme-color);
      }
      
      .color-preview {
        width: 24px;
        height: 24px;
        border-radius: 50%;
        box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.1);
      }
      
      .check-icon {
        position: absolute;
        bottom: 2px;
        right: 2px;
        width: 14px;
        height: 14px;
        background: var(--theme-color);
        border-radius: 50%;
        color: white;
        font-size: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }
  }
}
</style>

<style>
.theme-switcher-popover {
  padding: 16px !important;
}
</style>
