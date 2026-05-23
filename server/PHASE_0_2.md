# Fase 0.2 - Configuración de Base de Datos

Este documento describe cómo se configura y se inicializa la base de datos SQLite para el proyecto Your Own Boss.

## Estructura

```
server/
├── internal/
│   └── db/
│       ├── init.go           # Funciones de inicialización de BD
│       └── migrations/
│           └── schema.sql    # Esquema completo de la BD
├── setup.ps1                 # Script de setup (Windows)
├── setup.sh                  # Script de setup (Linux/Mac)
└── Makefile                  # Comandos útiles
```

## Archivos Principales

### 1. `internal/db/init.go`

Contiene las funciones para inicializar y gestionar la base de datos:

- **`InitDatabase(dbPath string) (*sql.DB, error)`**
  - Crea el archivo de BD si no existe
  - Ejecuta el schema.sql automáticamente
  - Habilita foreign keys
  - Retorna conexión abierta a la BD

- **`GetDatabase(dbPath string) (*sql.DB, error)`**
  - Abre una conexión existente a la BD
  - NO crea la BD ni ejecuta migraciones
  - Útil para conexiones posteriores

- **`ResetDatabase(dbPath string) (*sql.DB, error)`**
  - Elimina la BD existente
  - Reinicializa desde cero
  - Útil para desarrollo y pruebas

### 2. `internal/db/migrations/schema.sql`

Archivo SQL que define toda la estructura de la base de datos:

- Tablas de datos maestros (resources, buildings, processes)
- Tablas de usuarios y autenticación
- Tablas de empresas e inventario
- Tablas de producción y venta
- Tablas de rate limiting y auditoría
- Índices para optimización
- Triggers para soft-delete automático

**Características:**
- Sin migraciones versiona das: el schema es la única fuente de verdad
- Triggers automáticos para soft-delete
- Foreign keys habilitadas
- Indices en columnas frecuentemente consultadas

### 3. `cmd/api/main.go` (Integración)

El servidor inicializa la BD automáticamente al arrancar:

```go
// Obtener ruta de BD desde ENV (default: ./yourownboss.db)
dbPath := os.Getenv("DATABASE_URL")
if dbPath == "" {
    dbPath = "./yourownboss.db"
}

// Inicializar BD
dbConn, err := db.InitDatabase(dbPath)
if err != nil {
    logger.Fatal().Err(err).Msg("Error al inicializar BD")
    os.Exit(1)
}
defer dbConn.Close()
```

## Setup Initial

### Opción 1: Usar Script de Setup (Recomendado)

#### Windows (PowerShell):
```powershell
cd server
.\setup.ps1
```

El script:
1. Verifica que estamos en el directorio correcto
2. Copia `.env.example` a `.env` (si no existe)
3. Descarga dependencias Go
4. Pregunta si deseas limpiar la BD anterior
5. Verifica archivos necesarios
6. Compila el servidor para verificar todo funciona

#### Linux/Mac (Bash):
```bash
cd server
chmod +x setup.sh
./setup.sh
```

### Opción 2: Setup Manual

```bash
cd server

# 1. Copiar .env.example a .env
cp .env.example .env

# 2. Descargar dependencias
go mod download

# 3. Compilar
go build -o yourownboss-api ./cmd/api

# 4. Ejecutar (la BD se crea automáticamente)
./yourownboss-api
```

### Opción 3: Usar Makefile

```bash
cd server

# Build y run
make run

# Build solo
make build

# Limpiar BD
make reset-db

# Limpiar todo
make clean
```

## Variables de Entorno (Database)

En `.env`:

```
# Ruta de la BD SQLite (default: ./yourownboss.db)
DATABASE_URL=./yourownboss.db
```

## Comportamiento Automático

Cuando el servidor inicia:

1. ✅ Lee `DATABASE_URL` del `.env` (o usa default `./yourownboss.db`)
2. ✅ Crea el archivo de BD si no existe
3. ✅ Ejecuta `schema.sql` automáticamente (solo la primera vez)
4. ✅ Habilita foreign keys
5. ✅ Prueba conexión (si falla, el servidor no inicia)
6. ✅ Logs del progreso en consola

## Para Desarrollo

### Limpiar BD y Reiniciar

```bash
make reset-db
go run ./cmd/api
```

O sin Makefile:

```bash
rm -f yourownboss.db yourownboss.db-shm yourownboss.db-wal
go run ./cmd/api
```

### Verificar Estructura de BD

```bash
# Instalar sqlite3 CLI si no está
# macOS: brew install sqlite
# Windows: https://www.sqlite.org/download.html
# Linux: apt-get install sqlite3

sqlite3 yourownboss.db ".tables"
sqlite3 yourownboss.db ".schema"
```

### Logs de Inicialización

El servidor loguea el progreso:

```
INFO Inicializando base de datos...
INFO Base de datos inicializada exitosamente
```

Si hay errores:

```
ERROR error al inicializar base de datos: [detalle del error]
```

## Características Implementadas

✅ Inicialización automática de BD en primer arranque
✅ Schema SQL como única fuente de verdad
✅ Foreign keys habilitadas por defecto
✅ Soft-delete con triggers automáticos
✅ Índices para optimización
✅ Función para resetear BD (desarrollo)
✅ Scripts de setup para Windows y Unix
✅ Integración en main.go

## Próximas Fases

- **Fase 1**: Autenticación y Usuarios
  - Tablas users, user_sessions ya están creadas
  - Se implementarán repositories y servicios

- **Fase 2**: Empresas y Dinero
  - Tablas companies, company_inventory ya existen
  - Se implementarán servicios de empresa

## Documentación

- [Schema SQL Completo](./internal/db/migrations/schema.sql)
- [README Backend](./README.md)
- [Especificación Completa](../../docs/YOUROWNBOSS.md)
