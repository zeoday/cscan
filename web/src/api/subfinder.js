import request from './request'

// 获取Subfinder数据源列表
export function getSubfinderProviderList() {
  return request.post('/subfinder/provider/list', {})
}

// 保存Subfinder数据源配置
export function saveSubfinderProvider(data) {
  return request.post('/subfinder/provider/save', data)
}

// 获取所有支持的数据源信息
export function getSubfinderProviderInfo() {
  return request.post('/subfinder/provider/info', {})
}
