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
  - [Base de datos](#base-de-datos)
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
- [Planificación de desarrollo](#planificación-de-desarrollo)
  - [Objetivo del MVP](#objetivo-del-mvp)
  - [Fases y entregables](#fases-y-entregables)
  - [Tareas técnicas por dominio](#tareas-técnicas-por-dominio)
  - [Prioridades y criterios de aceptación](#prioridades-y-criterios-de-aceptación)
  - [Roadmap tentativo](#roadmap-tentativo)
  - [Riesgos y mitigaciones](#riesgos-y-mitigaciones)
- [Decisiones consolidadas](#decisiones-consolidadas)
  - [Resumen de decisiones por área](#resumen-de-decisiones-por-área)
  - [Preguntas pendientes para siguiente fase](#preguntas-pendientes-para-siguiente-fase)


## Sobre el juego

Your Own Boss es un juego web y móvil de gestión y producción. El juego consiste en producir recursos a partir de otros recursos, venderlos y con el dinero comprar nuevos edificios de producción y venta para ganar más dinero. El foco de diseño es un juego casual de un jugador que no requiere estar conectado mucho tiempo con un sistema para aprovechar al máximo el tiempo que el usuario no podrá estar conectado.

### Dinámica del juego

**Usuarios y empresas**
- El jugador registra un usuario y crea una empresa con una cantidad de dinero inicial configurado por el admin (valor en `.env`). Un usuario solo puede tener una empresa.
- El jugador debe seleccionar un huso horario y si desea modificar el huso horario no podrá modificarlo otra vez hasta pasados 30 días.

**Recursos**
- Los recursos se compran en el mercado. Este mercado tiene precios estáticos, es ilimitado y no depende de otros jugadores. 
- Los recursos se venden por lotes y solo se puede vender múltiplos de esa cantidad de lote. Por ejemplo, si 3 unidades de agua se venden por 1 moneda, no es posible tener monedas decimales así que siempre se deben comprar y vender múltiplos de 3 unidades de agua.
- El inventario almacena los recursos de la empresa. Es ilimitado y los recursos no tienen caducidad.

**Producción y venta**
- Los edificios de producción y venta tienen un coste y tiempo de construcción. Cuando estén acabados, se podrá empezar a vender y producir. Una empresa puede tener varias instancias del mismo edificio de producción.
- Los procesos de producción determinan qué recursos se necesitan para producir otros recursos y cuál es el tiempo de producción. El proceso productivo se estructura en ciclos. Por ejemplo, un proceso de producción puede tener un ciclo de 3 segundos en el que se utilizan 5 tomates para fabricar 3 botes de tomates en conserva.
- El usuario inicia manualmente el proceso productivo indicando el tiempo o la cantidad de ciclos que desea producir. En ese momento, los recursos utilizados se extraen del inventario del jugador. Si el usuario no tiene los recursos suficientes, no es posible producir.
- A partir de ese momento, el edificio de producción estará produciendo y no podrá ocuparse en nada más. Cuando se acabe el tiempo de producción, el usuario podrá obtener los recursos producidos. En ese momento se añadirán a su inventario y el edificio de producción estará disponible de nuevo para iniciar otro proceso productivo.
- La mayoría de procesos se pueden iniciar y finalizar en cualquier momento y la única restricción es que la empresa tenga la cantidad de recursos necesaria. Sin embargo, hay algunos procesos que tienen ventanas horarias y no pueden iniciarse o acabar fuera de esas horas. Por ejemplo, si un proceso solo está disponible de 8 a 20, no se podrá iniciar antes de las 8 y no podrá terminar más tarde de las 20. Esta hora depende del huso horario almacenado en la información del usuario. 
- Cuando el edificio no está produciendo se puede subir de nivel. Los niveles actúan como multiplicadores: un edificio de nivel N produce y consume exactamente N veces la cantidad base por ciclo. Es equivalente a tener N fábricas iguales funcionando en paralelo. Se pueden subir varios niveles a la vez si se tiene el dinero y el tiempo necesario para mejorar el edificio es el mismo que para construirlo, independientemente de cuantos niveles se suban. 
- Los recursos obtenidos se pueden vender en el mercado o en edificios de venta. Los edificios de venta tienen las mismas características que los edificios de producción, pero en lugar de tener procesos productivos tienen procesos de venta en los que los recursos se transforman en dinero. Son más rentables que vender al mercado pero la venta no es inmediata. Al igual que en la producción, cuando se inicie un proceso de venta los recursos se extraen del inventario del jugador y al acabar el usuario podrá recoger el dinero.

### Monetización del juego

**Fase inicial**: sin monetización (juego es free-to-play sin anuncios).

**Monetización futura**: cuando el juego esté terminado y con base de usuarios grande, posibles anuncios no intrusivos o publicidad de marcas (ej. producir Fanta en lugar de "refresco genérico").

### Requisitos legales (GDPR)

El proyecto debe cumplir GDPR desde el inicio: consentimiento para trackers, borrado de datos a petición y políticas de privacidad documentadas.


## Backend

Es prioritario que el backend sea lo más rápido posible. Entiendo que SQLite no es la mejor opción pero no quiero tener postgresql que seguramente consuma más recursos de lo que lo hará Go con SQLite.

### Stack tecnológico

- Lenguaje: Go.
- Router ligero: `chi` (incluye middleware de rate limiting y CORS).
- Persistencia: SQLite usando `modernc.org/sqlite` para evitar cgo.
- Consultas generadas: `sqlc` para mantener consultas tipadas y seguras.
- Validación de input: `go-playground/validator` en los handlers.
- Hash de contraseñas: `argon2id` (más resistente a ataques GPU que bcrypt en 2026).
- Logger: `rs/zerolog`. Se guardarán archivos de log de 10MB con rotación
- Manejo de configuraciones: variables de entorno con un pequeño loader (`envconfig` o similar).

Razonamiento: esta pila mantiene el binario puro (sin CGO), consultas claras y testables.

Se plantea tener Redis o alguna herramienta que reduzca las lecturas de base de datos para mejorar el rendimiento.

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
- Tests: mockear interfaces de `repository` para probar `service` e mockear servicios para probar controladores.
- Migraciones en `internal/db/migrations`.

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

- El session token dura 1 minuto. En cada petición autenticada el middleware lo verifica sin necesidad de acceder a base de datos.
- Si el session token ha expirado, el cliente reintenta con el refresh token. El servidor valida el hash del refresh token contra la tabla `refresh_tokens` en BD, emite un nuevo par de cookies y revoca el token anterior.
- El refresh token tiene una expiración más larga, de 300 días, y queda invalidado en BD al renovarse o al hacer logout.
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

### Base de datos

- `modernc.org/sqlite` (sin CGO, fácil despliegue).
- `sqlc` para generar tipos y queries.
- Migraciones manuales en `internal/db/migrations`.
- Backups automáticos nocturnos (dump de SQLite a storage).
- Capa de abstracción DB para facilitar futura migración a PostgreSQL si es necesario.


## Frontend

### Stack y decisiones

- React + TypeScript + Vite.
- Estilos: componentes custom con SCSS (mejor organización que CSS plano para proyectos medianos: variables, anidamiento, mixins).
- Prioridad: versión web primero, luego PWA; aplicación móvil (React Native) como posibilidad futura.
- El juego debe poder jugarse sin conexión. Al reconectar, el servidor valida que todas las acciones que haya realizado el usuario sean posibles comprobando que los recursos producidos y el tiempo de producción es coherente.

### Estructura de carpetas

```
web/
	package.json
	vite.config.ts
	src/
		assets/
		layout/               # GameLayout, nav, footer
		pages/                # entradas de rutas (ProductionPage, ResourcesPage...)
		components/           # componentes UI genéricos (atoms, molecules)
		features/             # carpetas por dominio
			users/
				api.ts            # llamadas al backend (typed)
				hooks/
				components/
				pages/
				types.ts
			resources/
			production/
		services/             # api client, auth, storage
		contexts/             # AuthContext, ThemeContext
		hooks/                # hooks globales reutilizables
		lib/                  # utilidades, formatters
		styles/               # global css / tailwind
		types/                # tipos compartidos
		main.tsx
		App.tsx
	public/
	tests/
```

### Librerías y tooling

- `react-router` — routing.
- `react-query` (TanStack Query) — data fetching y caching.
- `react-hook-form` — formularios.
- `Vitest + Testing Library` — tests unitarios/integración.
- `Playwright` o `Cypress` — e2e.
- `ESLint + Prettier` — lint y formateo.
- `pnpm` — gestión de paquetes y monorepo.

### State management y caché

- `react-query` para estado asíncrono y caché de datos del servidor (inventario, edificios, procesos en curso).
- `React Context + reducer` para estado global de UI y sesión (usuario autenticado y preferencias del usuario).
- Estado local del temporizador de producción gestionado en el componente/hook correspondiente.

### API client y tipos

- Cliente centralizado en `services/api.ts` (baseURL, interceptors para JWT, renovación silenciosa de session token, manejo de errores).
- Tipos en `types/api.d.ts` alineados con los DTOs del backend.

### Offline y sincronización

La mayoría de acciones están disponibles sin conexión. Las únicas acciones que necesitan conexión son la autenticación y la creación de empresas.

**Frontend:**
- El cliente gestiona la cuenta atrás de los procesos de producción o venta activos. Al acabar el tiempo, el usuario emite la acción de obtener recursos o dinero al servidor. Esto reduce significativamente la carga de API.
- **Caché de datos maestros**: recursos, edificios, procesos y ritmos de venta se cargan al iniciar el juego (en el login o app start) y se cachean localmente, ya que estos datos no cambian frecuentemente.

**Sincronización offline → online:**
- Cuando el cliente se queda offline, coloca nuevas acciones en una cola local (transacciones locales en IndexedDB o localStorage).
- Al reconectar, el servidor procesa la cola validando que todas las acciones han sido posibles, tanto porque el usuario tiene los recursos o dinero necesarios como que el tiempo que ha pasado es correcto.
- Al acabar, se envía al usuario su estado actual y un registro de las acciones para indicar si han sido exitosas o erróneas.

### PWA y móvil

- Manifest + Service Worker (Workbox o Vite plugin) para PWA.
- Separar UI (components) de lógica (hooks/services).
- Mantener pruebas para flujos críticos (inicio de producción, recoger producto, ventas).

## Esquema de base de datos

- Las cantidades de recursos y precios usan valores enteros para evitar fallos por operaciones con coma flotante
- Los ids son UUIDs para que se puedan generar desde el cliente
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

-- roles: 'P' (player) | 'A' (admin)
CREATE TABLE users (
  id                              TEXT     PRIMARY KEY,
  username                        TEXT     NOT NULL UNIQUE,
  email                           TEXT     NOT NULL UNIQUE,
  password_hash                   TEXT     NOT NULL,
  role                            TEXT     NOT NULL DEFAULT 'P',
  timezone                        TEXT     NOT NULL,
  last_timezone_modification_at   DATETIME,
  created_at                      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted                      INTEGER  NOT NULL DEFAULT 0,
  deleted_at                      DATETIME
);

-- El token se almacena como hash (nunca en claro).
-- revoked_at != NULL → token invalidado (logout o rotación).
CREATE TABLE refresh_tokens (
  id          TEXT     PRIMARY KEY,
  user_id     TEXT     NOT NULL REFERENCES users(id),
  token_hash  TEXT     NOT NULL UNIQUE,
  expires_at  DATETIME NOT NULL,
  revoked_at  DATETIME,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted  INTEGER  NOT NULL DEFAULT 0,
  deleted_at  DATETIME
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

-- status: 'C' (constructing) | 'I' (idle) | 'P' (producing)
CREATE TABLE company_production_buildings (
  id                     TEXT     PRIMARY KEY,
  company_id             TEXT     NOT NULL REFERENCES companies(id),
  production_building_id TEXT     NOT NULL REFERENCES production_buildings(id),
  level                  INTEGER  NOT NULL DEFAULT 1,
  status                 TEXT     NOT NULL DEFAULT 'C',
  construction_ends_at   DATETIME,
  created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted             INTEGER  NOT NULL DEFAULT 0,
  deleted_at             DATETIME
);

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
  deleted_at         DATETIME
);

-- VENTA

-- status: 'C' (constructing) | 'I' (idle) | 'S' (selling)
CREATE TABLE company_sale_buildings (
  id                 TEXT     PRIMARY KEY,
  company_id         TEXT     NOT NULL REFERENCES companies(id),
  sale_building_id   TEXT     NOT NULL REFERENCES sale_buildings(id),
  level              INTEGER  NOT NULL DEFAULT 1,
  status             TEXT     NOT NULL DEFAULT 'C',
  construction_ends_at DATETIME,
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted         INTEGER  NOT NULL DEFAULT 0,
  deleted_at         DATETIME
);

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
  deleted_at                DATETIME
);

-- Triggers to convert DELETE into soft-delete

CREATE TRIGGER users_before_delete
BEFORE DELETE ON users
FOR EACH ROW
BEGIN
  UPDATE users
  SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
  WHERE id = OLD.id;

  -- Propagate soft-delete to direct and indirect child rows
  UPDATE refresh_tokens
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
- POST /api/v1/auth/register — Registra un nuevo usuario y crea la empresa inicial.
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

Registra un nuevo usuario.

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

Response 200 OK:

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
      "status": "P",
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
    "status": "P",
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

**POST /api/v1/company/production/buildings/:id/start**

Inicia la producción en un edificio

Request:
```json
{
  "process_id": "proc-1",
  "cycles": 3
}
```

Response 200 OK

**POST /api/v1/company/production/buildings/:id/collect**

Recoge los recursos producidos por un proceso productivo que ha terminado.

Request: {}

Response 200 OK

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
      "status": "S",
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
    "status": "S",
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

**POST /api/v1/company/sale/buildings/:id/start**

Inicia el proceso de venta de un edificio de venta

Request:
```json
{
  "resource_id": "res-juice",
  "units": 10
}
```

Response 200 OK

**POST /api/v1/company/sale/buildings/:id/collect**

Recoge los beneficios de un proceso de venta que ha terminado.

Request: {}

Response 200 OK
