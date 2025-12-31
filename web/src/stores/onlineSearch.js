import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useOnlineSearchStore = defineStore('onlineSearch', () => {
  // 搜索表单
  const searchForm = ref({
    source: 'fofa',
    query: '',
    page: 1,
    size: 50
  })

  // 搜索结果
  const tableData = ref([])
  const total = ref(0)

  // 保存搜索状态
  function saveState(form, data, totalCount) {
    searchForm.value = { ...form }
    tableData.value = data
    total.value = totalCount
  }

  // 清空状态
  function clearState() {
    searchForm.value = {
      source: 'fofa',
      query: '',
      page: 1,
      size: 50
    }
    tableData.value = []
    total.value = 0
  }

  return {
    searchForm,
    tableData,
    total,
    saveState,
    clearState
  }
})
