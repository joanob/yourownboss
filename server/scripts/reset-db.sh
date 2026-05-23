#!/bin/bash

# Script para resetear la base de datos (Linux/macOS)
# Eliminará la BD y ejecutará schema.sql nuevamente

echo "================================"
echo "Reset de Base de Datos"
echo "================================"

# Obtener ruta del directorio del proyecto
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DB_PATH="$PROJECT_ROOT/yourownboss.db"

# Verificar si DB existe
if [ -f "$DB_PATH" ]; then
    echo "Eliminando archivo de BD: $DB_PATH"
    rm -f "$DB_PATH"
fi

# Eliminar archivos WAL y SHM si existen
WAL_PATH="${DB_PATH}-wal"
SHM_PATH="${DB_PATH}-shm"

if [ -f "$WAL_PATH" ]; then
    echo "Eliminando archivo WAL: $WAL_PATH"
    rm -f "$WAL_PATH"
fi

if [ -f "$SHM_PATH" ]; then
    echo "Eliminando archivo SHM: $SHM_PATH"
    rm -f "$SHM_PATH"
fi

echo ""
echo "✓ Base de datos reseteada exitosamente"
echo "  Próximo paso: ejecutar 'make run' para inicializar la BD con el schema"
echo ""
