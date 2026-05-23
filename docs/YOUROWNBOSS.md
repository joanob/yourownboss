# Your Own Boss

## Índice

- [Sobre el juego](#sobre-el-juego)
  - [Dinámica del juego](#dinámica-del-juego)
  - [Monetización del juego](#monetización-del-juego)
  - [Requisitos legales (GDPR)](#requisitos-legales-gdpr)
- [Backend](#backend)
  - [Stack tecnológico](#stack-tecnológico)
  - [Arquitectura y estructura de carpetas](#arquitectura-y-estructura-de-carpetas)
  - [DBO / Model / DTO](#dbo--model--dto)
  - [Inyección de dependencias](#inyección-de-dependencias)
  - [API](#api)
  - [Rate Limiting y Timeouts](#rate-limiting-y-timeouts)
  - [Error Codes](#error-codes)
  - [Base de datos](#base-de-datos)
  - [Cache en Memoria](#cache-en-memoria)
- [Frontend](#frontend)
  - [Stack y decisiones](#stack-y-decisiones)
  - [Estructura de carpetas](#estructura-de-carpetas)
  - [Librerías y tooling](#librerías-y-tooling)
  - [State management y caché](#state-management-y-caché)
  - [API client y tipos](#api-client-y-tipos)
  - [Offline y sincronización](#offline-y-sincronización)
  - [PWA y móvil](#pwa-y-móvil)
- [Esquema de base de datos](#esquema-de-base-de-datos)
- [Endpoints](#endpoints)
  - [Resumen de endpoints](#resumen-de-endpoints)
  - [Game](#game)
  - [Users](#users)
  - [Company](#company)
  - [Market](#market)
  - [Production](#production)
  - [Sale](#sale)
- [Plan](#plan)
  - [Patrones Arquitectónicos (Aplicar a todas las fases)](#patrones-arquitectónicos-aplicar-a-todas-las-fases)
  - [Fase 0: Inicialización (Infraestructura Base)](#fase-0-inicialización-infraestructura-base)
  - [Fase 1: Backend - Autenticación y Usuarios](#fase-1-backend---autenticación-y-usuarios)
  - [Fase 2: Backend - Empresas y Gestión de Dinero](#fase-2-backend---empresas-y-gestión-de-dinero)
  - [Fase 3: Backend - Sistema de Recursos Maestros](#fase-3-backend---sistema-de-recursos-maestros)
  - [Fase 4: Backend - Mercado (Compra/Venta)](#fase-4-backend---mercado-compraventa)
  - [Fase 5: Backend - Edificios de Producción](#fase-5-backend---edificios-de-producción)
  - [Fase 6: Backend - Edificios de Venta](#fase-6-backend---edificios-de-venta)
  - [Fase 7: Backend - Optimizaciones y Tests Finales](#fase-7-backend---optimizaciones-y-tests-finales)


## Sobre el juego

Your Own Boss es un juego web y móvil de gestión y producción. El juego consiste en producir recursos a partir de otros recursos, venderlos y con el dinero comprar nuevos edificios de producción y venta para ganar más dinero. El foco de diseño es un juego casual de un jugador que no requiere estar conectado mucho tiempo con un sistema para aprovechar al máximo el tiempo que el usuario no podrá estar conectado.

### Dinámica del juego

**Usuarios y empresas**
- El jugador registra un usuario y crea una empresa. Un usuario solo puede tener una empresa.
- El jugador debe seleccionar un huso horario y si desea modificar el huso horario no podrá modificarlo otra vez hasta pasados 30 días.
- Al crear la empresa empieza con una cantidad de dinero inicial.

**Recursos**
- Los recursos se compran en el mercado. Este mercado tiene precios estáticos, es ilimitado y no depende de otros jugadores. 
- Los recursos se venden por lotes y solo se puede vender múltiplos de esa cantidad de lote. Por ejemplo, si 3 unidades de agua se venden por 1 moneda, no es posible tener monedas decimales así que siempre se deben comprar y vender múltiplos de 3 unidades de agua.
- El inventario almacena los recursos de la empresa. Es ilimitado y los recursos no tienen caducidad. Al crear la empresa el inventario está vacío
- Cuando se crea una empresa, el inventario está vacío. Cuando el usuario compra un recurso por primera vez, se crea un registro. Si después vende todo (quantity = 0), el registro se mantiene con quantity = 0.

**Producción**
- Existen distintos edificios de producción. Cada edificio de producción tiene unos procesos productivos.
- Los procesos de producción determinan qué recursos se necesitan para producir otros recursos y cuál es el tiempo de producción. El proceso productivo se estructura en ciclos. Por ejemplo, un proceso de producción puede tener un ciclo que dura 3 segundos en el que se utilizan 5 tomates para fabricar 3 botes de tomates en conserva.
- Una empresa puede tener varias instancias del mismo edificio de producción y en cada uno estar produciendo o vendiendo unos recursos.
- El usuario inicia manualmente el proceso productivo indicando la cantidad de ciclos que desea producir. En ese momento, los recursos utilizados se extraen del inventario del jugador y se calcula el momento en el que habrá acabado la producción utilizando el número de ciclos y el tiempo que dura cada ciclo.
- Si en el momento de iniciar la producción el usuario no dispone de las cantidades necesarias de recursos, no empezará la producción.
- Un edificio que está produciendo no puede utilizarse para nada más hasta que la producción termine. No se puede interrumpir la producción.
- Cuando se acabe el tiempo de producción, el usuario podrá obtener los recursos producidos. En ese momento se añadirán a su inventario y el edificio de producción estará disponible de nuevo para iniciar otro proceso productivo.
- La mayoría de procesos se pueden iniciar y finalizar a cualquier hora y la única restricción es que la empresa tenga la cantidad de recursos necesaria. Sin embargo, hay algunos procesos que tienen ventanas horarias y no pueden iniciarse o acabar fuera de esas horas. . Por ejemplo, si un proceso solo está disponible de 8 a 20 UTC, no se podrá iniciar antes de las 8 UTC y no podrá terminar más tarde de las 20 UTC, independientemente del timezone del usuario. Esta ventana siempre será en el mismo día, nunca empezará un día y se acabará al día siguiente.
- La validación de que la producción está dentro de la ventana del usuario se hará utilizando la hora del timezone del usuario. El usuario puede cambiar su timezone como máximo cada 30 días, pero esto solo afecta a futuros procesos (no invalida ni modifica procesos en curso). Si un proceso ya está programado para terminar a las 16:53 UTC, seguirá terminando a esa hora aunque el usuario cambie su timezone después.
- Cuando el edificio está parado se puede subir de nivel. Los niveles actúan como multiplicadores: un edificio de nivel N produce y consume exactamente N veces la cantidad base por ciclo. Es equivalente a tener N fábricas iguales funcionando en paralelo. Se pueden subir varios niveles a la vez si se tiene el dinero. El coste por nivel es lineal: si subir de nivel 1 a 2 cuesta 100 monedas, subir de nivel 4 a 5 también cuesta 100 monedas. Por lo tanto, subir 2 niveles (ej. de 1 a 3) cuesta 200 monedas. El tiempo de mejora es igual al tiempo de construccion, independientemente de cuantos niveles se suban.

**Venta**
- Los edificios de venta tienen las mismas características que los edificios de producción, pero en lugar de producir recursos en procesos productivos los venden.
- Los edificios de venta tienen una lista de recursos que pueden vender, el ritmo al que los venden y el precio por unidad o lote.
- Los edificios de venta no tienen ventanas de venta, pueden funcionar a cualquier hora.
- Para iniciar una venta, el usuario elige el número de unidades o lotes que se van a vender. Si en el inventario existen suficientes recursos, se toman del inventario y se inicia la venta.
- Cuando la venta termine, el usuario puede obtener el dinero conseguido y se sumará al balance de su empresa.

### Monetización del juego

**Fase inicial**: sin monetización (juego es free-to-play sin anuncios).

**Monetización futura**: cuando el juego esté terminado y con base de usuarios grande, posibles anuncios no intrusivos o publicidad de marcas (ej. producir Fanta en lugar de "refresco genérico").

### Requisitos legales (GDPR)

El proyecto debe cumplir GDPR desde el inicio: consentimiento para trackers, borrado de datos a petición y políticas de privacidad documentadas.


## Backend

Es prioritario que el backend sea lo más rápido posible. Entiendo que SQLite no es la mejor opción pero no quiero tener PostgreSQL que seguramente consuma más recursos de lo que lo hará Go con SQLite.

### Stack tecnológico

- Lenguaje: Go.
- Router ligero: `chi` (incluye middleware de rate limiting y CORS).
- Persistencia: SQLite usando `modernc.org/sqlite` para evitar cgo.
- Consultas generadas: `sqlc` para mantener consultas tipadas y seguras.
- Validación de input: `go-playground/validator` en los handlers.
- Hash de contraseñas: `argon2id` (más resistente a ataques GPU que bcrypt en 2026).
- Logger: `rs/zerolog`. Se guardarán archivos de log rotados en carpeta `./logs/` con máximo 100 archivos de 10MB cada uno (1GB total). La compresión de archivos rotados se decide según mejor rendimiento.
- Manejo de configuraciones: variables de entorno usando `godotenv` (cargará `.env` al iniciar).

Razonamiento: esta pila mantiene el binario puro (sin CGO), consultas claras y testables.

Se plantea tener Redis o alguna herramienta propia que reduzca las lecturas de base de datos para mejorar el rendimiento.

### Arquitectura y estructura de carpetas

Organizar `internal/` por dominios (cada dominio es un paquete independiente) y dentro de cada dominio mantener las subcarpetas por responsabilidad: `http`, `service`, `repository`, `models` y `sql`. La carpeta `sql` tendrá las queries generadas por sqlc.

```
server/
	cmd/api/
	internal/
		users/
			http/          # handlers, rutas, validaciones de request
			service/       # lógica de negocio, orquestación
			repository/    # sqlc queries y adaptadores DB
			models/        # modelos/domain types y validaciones
			sql/           # .sql y generated code by sqlc
		resources/
			http/
			service/
			repository/
			models/
		production/
			http/
			service/
			repository/
			models/
		auth/            # middleware, jwt helpers, shared auth types
		db/              # migrations, schema.sql, conexión DB global
		pkg/             # utilidades compartidas (evitar lógica de dominio aquí)
```

Buenas prácticas:

- Dependencias acíclicas: mantener el flujo `http -> service -> repository`. Nunca importar `http` desde `service` ni `repository` desde `http`.
- `models` dentro del dominio contienen tipos de dominio puros; los DTOs para la API pueden vivir en `http` o en `models/dto`.
- `sqlc` por dominio
- Tipos compartidos (errores, utilidades) en `internal/pkg` o `internal/shared`.
- Wiring en `cmd/api/main.go`:
- Tests: mockear interfaces de `repository` para probar `service` e mockear `service` en `http`.

### DBO / Model / DTO

- **DBO** (DB Object): estructuras exactamente mapeadas a tablas/columnas (sqlc). No exponer directamente a la API.
- **Model**: estructuras internas enriquecidas con reglas de negocio (p. ej. `Building` con su lista de `Process`).
- **DTO** (Data Transfer Object): estructuras para la API pública (request/response), solo lo necesario para el cliente.

Flujo: `DB <-> DBO (persistencia) → Model (reglas) → DTO (respuesta API)`

Ejemplo: `BuildingDBO` (tabla), `Building` (modelo con procesos cargados), `BuildingDTO` (id, nombre, procesos: []ProcessDTO).

### Inyección de dependencias

Constructor injection con interfaces:

```go
// Interfaz en la capa de servicio
type UserRepo interface {
    GetByID(ctx context.Context, id int64) (*UserDBO, error)
}

// Wiring en main.go
repo    := repository.NewUserRepo(db)
svc     := service.NewUserService(repo)
handler := http.NewUserHandler(svc)
```

Esto permite testear `service` inyectando un mock de `UserRepo`. No se recomienda DI container (fx, wire) hasta que haya necesidad real.

### API

Autenticación: sesión JWT en cookie httpOnly de corta duración (1 minuto) + refresh token en segunda cookie httpOnly.

- El session token dura el tiempo configurado en ENV `JWT_SESSION_EXPIRY` (default: 60 segundos). Contiene `user_id`, `session_id`, `company_id`. En cada petición autenticada el middleware lo verifica sin acceder a BD (JWT se valida localmente).
- Si el session token caduca, el middleware utiliza el refresh token. Este token contiene `user_id`, `session_id`, `verification_string`.
- El servidor valida que la sesión con ese `session_id` sigue activa verificando el `verification_string` en cache (rápido) o BD (fallback). Si es válida, emite un nuevo session token y ejecuta la acción original.
- El refresh token tiene duración configurada en ENV `JWT_REFRESH_EXPIRY` (default: 25920000 segundos = 300 días). Al logout o si el usuario revoca la sesión manualmente, se marca `revoked_at` en la tabla `user_sessions` y se invalida en cache.
- Si se detecta que el refresh token es inválido (revocado), el usuario debe hacer login nuevamente.
- El servidor usa códigos HTTP semánticos para indicar el resultado (200, 201, 400, 401, 403, 404, 409…). Cuando hay error, el body incluye el detalle:

```json
{
    "data": "...",
    "error": {
        "code": "string",
        "message": "string"
    }
}
```

CORS configurado en el middleware de `chi` para aceptar peticiones del dominio del frontend (proyectos completamente separados).

### Rate Limiting y Timeouts

**Rate Limiting** (solo en caso de fallos):
- **Login fallidos**: máximo 5 intentos fallidos por usuario cada 15 minutos. Después bloquear durante 15 min.
- **Market buy/sell**: máximo 20 requests exitosos por minuto por usuario.
- **Production start**: máximo 20 requests exitosos por minuto por usuario.
- El rate limiting cuenta solo requests fallidos/exitosos, no reintento de validación simple.

**Request Timeout**: Todas las requests tienen timeout de 30 segundos.

### Error Codes

Los errores devuelven un objeto con `code` y `message`. Códigos estándar:

**Errores 400 (Bad Request)**
- `INVALID_INPUT` - Input no pasa validación
- `INVALID_QUANTITY_MULTIPLE` - Cantidad no es múltiplo válido de venta
- `INVALID_TIMEZONE` - Timezone no es válido
- `TIMEZONE_RECENTLY_CHANGED` - No puede cambiar timezone (menos de 30 días)
- `PROCESS_NOT_AVAILABLE_NOW` - Proceso tiene ventana horaria y no está disponible ahora

**Errores 401 (Unauthorized)**
- `UNAUTHORIZED` - No tiene sesión válida
- `INVALID_CREDENTIALS` - Username/password incorrectos
- `SESSION_EXPIRED` - Sesión ha sido revocada

**Errores 403 (Forbidden)**
- `INSUFFICIENT_PERMISSIONS` - Usuario no es admin
- `COMPANY_NOT_FOUND` - Usuario no tiene empresa

**Errores 404 (Not Found)**
- `USER_NOT_FOUND` - Usuario no existe
- `RESOURCE_NOT_FOUND` - Recurso no existe
- `BUILDING_NOT_FOUND` - Edificio no existe
- `COMPANY_NOT_FOUND` - Empresa no existe
- `PROCESS_NOT_FOUND` - Proceso no existe

**Errores 409 (Conflict)**
- `INSUFFICIENT_FUNDS` - Dinero insuficiente para la operación
- `INSUFFICIENT_INVENTORY` - Inventario insuficiente para la operación
- `BUILDING_NOT_IDLE` - Edificio está en construcción o producción
- `BUILDING_NOT_PRODUCING` - Edificio no tiene producción activa
- `COMPANY_ALREADY_EXISTS` - Usuario ya tiene empresa
- `USERNAME_ALREADY_EXISTS` - Username ya está en uso
- `EMAIL_ALREADY_EXISTS` - Email ya está en uso

### Base de datos

- `modernc.org/sqlite` (sin CGO, fácil despliegue).
- `sqlc` para generar tipos y queries.
- **Migraciones**: `schema.sql` es la única fuente de verdad. Las migraciones son manuales si en el futuro hubiera cambios en el esquema. En Fase 0.2 se ejecutará `schema.sql` una sola vez para inicializar la BD.
- Backups automáticos nocturnos (dump de SQLite a storage).
- Capa de abstracción DB para facilitar futura migración a PostgreSQL si es necesario.

### Cache en Memoria

**Gamedata (Recursos, Edificios, Procesos)**:
- **REQUERIDO**: Cargar datos maestros en memoria al iniciar la app (tabla `resources`, `production_buildings`, `sale_buildings`, `production_processes`, etc.). El servidor NO inicia si no puede cargar los datos maestros.
- **Validación**: Al cargar el archivo JSON, validar que:
  - La estructura tiene todos los campos requeridos
  - No hay duplicados (master_id únicos)
  - No hay límite de cantidad de recursos/edificios
- **Warm-up**: El servidor debe esperar a que el cache de gamedata esté completamente cargado antes de aceptar requests. Los health checks también esperarán.
- **Estructura**: Tres maps separados (NO interface{}):
  - `resources`: `map[string]Resource` - misma estructura que endpoint GET /api/v1/resources
  - `productionBuildings`: `map[string]ProductionBuilding` - misma estructura que endpoint GET /api/v1/production/buildings
  - `saleBuildings`: `map[string]SaleBuilding` - misma estructura que endpoint GET /api/v1/sale/buildings
  - Cada map protegido con mutex para concurrencia segura
- Estrategia: lookup en memoria primero, si no existe caer a BD como fallback
- Invalidación: solo manual (admin importa nuevos datos) o al reiniciar la app
- **Precios siempre actuales**: Cuando se compra/vende, se usa el precio actual del cache (aunque haya cambiado desde que se inició el proceso)
- Ventajas: evita select en cada operación (compra, producción), datos siempre consistentes, bajo costo de memoria

**JWT de Sesión y Refresco**:
- El sistema emite dos JWT diferentes:
  - **Session Token**: contiene `user_id`, `session_id`, `company_id` (nullable, null si usuario no tiene empresa). Duración 1 minuto. Se valida sin acceso a BD (validación local por firma JWT). En cada request, se valida la firma y expiry.
  - **Refresh Token**: contiene `user_id`, `session_id`, `verification_string`. Duración 300 días. Se almacena como hash en BD.
- **Flujo de renovación**: Cuando el session token expira (401), el middleware utiliza el refresh token para validar que la sesión sigue activa. Busca la sesión en cache por `session_id + verification_string`. Si está válida, emite un nuevo session token y ejecuta la acción original automáticamente (transparente al cliente).
- **Estructura de sesión en cache**: `map[sessionID]SessionData` con `verification_string`, `user_id`, `company_id`, `expires_at`, `revoked_at`
- **Validación**:
  - Para session token: solo verificar firma JWT (sin BD)
  - Para refresh token: buscar en cache primero por `session_id + verification_string`, si no existe buscar en BD, verificar que no está revocado
- **Revocación**: Al logout o revoke manual, marcar `revoked_at` en BD e invalidar en cache. El próximo request con ese refresh token fallará.
- **Limpieza automática**: LRU cleaner ejecuta cada 6 horas para eliminar sesiones expiradas de cache
- **Company ID nullable**: Si usuario borra su empresa, `company_id` se pone a null en el session token (sin necesidad de revocar sesiones). El usuario puede volver a crear empresa después.
- Ventajas: evita queries a BD en cada request (session token valido localmente), pero mantiene auditoría en BD y capacidad de revocación

## Esquema de base de datos

- Las cantidades de recursos y precios usan valores enteros para evitar fallos por operaciones con coma flotante
- Los ids son UUIDs para que se puedan generar desde el cliente o desde el servidor
- Los campos `master_id` son identificadores textuales únicos para datos maestros importables.

```sql
-- DATOS MAESTROS

CREATE TABLE resources (
  id              TEXT PRIMARY KEY,
  master_id       TEXT    NOT NULL UNIQUE,
  name            TEXT    NOT NULL,
  market_price    INTEGER NOT NULL,
  market_sale_qty INTEGER NOT NULL
);

CREATE TABLE production_buildings (
  id                  TEXT PRIMARY KEY,
  master_id           TEXT    NOT NULL UNIQUE,
  name                TEXT    NOT NULL,
  construction_cost   INTEGER NOT NULL,
  construction_time_s INTEGER NOT NULL
);

-- Both window_start_hour and window_end_hour must be null or have value
-- window_start_hour y window_end_hour son segundos desde medianoche UTC (0-86400)
-- Ejemplo: 8 AM UTC = 28800, 20 PM UTC = 72000
-- Las ventanas horarias son SIEMPRE en UTC, independiente del timezone del usuario
-- El usuario puede iniciar un proceso si ahora está dentro de la ventana (en UTC)
-- Si un proceso ya está en curso, NO se invalida aunque el usuario cambie timezone después
CREATE TABLE production_processes (
  id                     TEXT PRIMARY KEY,
  master_id              TEXT    NOT NULL UNIQUE,
  production_building_id TEXT    NOT NULL REFERENCES production_buildings(id),
  name                   TEXT    NOT NULL,
  cycle_time_s           INTEGER NOT NULL,
  window_start_hour      INTEGER,
  window_end_hour        INTEGER
);

CREATE TABLE production_process_resources (
  process_id  TEXT    NOT NULL REFERENCES production_processes(id),
  resource_id TEXT    NOT NULL REFERENCES resources(id),
  is_output   BOOLEAN NOT NULL,
  quantity    INTEGER NOT NULL,
  PRIMARY KEY (process_id, resource_id, is_output)
);

CREATE TABLE sale_buildings (
  id                  TEXT PRIMARY KEY,
  master_id           TEXT    NOT NULL UNIQUE,
  name                TEXT    NOT NULL,
  construction_cost   INTEGER NOT NULL,
  construction_time_s INTEGER NOT NULL
);

CREATE TABLE sale_resources (
  sale_building_id        TEXT    NOT NULL REFERENCES sale_buildings(id),
  resource_id             TEXT    NOT NULL REFERENCES resources(id),
  price_per_unit          INTEGER NOT NULL,
  units_sold_per_second   INTEGER NOT NULL,
  PRIMARY KEY (sale_building_id, resource_id)
);

-- USUARIOS Y AUTENTICACIÓN

-- roles: null (player) | 'A' (admin)
CREATE TABLE users (
  id                              TEXT     PRIMARY KEY,
  username                        TEXT     NOT NULL UNIQUE,
  email                           TEXT     NOT NULL UNIQUE,
  password_hash                   TEXT     NOT NULL,
  role                            TEXT     NOT NULL DEFAULT 'P',
  timezone                        TEXT,
  last_timezone_modification_at   DATETIME,
  created_at                      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted                      INTEGER  NOT NULL DEFAULT 0,
  deleted_at                      DATETIME
);

-- Tabla que almacena las sesiones activas de usuarios.
-- El token se almacena como hash (nunca en claro).
-- revoked_at != NULL → sesión invalidada (logout o revocación manual).
-- session_id es un UUID que identifica única sesión.
CREATE TABLE user_sessions (
  id                  TEXT     PRIMARY KEY,
  user_id             TEXT     NOT NULL REFERENCES users(id),
  session_id          TEXT     NOT NULL UNIQUE,
  verification_string TEXT     NOT NULL,
  token_hash          TEXT     NOT NULL UNIQUE,
  expires_at          DATETIME NOT NULL,
  revoked_at          DATETIME,
  created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted          INTEGER  NOT NULL DEFAULT 0,
  deleted_at          DATETIME
);

-- EMPRESAS

CREATE TABLE companies (
  id         TEXT     PRIMARY KEY,
  user_id    TEXT     NOT NULL UNIQUE REFERENCES users(id),
  name       TEXT     NOT NULL,
  money      INTEGER  NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted INTEGER  NOT NULL DEFAULT 0,
  deleted_at DATETIME
);

CREATE TABLE company_inventory (
  id          TEXT    PRIMARY KEY,
  company_id  TEXT    NOT NULL REFERENCES companies(id),
  resource_id TEXT    NOT NULL REFERENCES resources(id),
  quantity    INTEGER NOT NULL DEFAULT 0,
  is_deleted  INTEGER NOT NULL DEFAULT 0,
  deleted_at  DATETIME,
  UNIQUE (company_id, resource_id)
);

-- PRODUCCIÓN

-- Estados del edificio (inferidos por construction_ends_at y production_runs):
-- - Construcción: construction_ends_at > NOW()
-- - Idle: construction_ends_at <= NOW() AND no hay production_runs con is_collected=0
-- - Produciendo: hay production_runs con is_collected=0
CREATE TABLE company_production_buildings (
  id                     TEXT     PRIMARY KEY,
  company_id             TEXT     NOT NULL REFERENCES companies(id),
  production_building_id TEXT     NOT NULL REFERENCES production_buildings(id),
  level                  INTEGER  NOT NULL DEFAULT 1,
  construction_ends_at   DATETIME,
  created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted             INTEGER  NOT NULL DEFAULT 0,
  deleted_at             DATETIME
);

-- CONSTRAINT: Solo puede haber 1 production_run activo (is_collected=0) por company_building_id
CREATE TABLE production_runs (
  id                 TEXT     PRIMARY KEY,
  company_building_id TEXT    NOT NULL REFERENCES company_production_buildings(id),
  process_id         TEXT     NOT NULL REFERENCES production_processes(id),
  production_cycles  INTEGER  NOT NULL,
  started_at         DATETIME NOT NULL,
  ends_at            DATETIME NOT NULL,
  is_collected       INTEGER  NOT NULL DEFAULT 0,
  collected_at       DATETIME,
  is_deleted         INTEGER  NOT NULL DEFAULT 0,
  deleted_at         DATETIME,
  UNIQUE (company_building_id, is_collected) WHERE is_collected = 0
);

-- VENTA

-- Estados del edificio (inferidos por construction_ends_at y sale_runs):
-- - Construcción: construction_ends_at > NOW()
-- - Idle: construction_ends_at <= NOW() AND no hay sale_runs con is_collected=0
-- - Vendiendo: hay sale_runs con is_collected=0
CREATE TABLE company_sale_buildings (
  id                 TEXT     PRIMARY KEY,
  company_id         TEXT     NOT NULL REFERENCES companies(id),
  sale_building_id   TEXT     NOT NULL REFERENCES sale_buildings(id),
  level              INTEGER  NOT NULL DEFAULT 1,
  construction_ends_at DATETIME,
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted         INTEGER  NOT NULL DEFAULT 0,
  deleted_at         DATETIME
);

-- CONSTRAINT: Solo puede haber 1 sale_run activo (is_collected=0) por company_sale_building_id
CREATE TABLE sale_runs (
  id                        TEXT     PRIMARY KEY,
  company_sale_building_id  TEXT     NOT NULL REFERENCES company_sale_buildings(id),
  resource_id               TEXT     NOT NULL REFERENCES resources(id),
  units_to_sell             INTEGER  NOT NULL,
  started_at                DATETIME NOT NULL,
  ends_at                   DATETIME NOT NULL,
  is_collected              INTEGER  NOT NULL DEFAULT 0,
  collected_at              DATETIME,
  is_deleted                INTEGER  NOT NULL DEFAULT 0,
  deleted_at                DATETIME,
  UNIQUE (company_sale_building_id, is_collected) WHERE is_collected = 0
);

-- RATE LIMITING

-- Tabla para track de intentos fallidos de login (rate limiting)
-- Se usa para implementar: máximo 5 intentos fallidos por usuario cada 15 minutos
CREATE TABLE login_attempts (
  id             TEXT     PRIMARY KEY,
  user_id        TEXT,
  username       TEXT     NOT NULL,
  failed_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted     INTEGER  NOT NULL DEFAULT 0
);

CREATE INDEX idx_login_attempts_username ON login_attempts(username, failed_at);

-- AUDITORÍA

-- AUDITORÍA Y LOGGING
-- Tabla de auditoría para registrar cambios de negocio (dinero, inventario, timezone, etc.)
-- Diferente de logs de zerolog: estos registros son queryables y para compliance GDPR.
-- Notas sobre referencias:
--   - user_id: referencia a usuario. Aunque sea soft-delete, los registros de auditoría persisten.
--   - company_id: referencia a empresa. Similar a user_id.
-- Opciones para ON DELETE:
--   Opción profesional: usar ON DELETE SET NULL para que audit_log persista sin referencia a usuarios borrados.
--   Por ahora: SET NULL aplicado a ambas referencias.

CREATE TABLE audit_log (
  id             TEXT     PRIMARY KEY,
  user_id        TEXT     REFERENCES users(id) ON DELETE SET NULL,
  company_id     TEXT     REFERENCES companies(id) ON DELETE SET NULL,
  action         TEXT     NOT NULL,  -- ej: 'BUY_RESOURCE', 'START_PRODUCTION', 'COLLECT_PRODUCTION', 'CHANGE_TIMEZONE', etc.
  resource_type  TEXT,                -- ej: 'company', 'production_run', 'inventory', 'user', etc.
  resource_id    TEXT,                -- ID del recurso modificado
  changes        TEXT,                -- JSON con before/after de campos modificados
  timestamp      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ÍNDICES PARA OPTIMIZACIÓN DE QUERIES
-- Búsquedas frecuentes y relaciones
CREATE INDEX idx_companies_user_id ON companies(user_id);
CREATE INDEX idx_production_buildings_company ON company_production_buildings(company_id);
CREATE INDEX idx_sale_buildings_company ON company_sale_buildings(company_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id, revoked_at);
CREATE INDEX idx_audit_log_company ON audit_log(company_id, timestamp);
CREATE INDEX idx_audit_log_user ON audit_log(user_id, timestamp);

-- Triggers to convert DELETE into soft-delete

CREATE TRIGGER users_before_delete
BEFORE DELETE ON users
FOR EACH ROW
BEGIN
  UPDATE users
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  -- Propagate soft-delete to direct and indirect child rows
  UPDATE user_sessions
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE user_id = OLD.id;

  UPDATE companies
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE user_id = OLD.id;

  UPDATE company_inventory
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);

  UPDATE company_production_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);

  UPDATE company_sale_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);

  UPDATE production_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_building_id IN (
    SELECT id FROM company_production_buildings
    WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id)
  );

  UPDATE sale_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_sale_building_id IN (
    SELECT id FROM company_sale_buildings
    WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id)
  );

  SELECT RAISE(IGNORE);
END;

CREATE TRIGGER companies_before_delete
BEFORE DELETE ON companies
FOR EACH ROW
BEGIN
  UPDATE companies
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  -- Propagate soft-delete to direct and indirect child tables
  UPDATE company_inventory
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id = OLD.id;

  UPDATE company_production_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id = OLD.id;

  UPDATE company_sale_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_id = OLD.id;

  UPDATE production_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_building_id IN (
    SELECT id FROM company_production_buildings WHERE company_id = OLD.id
  );

  UPDATE sale_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_sale_building_id IN (
    SELECT id FROM company_sale_buildings WHERE company_id = OLD.id
  );

  SELECT RAISE(IGNORE);
END;

CREATE TRIGGER company_production_buildings_before_delete
BEFORE DELETE ON company_production_buildings
FOR EACH ROW
BEGIN
  UPDATE company_production_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  -- Propagate soft-delete to production runs belonging to this company building
  UPDATE production_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_building_id = OLD.id;

  SELECT RAISE(IGNORE);
END;

CREATE TRIGGER company_sale_buildings_before_delete
BEFORE DELETE ON company_sale_buildings
FOR EACH ROW
BEGIN
  UPDATE company_sale_buildings
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  -- Propagate soft-delete to sale runs belonging to this company sale building
  UPDATE sale_runs
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE company_sale_building_id = OLD.id;

  SELECT RAISE(IGNORE);
END;
```


## Endpoints

### Resumen de endpoints

_Nota: No hay paginación en ningún endpoint GET. Todas las listas devuelven todos los registros._

**Game**
- GET /api/v1/status — Endpoint de salud y estado básico del servicio.
- GET /api/v1/gamedata — Devuelve datos maestros necesarios al cliente (resources, buildings).
- POST /api/v1/gamedata - Cargar json con los datos maestros. Solo disponible para admins.
- GET /api/v1/resources — Lista de recursos maestros disponibles en el juego.
- GET /api/v1/resources/:resource_id — Detalle de un recurso maestro.
- GET /api/v1/production/buildings — Tipos de edificios de producción con sus procesos productivos y sus recursos.
- GET /api/v1/production/buildings/:building_id — Detalle de un edificio de producción con sus procesos productivos y sus recursos.
- GET /api/v1/sale/buildings — Tipos de edificios de venta disponibles y los recursos que venden.
- GET /api/v1/sale/buildings/:building_id — Detalle de un edificio de venta y los recursos que vende.

**Users**
- POST /api/v1/auth/register — Registra un nuevo usuario. NO crea la empresa automáticamente (se crea con POST /api/v1/company después).
- POST /api/v1/auth/login — Autentica al usuario y emite cookies de sesión/refresh.
- POST /api/v1/auth/logout — Revoca el refresh token y elimina cookies del cliente.
- GET /api/v1/users/me — Obtiene el perfil del usuario autenticado.
- PUT /api/v1/users/me — Actualiza datos del usuario (perfil, timezone, preferencias).

**Company**
- GET /api/v1/company — Obtiene datos de la empresa del usuario.
- POST /api/v1/company — Crea una nueva empresa para el usuario.
- PUT /api/v1/company — Actualiza datos de la empresa del usuario (nombre, ajustes).
- DELETE /api/v1/company — Elimina (soft-delete) la empresa.
- GET /api/v1/company/inventory — Lista inventario de recursos de la empresa.

**Market**
- POST /api/v1/market/buy — Comprar recursos al mercado (ajusta inventario y dinero).
- POST /api/v1/market/sell — Vender recursos al mercado (ajusta inventario y dinero).

**Production**
- GET /api/v1/company/production/buildings — Instancias de producción de la empresa.
- POST /api/v1/company/production/buildings — Construir una nueva instancia.
- POST /api/v1/company/production/buildings/:id/upgrade — Subir nivel del edificio.
- POST /api/v1/company/production/buildings/:id/start — Iniciar un proceso productivo.
- POST /api/v1/company/production/buildings/:id/collect — Recoger productos al finalizar.

**Sale**
- GET /api/v1/company/sale/buildings — Instancias de venta de la empresa.
- POST /api/v1/company/sale/buildings — Construir nueva instancia de venta.
- POST /api/v1/company/sale/buildings/:id/upgrade — Subir nivel del edificio de venta.
- POST /api/v1/company/sale/buildings/:id/start — Iniciar un proceso de venta.
- POST /api/v1/company/sale/buildings/:id/collect — Cobrar ingresos al finalizar la venta.

### Game

**GET /api/v1/status**

Informa del estado del servidor, su versión y el estado de la conexión a base de datos.
No incluye estado del cache de gamedata (se asume que si el servidor está arriba, el cache está correcto).

Response 200 OK
```json
{
  "data": {
    "status": "ok",
    "version": "0.1.0",
    "db_connected": true
  }
}
```

**GET /api/v1/gamedata**

Devuelve todos los datos del juego: recursos, edificios de producción con sus procesos y recursos y edificios de venta con sus recursos.

Response 200 OK
```json
{
  "data": {
    "resources": [
      {
        "id": "res-water",
        "master_id": "water",
        "name": "Water",
        "market_price": 10,
        "market_sale_qty": 3
      }
    ],
    "production_buildings": [
      {
        "id": "pb-1",
        "master_id": "pb_basic",
        "name": "Factory",
        "construction_cost": 100,
        "construction_time_s": 120,
        "processes": [
          {
            "id": "proc-1",
            "master_id": "proc_tomato_can",
            "name": "Tomato Can",
            "cycle_time_s": 3,
            "inputs": [
              {
                "resource_id": "res-tomato",
                "quantity": 5
              }
            ],
            "outputs": [
              {
                "resource_id": "res-tomato_can",
                "quantity": 3
              }
            ]
          }
        ]
      }
    ],
    "sale_buildings": [
      {
        "id": "sb-1",
        "master_id": "sb_shop",
        "name": "Shop",
        "construction_cost": 50,
        "construction_time_s": 120,
        "resources": [
          {
            "resource_id": "res-tomato",
            "units_sold_per_second": 1,
            "price_per_unit": 15
          }
        ]
      }
    ]
  }
}
```

**POST /api/v1/gamedata** (admin)

Carga los datos del juego al servidor: recursos, edificios de producción con sus procesos y recursos y edificios de venta con sus recursos.

Request
```json
{
  "resources": [
    {
      "id": "res-water",
      "master_id": "water",
      "name": "Water",
      "market_price": 10,
      "market_sale_qty": 3
    }
  ],
  "production_buildings": [
    {
      "id": "pb-1",
      "master_id": "pb_basic",
      "name": "Factory",
      "construction_cost": 100,
      "construction_time_s": 120,
      "processes": [
        {
          "id": "proc-1",
          "master_id": "proc_tomato_can",
          "name": "Tomato Can",
          "cycle_time_s": 3,
          "window_start_hour": 5000,
          "window_end_hour": 6000,
          "inputs": [
            {
              "resource_id": "res-tomato",
              "quantity": 5
            }
          ],
          "outputs": [
            {
              "resource_id": "res-tomato_can",
              "quantity": 3
            }
          ]
        }
      ]
    }
  ],
  "sale_buildings": [
    {
      "id": "sb-1",
      "master_id": "sb_shop",
      "name": "Shop",
      "construction_cost": 50,
      "construction_time_s": 120,
      "resources": [
        {
          "resource_id": "res-tomato",
          "units_sold_per_second": 1,
          "price_per_unit": 15
        }
      ]
    }
  ]
}
```

Response 201 Created
```json
{
  "data": {
    "imported": true,
    "counts": {
      "resources": 10,
      "buildings": 5,
      "processes": 12
    }
  }
}
```

**GET /api/v1/resources**

Devuelve la lista de recursos.

Response 200 OK
```json
{
  "data": [
    {
      "id": "res-water",
      "master_id": "water",
      "name": "Water",
      "market_price": 10,
      "market_sale_qty": 3
    }
  ]
}
```

**GET /api/v1/production/buildings**

Devuelve la lista de edificios de producción con sus procesos y recursos.

Response 200 OK
```json
{
  "data": [
    {
      "id": "pb-1",
      "master_id": "pb_basic",
      "name": "Factory",
      "construction_cost": 100,
      "construction_time_s": 600,
      "processes": [
        {
          "id": "proc-1",
          "name": "Tomato Can",
          "cycle_time_s": 3,
          "window_start_hour": 5000,
          "window_end_hour": 6000,
          "inputs": [
            {
              "resource_id": "res-tomato",
              "quantity": 5
            }
          ],
          "outputs": [
            {
              "resource_id": "res-tomato_can",
              "quantity": 3
            }
          ]
        }
      ]
    }
  ]
}
```

**GET /api/v1/sale/buildings**

Devuelve la lista de edificios de venta y sus recursos.

Response 200 OK
```json
{
  "data": [
    {
      "id": "sb-1",
      "master_id": "sb_shop",
      "name": "Shop",
      "construction_cost": 50,
      "resources": [
        {
          "resource_id": "res-tomato",
          "units_sold_per_second": 1,
          "price_per_unit": 15
        }
      ]
    }
  ]
}
```

### Users

**POST /api/v1/auth/register**

Registra un nuevo usuario. NO crea la empresa automáticamente (se crea con POST /api/v1/company después).

Request 201 Created
```json
{
  "username": "player1",
  "email": "player1@example.com",
  "password": "s3cret",
  "timezone": "Europe/Madrid"
}
```

Response:
```json
{
  "data": {
    "id": "uuid-user-1",
    "username": "player1",
    "email": "player1@example.com",
    "timezone": "Europe/Madrid",
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

**POST /api/v1/auth/login**

Inicia sesión a un usuario.

Request:
```json
{
  "username": "player1",
  "password": "s3cret"
}
```

Response 201 Created (sets cookies):
```json
{
  "data": {
    "user": {
      "id": "uuid-user-1",
      "username": "player1",
      "email": "player1@example.com",
      "timezone": "Europe/Madrid",
      "created_at": "2026-05-01T10:00:00Z"
    },
    "session_expires_at": "2026-05-17T12:34:56Z"
  }
}
```

**POST /api/v1/auth/logout**

Finaliza la sesión de un usuario.

Request: {}

Response 200 OK
```json
{
  "data": {
    "logged_out": true,
    "user_id": "uuid-user-1"
  }
}
```

**GET /api/v1/users/me**

Devuelve los datos del usuario a partir de su sesión.

Response 200 OK:
```json
{
  "data": {
    "id": "uuid-user-1",
    "username": "player1",
    "email": "player1@example.com",
    "timezone": "Europe/Madrid",
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

**PUT /api/v1/users/me**

Actualiza los datos del usuario.

Request:
```json
{
  "timezone": "Europe/Madrid"
}
```

Response 200 OK:
```json
{
  "data": {
    "id": "uuid-user-1",
    "username": "player1",
    "email": "player1@example.com",
    "timezone": "Europe/Madrid",
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

### Company

**GET /api/v1/company**

Obtiene la empresa del usuario.

Response 200 OK
```json
{
  "data": {
    "id": "uuid-co-1",
    "user_id": "uuid-user-1",
    "name": "Mi Empresa",
    "money": 1000,
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

**POST /api/v1/company**

Crea una empresa.

Request
```json
{
  "name": "Mi Empresa"
}
```

Response 201 Created:
```json
{
  "data": {
    "id": "uuid-co-1",
    "user_id": "uuid-user-1",
    "name": "Mi Empresa",
    "money": 1000,
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

**PUT /api/v1/company**

Modifica el nombre de la empresa del usuario.

Request:
```json
{
  "name": "Nuevo Nombre"
}
```

Response 200 OK:
```json
{
  "data": {
    "id": "uuid-co-1",
    "user_id": "uuid-user-1",
    "name": "Nuevo Nombre",
    "money": 1000,
    "created_at": "2026-05-01T10:00:00Z"
  }
}
```

**DELETE /api/v1/company**

Elimina la empresa del usuario.

Request: {}

Response 200 OK

**GET /api/v1/company/inventory**

Devuelve el estado del inventario de la empresa del usuario.

Response:
```json
{
  "data": [
    {
      "resource_id": "res-water",
      "quantity": 120
    },
    {
      "resource_id": "res-tomato",
      "quantity": 30
    }
  ]
}
```

### Market

**POST /api/v1/market/buy**

Compra recursos. Los recursos se añaden al inventario y se resta su coste del dinero de la empresa

Request:
```json
{
  "resource_id": "res-water",
  "quantity": 30
}
```

Response 200 OK
```json
{
  "data": {
    "company": {
      "id": "uuid-co-1",
      "user_id": "uuid-user-1",
      "name": "Mi Empresa",
      "money": 700,
      "created_at": "2026-05-01T10:00:00Z"
    },
    "inventory": [
      {
        "resource_id": "res-water",
        "quantity": 150
      }
    ]
  }
}
```

**POST /api/v1/market/sell**

Vende recursos. Los recursos se restan del inventario y se suman las ganancias al dinero de la empresa

Request:
```json
{
  "resource_id": "res-tomato",
  "quantity": 9
}
```

Response 200 OK
```json
{
  "data": {
    "company": {
      "id": "uuid-co-1",
      "user_id": "uuid-user-1",
      "name": "Mi Empresa",
      "money": 1015,
      "created_at": "2026-05-01T10:00:00Z"
    },
    "inventory": [
      {
        "resource_id": "res-tomato",
        "quantity": 21
      }
    ]
  }
}
```

### Production

**GET /api/v1/company/production/buildings**

Devuelve la lista de edificios de producción que tiene la empresa y su estado actual.

Response 200 OK
```json
{
  "data": [
    {
      "id": "cb-1",
      "production_building_id": "pb-1",
      "level": 1,
      "construction_ends_at": "2026-05-01T10:00:00Z",
      "active_run": {
        "id": "run-1",
        "process_id": "proc-1",
        "production_cycles": 3,
        "started_at": "2026-05-17T12:00:00Z",
        "ends_at": "2026-05-17T12:00:09Z",
        "is_collected": false
      }
    }
  ]
}
```

**POST /api/v1/company/production/buildings**

Construye un nuevo edificio de producción. Devuelve los datos del edificio de producción creado

Request 201 Created
```json
{
  "production_building_id": "pb-1"
}
```

Response:
```json
{
  "data": {
    "id": "cb-1",
    "production_building_id": "pb-1",
    "level": 1,
    "construction_ends_at": "2026-05-22T14:45:00Z",
    "active_run": null
  }
}
```

**POST /api/v1/company/production/buildings/:id/upgrade**

Sube de nivel un edificio de producción de la empresa

Request:
```json
{
  "levels": 1
}
```

Response 200 OK
```json
{
  "data": {
    "building": {
      "id": "cb-1",
      "production_building_id": "pb-1",
      "level": 2,
      "construction_ends_at": "2026-05-22T14:45:00Z",
      "active_run": null
    },
    "company": {
      "id": "uuid-co-1",
      "money": 900
    }
  }
}
```

**POST /api/v1/company/production/buildings/:id/start**

Inicia la producción en un edificio. Solo se puede si no hay `production_runs` activo (is_collected=0) para este building.

Request:
```json
{
  "process_id": "proc-1",
  "cycles": 3
}
```

Nota: El cliente envía el número de `cycles` (ciclos), no la duración. El servidor calcula:
- `total_duration = cycles * cycle_time_s`
- `ends_at = now + total_duration`
- Se verifica que el proceso esté disponible en la ventana horaria (UTC)
- Se verifica que el usuario tiene suficientes recursos
- Los recursos se extraen del inventario inmediatamente
- **Precios**: Se usan los precios ACTUALES del cache de gamedata (si cambiaron desde que se inició el request, se usa el nuevo precio)

Response 200 OK
```json
{
  "data": {
    "building": {
      "id": "cb-1",
      "production_building_id": "pb-1",
      "level": 1,
      "construction_ends_at": null,
      "active_run": {
        "id": "run-1",
        "process_id": "proc-1",
        "production_cycles": 3,
        "started_at": "2026-05-22T14:30:00Z",
        "ends_at": "2026-05-22T14:30:09Z",
        "is_collected": false
      }
    },
    "inventory": [
      {
        "resource_id": "res-tomato",
        "quantity": 45
      }
    ]
  }
}
```

**POST /api/v1/company/production/buildings/:id/collect**

Recoge los recursos producidos por un proceso productivo que ha terminado.

Request: {}

Response 200 OK
```json
{
  "data": {
    "building": {
      "id": "cb-1",
      "production_building_id": "pb-1",
      "level": 1,
      "construction_ends_at": null,
      "active_run": null
    },
    "inventory": [
      {
        "resource_id": "res-tomato_can",
        "quantity": 10
      }
    ]
  }
}
```

### Sale

**GET /api/v1/company/sale/buildings**

Devuelve la lista de edificios de venta que tiene la empresa y su estado actual.

Response 200 OK
```json
{
  "data": [
    {
      "id": "csb-1",
      "sale_building_id": "sb-1",
      "level": 1,
      "construction_ends_at": "2026-05-01T10:00:00Z",
      "active_run": {
        "id": "srun-1",
        "resource_id": "res-juice",
        "units_to_sell": 10,
        "started_at": "2026-05-17T12:00:00Z",
        "ends_at": "2026-05-17T12:00:10Z",
        "is_collected": false
      }
    }
  ]
}
```

**POST /api/v1/company/sale/buildings**

Construye un edificio de venta. Devuelve los datos del edificio de venta

Request 201 Created
```json
{
  "sale_building_id": "sb-1"
}
```

Response:
```json
{
  "data": {
    "id": "csb-1",
    "sale_building_id": "sb-1",
    "level": 1,
    "construction_ends_at": "2026-05-22T14:45:00Z",
    "active_run": null
  }
}
```

**POST /api/v1/company/sale/buildings/:id/upgrade**

Sube de nivel un edificio de venta.

Request:
```json
{
  "levels": 1
}
```

Response 200 OK
```json
{
  "data": {
    "building": {
      "id": "csb-1",
      "sale_building_id": "sb-1",
      "level": 2,
      "construction_ends_at": "2026-05-22T14:45:00Z",
      "active_run": null
    },
    "company": {
      "id": "uuid-co-1",
      "money": 950
    }
  }
}
```

**POST /api/v1/company/sale/buildings/:id/start**

Inicia el proceso de venta de un edificio de venta. Solo se puede si no hay `sale_runs` activo (is_collected=0) para este building.

Request:
```json
{
  "resource_id": "res-juice",
  "units": 10
}
```

Nota: El servidor calcula la duración automáticamente:
- `duration = units / units_sold_per_second`
- `ends_at = now + duration`
- Se verifica que el usuario tiene suficientes recursos
- Los recursos se extraen del inventario inmediatamente
- Al recoger, el dinero se calcula como `units * price_per_unit` (precio ACTUAL del cache)
- **Precios**: Se usan los precios ACTUALES del cache de gamedata (si cambiaron, se usa el nuevo precio)

Response 200 OK
```json
{
  "data": {
    "building": {
      "id": "csb-1",
      "sale_building_id": "sb-1",
      "level": 1,
      "construction_ends_at": null,
      "active_run": {
        "id": "srun-1",
        "resource_id": "res-juice",
        "units_to_sell": 10,
        "started_at": "2026-05-17T12:00:00Z",
        "ends_at": "2026-05-17T12:00:10Z",
        "is_collected": false
      }
    },
    "inventory": [
      {
        "resource_id": "res-juice",
        "quantity": 40
      }
    ]
  }
}
```

**POST /api/v1/company/sale/buildings/:id/collect**

Recoge los beneficios de un proceso de venta que ha terminado.

Request: {}

Response 200 OK
```json
{
  "data": {
    "building": {
      "id": "csb-1",
      "sale_building_id": "sb-1",
      "level": 1,
      "construction_ends_at": null,
      "active_run": null
    },
    "company": {
      "id": "uuid-co-1",
      "money": 1150
    }
  }
}
```

## Plan

### Patrones Arquitectónicos (Aplicar a todas las fases)

**Logging**: Todo lo importante debe ser logeado con zerolog:
- Inicio de operaciones críticas (transacciones, operaciones de dinero)
- Entrada a handlers HTTP (método, ruta, usuario)
- Cambios de estado importantes (crear building, terminar producción)
- Errores y excepciones
- Usar niveles: Info para operaciones normales, Error para problemas, Warn para situaciones anomalas

**Validación de datos**:
- Validar TODOS los inputs con go-playground/validator
- Validar en el handler (nivel HTTP) antes de llamar al service
- Validar en el service (logica de negocio) antes de modificar BD
- Devolver errores HTTP 400 con mensaje claro si falla

**Prevención de race conditions**:
- Usar transacciones SQL para cualquier operación que modifique múltiples tablas
- Usar pessimistic locking (SELECT ... FOR UPDATE) para operaciones críticas de dinero
- Versioning optimista en campos sensibles si es necesario
- En cada fase, especificar qué operaciones deben ser atomicas

---

### Fase 0: Inicialización (Infraestructura Base)

**Objetivo**: Establecer la estructura de proyectos, herramientas y configuración inicial.

#### 0.1 Configuración del Backend (Go)

- Crear estructura de carpetas según arquitectura definida (`server/cmd/api`, `server/internal/`)
- Inicializar módulo Go (`go mod init`)
- Instalar dependencias principales:
  - `chi` para router
  - `modernc.org/sqlite` para BD
  - `sqlc` para generación de queries
  - `go-playground/validator` para validación
  - `golang-jwt/jwt` para JWT
  - `rs/zerolog` para logging
  - `joho/godotenv` para cargar `.env`
  - `crypto/argon2` para hash de contraseñas
- Crear `internal/pkg/logger.go`:
  - Configurar zerolog con nivel configurable desde ENV
  - Crear logger global accesible en toda la app
  - Configurar rotación de logs en carpeta `./logs/`
  - Máximo 100 archivos de 10MB cada uno (1GB total)
  - Configurar rotación automática
- Crear `internal/pkg/validator.go`:
  - Wrapper sobre go-playground/validator
  - Custom validators (validar multiplos, ranges, etc.)
- Crear `internal/pkg/cache/gamedata.go`:
  - Cache en memoria para recursos, edificios de producción/venta, procesos
  - Estructura: `map[string]interface{}` con mutex
  - Métodos: `LoadGamedata()`, `GetResource()`, `GetProductionBuilding()`, `GetSaleBuilding()`, `GetProcess()`
  - Estrategia fallback: si no encuentra en cache, ir a BD
  - Inicializarse en `main()` antes de escuchar peticiones
- Crear `internal/pkg/cache/sessions.go`:
  - Cache en memoria para sesiones de usuario con TTL
  - Estructura: `map[tokenHash]SessionCache` con timestamp de expiración
  - Métodos: `SetSession()`, `GetSession()`, `InvalidateSession()`
  - Limpiador automático (goroutine) que ejecuta cada 6 horas para limpiar expirados
  - Inicializarse en `main()` antes de escuchar peticiones
- Crear archivos de configuración básicos:
  - `.env.example` con variables de entorno necesarias
  - `go.mod` y `go.sum`
  - README con instrucciones de setup

#### 0.2 Configuración de Base de Datos

- Crear directorio `server/internal/db/migrations`
- Crear archivo `schema.sql` con esquema completo (copiar del documento)
- Crear script de inicialización de BD
- Configurar `sqlc.yaml` para generación de queries

#### 0.3 Inicialización de Cache de Gamedata

- En `cmd/api/main.go`:
  - Al iniciar, cargar archivo JSON de gamedata (ubicación configurable por ENV variable `GAMEDATA_FILE`)
  - **Validación**: Verificar que la estructura JSON es válida, contiene todos los campos requeridos, no hay duplicados (master_ids únicos), no hay límite de cantidad de recursos/edificios
  - Si archivo JSON no existe o es inválido, **fallar el inicio** (no puede operar sin datos maestros)
  - **Warm-up obligatorio**: El servidor debe esperar a que el cache de gamedata esté completamente cargado en memoria antes de aceptar cualquier request HTTP
  - Cargar datos en BD (upsert) y precarga en cache
  - Loguear al inicio: "Gamedata loaded: X recursos, Y edificios producción, Z edificios venta"
- Crear archivo `config/gamedata.json` de ejemplo con estructura de datos maestros (mismo formato que endpoints GET: resources, production_buildings, sale_buildings)
- Crear repository para precargar datos:
  - `internal/resources/repository/resource_repo.go` con `GetAll()` para cargar todos los recursos
  - Similar para producción buildings, sale buildings, procesos

---

### Fase 1: Backend - Autenticación y Usuarios

**Objetivo**: Implementar sistema completo de autenticación y gestión de usuarios.
**Dependencias**: Fase 0

#### 1.1 Configuración de Base de Datos y Migraciones

- Ejecutar migraciones para crear todas las tablas
- Verificar que la BD está correctamente inicializada
- Crear script para reset de BD (para desarrollo)

#### 1.2 Modelos y Tipos Base

- Crear `internal/pkg/errors.go` con tipos de error personalizados
- Crear `internal/auth/models.go` con tipos de JWT
- Crear `internal/users/models/user.go` con estructura User (DBO, Model, DTO)
- Crear `internal/users/models/refresh_token.go`

#### 1.3 Configuración de JWT y Seguridad

- Crear `internal/auth/jwt.go`:
  - Generar claves de firma (cargar de variables de entorno)
  - Implementar `GenerateSessionToken()` (duración 1 minuto)
  - Implementar `GenerateRefreshToken()` (duración 300 días)
  - Implementar `ValidateSessionToken()` y `ValidateRefreshToken()`
- Crear `internal/auth/password.go`:
  - Implementar `HashPassword()` con argon2id
  - Implementar `VerifyPassword()`

#### 1.4 Repository de Usuarios

- Crear `internal/users/repository/user_repo.go`:
  - `CreateUser()`
  - `GetUserByID()`
  - `GetUserByUsername()`
  - `GetUserByEmail()`
  - `UpdateUserTimezone()`
  - `SoftDeleteUser()`
- Generar queries con sqlc en `internal/users/sql/queries.sql`

#### 1.5 Repository de Sesiones de Usuario

- Crear `internal/auth/repository/user_session_repo.go`:
  - `CreateSession()` - crear nueva sesión con session_id, verification_string, token_hash
  - `GetSessionByID()`
  - `GetSessionByHash()` - búsqueda por token_hash
  - `RevokeSession()` - marcar como revocada
  - `DeleteExpiredSessions()`
- Generar queries con sqlc

#### 1.6 Service de Usuarios

- Crear `internal/users/service/user_service.go`:
  - `Register()` - crear usuario, validar datos, hashear contraseña
  - `UpdateTimezone()` - validar cambio de timezone (30 días)
  - `GetUserProfile()`
  - `DeleteUser()`

#### 1.7 Service de Autenticación

- Crear `internal/auth/service/auth_service.go`:
  - `Login()` - validar credenciales, generar tokens, guardar en cache de sesiones
  - `RefreshSession()` - buscar token en cache de sesiones, si no existe o expiró buscar en BD, emitir nuevos
  - `Logout()` - revocar refresh token en BD e invalidar en cache

#### 1.8 Handlers HTTP

- Crear `internal/auth/http/routes.go` - registrar rutas
- Crear `internal/auth/http/register_handler.go`:
  - Validar input con validator
  - Llamar a `AuthService.Register()`
  - Devolver user DTO
- Crear `internal/auth/http/login_handler.go`:
  - Validar credenciales
  - Llamar a `AuthService.Login()`
  - Setear cookies httpOnly
  - Devolver user + session_expires_at
- Crear `internal/auth/http/logout_handler.go`:
  - Revocar token
  - Limpiar cookies
- Crear `internal/users/http/routes.go`
- Crear `internal/users/http/get_me_handler.go` - GET /api/v1/users/me
- Crear `internal/users/http/update_me_handler.go` - PUT /api/v1/users/me

#### 1.9 Middleware de Autenticación y Logging

- Crear `internal/auth/http/middleware.go`:
  - `AuthMiddleware()` - validar session token
  - `OptionalAuthMiddleware()` - permitir no autenticados
  - Manejo de tokens expirados (retry con refresh token)
- Crear `internal/http/middleware.go`:
  - `LoggingMiddleware()` - loguear método, ruta, usuario, status code
  - `RecoveryMiddleware()` - capturar panics y loguear errores

#### 1.10 Tests Unitarios

- Tests para `jwt.go`
- Tests para `password.go`
- Tests para servicios (con mocks de repository)
- Tests para handlers (con mocks de servicios)
- Tests para validación (validators personalizados)
- Verificar que: validaciones rechazan datos inválidos, logs se escriben correctamente

#### 1.11 Integración en main.go

- Configurar logger con zerolog (niveles: Info para operaciones normales, Warn para anomalías, Error para problemas)
- Cargar variables de entorno desde `.env` (usar valores por defecto si no existen):
  - `PORT` (default: 8080) - puerto HTTP del servidor
  - `LOG_LEVEL` (default: "info") - nivel de logging para zerolog (debug, info, warn, error)
  - `DATABASE_URL` (default: "./data/game.db") - ruta a BD SQLite
  - `ADMIN_USERNAME` (default: "") - username del admin (si vacío, no crear admin)
  - `ADMIN_PASSWORD` (default: "") - contraseña del admin

---

## Respuestas Finales a Preguntas Técnicas

Esta sección documenta todas las decisiones técnicas confirmadas para la especificación final del proyecto.

### 1. Estructura de Cache de Gamedata

**Pregunta**: ¿Usar map[string]interface{} genérico o typed maps?

**Respuesta CONFIRMADA**: Usar **THREE TYPED MAPS** (NO genéricos):
- `resources`: map[string]Resource (matching GET /api/v1/resources response)
- `productionBuildings`: map[string]ProductionBuilding (matching GET /api/v1/production/buildings)
- `saleBuildings`: map[string]SaleBuilding (matching GET /api/v1/sale/buildings)
- Cada mapa protegido con mutex para acceso concurrente
- **Beneficio**: Type safety sin casting, IDE auto-completion, mejor rendimiento

**Implementación**:
```go
type GamedataCache struct {
	resources            map[string]Resource
	productionBuildings  map[string]ProductionBuilding
	saleBuildings        map[string]SaleBuilding
	mu                   sync.RWMutex
}
```

### 2. Unicidad de Producción/Venta Activa por Edificio

**Pregunta**: ¿Cómo prevenir múltiples runs simultáneos del mismo edificio?

**Respuesta CONFIRMADA**: Usar **UNIQUE constraints en la database**:
- `production_runs`: UNIQUE (company_building_id, is_collected) WHERE is_collected = 0
- `sale_runs`: UNIQUE (company_sale_building_id, is_collected) WHERE is_collected = 0
- **Beneficio**: Previene race conditions a nivel SQL, no depende de validación en application code

**Implementación en schema**:
```sql
UNIQUE (company_building_id, is_collected) WHERE is_collected = 0
UNIQUE (company_sale_building_id, is_collected) WHERE is_collected = 0
```

### 3. Inferencia de Estado de Edificios

**Pregunta**: ¿Almacenar status como campo o inferirlo de timestamps y runs?

**Respuesta CONFIRMADA**: **INFERIR estado, NO almacenar en database**:
- **Construction** (En construcción): `construction_ends_at > NOW()`
- **Idle** (Inactivo): `construction_ends_at <= NOW() AND no active runs (is_collected=0)`
- **Producing/Selling** (Produciendo/Vendiendo): exists active run con `is_collected=0`
- Campo `status` **removido** de `company_production_buildings` y `company_sale_buildings`
- **Beneficio**: Single source of truth, elimina bugs de sincronización, estado siempre actual

**Ejemplos de cálculo en API response**:
```go
// Para production buildings
if building.construction_ends_at > now {
    status = "Construction"
} else if hasActiveRun(building.id) {
    status = "Producing"
} else {
    status = "Idle"
}
```

### 4. Configurabilidad de JWT Expiry

**Pregunta**: ¿Hardcoded 60s + 300d o configurable via ENV?

**Respuesta CONFIRMADA**: **Configurable via ENV variables**:
- `JWT_SESSION_EXPIRY`: default 60 segundos (sesión)
- `JWT_REFRESH_EXPIRY`: default 25920000 segundos (300 días)
- Cargadas en startup desde `.env` con `godotenv`
- **Beneficio**: Flexibilidad sin recompilación, ajustes en producción sin rebuild

**Implementación**:
```go
sessionExpiry := os.Getenv("JWT_SESSION_EXPIRY")
if sessionExpiry == "" {
    sessionExpiry = "60"  // default 60 seconds
}

refreshExpiry := os.Getenv("JWT_REFRESH_EXPIRY")
if refreshExpiry == "" {
    refreshExpiry = "25920000"  // default 300 days
}
```

### 5. Rate Limiting con Persistencia

**Pregunta**: ¿In-memory o persistent storage para intentos fallidos de login?

**Respuesta CONFIRMADA**: **Database table login_attempts** (persistent):
- Tabla: `login_attempts(id, user_id, username, failed_at, is_deleted)`
- Índice: `(username, failed_at)` para queries rápidas
- Regla: máximo 5 intentos fallidos por username en 15 minutos
- Query: `SELECT COUNT(*) FROM login_attempts WHERE username = ? AND failed_at > NOW() - 900`
- **Beneficio**: Persiste en restarts, auditable para GDPR, queryable para analytics

**Schema confirmado**:
```sql
CREATE TABLE login_attempts (
  id         TEXT     PRIMARY KEY,
  user_id    TEXT,
  username   TEXT     NOT NULL,
  failed_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted INTEGER  NOT NULL DEFAULT 0
);
CREATE INDEX idx_login_attempts_username ON login_attempts(username, failed_at);
```

### 6. Estrategia de Precios en Operaciones

**Pregunta**: ¿Precio del momento de request start o moment de ejecución (collect)?

**Respuesta CONFIRMADA**: **ALWAYS use CURRENT gamedata cache price**:
- Al hacer BUY: usa precio ACTUAL del cache en momento del request
- Al hacer SELL (market): usa precio ACTUAL del cache en momento del request  
- Al hacer COLLECT (production/sale): usa precio ACTUAL del cache en momento del collect
- Si gamedata cambió desde request start → usa nuevo precio (sin lock)
- **Beneficio**: Admin puede ajustar precios en tiempo real, evita gaming/arbitrage, game economía dinámica

**Notas de implementación en endpoints**:
- En POST /api/v1/market/buy: usar `gamedata.resources[resourceId].current_price`
- En POST /api/v1/market/sell: usar `gamedata.resources[resourceId].current_price`
- En POST /api/v1/company/production/buildings/:id/collect: usar precios ACTUALES
- En POST /api/v1/company/sale/buildings/:id/collect: usar precios ACTUALES
- Endpoints marcados con: "Se usan los precios ACTUALES del cache de gamedata"

---

## Status Final del Documento

**✅ Documento 100% FINALIZADO y LISTO PARA FASE 0**

Todas las actualizaciones completadas:
- ✅ Estructura de typed maps para gamedata cache
- ✅ UNIQUE constraints en production_runs y sale_runs
- ✅ Status field removido de company_production_buildings y company_sale_buildings
- ✅ Comentarios de inferencia de estado agregados en schema
- ✅ Tabla login_attempts con index para rate limiting
- ✅ Todas las notas de "precios actuales" en endpoints
- ✅ Sección de Respuestas Finales documentada

**No hay preguntas pendientes. El documento es completo y no ambiguo.**

Próximo paso: **FASE 0: Inicialización - Infraestructura Base**
  - `JWT_SECRET` (default: generar aleatorio si vacío) - clave para firmar JWTs (mínimo 32 caracteres)
  - `JWT_SESSION_EXPIRY` (default: 60) - duración del session token en segundos
  - `JWT_REFRESH_EXPIRY` (default: 25920000) - duración del refresh token en segundos (300 días)
  - `GAMEDATA_FILE` (default: "./config/gamedata.json") - ruta al archivo JSON de gamedata (requerido, server falla si no existe)
  - `APP_VERSION` (default: "0.1.0") - versión de la aplicación
  - `INITIAL_COMPANY_MONEY` (default: 1000) - dinero inicial cuando se crea empresa
  - `CORS_ORIGIN` (default: "*") - origen permitido para CORS (ej: "http://localhost:3000")
  - `SESSION_CACHE_TTL` (default: 21600) - TTL de sesiones en caché en segundos (6 horas)
  - `SESSION_CACHE_CLEANUP_INTERVAL` (default: 21600) - intervalo de limpieza de sesiones expiradas en segundos (6 horas)
- Si el admin no existe en BD, crearlo automáticamente con contraseña hasheada y role 'A'
- Inicializar conexión a BD
- Cargar gamedata desde archivo JSON (ver Fase 0.3) e inicializar cache (falla si no puede cargar)
- Inicializar limpiador de sesiones (LRU cleaner que ejecuta cada 6 horas)
- Inyectar dependencias
- Registrar rutas de auth y usuarios
- Registrar middleware global (logging, CORS, recovery, error handling)
- Escuchar en puerto configurado por ENV (ej: `PORT=8080`)

---

### Fase 2: Backend - Empresas y Gestión de Dinero

**Objetivo**: Sistema de empresas y gestión de dinero del jugador.
**Dependencias**: Fase 1

#### 2.1 Modelos de Empresa

- Crear `internal/company/models/company.go` (DBO, Model, DTO)
- Crear `internal/company/models/inventory.go`

#### 2.2 Repository de Empresas

- Crear `internal/company/repository/company_repo.go`:
  - `CreateCompany()`
  - `GetCompanyByUserID()`
  - `UpdateCompanyName()`
  - `UpdateCompanyMoney()` (con transacción)
  - `SoftDeleteCompany()`
- Crear `internal/company/repository/inventory_repo.go`:
  - `GetInventory()`
  - `GetInventoryItem()`
  - `UpdateInventoryItem()`
  - `AddToInventory()` (suma cantidad)
  - `RemoveFromInventory()` (resta cantidad, validar suficiencia)

#### 2.3 Service de Empresas

- Crear `internal/company/service/company_service.go`:
  - `CreateCompany()` - crear con dinero inicial (del .env)
  - `GetCompany()`
  - `UpdateCompanyName()`
  - `DeleteCompany()`

#### 2.4 Service de Inventario

- Crear `internal/company/service/inventory_service.go`:
  - `GetInventory()`
  - `AddResource()` - validación de cantidad, usar transacción con INSERT OR UPDATE (UPSERT)
  - `RemoveResource()` - validación de disponibilidad, usar transacción con SELECT ... FOR UPDATE pessimista
  - `TransferResource()` (usado en compra/venta) - operación atómica con locks
  - **IMPORTANTE**: Todas las operaciones de inventario deben ser transacciones. Usar SELECT ... FOR UPDATE para recursos críticos.

#### 2.5 Handlers HTTP

- Crear `internal/company/http/routes.go`
- GET `/api/v1/company` - obtener empresa
- POST `/api/v1/company` - crear empresa
- PUT `/api/v1/company` - actualizar nombre
- DELETE `/api/v1/company` - eliminar
- GET `/api/v1/company/inventory` - listar inventario

#### 2.6 Tests Unitarios

- Tests para servicios de empresa e inventario
- Tests para handlers

---

### Fase 3: Backend - Sistema de Recursos Maestros

**Objetivo**: Cargar datos maestros (recursos, edificios, procesos) que definen la economía del juego.
**Dependencias**: Fase 2

#### 3.1 Modelos de Recursos

- Crear `internal/resources/models/resource.go` (DBO, Model, DTO)

#### 3.2 Repository de Recursos

- Crear `internal/resources/repository/resource_repo.go`:
  - `CreateResource()`
  - `GetResourceByID()`
  - `GetAllResources()`
  - `UpsertResource()` (para importación)

#### 3.3 Modelos de Edificios de Producción

- Crear `internal/production/models/building.go`
- Crear `internal/production/models/process.go`

#### 3.4 Repository de Edificios de Producción

- Crear `internal/production/repository/building_repo.go`:
  - `CreateProductionBuilding()`
  - `GetProductionBuildingByID()`
  - `GetAllProductionBuildings()`
  - `UpsertProductionBuilding()`
- Crear `internal/production/repository/process_repo.go`:
  - `CreateProductionProcess()`
  - `GetProcessByID()`
  - `GetProcessesByBuildingID()`
  - Métodos para recursos de proceso (input/output)

#### 3.5 Modelos de Edificios de Venta

- Crear `internal/sale/models/building.go`

#### 3.6 Repository de Edificios de Venta

- Crear `internal/sale/repository/building_repo.go`:
  - Similar a production buildings
- Crear `internal/sale/repository/resource_repo.go`:
  - Recursos que se pueden vender en edificios

#### 3.7 Service de Gamedata

- Crear `internal/gamedata/service/gamedata_service.go`:
  - `ImportGameData()` - transacción que:
    - Valida la estructura JSON (todos los campos requeridos, sin duplicados en master_ids)
    - Inserta/actualiza todos los datos maestros en BD
    - Si cualquier error ocurre en la transacción, hace rollback automático
    - Loguea todos los errores en archivos con detalles específicos
    - Si es exitoso, recarga el cache en memoria desde BD
  - `GetGameData()` - devuelve todos los datos maestros (desde cache)
  - `RefreshCache()` - recarga cache desde BD (útil después de importar datos)

#### 3.8 Handlers HTTP

- GET `/api/v1/gamedata` - obtener todos los datos maestros (desde cache)
- POST `/api/v1/gamedata` - importar datos (solo admin), actualiza cache al terminar
- GET `/api/v1/resources` - listar recursos (desde cache)
- GET `/api/v1/resources/:id` - detalle recurso (desde cache, fallback a BD si no existe)
- GET `/api/v1/production/buildings` - listar edificios producción (desde cache)
- GET `/api/v1/production/buildings/:id` - detalle (desde cache)
- GET `/api/v1/sale/buildings` - listar edificios venta (desde cache)
- GET `/api/v1/sale/buildings/:id` - detalle (desde cache)

#### 3.9 Middleware de Admin

- Crear `internal/auth/http/admin_middleware.go`:
  - Validar que usuario tiene role 'A' (admin)

#### 3.10 Tests

- Tests para servicios y handlers

---

### Fase 4: Backend - Mercado (Compra/Venta)

**Objetivo**: Implementar compra y venta de recursos en el mercado.
**Dependencias**: Fase 3

#### 4.1 Service del Mercado

- Crear `internal/market/service/market_service.go`:
  - `BuyResource()` - obtener precio desde cache de gamedata, validar dinero, actualizar inventario y dinero en transacción
  - `SellResource()` - obtener precio desde cache, validar cantidad (múltiplos), actualizar dinero e inventario en transacción
  - **CRÍTICO**: Usar transacciones SQL para TODOS los updates (aunque sean 2 tablas). Usar SELECT ... FOR UPDATE en `companies.money` para evitar race conditions
  - Loguear cada compra/venta con ID de usuario, recurso, cantidad, dinero antes/después en audit_log
  - Cache fallback: si no encuentra recurso en cache, obtener de BD
  - **Ejemplo transacción BUY**:
    ```sql
    BEGIN TRANSACTION;
      SELECT money FROM companies WHERE id = ? FOR UPDATE;  -- lock pessimista
      IF money < cost THEN ROLLBACK, error INSUFFICIENT_FUNDS;
      UPDATE companies SET money = money - cost WHERE id = ?;
      INSERT INTO company_inventory (company_id, resource_id, quantity) VALUES (?, ?, ?)
        ON CONFLICT DO UPDATE SET quantity = quantity + ?;
      INSERT INTO audit_log (...) VALUES (...);
    COMMIT;
    ```

#### 4.2 Handlers HTTP

- POST `/api/v1/market/buy`:
  - Validar cantidad es múltiplo de `market_sale_qty`
  - Validar dinero suficiente
  - Llamar a service
  - Devolver company + inventory actualizado
- POST `/api/v1/market/sell`:
  - Validar cantidad es múltiplo
  - Validar inventario suficiente
  - Llamar a service
  - Devolver company + inventory actualizado

#### 4.3 Tests

- Tests con múltiples casos de error (dinero insuficiente, cantidad inválida, etc.)

---

### Fase 5: Backend - Edificios de Producción

**Objetivo**: Sistema completo de producción (construir, subir nivel, iniciar, recoger).
**Dependencias**: Fase 4

#### 5.1 Modelos de Empresa

- Crear `internal/production/models/company_building.go`
- Crear `internal/production/models/production_run.go`

#### 5.2 Repository

- Crear `internal/production/repository/company_building_repo.go`:
  - `CreateCompanyProductionBuilding()` - con status 'C' (construcción)
  - `GetCompanyProductionBuilding()`
  - `GetCompanyProductionBuildings()`
  - `UpdateBuildingStatus()`
  - `UpdateBuildingLevel()`
- Crear `internal/production/repository/production_run_repo.go`:
  - `CreateProductionRun()`
  - `GetProductionRun()`
  - `GetActiveRun()` - run no recogido
  - `MarkRunAsCollected()`

#### 5.3 Service de Producción

- Crear `internal/production/service/production_service.go`:
  - `BuildProductionBuilding()` - obtener costo desde cache de gamedata, crear con transacción (restar dinero, crear building con status C), loguear construcción
  - `UpgradeBuilding()` - obtener costo de upgrade desde cache, validar idle, restar dinero en transacción, actualizar nivel y status, loguear
  - `StartProduction()` - obtener proceso desde cache (fallback BD), validar recursos, hora si aplica, crear run en transacción, cambiar status a P, restar recursos
  - `CollectProduction()` - validar que run está lista, agregar recursos en transacción, marcar como recogido, cambiar status a I
  - **CRÍTICO**: Cada operación es una transacción separada, loguear timestamps para auditoría

#### 5.4 Validadores

- Crear `internal/production/validators.go`:
  - `ValidateProcessAvailable()` - obtener proceso desde cache, verificar ventana horaria
  - `ValidateResourcesAvailable()` - verificar que hay suficientes en inventario

#### 5.5 Handlers HTTP

- POST `/api/v1/company/production/buildings`:
  - Validar dinero suficiente
  - Llamar a service
  - Devolver building creado
- GET `/api/v1/company/production/buildings` - listar buildings de la empresa
- POST `/api/v1/company/production/buildings/:id/upgrade`:
  - Validar building está idle
  - Validar dinero suficiente
  - Devolver building actualizado + company
- POST `/api/v1/company/production/buildings/:id/start`:
  - Validar building está idle
  - Validar recursos suficientes
  - Validar ventana horaria si aplica
  - Devolver building + inventory actualizado
- POST `/api/v1/company/production/buildings/:id/collect`:
  - Validar que hay run activo y está completo
  - Devolver building + inventory actualizado

#### 5.6 Tests

- Tests para cada endpoint con múltiples escenarios

---

### Fase 6: Backend - Edificios de Venta

**Objetivo**: Sistema de venta en edificios (similar a producción pero genera dinero).
**Dependencias**: Fase 5

#### 6.1 Modelos

- Crear `internal/sale/models/company_building.go`
- Crear `internal/sale/models/sale_run.go`

#### 6.2 Repository

- Similar a producción pero para venta

#### 6.3 Service

- Crear `internal/sale/service/sale_service.go`:
  - `BuildSaleBuilding()` - obtener costo desde cache, crear en transacción
  - `UpgradeBuilding()` - obtener costo desde cache, actualizar en transacción
  - `StartSale()` - obtener recurso a vender desde cache, validar disponibilidad, crear run en transacción, restar inventario
  - `CollectSale()` - obtener precio de venta desde cache, agregar dinero en transacción, marcar como recogido

#### 6.4 Handlers HTTP

- POST `/api/v1/company/sale/buildings`
- GET `/api/v1/company/sale/buildings`
- POST `/api/v1/company/sale/buildings/:id/upgrade`
- POST `/api/v1/company/sale/buildings/:id/start`
- POST `/api/v1/company/sale/buildings/:id/collect`

#### 6.5 Tests

---

### Fase 7: Backend - Optimizaciones y Tests Finales

**Objetivo**: Realizar tests completos, optimizaciones y preparar para producción.
**Dependencias**: Fase 6

#### 7.1 Tests de Integración

- Tests end-to-end de flujos completos (crear usuario → crear empresa → comprar → producir → vender)
- Tests de transacciones en casos de error

#### 7.2 Validaciones Globales

- Verificar que TODOS los endpoints validan input con validator
- Tests de race conditions:
  - Simular múltiples requests simultáneos de compra/venta
  - Verificar que dinero nunca se duplica o desaparece
  - Verificar que inventario es correcto
  - Usar herramientas como `go test -race` para detectar race conditions

#### 7.3 Logging y Monitoreo

- Auditoría: Crear tabla `audit_log` para registrar:
  - Cambios de dinero (usuario, cantidad antes/después, timestamp)
  - Cambios de inventario (usuario, recurso, cantidad antes/después)
  - Creaciones/actualizaciones de buildings
- Implementar logging en TODOS los servicios (info en operaciones normales, error en problemas)
- Crear endpoint GET `/api/v1/status` - health check

#### 7.4 Documentación

- Generar documentación de endpoints (si se requiere OpenAPI/Swagger)
- Actualizar README con instrucciones de ejecución

#### 7.5 Build y Deploy

- Crear Dockerfile (opcional)
- Script de build para producción

---

