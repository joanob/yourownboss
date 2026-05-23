# Fase 0.3 - Inicialización de Cache de Gamedata

## Descripción

En esta fase se implementa el sistema de carga y caché de datos maestros (gamedata) del juego. Los datos maestros son inmutables dentro de una sesión del servidor y se cargan al inicio.

## Estructura

```
server/
├── internal/
│   └── gamedata/
│       ├── repository/
│       │   └── gamedata_repo.go    # Carga desde JSON
│       └── service/
│           └── gamedata_service.go  # Orquestación
├── config/
│   └── gamedata.json               # Datos maestros (ejemplo)
└── cmd/api/main.go                 # Integración en startup
```

## Componentes

### 1. `internal/gamedata/repository/gamedata_repo.go`

**GamedataRepository** maneja la lectura y validación del archivo JSON.

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

El servidor carga gamedata al iniciar:

```go
// 1. Crear servicio
gamedataSvc := service.NewGamedataService(gamedataFilePath, gamedataCache)

// 2. Cargar datos
if err := gamedataSvc.Load(); err != nil {
    logger.Fatal().Err(err).Msg("Error al cargar gamedata")
    os.Exit(1)
}
```

**Comportamiento:**
1. Lee ruta de `GAMEDATA_FILE` (default: `./config/gamedata.json`)
2. Lee y parsea JSON
3. Valida estructura completa
4. Almacena en cache thread-safe
5. Loguea cantidad de datos cargados

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

El cache es **read-only** después del startup:

```go
// Thread-safe read
resource, exists := gamedataCache.GetResource("res-water")

// Para modificar gamedata, necesitarías:
// 1. Parar el servidor
// 2. Editar gamedata.json
// 3. Reiniciar el servidor
```

## Logging

El servicio loguea todo el proceso:

```
INFO Cargando datos maestros...
INFO Gamedata cargado desde archivo resources=3 production_buildings=1 sale_buildings=1
INFO Validación de gamedata exitosa
INFO Gamedata cargado exitosamente resources=3 production_buildings=1 sale_buildings=1
```

Si hay errores:

```
ERROR error al cargar gamedata: gamedata no válido: recurso res-water: master_id duplicado
```

## Para Desarrollo

### Validar gamedata.json

El servicio valida automáticamente. Si hay errores, el servidor no inicia:

```
ERROR error al cargar gamedata: gamedata no válido: [detalle del error]
ERROR Error al cargar datos maestros. El servidor no puede iniciar sin gamedata
```

### Editar gamedata.json

1. Editar `config/gamedata.json`
2. Verificar que la estructura es válida
3. Reiniciar el servidor

El servidor revalidará automáticamente.

## Características Implementadas

✅ Repository para cargar datos desde JSON
✅ Validación exhaustiva de estructura y consistencia
✅ Service para orquestación
✅ Integración en main.go
✅ Thread-safe cache
✅ Logging completo
✅ Ejemplo en `config/gamedata.json`
✅ Soporte para ventanas horarias en procesos

## Próximas Fases

- **Fase 1**: Autenticación y Usuarios
  - Implementar JWT
  - Crear tablas de sesiones
  - Handlers de auth

## Documentación

- [Cache Gamedata](../pkg/cache/gamedata.go)
- [Especificación Completa](../../docs/YOUROWNBOSS.md)
- [README Backend](../README.md)
