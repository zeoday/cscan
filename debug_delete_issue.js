// 调试删除问题的脚本
// 在浏览器控制台运行此脚本来检查删除请求

console.log('=== 删除问题调试脚本 ===');

// 1. 检查当前工作空间ID
const workspaceStore = window.localStorage.getItem('workspace-store');
if (workspaceStore) {
    const parsed = JSON.parse(workspaceStore);
    console.log('当前工作空间ID:', parsed.currentWorkspaceId);
    
    if (!parsed.currentWorkspaceId || parsed.currentWorkspaceId === 'all') {
        console.error('❌ 问题发现: 工作空间ID为空或为"all"，这会导致删除失败');
        console.log('解决方案: 请选择一个具体的工作空间');
    } else {
        console.log('✅ 工作空间ID正常');
    }
} else {
    console.error('❌ 未找到工作空间存储');
}

// 2. 检查用户token
const userStore = window.localStorage.getItem('user-store');
if (userStore) {
    const parsed = JSON.parse(userStore);
    if (parsed.token) {
        console.log('✅ 用户token存在');
    } else {
        console.error('❌ 用户token缺失');
    }
} else {
    console.error('❌ 未找到用户存储');
}

// 3. 监听删除请求
const originalFetch = window.fetch;
window.fetch = function(...args) {
    const url = args[0];
    const options = args[1] || {};
    
    if (url.includes('/delete') || url.includes('/batchDelete') || url.includes('/clear')) {
        console.log('🔍 删除请求拦截:', {
            url: url,
            method: options.method,
            headers: options.headers,
            body: options.body
        });
    }
    
    return originalFetch.apply(this, args).then(response => {
        if (url.includes('/delete') || url.includes('/batchDelete') || url.includes('/clear')) {
            response.clone().json().then(data => {
                console.log('📥 删除响应:', data);
                if (data.code !== 0) {
                    console.error('❌ 删除失败:', data.msg);
                }
            });
        }
        return response;
    });
};

console.log('✅ 调试脚本已启动，现在可以尝试删除操作');