import request from './request'

// 目录扫描字典列表
export function getDirScanDictList(data) {
  return request.post('/dirscan/dict/list', data)
}

// 保存目录扫描字典
export function saveDirScanDict(data) {
  return request.post('/dirscan/dict/save', data)
}

// 删除目录扫描字典
export function deleteDirScanDict(data) {
  return request.post('/dirscan/dict/delete', data)
}

// 清空目录扫描字典（非内置）
export function clearDirScanDict() {
  return request.post('/dirscan/dict/clear')
}

// 获取启用的目录扫描字典列表（用于任务创建时选择）
export function getDirScanDictEnabledList() {
  return request.post('/dirscan/dict/enabled')
}

// ==================== 目录扫描结果 API ====================

// 目录扫描结果列表
export function getDirScanResultList(data) {
  return request.post('/dirscan/result/list', data)
}

// 目录扫描结果统计
export function getDirScanResultStat(data) {
  return request.post('/dirscan/result/stat', data)
}

// 删除目录扫描结果
export function deleteDirScanResult(data) {
  return request.post('/dirscan/result/delete', data)
}

// 批量删除目录扫描结果
export function batchDeleteDirScanResult(data) {
  return request.post('/dirscan/result/batchDelete', data)
}

// 清空目录扫描结果
export function clearDirScanResult(data) {
  return request.post('/dirscan/result/clear', data)
}
