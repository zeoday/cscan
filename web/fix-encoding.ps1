# Comprehensive encoding fix script for Vue files
$ErrorActionPreference = "Stop"

$fileFixes = @{
    'src\components\asset\AssetAllView.vue' = @(
        @('确定删除该资产吗？?', '确定删除该资产吗？')
    )
    'src\components\asset\IPView.vue' = @(
        @('共 {{ pagination.total }} 个IP', '共 {{ pagination.total }} 个IP'),
        @('确定删除该IP及其所有资产吗？?', '确定删除该IP及其所有资产吗？')
    )
    'src\components\asset\DomainView.vue' = @(
        @('未解析', '未解析')
    )
    'src\views\Task.vue' = @(
        @('<Plus /\>', '<Plus />')
    )
    'src\views\Poc.vue' = @(
        @('{{ templateStats.info || 0 }}\<', '{{ templateStats.info || 0 }}</el-tag>'),
        @('label="状态"', 'label="状态"')
    )
    'src\views\Fingerprint.vue' = @(
        @('label="状态"', 'label="状态"')
    )
    'src\views\Worker.vue' = @(
        @('<Refresh />\<', '<Refresh /></el-icon>')
    )
    'src\views\Settings.vue' = @(
        @('label="数据源"', 'label="数据源"')
    )
}

$baseDir = "D:\cscan\cscan\web"
$fixedCount = 0
$notFoundCount = 0

foreach ($file in $fileFixes.Keys) {
    $fullPath = Join-Path $baseDir $file
    
    if (-not (Test-Path $fullPath)) {
        Write-Host "File not found: $fullPath" -ForegroundColor Yellow
        continue
    }
    
    Write-Host "`nProcessing: $file" -ForegroundColor Cyan
    $content = [System.IO.File]::ReadAllText($fullPath, [System.Text.Encoding]::UTF8)
    $originalContent = $content
    
    foreach ($fix in $fileFixes[$file]) {
        $oldStr = $fix[0]
        $newStr = $fix[1]
        
        if ($content -match [regex]::Escape($oldStr)) {
            $content = $content -replace [regex]::Escape($oldStr), $newStr
            Write-Host "  ✓ Fixed: $($oldStr.Substring(0, [Math]::Min(50, $oldStr.Length)))..." -ForegroundColor Green
            $fixedCount++
        } else {
            Write-Host "  ✗ Pattern not found: $($oldStr.Substring(0, [Math]::Min(50, $oldStr.Length)))..." -ForegroundColor Red
            $notFoundCount++
        }
    }
    
    if ($content -ne $originalContent) {
        [System.IO.File]::WriteAllText($fullPath, $content, [System.Text.Encoding]::UTF8)
        Write-Host "  → Saved changes" -ForegroundColor Green
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Fixed: $fixedCount patterns" -ForegroundColor Green
Write-Host "Not found: $notFoundCount patterns" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan
