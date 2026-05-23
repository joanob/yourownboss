# Script para resetear la base de datos (Windows PowerShell)
# Eliminará la BD y ejecutará schema.sql nuevamente

Write-Host "================================" -ForegroundColor Cyan
Write-Host "Reset de Base de Datos" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# Obtener ruta del directorio del proyecto
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptPath
$dbPath = Join-Path $projectRoot "yourownboss.db"

# Verificar si DB existe
if (Test-Path $dbPath) {
    Write-Host "Eliminando archivo de BD: $dbPath" -ForegroundColor Yellow
    Remove-Item $dbPath -ErrorAction SilentlyContinue
}

# Eliminar archivos WAL y SHM si existen
$walPath = "$dbPath-wal"
$shmPath = "$dbPath-shm"

if (Test-Path $walPath) {
    Write-Host "Eliminando archivo WAL: $walPath" -ForegroundColor Yellow
    Remove-Item $walPath -ErrorAction SilentlyContinue
}

if (Test-Path $shmPath) {
    Write-Host "Eliminando archivo SHM: $shmPath" -ForegroundColor Yellow
    Remove-Item $shmPath -ErrorAction SilentlyContinue
}

Write-Host "`nBase de datos reseteada exitosamente" -ForegroundColor Green
Write-Host "Próximo paso: ejecutar 'make run' para inicializar la BD con el schema" -ForegroundColor Green
Write-Host ""
