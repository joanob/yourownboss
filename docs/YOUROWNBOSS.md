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
	- [PWA y móvil](#pwa-y-móvil)
- [Simulaciones](#simulaciones)
	- [Objetivo](#objetivo)
	- [Modelo de datos](#modelo-de-datos)
	- [Estructura del proyecto](#estructura-del-proyecto)
	- [Proceso de fondo](#proceso-de-fondo)
- [Planificación de desarrollo](#planificación-de-desarrollo)
	- [Objetivo del MVP](#objetivo-del-mvp)
	- [Fases y entregables](#fases-y-entregables)
	- [Tareas técnicas por dominio](#tareas-técnicas-por-dominio)
	- [Prioridades y criterios de aceptación](#prioridades-y-criterios-de-aceptación)
	- [Roadmap tentativo](#roadmap-tentativo)
	- [Riesgos y mitigaciones](#riesgos-y-mitigaciones)


## Sobre el juego

Your Own Boss es un juego web y móvil idle de gestión y producción: producir recursos a partir de otros recursos, venderlos y con el dinero comprar edificios más avanzados. El foco de diseño es juego casual: sesiones cortas (p. ej. 5 minutos) cada varias horas.

### Dinámica del juego

- El jugador registra un usuario y crea una empresa con dinero inicial.
- Compra recursos y edificios de producción.
- Inicia procesos productivos en edificios; cuando el ciclo termina, el jugador pulsa "Obtener" para recoger la salida.
- Existen edificios de venta (venden a ritmo limitado) y un mercado invariable para compra/venta inmediata.

### Estructura de datos

- Usuarios: id, username, password, ...
- Empresas: id, id usuario, nombre, dinero, ...
- Recurso: id, nombre, categoría, precio_mercado (precio por unidad o por paquete).
- Edificio: id, nombre, coste_compra, tiempo_construcción, lista_de_procesos.
- Proceso: id, nombre, tiempo_ciclo (s), ventana_producción (horas o reglas), recursos_entrada, recursos_salida, output_por_ciclo.

Ejemplos prácticos están en el repositorio (migraciones y JSON de ejemplo). En la implementación, la tabla SQL puede reflejar DBOs (persistencia), los modelos en memoria pueden agregar estructuras anidadas (edificio -> lista procesos) y los DTOs exponen lo necesario a la API.

### Monetización

En la primera fase no habrá monetización. Cuando esté todo terminado, perfecto y haya una base de usuarios grande se plantea añadir anuncios no intrusivos. Si el juego consigue llamar la atención de marcas, la mejor forma de publicidad sería que se pudieran producir productos de esas marcas. Por ejemplo, en lugar de producir refresco que se pueda producir la nueva Fanta sabor melón.

### Requisitos legales (GDPR)

El proyecto debe cumplir GDPR desde el inicio: consentimiento para trackers, borrado de datos a petición y políticas de privacidad documentadas.


## Backend

### Stack tecnológico

- Lenguaje: Go.
- Router ligero: `chi`.
- Persistencia: SQLite usando `modernc.org/sqlite` para evitar cgo.
- Consultas generadas: `sqlc` para mantener consultas tipadas y seguras.
- Logger: `rs/zerolog`.
- Manejo de configuraciones: `envconfig` o simplemente `env` + `flag` con un pequeño loader.

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

Autenticación: JWT en cookies httpOnly.

Formato de respuesta cuando hay body:

```json
{
    "value": "...",
    "error": {
        "code": "string",
        "message": "string"
    }
}
```

### Base de datos

- `modernc.org/sqlite` (sin CGO, fácil despliegue).
- `sqlc` para generar tipos y queries.
- Migraciones manuales en `internal/db/migrations`.
- Backups automáticos nocturnos (dump de SQLite a storage).
- Capa de abstracción DB para facilitar futura migración a PostgreSQL si es necesario.


## Frontend

### Stack y decisiones

- React + TypeScript + Vite.
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

### API client y tipos

- Cliente centralizado en `services/api.ts` (baseURL, interceptors para JWT, manejo de errores).
- Tipos en `types/api.d.ts` alineados con los DTOs del backend.

### PWA y móvil

- Manifest + Service Worker (Workbox o Vite plugin) para PWA.
- Separar UI (components) de lógica (hooks/services).
- Mantener pruebas para flujos críticos (inicio de producción, recoger producto, ventas).

## Simulaciones

El simulador es un proyecto completamente independiente del juego, con su propia base de datos y su propio servidor.

### Objetivo

Probar diferentes valores de tiempo de producción, cantidad de recursos y precio de recursos para buscar un equilibrio entre todos los procesos productivos de manera que todos tengan un beneficio similar.

Ejemplo: simular el cultivo de tomates para un tiempo de entre 1 y 30 segundos, entre 1 y 5 semillas con un precio de entre 1 y 10 y entre 1 y 3 tomates con un precio entre 2 y 200. El programa genera todas las combinaciones posibles, calcula el beneficio de cada una y guarda en base de datos las que superen un umbral especificado. Los resultados se pueden consultar ordenados por diferentes parámetros.

### Modelo de datos

- `resources` — datos maestros de recursos.
- `processes` — datos maestros de procesos.
- `simulations` — id, process_id, beneficio_calculado, tiempo_fabricación.
- `simulation_resources` — id, simulation_id, resource_id, tipo (entrada/salida), cantidad, precio.
- `simulation_jobs` — id, process_id, status, started_at, finished_at, last_checkpoint_index, total_combinations.

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

## Planificación de desarrollo

### Objetivo del MVP

Entregar un MVP jugable (web) que permita registro de usuario, creación de empresas, compra/venta de recursos, compra de edificios, iniciar procesos productivos y recoger producción. El simulador se desarrolla en paralelo para apoyar el balance y tuning de la economía del juego.

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

