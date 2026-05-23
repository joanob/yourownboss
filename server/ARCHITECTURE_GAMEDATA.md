# Arquitectura Refactorizada: Gamedata con BD

## Cambios Principales

Se refactorizó la carga de datos maestros (gamedata) para usar la base de datos como fuente de verdad:

### Flujo Anterior (0.3)
```
JSON → Cache (JSON se lee en cada startup)
```

### Flujo Nuevo (Mejorado)
```
JSON → BD (primera vez o importación admin)
  ↓
Cache ← BD (select al startup)
  ↓
Endpoints Admin → BD → Cache (refresh)
```

## Componentes Nuevos

### 1. Repository para Recursos
**`internal/gamedata/repository/resource_repo.go`**
- `GetAll()` - SELECT desde BD
- `InsertBatch()` - INSERT múltiples en transacción
- `Count()` - Verifica si hay datos en BD

### 2. Repository para Edificios de Producción
**`internal/gamedata/repository/production_building_repo.go`**
- `GetAll()` - SELECT edificios + procesos + recursos
- `InsertBatch()` - INSERT en transacción (3 operaciones)
- `Count()` - Cuenta edificios en BD

### 3. Repository para Edificios de Venta
**`internal/gamedata/repository/sale_building_repo.go`**
- `GetAll()` - SELECT edificios + recursos
- `InsertBatch()` - INSERT en transacción
- `Count()` - Cuenta edificios en BD

## GamedataService Refactorizado

**Constructor:**
```go
NewGamedataService(filePath string, db *sql.DB)
```

**Métodos:**

1. **`Load()`** - Asegura que BD tiene datos:
   - Verifica si BD está vacía (Count)
   - Si vacía, importa desde JSON
   - NO sincroniza cache
   - Responsabilidad: **JSON → BD**

2. **`RefreshCache(gameCache *GamedataCache)`** - Sincroniza cache desde BD:
   - Recibe cache como parámetro
   - Carga recursos desde BD
   - Carga edificios de producción desde BD
   - Carga edificios de venta desde BD
   - Actualiza cache
   - Responsabilidad: **BD → Cache**

3. **`importFromFile()`** - Importación desde JSON:
   - Lee JSON
   - Valida estructura
   - Inserta en BD (3 transacciones)

## Flujo de Inicialización en main.go

```go
// 1. Inicializar BD (schema.sql)
dbConn, _ := db.InitDatabase(dbPath)

// 2. Crear caches
gamedataCache := cache.NewGamedataCache()
sessionCache := cache.NewSessionCache()

// 3. Crear servicio
gamedataSvc := service.NewGamedataService(filePath, dbConn, cache)

// 4. PRIMERO: Asegurar que BD tiene datos (importa si vacía)
gamedataSvc.Load()  // JSON → BD (si es necesario)

// 5. SEGUNDO: Sincronizar cache desde BD
gamedataSvc.RefreshCache()  // BD → Cache

// 6. Servidor listo
```

**Separación de responsabilidades:**
- `Load()` = asegurar BD tiene datos (JSON → BD si necesario)
- `RefreshCache()` = sincronizar cache desde BD

## Ventajas de Esta Arquitectura

✅ **BD como fuente de verdad** - Datos maestros persistidos
✅ **Importación única** - JSON se importa una sola vez a BD
✅ **Actualizaciones sin restart** - Endpoint admin puede actualizar datos
✅ **Cache consistente** - Siempre derivado de BD
✅ **Transacciones seguras** - Inserciones atómicas en BD
✅ **Escalabilidad** - Preparado para multi-instancia (cache desde BD compartida)

## Casos de Uso Futuros

### Caso 1: Resincronizar cache (Admin)
```go
// En handler admin POST /api/v1/admin/gamedata/refresh
if err := gamedataSvc.RefreshCache(gamedataCache); err != nil {
    // error handling
}
```

**Útil si:**
- Se modifican datos maestros en BD
- Se quiere actualizar cache sin reiniciar servidor

### Caso 2: Importar nuevos datos desde JSON (Admin)
```go
// En handler admin POST /api/v1/admin/gamedata/import
if err := gamedataSvc.Load(); err != nil {  // Importa a BD si está vacía
    // error handling
}
if err := gamedataSvc.RefreshCache(gamedataCache); err != nil {  // Sincroniza cache
    // error handling
}
```

**Nota:** `Load()` solo importa si BD está vacía. Para forzar importación:
```go
// Primero vaciar BD (eliminar registros)
// Luego llamar Load()
```

## Logs de Inicialización

```
INFO Verificando datos maestros en BD...
INFO BD vacía, importando datos desde JSON...
INFO Validación de gamedata exitosa
INFO Recursos importados a BD count=5
INFO Edificios de producción importados a BD count=2
INFO Edificios de venta importados a BD count=2
INFO Sincronizando cache desde BD...
INFO Cache sincronizado exitosamente resources=5 production_buildings=2 sale_buildings=2
```

O si BD ya tiene datos:
```
INFO Verificando datos maestros en BD...
INFO Datos maestros encontrados en BD resources_in_db=5
INFO Sincronizando cache desde BD...
INFO Cache sincronizado exitosamente resources=5 production_buildings=2 sale_buildings=2
```

## Variables de Entorno

Las mismas de antes:
```
GAMEDATA_FILE=./config/gamedata.json
DATABASE_URL=./yourownboss.db
```

## Próximos Pasos

- **Endpoints Admin para Gamedata**
  - POST /api/v1/admin/gamedata/import - Importar desde JSON
  - POST /api/v1/admin/gamedata/refresh - Resincronizar cache
  - GET /api/v1/admin/gamedata/status - Estado de cache y BD

- **Fase 1**: Autenticación y Usuarios
