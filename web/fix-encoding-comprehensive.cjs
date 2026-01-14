const fs = require('fs');
const path = require('path');

// 获取所有Vue文件
function getAllVueFiles(dir) {
  const files = [];
  const items = fs.readdirSync(dir);
  
  for (const item of items) {
    const fullPath = path.join(dir, item);
    const stat = fs.statSync(fullPath);
    
    if (stat.isDirectory() && !item.startsWith('.') && item !== 'node_modules') {
      files.push(...getAllVueFiles(fullPath));
    } else if (item.endsWith('.vue')) {
      files.push(fullPath);
    }
  }
  
  return files;
}

// 修复文件编码
function fixFile(filePath) {
  try {
    let content = fs.readFileSync(filePath, 'utf8');
    let changed = false;
    const originalContent = content;
    
    // 1. 修复所有的乱码字符
    content = content.replace(/�/g, '');
    
    // 2. 修复常见的中文字符编码问题
    const charFixes = {
      '根域�?': '根域名',
      '已解�?': '已解析',
      '未解�?': '未解析',
      '个域�?': '个域名',
      '个站�?': '个站点',
      '状�?': '状态',
      '响应�?': '响应头',
      '选中�?': '选中的',
      '详情对话�?': '详情对话框',
      'Worker状�?': 'Worker状态',
      '请输入名�?': '请输入名称',
      '请输入描�?': '请输入描述',
      '请输入应用名�?': '请输入应用名称',
      '请输入查询语�?': '请输入查询语句',
      '请输入扫描目�?': '请输入扫描目标',
      '请输入任务名�?': '请输入任务名称',
      '审计日志已清�?': '审计日志已清空',
      '内置指�?': '内置指纹',
      '自定义指�?': '自定义指纹',
      '必需�?': '必需）',
      '分钟�?': '分钟前',
      '开放端?': '开放端口',
      '端口?': '端口"'
    };
    
    for (const [wrong, correct] of Object.entries(charFixes)) {
      content = content.replace(new RegExp(wrong.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g'), correct);
    }
    
    // 3. 修复引号问题
    content = content.replace(/'确定删除该资产吗\?\?/g, "'确定删除该资产吗？'");
    content = content.replace(/'确定删除该记录吗\?\?/g, "'确定删除该记录吗？'");
    content = content.replace(/'确定删除该域名吗\?\?/g, "'确定删除该域名吗？'");
    content = content.replace(/'确定删除该站点吗\?\?/g, "'确定删除该站点吗？'");
    content = content.replace(/'确定删除该漏洞记录吗\?\?/g, "'确定删除该漏洞记录吗？'");
    content = content.replace(/'确定删除该IP及其所有资产吗\?\?/g, "'确定删除该IP及其所有资产吗？'");
    
    // 4. 修复标签问题
    content = content.replace(/>�?</g, '>新<');
    content = content.replace(/>?\/el-tag>/g, '>新</el-tag>');
    content = content.replace(/class="new-tag">?\/el-tag>/g, 'class="new-tag">新</el-tag>');
    
    // 5. 修复属性问题
    content = content.replace(/placeholder="端口?/g, 'placeholder="端口"');
    content = content.replace(/label="根域�?"/g, 'label="根域名"');
    content = content.replace(/placeholder="根域�?"/g, 'placeholder="根域名"');
    content = content.replace(/label="状�?"/g, 'label="状态"');
    content = content.replace(/label="响应�?"/g, 'label="响应头"');
    
    // 6. 修复div标签问题
    content = content.replace(/开放端?\/div>/g, '开放端口</div>');
    
    // 7. 修复字符串模板问题
    content = content.replace(/共 \$\{/g, '共 ${');
    content = content.replace(/选中�?\$\{/g, '选中的${');
    
    // 8. 修复特殊的双问号问题
    content = content.replace(/\?\?/g, '？');
    
    if (content !== originalContent) {
      fs.writeFileSync(filePath, content, 'utf8');
      console.log(`Fixed: ${filePath}`);
      return true;
    }
    
    return false;
  } catch (error) {
    console.error(`Error fixing ${filePath}:`, error.message);
    return false;
  }
}

// 主函数
function main() {
  const srcDir = path.join(__dirname, 'src');
  const vueFiles = getAllVueFiles(srcDir);
  
  console.log(`Found ${vueFiles.length} Vue files`);
  
  let fixedCount = 0;
  for (const file of vueFiles) {
    if (fixFile(file)) {
      fixedCount++;
    }
  }
  
  console.log(`Fixed ${fixedCount} files`);
}

main();