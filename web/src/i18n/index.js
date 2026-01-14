import { createI18n } from 'vue-i18n'

// 导入语言文件
import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'

export const SUPPORT_LOCALES = ['zh-CN', 'en-US']
export const DEFAULT_LOCALE = 'zh-CN'

// 获取浏览器语言
function getDefaultLocale() {
  const locale = navigator.language
  if (locale.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

// 创建 i18n 实例
export const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || getDefaultLocale(),
  fallbackLocale: DEFAULT_LOCALE,
  globalInjection: true,
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

// 切换语言
export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.lang = locale
}

// 获取当前语言
export function getLocale() {
  return i18n.global.locale.value
}

// 安装插件
export function setupI18n(app) {
  app.use(i18n)
}

export default i18n