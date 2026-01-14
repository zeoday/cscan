import request from './request'

// 获取定时任务列表
export function getCronTaskList(data) {
  return request.post('/task/cron/list', data)
}

// 保存定时任务
export function saveCronTask(data) {
  return request.post('/task/cron/save', data)
}

// 开关定时任务
export function toggleCronTask(data) {
  return request.post('/task/cron/toggle', data)
}

// 删除定时任务
export function deleteCronTask(data) {
  return request.post('/task/cron/delete', data)
}

// 批量删除定时任务
export function batchDeleteCronTask(data) {
  return request.post('/task/cron/batchDelete', data)
}

// 立即执行定时任务
export function runCronTaskNow(data) {
  return request.post('/task/cron/runNow', data)
}

// 验证Cron表达式
export function validateCronSpec(data) {
  return request.post('/task/cron/validate', data)
}
