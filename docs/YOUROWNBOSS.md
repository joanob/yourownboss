# Your Own Boss

## Índice

- [Sobre el juego](#sobre-el-juego)
	- [Dinámica del juego](#dinámica-del-juego)
	- [Estructura de datos](#estructura-de-datos)
	- [Monetización](#monetización)
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
- [Simulaciones](#simulaciones)
	- [Objetivo](#objetivo)
	- [Modelo de datos](#modelo-de-datos)
	- [Estructura del proyecto](#estructura-del-proyecto)
	- [Proceso de fondo](#proceso-de-fondo)
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

Your Own Boss es un juego web y móvil idle de gestión y producción: producir recursos a partir de otros recursos, venderlos y con el dinero comprar edificios más avanzados. El foco de diseño es juego casual: sesiones cortas (p. ej. 5 minutos) cada varias horas.

### Dinámica del juego

- El jugador registra un usuario y crea una empresa con el dinero inicial configurado por el admin (valor en `.env`). **Un usuario solo puede tener una empresa** (UNIQUE(user_id) en la tabla).
- Compra recursos en el mercado y adquiere edificios de producción. Los edificios tienen un coste de dinero y un tiempo de construcción; hasta que no finaliza el edificio no está disponible.
- Selecciona un proceso productivo, elige cuántos ciclos quiere producir y arranca la producción (sin límite de ciclos). No se puede recolectar hasta que hayan transcurrido todos los ciclos (`cycle_time_s × ciclos`).
- Cuando el tiempo termina, el jugador pulsa "Obtener" para añadir los recursos al inventario.
- Los edificios tienen niveles. El nivel actúa como multiplicador: un edificio de nivel N produce y consume exactamente N veces la cantidad base por ciclo (el tiempo de ciclo no cambia). Es equivalente a tener N fábricas iguales funcionando en paralelo. **Subir de nivel cuesta dinero (lineal: `purchase_cost × (nuevo_nivel - nivel_actual)`) y siempre toma el mismo tiempo (igual al `construction_time_s` del tipo, independientemente de cuántos niveles se suban)**. Solo se puede mejorar cuando el edificio no está produciendo (status = `idle`).
- Existen **edificios de venta** (entidad distinta a los de producción). Tienen una lista de recursos que pueden vender, precio de venta por unidad y ritmo máximo de venta (p. ej. vender 10 unidades/segundo × nivel). El usuario decide manualmente cuántas unidades vender desde su inventario. Las unidades se restan del inventario **al iniciar la venta** (no después). Después de que pase el tiempo necesario, recolecta el dinero. **Solo una `sale_run` activa por edificio de venta.**
- También existe un **mercado con precios fijos** para compra/venta inmediata. Los precios son estáticos, sin fluctuación dinámica. El mercado es ilimitado: se pueden comprar/vender recursos directamente sin restricción de cantidad o horario (distinto de los edificios de venta que tienen ritmo).
- El inventario es ilimitado y los recursos no caducan.
- **Ciclos y múltiplos**: el juego funciona en ciclos. Si un proceso produce 3 unidades por ciclo, las cantidades siempre serán múltiplos de 3. Lo mismo aplica para venta: si un edificio de venta vende 3 unidades por ciclo, solo se pueden vender múltiplos de 3.
- Algunos procesos tienen una ventana horaria: solo pueden iniciarse y ejecutarse dentro de ese rango (p. ej. electricidad solar: 08:00–20:00). **La validación es en la hora local del jugador según su timezone. Si el proceso empieza antes o termina después de la ventana, se rechaza.** El jugador almacena su timezone al registrarse y puede cambiarla solo una vez al mes para evitar abusos.
- **Validación de recursos**: al iniciar producción, se valida que la empresa tenga suficientes recursos de entrada. No se pueden quedar en rojo (saldo negativo).

### Estructura de datos (entidades principales)

Esta sección describe las entidades conceptuales. El esquema SQL completo está en [Esquema de base de datos](#esquema-de-base-de-datos).

- **users** — usuario del sistema. Roles: `player` y `admin`.
- **companies** — empresa de un usuario, con saldo de dinero.
- **resources** — catálogo maestro de recursos (importable vía endpoint JSON). Cada recurso tiene un `master_id` textual único para facilitar imports y ediciones.
- **production_buildings** — catálogo maestro de tipos de edificio (importable vía endpoint JSON). Definen coste de compra y tiempo de construcción en segundos.
- **company_buildings** — instancias de edificio de una empresa. Una empresa puede tener múltiples instancias del mismo `building_type` (sin límite). Solo pueden tener una `production_run` activa a la vez. Tienen nivel (multiplicador de entradas y salidas), estado (`constructing` / `idle` / `producing`) y `ready_at` mientras construyen o suben de nivel. Subir de nivel cuesta dinero (lineal: `purchase_cost × (nuevo_nivel - nivel_actual)`) y tiene tiempo de construcción (igual al `construction_time_s` del tipo, independientemente de cuántos niveles se suban).
- **sale_building_types** — catálogo maestro de tipos de edificio de venta. Define costes y tiempo de construcción, equivalente a `building_types`.
- **company_sale_buildings** — instancias de edificio de venta de una empresa. Tienen nivel (multiplicador del ritmo de venta), estado (`constructing` / `idle` / `selling`) y `ready_at`. Subir de nivel cuesta dinero y tiempo (igual que en edificios de producción). Solo pueden tener una `sale_run` activa a la vez.
- **sale_building_type_resources** — configuración maestro: qué recursos vende cada tipo de edificio de venta, a qué precio por unidad y a qué ritmo en unidades/segundo (p. ej. 10 unidades/segundo a nivel 1).
- **sale_runs** — ejecuciones de venta. Al iniciar, se quitan las unidades del inventario. Registra unidades a vender, precio_por_unidad capturado al inicio, `started_at`, `collect_at` = `started_at + (units_to_sell / (rate_per_second × level))` y dinero ganado. Aunque el admin cambió el precio después, la venta usa el precio capturado al iniciar.
- **processes** — catálogo maestro de procesos ligados a un `building_type`. Definen `cycle_time_s`, recursos de entrada/salida (cantidad base nivel 1), y opcionalmente `window_start_hour` / `window_end_hour`.
- **company_inventory** — stock de recursos por empresa. Sin límite ni caducidad.
- **production_runs** — ejecución de un proceso en un edificio concreto. Registra ciclos solicitados, `started_at`, `collect_at` = `started_at + cycle_time_s × cycles × level` y estado (`running` / `ready` / `collected`).
- **refresh_tokens** — tokens de refresco vinculados a sesiones de usuario, con hash almacenado en BD.

### Monetización y economía

**Fase inicial**: sin monetización (juego es free-to-play sin anuncios).

**Dinero en juego**:
- Dinero inicial: configurable en `.env`, mismo para todos los usuarios, sin límite máximo (hasta `int64`).
- Fuentes: venta de recursos en el mercado, venta de recursos en edificios de venta, algunos procesos pueden tener ganancia neutra o negativa.
- Uso: comprar edificios, subir niveles, comprar recursos en el mercado.

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

Organizar `internal/` por dominios (cada dominio es un paquete independiente) y dentro de cada dominio mantener las subcarpetas por responsabilidad: `http`, `service`, `repository`, `models` (y opcional `sql` para queries generadas).

```
server/
	cmd/api/
	internal/
		users/
			http/          # handlers, rutas, validaciones de request
			service/       # lógica de negocio, orquestación
			repository/    # sqlc queries y adaptadores DB
			models/        # modelos/domain types y validaciones
			sql/           # (opcional) .sql y generated code by sqlc
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
- `sqlc` por dominio: `internal/production/repository/queries.sql` → package `repository`.
- Tipos compartidos (errores, utilidades) en `internal/pkg` o `internal/shared`.
- Wiring en `cmd/api/main.go`:

```go
db := openDB(cfg)
usersRepo    := users_repository.New(db)
usersSvc     := users_service.New(usersRepo)
usersHandler := users_http.New(usersSvc)

router.Mount("/api/v1/users", usersHandler.Routes())
```

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

- El **session token** dura 1 minuto. En cada petición autenticada el middleware lo verifica.
- Si el session token ha expirado, el cliente reintenta con el **refresh token**. El servidor valida el hash del refresh token contra la tabla `refresh_tokens` en BD, emite un nuevo par de cookies y revoca el token anterior.
- El refresh token tiene una expiración más larga (p. ej. 30 días) y queda invalidado en BD al renovarse o al hacer logout.
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
- Estilos: componentes custom con **SCSS** (mejor organización que CSS plano para proyectos medianos: variables, anidamiento, mixins).
- Prioridad: versión web primero, luego PWA; aplicación móvil (React Native) como posibilidad futura.
- El juego debe poder jugarse sin conexión. Al reconectar, el servidor valida que no se haya manipulado el tiempo local para producir más rápido.

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
- `React Context + reducer` para estado global de UI y sesión (usuario autenticado).
- Estado local del temporizador de producción gestionado en el componente/hook correspondiente.

### API client y tipos

- Cliente centralizado en `services/api.ts` (baseURL, interceptors para JWT, renovación silenciosa de session token, manejo de errores).
- Tipos en `types/api.d.ts` alineados con los DTOs del backend.

### Offline y sincronización

Acciones disponibles sin conexión: iniciar proceso productivo, recolectar producción, consultar inventario y estado de edificios.

**Frontend:**
- El cliente gestiona los `collect_at` localmente y cuenta atrás. No hace polling al servidor; solo llama al servidor cuando el jugador pulsa "Obtener".
- Esto reduce significativamente la carga de API.
- **Caché de datos maestros**: recursos, edificios, procesos y ritmos de venta se cargan al iniciar el juego (en el login o app start) y se cachean localmente, ya que estos datos no cambian frecuentemente.

**Sincronización offline → online:**
- Estrategia: **Cola de acciones pendientes** (optimistic updates con rollback si el servidor rechaza).
- Cuando el cliente se queda offline, coloca nuevas acciones en una cola local (transacciones locales en IndexedDB o localStorage).
- Al reconectar, replaya la cola contra el servidor. Si una acción falla, se marca y el usuario recibe una notificación.
- Diseño completo pendiente en sesión específica (manejo de conflictos, versionado, etc).

### PWA y móvil

- Manifest + Service Worker (Workbox o Vite plugin) para PWA.
- Separar UI (components) de lógica (hooks/services).
- Mantener pruebas para flujos críticos (inicio de producción, recoger producto, ventas).

## Simulaciones

El simulador es un proyecto completamente independiente del juego, con su propia base de datos y su propio servidor.

### Objetivo

Probar diferentes valores de tiempo de ciclo, cantidad de recursos y precio para buscar un equilibrio entre todos los procesos productivos de manera que todos ofrezcan un beneficio/hora similar.

**Fórmula de beneficio:**

```
beneficio_por_ciclo = (Σ precio_salida × cantidad_salida) - (Σ precio_entrada × cantidad_entrada)
beneficio_por_hora  = beneficio_por_ciclo / (cycle_time_s / 3600)
```

Solo se persisten en BD los resultados cuyo `beneficio_por_hora` estén dentro del abanico especificado al lanzar el job (beneficios mínimos y beneficios máximos). Esto controla el volumen de resultados almacenados sin necesidad de limitar el número de combinaciones.

El simulador **solo simula procesos de producción**, no ventas del inventario ni edificios de venta.

**Visualización y aplicación**: Los resultados se visualizan directamente en el simulador. El admin revisa los resultados y modifica manualmente los datos maestros en `yourownboss` (no hay carga/import automático desde el simulador jamás).

### Modelo de datos

Los datos maestros se importan desde JSON vía endpoint. Cada entidad tiene un `master_id` textual único para facilitar imports, ediciones y referencias entre el juego y el simulador.

- `resources` — datos maestros de recursos (master_id, nombre, categoría, precio_mercado).
- `processes` — datos maestros de procesos (master_id, nombre, cycle_time_s, inputs, outputs).
- `simulations` — id, process_id, beneficio_por_hora, cycle_time_s.
- `simulation_resources` — id, simulation_id, resource_master_id, tipo (entrada/salida), cantidad, precio.
- `simulation_jobs` — id, process_id, min_profit_per_hour, status, started_at, finished_at, last_checkpoint_index, total_combinations.

### Estructura del proyecto

```
simulation_server/
	cmd/main.go              # arranque y wiring (config, DB, logger)
	internal/
		api/                 # handlers HTTP (lanzar simulación, consultar resultados)
		service/             # orquestador (validación, preparar trabajos)
		worker/              # motor: worker pool, generación de combinaciones, checkpointing
		repository/          # adaptadores DB (insert batch, queries, transacciones)
		types/               # DTOs compartidos (SimulationRequest, Combination, Result)
		db/                  # migraciones y helpers de conexión
		tools/               # (opcional) utilidades (csv export, compress)
```

### Proceso de fondo

La tabla `simulation_jobs` permite:
- Consultar progreso de un job en curso.
- Cancelación controlada.
- Registro de errores resumidos.

Las combinaciones se generan en streaming para no saturar memoria. Se usan goroutines y se persiste en base de datos cada X registros (batch insert). El frontend del simulador puede ser simple (templ u otro servidor de plantillas en Go).

## Esquema de base de datos

Schema SQL del servidor de juego (`server/`). Las cantidades de recursos y precios usan valores enteros para evitar fallos por operaciones con coma flotante. Los campos `master_id` son identificadores textuales únicos para datos maestros importables.

```sql
-- ─── DATOS MAESTROS ───────────────────────────────────────────────────────────

CREATE TABLE resources (
    id           INTEGER PRIMARY KEY,
    master_id    TEXT    NOT NULL UNIQUE,
    name         TEXT    NOT NULL,
    category     TEXT    NOT NULL,
    market_price INTEGER NOT NULL
);

CREATE TABLE building_types (
    id                   INTEGER PRIMARY KEY,
    master_id            TEXT    NOT NULL UNIQUE,
    name                 TEXT    NOT NULL,
    purchase_cost        INTEGER NOT NULL,
    construction_time_s  INTEGER NOT NULL   -- nivel 1; el upgrade también usa este valor × nivel
);

-- Un proceso pertenece a un building_type.
-- window_start_hour / window_end_hour: rango horario en que puede ejecutarse (0–23).
-- NULL en ambos = sin restricción horaria.
CREATE TABLE processes (
    id                  INTEGER PRIMARY KEY,
    master_id           TEXT    NOT NULL UNIQUE,
    building_type_id    INTEGER NOT NULL REFERENCES building_types(id),
    name                TEXT    NOT NULL,
    cycle_time_s        INTEGER NOT NULL,
    window_start_hour   INTEGER,
    window_end_hour     INTEGER
);

-- Recursos de entrada de un proceso (cantidad base para nivel 1).
CREATE TABLE process_inputs (
    id          INTEGER PRIMARY KEY,
    process_id  INTEGER NOT NULL REFERENCES processes(id),
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    quantity    REAL    NOT NULL
);

-- Recursos de salida de un proceso (cantidad base para nivel 1).
CREATE TABLE process_outputs (
    id          INTEGER PRIMARY KEY,
    process_id  INTEGER NOT NULL REFERENCES processes(id),
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    quantity    REAL    NOT NULL
);

-- ─── USUARIOS Y AUTENTICACIÓN ─────────────────────────────────────────────────

CREATE TABLE users (
    id            INTEGER  PRIMARY KEY,
    username      TEXT     NOT NULL UNIQUE,
    email         TEXT     NOT NULL UNIQUE,
    password_hash TEXT     NOT NULL,          -- argon2id
    role          TEXT     NOT NULL DEFAULT 'player',  -- 'player' | 'admin'
    timezone      TEXT     NOT NULL,          -- ej: 'Europe/Madrid', 'America/New_York'
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- El token se almacena como hash (nunca en claro).
-- revoked_at != NULL → token invalidado (logout o rotación).
CREATE TABLE refresh_tokens (
    id          INTEGER  PRIMARY KEY,
    user_id     INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT     NOT NULL UNIQUE,
    expires_at  DATETIME NOT NULL,
    revoked_at  DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── EMPRESAS ─────────────────────────────────────────────────────────────────

CREATE TABLE companies (
    id         INTEGER  PRIMARY KEY,
    user_id    INTEGER  NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT     NOT NULL,
    money      REAL     NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── INVENTARIO ───────────────────────────────────────────────────────────────

-- Sin límite de cantidad. Un registro por (empresa, recurso).
CREATE TABLE company_inventory (
    id          INTEGER PRIMARY KEY,
    company_id  INTEGER NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    resource_id INTEGER NOT NULL REFERENCES resources(id),
    quantity    REAL    NOT NULL DEFAULT 0,
    UNIQUE(company_id, resource_id)
);

-- ─── EDIFICIOS ────────────────────────────────────────────────────────────────

-- status: 'constructing' | 'idle' | 'producing'
-- ready_at: timestamp en que termina construcción o upgrade de nivel (NULL si idle/producing).
-- El nivel multiplica tanto inputs como outputs de cada ciclo.
CREATE TABLE company_buildings (
    id               INTEGER  PRIMARY KEY,
    company_id       INTEGER  NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    building_type_id INTEGER  NOT NULL REFERENCES building_types(id),
    level            INTEGER  NOT NULL DEFAULT 1,
    status           TEXT     NOT NULL DEFAULT 'constructing',
    ready_at         DATETIME,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── PRODUCCIÓN ───────────────────────────────────────────────────────────────

-- collect_at = started_at + (cycle_time_s * cycles_requested * level)
-- status: 'running' | 'ready' | 'collected'
CREATE TABLE production_runs (
    id                   INTEGER  PRIMARY KEY,
    company_building_id  INTEGER  NOT NULL REFERENCES company_buildings(id) ON DELETE CASCADE,
    process_id           INTEGER  NOT NULL REFERENCES processes(id),
    cycles_requested     INTEGER  NOT NULL,
    started_at           DATETIME NOT NULL,
    collect_at           DATETIME NOT NULL,
    status               TEXT     NOT NULL DEFAULT 'running',
    collected_at         DATETIME
);

-- ─── EDIFICIOS DE VENTA ──────────────────────────────────────────────────────

-- Catálogo maestro de tipos de edificio de venta (equivalente a building_types pero para venta).
CREATE TABLE sale_building_types (
    id                   INTEGER PRIMARY KEY,
    master_id            TEXT    NOT NULL UNIQUE,
    name                 TEXT    NOT NULL,
    purchase_cost        INTEGER NOT NULL,
    construction_time_s  INTEGER NOT NULL
);

-- Instancias de edificio de venta (equivalente a company_buildings pero para venta).
-- status: 'constructing' | 'idle' | 'selling'
-- El nivel multiplica el ritmo de venta.
CREATE TABLE company_sale_buildings (
    id                   INTEGER  PRIMARY KEY,
    company_id           INTEGER  NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    sale_building_type_id INTEGER NOT NULL REFERENCES sale_building_types(id),
    level                INTEGER  NOT NULL DEFAULT 1,
    status               TEXT     NOT NULL DEFAULT 'constructing',
    ready_at             DATETIME,
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Configuración maestro: qué recursos puede vender cada tipo de edificio de venta.
-- rate_per_second: ritmo base (nivel 1) en unidades por segundo.
CREATE TABLE sale_building_type_resources (
    id                   INTEGER PRIMARY KEY,
    sale_building_type_id INTEGER NOT NULL REFERENCES sale_building_types(id),
    resource_id          INTEGER NOT NULL REFERENCES resources(id),
    price_per_unit       INTEGER NOT NULL,
    rate_per_second      REAL    NOT NULL
);

-- Ejecuciones de venta (análogo a production_runs pero para edificios de venta).
-- collect_at = started_at + (units_to_sell / (rate_per_second * level))
-- status: 'selling' | 'ready' | 'collected'
CREATE TABLE sale_runs (
    id                   INTEGER  PRIMARY KEY,
    company_sale_building_id INTEGER NOT NULL REFERENCES company_sale_buildings(id) ON DELETE CASCADE,
    resource_id          INTEGER  NOT NULL REFERENCES resources(id),
    units_to_sell        REAL     NOT NULL,
    started_at           DATETIME NOT NULL,
    collect_at           DATETIME NOT NULL,
    status               TEXT     NOT NULL DEFAULT 'selling',
    collected_at         DATETIME,
    money_earned         REAL
);
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