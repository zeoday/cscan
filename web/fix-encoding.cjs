const fs = require('fs');
const path = require('path');

// 定义需要修复的文件模式和替换规则
const fixes = [
  // 通用的编码修复
  { from: /�/g, to: '' },
  { from: /\?\?/g, to: '？' },
  { from: /根域�\?/g, to: '根域名' },
  { from: /已解�\?/g, to: '已解析' },
  { from: /未解�\?/g, to: '未解析' },
  { from: /个域�\?/g, to: '个域名' },
  { from: /个站�\?/g, to: '个站点' },
  { from: /状�\?/g, to: '状态' },
  { from: /响应�\?/g, to: '响应头' },
  { from: /选中�\?/g, to: '选中的' },
  { from: /详情对话�\?/g, to: '详情对话框' },
  { from: /Worker状�\?/g, to: 'Worker状态' },
  { from: /请输入名�\?/g, to: '请输入名称' },
  { from: /请输入描�\?/g, to: '请输入描述' },
  { from: /请输入应用名�\?/g, to: '请输入应用名称' },
  { from: /请输入查询语�\?/g, to: '请输入查询语句' },
  { from: /请输入扫描目�\?/g, to: '请输入扫描目标' },
  { from: /请输入任务名�\?/g, to: '请输入任务名称' },
  { from: /审计日志已清�\?/g, to: '审计日志已清空' },
  { from: /内置指�\?/g, to: '内置指纹' },
  { from: /自定义指�\?/g, to: '自定义指纹' },
  { from: /必需�\?/g, to: '必需）' },
  { from: /分钟�\?/g, to: '分钟前' },
  
  // 修复引号问题
  { from: /'确定删除该资产吗\?\?/g, to: "'确定删除该资产吗？'" },
  { from: /'确定删除该记录吗\?\?/g, to: "'确定删除该记录吗？'" },
  { from: /'确定删除该域名吗\?\?/g, to: "'确定删除该域名吗？'" },
  { from: /'确定删除该站点吗\?\?/g, to: "'确定删除该站点吗？'" },
  { from: /'确定删除该漏洞记录吗\?\?/g, to: "'确定删除该漏洞记录吗？'" },
  { from: /'确定删除该IP及其所有资产吗\?\?/g, to: "'确定删除该IP及其所有资产吗？'" },
  
  // 修复属性名问题
  { from: /label="根域�\?"/g, to: 'label="根域名"' },
  { from: /placeholder="根域�\?"/g, to: 'placeholder="根域名"' },
  { from: /label="状�\?"/g, to: 'label="状态"' },
  { from: /label="响应�\?"/g, to: 'label="响应头"' },
  
  // 修复标签内容
  { from: />�\?</g, to: '>新<' },
  { from: />已解�\?</g, to: '>已解析<' },
  { from: />未解�\?</g, to: '>未解析<' },
  { from: />状�\?</g, to: '>状态<' },
  
  // 修复字符串模板
  { from: /共 \$\{/g, to: '共 ${' },
  { from: /\$\{.*?\} 个域�\?/g, to: (match) => match.replace('个域�?', '个域名') },
  { from: /\$\{.*?\} 个站�\?/g, to: (match) => match.replace('个站�?', '个站点') },
  { from: /选中�\?\$\{/g, to: '选中的${' }
];

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
    
    for (const fix of fixes) {
      const newContent = content.replace(fix.from, fix.to);
      if (newContent !== content) {
        content = newContent;
        changed = true;
      }
    }
    
    if (changed) {
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