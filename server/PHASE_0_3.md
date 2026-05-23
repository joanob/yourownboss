# Fase 0.3 - Inicialización de Cache de Gamedata (ACTUALIZADA)

## Descripción

Fase refactorizada que implementa un sistema de carga de datos maestros donde:
1. **BD es la fuente de verdad** para datos maestros
2. **Cache sincroniza desde BD** mediante SELECT
3. **Endpoints admin** pueden actualizar datos sin reiniciar

## Arquitectura

```
JSON
 ↓
[Load()]
   Verifica BD → Importa si vacía
   Responsabilidad: JSON → BD
 ↓
[RefreshCache()]
   Carga desde BD → Actualiza cache
   Responsabilidad: BD → Cache
 ↓
[Repositories BD]
   - ResourceRepo → SELECT/INSERT recursos
   - ProductionBuildingRepo → SELECT/INSERT edificios + procesos
   - SaleBuildingRepo → SELECT/INSERT edificios venta
 ↓
[Cache: GamedataCache]
   (thread-safe, READ-ONLY después de startup)
 ↓
[Handlers REST]
   (leen del cache, rápido)
```

## Estructura

```
server/
├── internal/
│   └── gamedata/
│       ├── repository/
│       │   ├── gamedata_repo.go              # Carga desde JSON
│       │   ├── resource_repo.go              # SELECT/INSERT recursos
│       │   ├── production_building_repo.go   # SELECT/INSERT edificios prod
│       │   └── sale_building_repo.go         # SELECT/INSERT edificios venta
│       └── service/
│           └── gamedata_service.go           # Orquestación
├── config/
│   └── gamedata.json                         # Datos maestros (ejemplo)
└── cmd/api/main.go                           # Integración en startup
```

## Componentes Detallados

### 1. `internal/gamedata/repository/gamedata_repo.go`

Lee y valida JSON. Métodos:
- `LoadFromFile()` - Lee JSON, valida estructura
- `validate()` - Validación exhaustiva

### 2. `internal/gamedata/repository/resource_repo.go` ✨ NUEVO

Maneja recursos en BD.
- `GetAll()` - SELECT recursos desde BD
- `InsertBatch()` - INSERT múltiples en transacción
- `Count()` - Cuenta cuántos hay en BD

### 3. `internal/gamedata/repository/production_building_repo.go` ✨ NUEVO

Maneja edificios de producción en BD.
- `GetAll()` - SELECT edificios + procesos + recursos con JOINs
- `InsertBatch()` - INSERT en transacción (edificios, procesos, recursos)
- `Count()` - Cuenta cuántos hay en BD

### 4. `internal/gamedata/repository/sale_building_repo.go` ✨ NUEVO

Maneja edificios de venta en BD.
- `GetAll()` - SELECT edificios + recursos con JOINs
- `InsertBatch()` - INSERT en transacción
- `Count()` - Cuenta cuántos hay en BD

### 5. `internal/gamedata/service/gamedata_service.go`

Orquestación de carga y cache.

**Constructor:**
```go
NewGamedataService(filePath string, db *sql.DB)
```

**Métodos:**

1. **`Load()`** - Asegura que BD tiene datos:
   - Verifica si BD está vacía (Count)
   - Si vacía, importa desde JSON
   - NO sincroniza cache
   - **Responsabilidad: JSON → BD**

2. **`RefreshCache(gameCache *GamedataCache)`** - Sincroniza cache desde BD:
   - Recibe cache como parámetro (pasada desde main.go)
   - Carga recursos desde BD
   - Carga edificios de producción desde BD
   - Carga edificios de venta desde BD
   - Actualiza cache
   - **Responsabilidad: BD → Cache**

3. **`importFromFile()`** - Importación interna desde JSON a BD

## Flujo de Carga al Iniciar

```
main.go
  ↓
1. dbConn, _ := db.InitDatabase(dbPath)
   (BD con schema)
  ↓
2. gamedataCache := cache.NewGamedataCache()
   (Cache vacía creada)
  ↓
3. gamedataSvc := service.NewGamedataService(filePath, dbConn)
   (Servicio sin cache, cache se pasa por parámetro)
  ↓
4. gamedataSvc.Load()
   ├─ Verifica BD.Count()
   ├─ Si 0: Importa JSON a BD
   │  ├─ Carga JSON
   │  ├─ Valida estructura
   │  └─ InsertBatch() (3 transacciones)
   └─ NO sincroniza cache
  ↓
5. gamedataSvc.RefreshCache(gamedataCache)
   ├─ Carga desde BD (SELECT)
   └─ Actualiza cache (pasada como parámetro)
  ↓
Cache listo para handlers
```

## Comportamiento

### Primera ejecución:
```
BD vacía
  ↓
Load() → Importa JSON a BD
  ↓
RefreshCache() → Sincroniza cache desde BD
```

### Ejecuciones posteriores:
```
BD tiene datos
  ↓
Load() → Saltea importación
  ↓
RefreshCache() → Sincroniza cache desde BD
```

### Actualización de datos (endpoint admin futuro):
```
Handler admin edita BD
  ↓
Llama service.RefreshCache()
  ↓
Cache se actualiza sin reiniciar servidor
```

**Funciones principales:**
- `NewGamedataRepository(filePath)` - Constructor
- `LoadFromFile()` - Lee y parsea el JSON
- `validate(gamedata)` - Valida estructura y consistencia

**Validaciones implementadas:**

#### Recursos:
- ✅ master_id no vacío y único
- ✅ id no vacío
- ✅ name no vacío
- ✅ market_price >= 0
- ✅ market_sale_qty > 0

#### Edificios de Producción:
- ✅ master_id no vacío y único
- ✅ id no vacío
- ✅ name no vacío
- ✅ Cada proceso tiene master_id único
- ✅ cycle_time_s > 0
- ✅ window_start_hour/end_hour válidas (0-86400) si existen
- ✅ Cada proceso tiene al menos entrada y salida
- ✅ Todos los resource_id existen en tabla de recursos

#### Edificios de Venta:
- ✅ master_id no vacío y único
- ✅ id no vacío
- ✅ name no vacío
- ✅ Al menos un recurso para vender
- ✅ price_per_unit > 0
- ✅ units_sold_per_second > 0

### 2. `internal/gamedata/service/gamedata_service.go`

**GamedataService** orquesta la carga y almacenamiento en cache.

**Funciones principales:**
- `NewGamedataService(filePath, cache)` - Constructor
- `Load()` - Carga desde repository y almacena en cache
- `IsLoaded()` - Verifica si hay datos en cache
- `GetAllResources()` - Devuelve todos los recursos
- `GetAllProductionBuildings()` - Devuelve todos los edificios de producción
- `GetAllSaleBuildings()` - Devuelve todos los edificios de venta
- `GetResource(masterID)` - Busca recurso por master_id
- `GetProductionBuilding(masterID)` - Busca edificio de producción
- `GetSaleBuilding(masterID)` - Busca edificio de venta

### 3. `config/gamedata.json`

Archivo JSON que define todos los datos maestros del juego.

**Estructura:**
```json
{
  "resources": [...],
  "production_buildings": [...],
  "sale_buildings": [...]
}
```

## Integración en main.go

El servidor inicializa BD y gamedata en el siguiente orden:

```go
// 1. Inicializar BD (crea schema si no existe)
dbConn, err := db.InitDatabase(dbPath)
if err != nil {
    logger.Fatal().Err(err)
}
defer dbConn.Close()

// 2. Crear servicio con BD, archivo JSON y cache
gamedataSvc := service.NewGamedataService(gamedataFilePath, dbConn, gamedataCache)

// 3. Cargar/sincronizar datos
if err := gamedataSvc.Load(); err != nil {
    logger.Fatal().Err(err)
}
```

**Comportamiento:**
1. `Load()` verifica BD
2. Si vacía, importa desde JSON
3. Luego carga datos en cache desde BD

## Variables de Entorno

En `.env`:

```
# Ubicación del archivo de datos maestros
GAMEDATA_FILE=./config/gamedata.json
```

## Formato de gamedata.json

### Recursos

```json
{
  "resources": [
    {
      "id": "res-water",
      "master_id": "res-water",
      "name": "Water",
      "market_price": 10,
      "market_sale_qty": 1
    }
  ]
}
```

### Procesos de Producción

```json
{
  "production_buildings": [
    {
      "id": "pb-cannery",
      "master_id": "pb-cannery",
      "name": "Cannery",
      "construction_cost": 500,
      "construction_time_s": 10,
      "processes": [
        {
          "id": "proc-tomato-sauce",
          "master_id": "proc-tomato-sauce",
          "production_building_id": "pb-cannery",
          "name": "Make Tomato Sauce",
          "cycle_time_s": 3,
          "window_start_hour": null,
          "window_end_hour": null,
          "resources": [
            {
              "resource_id": "res-tomato",
              "is_output": false,
              "quantity": 5
            },
            {
              "resource_id": "res-tomato-sauce",
              "is_output": true,
              "quantity": 3
            }
          ]
        }
      ]
    }
  ]
}
```

### Edificios de Venta

```json
{
  "sale_buildings": [
    {
      "id": "sb-market",
      "master_id": "sb-market",
      "name": "Market Stand",
      "construction_cost": 100,
      "construction_time_s": 5,
      "resources": [
        {
          "resource_id": "res-water",
          "price_per_unit": 5,
          "units_sold_per_second": 2
        }
      ]
    }
  ]
}
```

## Comportamiento del Cache

El cache se sincroniza desde BD al startup:

```go
// Thread-safe read desde cache
resource, exists := gamedataCache.GetResource("res-water")
```

**Para actualizar datos:**

Opción 1: Editar BD y llamar `RefreshCache()` (sin restart):
```go
// En handler admin
if err := gamedataSvc.RefreshCache(); err != nil {
    // error handling
}
```

Opción 2: Importar desde JSON (BD vacía):
```go
// Se hace automáticamente en Load() si BD está vacía
```

Opción 3: Reiniciar servidor:
```bash
# Cambia gamedata.json, reinicia
./yourownboss-api
```

## Logging

### Primera ejecución (BD vacía):
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

### Ejecuciones posteriores (BD con datos):
```
INFO Verificando datos maestros en BD...
INFO Datos maestros encontrados en BD resources_in_db=5
INFO Sincronizando cache desde BD...
INFO Cache sincronizado exitosamente resources=5 production_buildings=2 sale_buildings=2
```

### Si hay errores:
```
ERROR error al cargar datos maestros: error importando datos: error al cargar JSON: no se puede leer archivo gamedata
ERROR Error al cargar datos maestros. El servidor no puede iniciar sin gamedata
```

## Para Desarrollo

### Workflow típico:

1. **Primera ejecución:**
   ```bash
   rm -f yourownboss.db  # Opcional: limpiar BD anterior
   go run ./cmd/api
   # → JSON se importa a BD automáticamente
   ```

2. **Ejecuciones posteriores:**
   ```bash
   go run ./cmd/api
   # → Cache se sincroniza desde BD (sin reimportar JSON)
   ```

3. **Cambiar datos:**
   - Opción A: Editar BD directamente y llamar endpoint refresh
   - Opción B: Editar `gamedata.json`, limpiar BD y reiniciar
   - Opción C: Llamar endpoint admin para importar nuevamente

### Validar gamedata.json

El servicio valida automáticamente en Load(). Si hay errores:

```
ERROR error al cargar datos maestros: error importando datos: gamedata no válido: [detalle]
```

### Limpiar BD y reimportar

```bash
# Windows
del yourownboss.db yourownboss.db-shm yourownboss.db-wal

# Linux/Mac
rm -f yourownboss.db yourownboss.db-shm yourownboss.db-wal

# Luego reiniciar
go run ./cmd/api
```

### Inspeccionar BD

```bash
# Listar tablas
sqlite3 yourownboss.db ".tables"

# Ver recursos
sqlite3 yourownboss.db "SELECT master_id, name FROM resources"

# Ver edificios
sqlite3 yourownboss.db "SELECT master_id, name FROM production_buildings"
```

## Características Implementadas

✅ Repository pattern para carga desde JSON
✅ Repositories para SELECT/INSERT en BD
✅ Validación exhaustiva de estructura JSON
✅ Importación automática JSON → BD (primera ejecución)
✅ Sincronización de cache desde BD
✅ BD como fuente de verdad
✅ Transacciones atómicas para inserciones
✅ Thread-safe cache
✅ Logging completo
✅ Método RefreshCache() para resincronizar sin restart
✅ Soporte para ventanas horarias en procesos
✅ Preparado para endpoints admin de actualización de datos

## Próximas Fases

- **Endpoints Admin de Gamedata**
  - POST /api/v1/admin/gamedata/import - Importar desde JSON
  - POST /api/v1/admin/gamedata/refresh - Resincronizar cache
  - GET /api/v1/admin/gamedata/status - Estado actual

- **Fase 1**: Autenticación y Usuarios
  - Implementar JWT
  - Crear tablas de sesiones
  - Handlers de auth

## Documentación Relacionada

- [Arquitectura Gamedata](./ARCHITECTURE_GAMEDATA.md)
- [Cache Gamedata](../pkg/cache/gamedata.go)
- [Especificación Completa](../../docs/YOUROWNBOSS.md)
  - Implementar JWT
  - Crear tablas de sesiones
  - Handlers de auth

## Documentación

- [Cache Gamedata](../pkg/cache/gamedata.go)
- [Especificación Completa](../../docs/YOUROWNBOSS.md)
- [README Backend](../README.md)
