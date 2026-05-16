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

- El jugador registra un usuario y crea una empresa con una cantidad de dinero inicial configurado por el admin (valor en `.env`). Un usuario solo puede tener una empresa.
- Los recursos se compran en el mercado. Este mercado tiene precios estáticos, es ilimitado y no depende de otros jugadores. Los recursos se venden por lotes y solo se puede vender múltiplos de esa cantidad de lote. Por ejemplo, si 3 unidades de agua se venden por 1 moneda, no es posible tener monedas decimales así que siempre se deben comprar y vender múltiplos de 3 unidades de agua.
- El inventario almacena los recursos de la empresa. Es ilimitado y los recursos no tienen caducidad.
- Los edificios de producción y venta tienen un coste y tiempo de construcción. Cuando estén acabados, se podrá empezar a vender y producir. Una empresa puede tener varias instancias del mismo edificio de producción.
- Los procesos de producción determinan qué recursos se necesitan para producir otros recursos y cuál es el tiempo de producción. El proceso productivo se estructura en ciclos. Por ejemplo, un proceso de producción puede tener un ciclo de 3 segundos en el que se utilizan 5 tomates para fabricar 3 botes de tomates en conserva.
- El usuario inicia manualmente el proceso productivo indicando el tiempo o la cantidad de ciclos que desea producir. En ese momento, los recursos utilizados se extraen del inventario del jugador. Si el usuario no tiene los recursos suficientes, no es posible producir.
- A partir de ese momento, el edificio de producción estará produciendo y no podrá ocuparse en nada más. Cuando se acabe el tiempo de producción, el usuario podrá obtener los recursos producidos. En ese momento se añadirán a su inventario y el edificio de producción estará disponible de nuevo para iniciar otro proceso productivo.
- La mayoría de procesos se pueden iniciar y finalizar en cualquier momento y la única restricción es que la empresa tenga la cantidad de recursos necesaria. Sin embargo, hay algunos procesos que tienen ventanas horarias y no pueden iniciarse o acabar fuera de esas horas. Por ejemplo, si un proceso solo está disponible de 8 a 20, no se podrá iniciar antes de las 8 y no podrá terminar más tarde de las 20. Esta hora depende del huso horario almacenado en la información del usuario. Para evitar abusos, el usuario solo puede cambiar su huso horario una vez cada 30 días.
- Cuando el edificio no está produciendo se puede subir de nivel. Los niveles actúan como multiplicadores: un edificio de nivel N produce y consume exactamente N veces la cantidad base por ciclo. Es equivalente a tener N fábricas iguales funcionando en paralelo. Se pueden subir varios niveles a la vez si se tiene el dinero y el tiempo necesario para mejorar el edificio es el mismo que para construirlo, independientemente de cuantos niveles se suban. 
- Los recursos obtenidos se pueden vender en el mercado o en edificios de venta. Los edificios de venta tienen las mismas características que los edificios de producción, pero en lugar de tener procesos productivos tienen procesos de venta en los que los recursos se transforman en dinero. Son más rentables que vender al mercado pero la venta no es inmediata.

### Monetización del juego

**Fase inicial**: sin monetización (juego es free-to-play sin anuncios).

**Monetización futura**: cuando el juego esté terminado y con base de usuarios grande, posibles anuncios no intrusivos o publicidad de marcas (ej. producir Fanta en lugar de "refresco genérico").

### Requisitos legales (GDPR)

El proyecto debe cumplir GDPR desde el inicio: consentimiento para trackers, borrado de datos a petición y políticas de privacidad documentadas.


## Backend

### Stack tecnológico

- Lenguaje: Go.
- Router ligero: `chi` (incluye middleware de rate limiting y CORS).
- Persistencia: SQLite usando `modernc.org/sqlite` para evitar cgo.
- Consultas generadas: `sqlc` para mantener consultas tipadas y seguras.
- Validación de input: `go-playground/validator` en los handlers.
- Hash de contraseñas: `argon2id` (más resistente a ataques GPU que bcrypt en 2026).
- Logger: `rs/zerolog`.
- Manejo de configuraciones: variables de entorno con un pequeño loader (`envconfig` o similar).

Razonamiento: esta pila mantiene el binario puro (sin CGO), consultas claras y testables.

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
    "value": "...",
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
- Al reconectar, el servidor procesa la cola validando que todas las acciones han sido posibles. Si una acción falla, se marca y el usuario recibe una notificación.
- Al acabar, se envía al usuario su estado actual sincronizado al servidor.

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
	id           TEXT PRIMARY KEY,
    master_id    TEXT    NOT NULL UNIQUE,
    name         TEXT    NOT NULL,
    category     TEXT    NOT NULL,
    market_price INTEGER NOT NULL,
	market_sale_qty INTEGER NOT NULL
);

CREATE TABLE production_buildings (
	id                   TEXT PRIMARY KEY,
    master_id            TEXT    NOT NULL UNIQUE,
    name                 TEXT    NOT NULL,
    construction_cost        INTEGER NOT NULL,
    construction_time_s  INTEGER NOT NULL
);

-- Both window_start_hour and window_end_hour must be null or have value
CREATE TABLE production_processes (
	id                  TEXT PRIMARY KEY,
    master_id           TEXT    NOT NULL UNIQUE,
	production_building_id    TEXT NOT NULL REFERENCES production_buildings(id),
    name                TEXT    NOT NULL,
    cycle_time_s        INTEGER NOT NULL,
    window_start_hour   INTEGER,
    window_end_hour     INTEGER
);

CREATE TABLE production_process_resources (
	process_id  TEXT NOT NULL REFERENCES production_processes(id),
	resource_id TEXT NOT NULL REFERENCES resources(id),
	is_output BOOLEAN NOT NULL,
    quantity    INTEGER    NOT NULL,
	PRIMARY KEY(process_id, resource_id, is_output)
);

CREATE TABLE sale_buildings (
	id                   TEXT PRIMARY KEY,
    master_id            TEXT    NOT NULL UNIQUE,
    name                 TEXT    NOT NULL,
    construction_cost        INTEGER NOT NULL,
    construction_time_s  INTEGER NOT NULL
);

CREATE TABLE sale_resources (
	sale_building_id TEXT NOT NULL REFERENCES sale_buildings(id),
	resource_id          TEXT NOT NULL REFERENCES resources(id),
	price_per_unit       INTEGER NOT NULL,
    units_sold_per_second      INTEGER    NOT NULL,
	PRIMARY KEY(sale_building_id, resource_id)
);

-- USUARIOS Y AUTENTICACIÓN

-- roles: 'P' (player) | 'A' (admin)
CREATE TABLE users (
	id            TEXT  PRIMARY KEY,
    username      TEXT     NOT NULL UNIQUE,
    email         TEXT     NOT NULL UNIQUE,
    password_hash TEXT     NOT NULL,      
    role          TEXT     NOT NULL DEFAULT 'P',
    timezone      TEXT     NOT NULL,          
	last_timezone_modification_at DATETIME,
	created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

-- El token se almacena como hash (nunca en claro).
-- revoked_at != NULL → token invalidado (logout o rotación).
CREATE TABLE refresh_tokens (
	id          TEXT  PRIMARY KEY,
	user_id     TEXT  NOT NULL REFERENCES users(id),
	token_hash  TEXT     NOT NULL UNIQUE,
	expires_at  DATETIME NOT NULL,
	revoked_at  DATETIME,
	created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

-- EMPRESAS

CREATE TABLE companies (
	id         TEXT  PRIMARY KEY,
	user_id    TEXT  NOT NULL UNIQUE REFERENCES users(id),
    name       TEXT     NOT NULL,
    money      INTEGER     NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

CREATE TABLE company_inventory (
	id          TEXT PRIMARY KEY,
	company_id  TEXT NOT NULL REFERENCES companies(id),
	resource_id TEXT NOT NULL REFERENCES resources(id),
	quantity    INTEGER    NOT NULL DEFAULT 0,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME,
	UNIQUE(company_id, resource_id)
);

-- PRODUCCIÓN

-- status: 'C' (constructing) | 'I' (idle) | 'P' (producing)
CREATE TABLE company_production_buildings (
	id               TEXT  PRIMARY KEY,
	company_id       TEXT  NOT NULL REFERENCES companies(id),
	production_building_id TEXT  NOT NULL REFERENCES production_buildings(id),
    level            INTEGER  NOT NULL DEFAULT 1,
    status           TEXT     NOT NULL DEFAULT 'C',
    construction_ends_at         DATETIME,
	created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

CREATE TABLE production_runs (
	id                   TEXT  PRIMARY KEY,
	company_building_id  TEXT  NOT NULL REFERENCES company_production_buildings(id),
	process_id           TEXT  NOT NULL REFERENCES production_processes(id),
    production_cycles    INTEGER  NOT NULL,
    started_at           DATETIME NOT NULL,
    ends_at           DATETIME NOT NULL,
	is_collected INTEGER NOT NULL DEFAULT 0,
	collected_at         DATETIME,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

-- VENTA

-- status: 'C' (constructing) | 'I' (idle) | 'S' (selling)
CREATE TABLE company_sale_buildings (
	id               TEXT  PRIMARY KEY,
	company_id       TEXT  NOT NULL REFERENCES companies(id),
	sale_building_id TEXT  NOT NULL REFERENCES sale_buildings(id),
    level            INTEGER  NOT NULL DEFAULT 1,
    status           TEXT     NOT NULL DEFAULT 'C',
    construction_ends_at         DATETIME,
	created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

CREATE TABLE sale_runs (
	id                   TEXT  PRIMARY KEY,
	company_sale_building_id TEXT NOT NULL REFERENCES company_sale_buildings(id),
	resource_id          TEXT  NOT NULL REFERENCES resources(id),
    units_to_sell        INTEGER     NOT NULL,
    started_at           DATETIME NOT NULL,
    ends_at           DATETIME NOT NULL,
	is_collected INTEGER NOT NULL DEFAULT 0,
	collected_at         DATETIME,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

-- Triggers to convert DELETE into soft-delete

CREATE TRIGGER users_before_delete
BEFORE DELETE ON users
FOR EACH ROW
BEGIN
	UPDATE users SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	-- Propagate soft-delete to direct and indirect child rows
	UPDATE refresh_tokens SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE user_id = OLD.id;
	UPDATE companies SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE user_id = OLD.id;
	UPDATE company_inventory SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);
	UPDATE company_production_buildings SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);
	UPDATE company_sale_buildings SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id);
	UPDATE production_runs SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_building_id IN (SELECT id FROM company_production_buildings WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id));
	UPDATE sale_runs SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_sale_building_id IN (SELECT id FROM company_sale_buildings WHERE company_id IN (SELECT id FROM companies WHERE user_id = OLD.id));
	SELECT RAISE(IGNORE);
END;

CREATE TRIGGER companies_before_delete
BEFORE DELETE ON companies
FOR EACH ROW
BEGIN
	UPDATE companies SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	-- Propagate soft-delete to direct and indirect child tables
	UPDATE company_inventory SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_id = OLD.id;
	UPDATE company_production_buildings SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_id = OLD.id;
	UPDATE company_sale_buildings SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_id = OLD.id;
	UPDATE production_runs SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_building_id IN (SELECT id FROM company_production_buildings WHERE company_id = OLD.id);
	UPDATE sale_runs SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_sale_building_id IN (SELECT id FROM company_sale_buildings WHERE company_id = OLD.id);
	SELECT RAISE(IGNORE);
END;

CREATE TRIGGER company_production_buildings_before_delete
BEFORE DELETE ON company_production_buildings
FOR EACH ROW
BEGIN
	UPDATE company_production_buildings SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	-- Propagate soft-delete to production runs belonging to this company building
	UPDATE production_runs SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_building_id = OLD.id;
	SELECT RAISE(IGNORE);
END;

CREATE TRIGGER company_sale_buildings_before_delete
BEFORE DELETE ON company_sale_buildings
FOR EACH ROW
BEGIN
	UPDATE company_sale_buildings SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	-- Propagate soft-delete to sale runs belonging to this company sale building
	UPDATE sale_runs SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP WHERE company_sale_building_id = OLD.id;
	SELECT RAISE(IGNORE);
END;

```

## Planificación de desarrollo

### Objetivo del MVP

Entregar un juego web jugable que cubra el ciclo completo: registro → empresa → comprar recursos/edificios → producir → recolectar → vender. El simulador se desarrolla en paralelo desde el principio porque sus datos (precios, tiempos de ciclo) bloquean el balance del juego.

El código existente en `server/`, `simulation_server/` y `web/` es un prototipo previo sin la arquitectura definida aquí. **La Fase 0 incluye auditar ese código** para decidir qué reutilizar y qué reescribir.

**Definition of Done** para cualquier tarea: tests pasando + revisión manual del flujo + documentación actualizada (API.md o este archivo).

### Fases y entregables

- **Fase 0 — Preparación** (~1 semana)
	- Auditoría del código existente.
	- Migraciones iniciales alineadas con el esquema de esta documentación.
	- Seed data de recursos y procesos básicos.
	- `docs/API.md` con contratos de los endpoints del MVP.

- **Fase 1 — Núcleo backend + simulador básico** (~3 semanas, en paralelo)
	- Backend: auth (register/login/refresh), companies, resources, building_types, company_buildings, inventory.
	- Simulador: motor de cálculo, endpoint de import JSON, resultados básicos.
	- Los valores de precio y tiempo de ciclo del seed se balancean con el simulador.

- **Fase 2 — Producción backend** (~2 semanas)
	- Endpoints: `production_runs` (start/collect), ventana horaria, state machine.
	- Endpoint de compra/venta en mercado.
	- Edificios de venta (pendiente de definir mecánica completa).
	- Rate limiting en endpoints críticos.

- **Fase 3 — Frontend MVP** (~2–3 semanas)
	- Auth flow, dashboard empresa, inventario, edificios, producción (start/collect), mercado.
	- SCSS + componentes custom.
	- Integración con backend y manejo de errores.
	- Tests de flujos críticos.

- **Fase 4 — PWA y calidad** (~1–2 semanas)
	- Service Worker + manifest (acciones offline definidas).
	- CI: lint/test/build para todos los proyectos.
	- Entorno de staging.
	- Backups automáticos y checklist de despliegue.

- **Fase 5 — Pulido** (continuo)
	- Onboarding, UX polish.
	- Sincronización offline → online (diseño pendiente).
	- Métricas y telemetría básica.

### Tareas técnicas por dominio

- `users`: register/login, argon2id hash, refresh tokens, roles (`admin` / `player`).
- `companies`: create, balance, ownership checks.
- `resources` + `building_types`: endpoints de catálogo + endpoint de import JSON con `master_id`.
- `company_buildings`: comprar, estado constructing/idle/producing, nivel (upgrade).
- `production`: state machine runs (running → ready → collected), validación ventana horaria.
- `market`: compra/venta inmediata a precio fijo.
- `simulation`: motor de combinaciones, jobs, checkpointing, resultados.
- Infra: backups nocturnos, health-check, variables de entorno documentadas.

### Prioridades y criterios de aceptación

**Prioridad alta (MVP bloqueante)**
- Ciclo completo: registrarse → empresa → comprar recurso → comprar edificio → iniciar proceso → recolectar.
- Mercado básico operativo.
- Simulador con datos iniciales balanceados.

**Criterio de aceptación principal**: el usuario puede completar el ciclo completo sin errores y los tiempos de ciclo/precios están balanceados con el simulador.

### Roadmap tentativo

| Sprint | Duración | Contenido |
|--------|----------|-----------|
| 0 | 1 semana | Auditoría código, migraciones, seeds, API docs |
| 1 | 2 semanas | Auth + companies + catálogos + simulador básico |
| 2 | 2 semanas | Producción + mercado backend |
| 3 | 2–3 semanas | Frontend MVP |
| 4 | 1–2 semanas | PWA, CI, staging, backups |

Sin plazo fijo — ajustar según disponibilidad.

### Riesgos y mitigaciones

| Riesgo | Mitigación |
|--------|------------|
| SQLite con alta concurrencia en producción | Capa de abstracción DB + plan de migración a PostgreSQL documentado |
| Explosión combinatoria en el simulador | Solo se persisten resultados > umbral; sin otro límite artificial |
| Cheating offline (manipulación de hora) | Validaciones server-side en eventos críticos al sincronizar |
| Deuda técnica del código existente | Auditoría en Fase 0 antes de desarrollar sobre él |

### Fases y entregables (alto nivel)

- Fase 0 — Preparación (1 week)
	- Definir alcance MVP y criterios de aceptación.
	- Modelado de datos: tablas principales y migraciones iniciales.
	- `docs/API.md` con contratos básicos (auth, empresas, recursos, edificios, procesos, simulaciones).
	- Seed data inicial (`internal/db/seed.sql`).

- Fase 1 — Backend MVP (2–3 weeks)
	- Endpoints: auth (register/login), empresas (create, list), resources (list, prices), buildings (list, buy), production (start, collect), marketplace (buy/sell).
	- Implementar persistencia (SQLite + sqlc queries por dominio).
	- JWT auth con cookies httpOnly y middleware.
	- Tests unitarios para `service` y pruebas básicas de integración para endpoints críticos.
	- Documentación de API y ejemplos `curl`.

- Fase 2 — Frontend MVP (2–3 weeks)
	- Layout principal y `Auth` flow (login/register + persistencia de sesión).
	- Páginas: Dashboard/Empresa, Resources, Buildings, Production flow (start/collect), Market.
	- Integración con backend (`services/api.ts`) y manejo de errores/UX mínima.
	- PWA básico (manifest + service worker) y comportamiento offline mínimo para acciones no críticas.
	- Tests de componentes críticos (Vitest + Testing Library).

- Fase 3 — Simulaciones (2–4 weeks)
	- Implementar `simulation_server` con `simulation_jobs`, worker pool y checkpointing.
	- Endpoints para lanzar job, consultar progreso y descargar resultados.
	- Pruebas de performance y benchmarks para combinaciones grandes.
	- Export CSV/JSON de resultados y UI básica para revisar resultados (opcional en `simulation_web`).

- Fase 4 — Calidad, CI/CD y despliegue (1–2 weeks)
	- CI: `lint`, `test`, `build` para `server`, `simulation_server` y `web`.
	- Añadir linting, formateo y pre-commit hooks.
	- Dockerfile(s) para despliegue mínimo y checklist de despliegue.
	- Mecanismo de backup y plan de migración (SQLite -> PostgreSQL) documentado.

- Fase 5 — Pulido y métricas (continuo)
	- Telemetría (Prometheus metrics / logs), alertas básicas.
	- Ajustes de balance y economía usando los datos de `simulation_server`.
	- UX polish, tutorial inicial y sistema de onboarding.

### Tareas técnicas por dominio

- Backend (por dominio)
	- `users`: register/login, password hashing, tests, DTOs.
		- Roles iniciales: `admin` (gestión y mantenimiento) y `player` (usuario estándar). No se incluirá `tester` como rol separado en la primera versión; los testers pueden usar `admin` o cuentas específicas de prueba.
	- `companies`: create, balances, ownership checks.
	- `resources`: prices, market feed, seed data.
	- `production`: start process, process state machine (running/ready/collected), time window rules.
	- `simulation`: job queue, streaming generator, result persistence.

- Infra/Operaciones
	- Backups automáticos nocturnos para SQLite (dump y copia a storage).
	- Exponer métricas básicas y health-check endpoint.
	- Documentar variables de entorno y secretos necesarios.

### Prioridades y criterios de aceptación

- Prioridad alta
	- Registro/autenticación segura.
	- Comprar edificios y recursos.
	- Iniciar/recoger procesos productivos y ver inventario.
	- Mercado básico (compra/venta inmediata).

- Prioridad media
	- API documentada y seed data funcional.
	- Offline PWA mínimo (cache de assets + último estado conocido).
	- Simulador básico separable.

- Criterios de aceptación (ejemplos)
	- Usuario puede completar ciclo completo (registrarse → crear empresa → comprar recurso → iniciar proceso → recoger) sin errores.
	- API responde <500ms para endpoints críticos en entorno de dev.
	- Simulación de ejemplo completa y exportable a CSV.

### Roadmap tentativo

| Sprint | Duración | Contenido |
|--------|----------|-----------|
| 0 | 1 semana | Setup, DB, seeds, API docs |
| 1 | 2 semanas | Backend: usuarios, empresas, recursos, producción |
| 2 | 2 semanas | Frontend MVP + integración |
| 3 | 2 semanas | Simulador + benchmarks |
| 4 | 1–2 semanas | CI/CD, Docker, backups, pulido |

Sin plazo fijo — ajustar según disponibilidad.

### Riesgos y mitigaciones

| Riesgo | Mitigación |
|--------|------------|
| SQLite con alta concurrencia en producción | Capa de abstracción DB + plan de migración a PostgreSQL documentado |
| Explosión combinatoria en el simulador | Límites configurables, sampling, validaciones previas al lanzar job |
| Cheating offline (manipulación de hora) | Validaciones server-side en eventos críticos al sincronizar |

## Decisiones consolidadas

Este documento fija las decisiones sobre el juego y el simulador. Los detalles de backend (endpoints, rate limiting, arquitectura, CI/CD) se definirán en la siguiente fase. Los detalles de frontend (pantallas, componentes, UX) se definirán después.

### Resumen de decisiones por área

**Dinámica de juego**:
- Edificios pueden subirse de nivel solo cuando no están activos (idle). Tiempo de mejora siempre igual, independiente de niveles.
- Venta manual: recursos se restan al iniciar, no al terminar. Precio se captura al iniciar y no varía.
- Ciclos y múltiplos: cantidades siempre son múltiplos de ciclos (si produce 3/ciclo, solo múltiplos de 3).
- Procesos pueden tener ganancia neutra o negativa para mecánicas especiales.
- Dinero: hasta `int64`, sin techo adicional.
- Mercado: ilimitado para compra/venta inmediata, distinto de edificios de venta con ritmo.

**Timezone y validación**:
- Cada usuario elige timezone al registrarse.
- Puede cambiar solo una vez al mes.
- Ventana horaria: valida que inicio Y fin estén dentro del rango.

**Simulador**:
- Solo simula procesos de producción, no venta.
- Admin carga datos JSON manualmente (no hay import automático desde el juego).
- Múltiples jobs en paralelo.
- Visualización solo en web, sin exportación CSV.

**Logs**:
- Archivos de 10 MB con rotación.