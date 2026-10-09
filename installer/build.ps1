$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$iss  = Join-Path $PSScriptRoot 'setup.iss'

# ISCC only decodes non-ASCII script text correctly when the file carries a UTF-8 BOM.
$bytes = [IO.File]::ReadAllBytes($iss)
if ($bytes.Length -lt 3 -or $bytes[0] -ne 0xEF -or $bytes[1] -ne 0xBB -or $bytes[2] -ne 0xBF) {
    [IO.File]::WriteAllBytes($iss, [byte[]](0xEF, 0xBB, 0xBF) + $bytes)
    Write-Output 'bom-added'
}

$iscc = "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe"
if (-not (Test-Path $iscc)) { throw 'ISCC.exe not found; install Inno Setup 6' }

$exe = Join-Path $root 'build\bin\wr_tool.exe'
if (-not (Test-Path $exe)) { throw 'build\bin\wr_tool.exe missing; run: wails build' }

& $iscc $iss
if ($LASTEXITCODE -ne 0) { throw "ISCC failed with $LASTEXITCODE" }

Get-ChildItem (Join-Path $root 'build\installer') | Format-Table -AutoSize Name, Length, LastWriteTime
