<template>
  <el-dropdown @command="handleCommand" trigger="click">
    <div class="language-switcher">
      <el-icon><Position /></el-icon>
      <span>{{ getCurrentLanguageLabel() }}</span>
      <el-icon class="arrow"><ArrowDown /></el-icon>
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item 
          v-for="locale in localeStore.supportLocales" 
          :key="locale"
          :command="locale"
          :class="{ active: localeStore.currentLocale === locale }"
        >
          {{ getLanguageName(locale) }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { useLocaleStore } from '@/stores/locale'
import { Position, ArrowDown } from '@element-plus/icons-vue'

const localeStore = useLocaleStore()

function getCurrentLanguageLabel() {
  return localeStore.currentLocale === 'zh-CN' ? '中文' : 'EN'
}

function getLanguageName(locale) {
  const names = {
    'zh-CN': '简体中文',
    'en-US': 'English'
  }
  return names[locale] || locale
}

function handleCommand(locale) {
  localeStore.changeLocale(locale)
}
</script>

<style scoped>
.language-switcher {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  color: hsl(var(--muted-foreground));
  
  &:hover {
    background: hsl(var(--accent));
    color: hsl(var(--accent-foreground));
  }
  
  .el-icon {
    font-size: 16px;
  }
  
  .arrow {
    font-size: 12px;
  }
  
  span {
    font-size: 14px;
  }
}

:deep(.el-dropdown-menu__item.active) {
  color: hsl(var(--primary));
  background: hsl(var(--primary) / 0.1);
}
</style>

