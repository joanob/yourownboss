#!/bin/bash

# Setup del Backend - Your Own Boss
# Script para inicializar el entorno de desarrollo (Linux/Mac)

echo "=== Setup Backend - Your Own Boss ==="
echo ""

# Verificar que estamos en el directorio correcto
if [ ! -f "go.mod" ]; then
    echo "Error: go.mod no encontrado. Ejecuta este script desde el directorio 'server/'"
    exit 1
fi

# 1. Crear archivo .env si no existe
echo "[1/5] Verificando archivo .env..."
if [ ! -f ".env" ]; then
    echo "  → Copiando .env.example a .env"
    cp .env.example .env
    echo "  ✓ .env creado. Edítalo con tu configuración."
else
    echo "  ✓ .env ya existe"
fi

echo ""

# 2. Descargar dependencias
echo "[2/5] Descargando dependencias Go..."
go mod download
if [ $? -ne 0 ]; then
    echo "  ✗ Error al descargar dependencias"
    exit 1
fi
echo "  ✓ Dependencias descargadas"

echo ""

# 3. Limpiar BD anterior (opcional)
echo "[3/5] Verificando base de datos..."
if [ -f "yourownboss.db" ]; then
    echo "  ⚠ BD existente encontrada"
    read -p "  ¿Deseas limpiarla? (s/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        rm -f yourownboss.db yourownboss.db-shm yourownboss.db-wal
        echo "  ✓ BD limpiada"
    fi
else
    echo "  ✓ BD será creada automáticamente al iniciar"
fi

echo ""

# 4. Verificar archivos necesarios
echo "[4/5] Verificando archivos necesarios..."
requiredFiles=(
    "config/gamedata.json"
    "internal/db/migrations/schema.sql"
    ".env.example"
)

allExist=true
for file in "${requiredFiles[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✓ $file"
    else
        echo "  ✗ $file (FALTA)"
        allExist=false
    fi
done

if [ "$allExist" = false ]; then
    echo "  ⚠ Faltan archivos. Ejecuta desde el directorio 'server/'"
fi

echo ""

# 5. Compilar y probar
echo "[5/5] Compilando servidor..."
if go build -o test-build ./cmd/api; then
    rm -f test-build
    echo "  ✓ Compilación exitosa"
else
    echo "  ✗ Error en la compilación"
    exit 1
fi

echo ""
echo "=== Setup completado exitosamente ==="
echo ""
echo "Próximos pasos:"
echo "  1. Editar .env con tu configuración"
echo "  2. Ejecutar: go run ./cmd/api"
echo ""
