# Setup del Backend - Your Own Boss
# Script para inicializar el entorno de desarrollo

Write-Host "=== Setup Backend - Your Own Boss ===" -ForegroundColor Cyan
Write-Host ""

# Verificar que estamos en el directorio correcto
if (-not (Test-Path "go.mod")) {
    Write-Host "Error: go.mod no encontrado. Ejecuta este script desde el directorio 'server/'" -ForegroundColor Red
    exit 1
}

# 1. Crear archivo .env si no existe
Write-Host "[1/5] Verificando archivo .env..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    Write-Host "  → Copiando .env.example a .env" -ForegroundColor Cyan
    Copy-Item ".env.example" ".env"
    Write-Host "  ✓ .env creado. Edítalo con tu configuración." -ForegroundColor Green
} else {
    Write-Host "  ✓ .env ya existe" -ForegroundColor Green
}

Write-Host ""

# 2. Descargar dependencias
Write-Host "[2/5] Descargando dependencias Go..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ✗ Error al descargar dependencias" -ForegroundColor Red
    exit 1
}
Write-Host "  ✓ Dependencias descargadas" -ForegroundColor Green

Write-Host ""

# 3. Limpiar BD anterior (opcional)
Write-Host "[3/5] Verificando base de datos..." -ForegroundColor Yellow
if (Test-Path "yourownboss.db") {
    Write-Host "  ⚠ BD existente encontrada" -ForegroundColor Yellow
    $response = Read-Host "  ¿Deseas limpiarla? (s/n)"
    if ($response -eq "s" -or $response -eq "S") {
        Remove-Item "yourownboss.db" -ErrorAction SilentlyContinue
        Remove-Item "yourownboss.db-shm" -ErrorAction SilentlyContinue
        Remove-Item "yourownboss.db-wal" -ErrorAction SilentlyContinue
        Write-Host "  ✓ BD limpiada" -ForegroundColor Green
    }
} else {
    Write-Host "  ✓ BD será creada automáticamente al iniciar" -ForegroundColor Green
}

Write-Host ""

# 4. Verificar archivos necesarios
Write-Host "[4/5] Verificando archivos necesarios..." -ForegroundColor Yellow
$requiredFiles = @(
    "config/gamedata.json",
    "internal/db/migrations/schema.sql",
    ".env.example"
)

$allExist = $true
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "  ✓ $file" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $file (FALTA)" -ForegroundColor Red
        $allExist = $false
    }
}

if (-not $allExist) {
    Write-Host "  ⚠ Faltan archivos. Ejecuta desde el directorio 'server/'" -ForegroundColor Yellow
}

Write-Host ""

# 5. Compilar y probar
Write-Host "[5/5] Compilando servidor..." -ForegroundColor Yellow
go build -o test-build.exe ./cmd/api 2>&1 | Tee-Object -Variable buildOutput
if ($LASTEXITCODE -eq 0) {
    Remove-Item "test-build.exe" -ErrorAction SilentlyContinue
    Write-Host "  ✓ Compilación exitosa" -ForegroundColor Green
} else {
    Write-Host "  ✗ Error en la compilación" -ForegroundColor Red
    Write-Host "  Errores:" -ForegroundColor Red
    Write-Host $buildOutput -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Setup completado exitosamente ===" -ForegroundColor Green
Write-Host ""
Write-Host "Próximos pasos:" -ForegroundColor Cyan
Write-Host "  1. Editar .env con tu configuración"
Write-Host "  2. Ejecutar: go run ./cmd/api"
Write-Host ""
