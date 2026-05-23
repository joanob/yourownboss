# Fase 1.1 - Configuración de Base de Datos y Migraciones

## Objetivo
Verificar que la base de datos está correctamente inicializada con todas las tablas necesarias para Fase 1 (Autenticación y Usuarios).

## Estado de la BD

### Tablas Creadas (desde Fase 0.2)
- ✅ **users** - Información de usuarios con hash de contraseña
- ✅ **user_sessions** - Sesiones activas de usuarios con refresh tokens
- ✅ **resources** - Datos maestros de recursos
- ✅ **production_buildings** - Tipos de edificios de producción
- ✅ **production_processes** - Procesos productivos
- ✅ **production_process_resources** - Recursos de procesos
- ✅ **sale_buildings** - Tipos de edificios de venta
- ✅ **sale_resources** - Recursos de edificios de venta
- ✅ **companies** - Empresas de usuarios
- ✅ **company_inventory** - Inventario de empresas
- ✅ **company_production_buildings** - Instancias de producción
- ✅ **production_runs** - Ejecuciones de producción
- ✅ **company_sale_buildings** - Instancias de venta
- ✅ **sale_runs** - Ejecuciones de venta
- ✅ **login_attempts** - Rate limiting de intentos de login
- ✅ **audit_log** - Auditoría de cambios

### Características de la BD
- ✅ SQLite sin CGO (modernc.org/sqlite)
- ✅ Foreign keys habilitadas
- ✅ Triggers de soft-delete funcionales
- ✅ Índices optimizados para queries frecuentes
- ✅ Schema ejecutado automáticamente en `InitDatabase()`

## Procedimiento de Inicialización

### 1. Primera ejecución
```bash
# Windows
make run

# Linux/macOS
make run
```

En la primera ejecución:
1. `cmd/api/main.go` llama a `db.InitDatabase(dbPath)`
2. Se crea el archivo `yourownboss.db`
3. Se ejecuta `internal/db/migrations/schema.sql`
4. Se crean todas las tablas

**Esperado en los logs:**
```
...
info: Base de datos inicializada exitosamente {"path": "yourownboss.db"}
info: Verificación de datos maestros completada
...
```

### 2. Resetear BD (para desarrollo)

**Windows (PowerShell):**
```powershell
.\scripts\reset-db.ps1
make run
```

**Linux/macOS (Bash):**
```bash
chmod +x scripts/reset-db.sh
./scripts/reset-db.sh
make run
```

O usar el comando Makefile:
```bash
make reset-db
make run
```

### 3. Verificar integridad de BD

Para verificar que la BD está correctamente inicializada:

```bash
# Abrir SQLite CLI
sqlite3 yourownboss.db

# Dentro de SQLite, listar todas las tablas:
.tables

# Verificar schema de tabla específica:
.schema users
.schema user_sessions

# Verificar foreign keys están habilitadas:
PRAGMA foreign_keys;

# Salir
.quit
```

**Tablas esperadas:**
```
audit_log               companies
company_inventory       company_production_buildings
company_sale_buildings  login_attempts
production_building... (truncated for brevity)
resources              sale_buildings
sale_resources         sale_runs
user_sessions          users
```

## Validación de Migraciones

Todas las tablas requeridas fueron creadas en Fase 0.2 mediante `schema.sql`. 

Para verificar que se ejecutó correctamente:

```bash
# Contar tablas (debería ser 16)
sqlite3 yourownboss.db "SELECT COUNT(*) FROM sqlite_master WHERE type='table';"

# Contar índices
sqlite3 yourownboss.db "SELECT COUNT(*) FROM sqlite_master WHERE type='index';"

# Verificar triggers
sqlite3 yourownboss.db "SELECT COUNT(*) FROM sqlite_master WHERE type='trigger';"
```

## Cambios para Fase 1

Fase 1 NO requiere cambios al schema de BD. Todas las tablas necesarias (`users`, `user_sessions`, etc.) ya están creadas.

El trabajo de Fase 1 se enfoca en:
- ✅ Modelos de tipos (DBO, Model, DTO)
- ✅ Repositorios con queries en `sqlc`
- ✅ Servicios de autenticación y usuarios
- ✅ Handlers HTTP
- ✅ Middleware de autenticación

## Scripts Disponibles

- **`scripts/reset-db.ps1`** - Resetea BD en Windows (PowerShell)
- **`scripts/reset-db.sh`** - Resetea BD en Linux/macOS (Bash)
- **`make reset-db`** - Resetea BD (plataforma agnóstica)

## Próximos Pasos (Fase 1.2)

Una vez confirmado que la BD está correctamente inicializada:

1. ✅ BD inicializada y verificada
2. → Pasar a **Fase 1.2: Modelos y Tipos Base**
