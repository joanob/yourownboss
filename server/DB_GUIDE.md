# Guía de Desarrollo - Base de Datos

## Inicialización Automática

La base de datos se crea y se inicializa **automáticamente** cuando ejecutas `make run` por primera vez.

```bash
make run
```

En los logs verás:
```
info: Base de datos inicializada exitosamente {"path": "yourownboss.db"}
```

## Resetear BD para Desarrollo

Durante el desarrollo, es común necesitar resetear la BD para empezar desde cero.

### Opción 1: Usar Makefile (Recomendado)

```bash
make reset-db
make run
```

### Opción 2: Scripts por Plataforma

**Windows (PowerShell):**
```powershell
.\scripts\reset-db.ps1
make run
```

**Linux/macOS:**
```bash
./scripts/reset-db.sh
make run
```

### Opción 3: Manual con SQLite CLI

```bash
# Borrar archivos de BD
rm -f yourownboss.db yourownboss.db-shm yourownboss.db-wal

# Ejecutar servidor (reinicializa BD)
make run
```

## Verificar Integridad de BD

### Verificar tablas con Makefile

```bash
make db-verify
```

Output esperado:
```
Verifying database integrity...
16
audit_log               companies               company_inventory
company_production_buildings  company_sale_buildings      login_attempts
...
```

### Verificar manualmente con SQLite CLI

```bash
# Listar todas las tablas
sqlite3 yourownboss.db ".tables"

# Contar tablas (debería ser 16)
sqlite3 yourownboss.db "SELECT COUNT(*) FROM sqlite_master WHERE type='table';"

# Ver estructura de tabla específica
sqlite3 yourownboss.db ".schema users"

# Verificar que foreign keys están habilitadas
sqlite3 yourownboss.db "PRAGMA foreign_keys;"

# Ver índices
sqlite3 yourownboss.db "SELECT name FROM sqlite_master WHERE type='index';"
```

## Ver Contenido de BD

```bash
# Ver todos los usuarios
sqlite3 yourownboss.db "SELECT id, username, email FROM users;"

# Ver sesiones activas
sqlite3 yourownboss.db "SELECT * FROM user_sessions WHERE revoked_at IS NULL;"

# Ver empresas
sqlite3 yourownboss.db "SELECT id, user_id, name, money FROM companies;"

# Ver recursos (gamedata)
sqlite3 yourownboss.db "SELECT id, name, market_price FROM resources;"
```

## Limpiar BD (eliminar datos, mantener schema)

```bash
# Dentro de SQLite CLI:
sqlite3 yourownboss.db

# Limpiar datos pero mantener schema
DELETE FROM users;
DELETE FROM companies;
DELETE FROM resources;
DELETE FROM production_buildings;
DELETE FROM sale_buildings;
DELETE FROM audit_log;

.quit
```

## Exportar/Importar BD

### Hacer backup

```bash
cp yourownboss.db yourownboss.db.backup
```

### Restaurar desde backup

```bash
cp yourownboss.db.backup yourownboss.db
```

### Exportar a SQL script

```bash
sqlite3 yourownboss.db ".dump" > backup.sql
```

### Importar de SQL script

```bash
sqlite3 new-db.db < backup.sql
```

## Troubleshooting

### Error: "database is locked"

Significa que hay otra instancia del servidor corriendo o conexión pendiente.

```bash
# Matar proceso en Windows
taskkill /F /IM yourownboss-api.exe

# Matar proceso en Linux/macOS
pkill -f yourownboss-api

# Luego resetear
make reset-db
make run
```

### Error: "schema.sql not found"

Verificar que `internal/db/migrations/schema.sql` existe. Si no, restaurar desde git:

```bash
git checkout internal/db/migrations/schema.sql
```

### BD corrupta

```bash
make reset-db
make run
```

## Estructura de Directorios

```
server/
├── yourownboss.db           # BD actual (generada, NO commitear)
├── yourownboss.db-shm       # Archivo WAL (NO commitear)
├── yourownboss.db-wal       # Archivo WAL (NO commitear)
├── internal/
│   └── db/
│       ├── init.go          # Función InitDatabase()
│       └── migrations/
│           └── schema.sql   # Schema principal (commitear)
└── scripts/
    ├── reset-db.ps1         # Script reset (Windows)
    └── reset-db.sh          # Script reset (Linux/macOS)
```

## .gitignore

Asegurar que los archivos de BD no se commitean:

```
# En .gitignore
yourownboss.db
yourownboss.db-shm
yourownboss.db-wal
*.db-wal
*.db-shm
```

## Performance

SQLite es suficientemente rápido para desarrollo y pruebas. Para producción, ver YOUROWNBOSS.md sección de arquitectura.

**Tipas de performance:**
- Queries simples: < 1ms
- Transacciones: 1-5ms
- Triggers de soft-delete: 2-10ms dependiendo de cascadas

Si en desarrollo notas queries lentas:
1. Verificar índices con `EXPLAIN QUERY PLAN`
2. Checar que constraints no están generando cascadas innecesarias
3. Revisar tamaño de BD: `ls -lh yourownboss.db`

## Próximos Pasos

- **Fase 1.1**: ✅ Configuración de BD completada
- **Fase 1.2**: Modelos y tipos (User, UserSession, etc.)
- **Fase 1.3**: Autenticación y JWT
