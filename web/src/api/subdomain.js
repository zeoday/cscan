import request from './request'

// 子域名字典列表
export function getSubdomainDictList(data) {
  return request.post('/subdomain/dict/list', data)
}

// 保存子域名字典
export function saveSubdomainDict(data) {
  return request.post('/subdomain/dict/save', data)
}

// 删除子域名字典
export function deleteSubdomainDict(data) {
  return request.post('/subdomain/dict/delete', data)
}

// 清空子域名字典（非内置）
export function clearSubdomainDict() {
  return request.post('/subdomain/dict/clear')
}

// 获取启用的子域名字典列表（用于任务创建时选择）
export function getSubdomainDictEnabledList() {
  return request.post('/subdomain/dict/enabled')
}
