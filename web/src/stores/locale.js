import { defineStore } from 'pinia'
import { ref } from 'vue'
import { setLocale, getLocale, SUPPORT_LOCALES } from '@/i18n'

export const useLocaleStore = defineStore('locale', () => {
  const currentLocale = ref(getLocale())

  function changeLocale(locale) {
    if (SUPPORT_LOCALES.includes(locale)) {
      currentLocale.value = locale
      setLocale(locale)
    }
  }

  function toggleLocale() {
    const newLocale = currentLocale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
    changeLocale(newLocale)
  }

  return {
    currentLocale,
    changeLocale,
    toggleLocale,
    supportLocales: SUPPORT_LOCALES
  }
})