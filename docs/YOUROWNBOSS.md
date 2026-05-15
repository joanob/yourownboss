# Your Own Boss

## Índice

- [Sobre el juego](#sobre-el-juego)
- [Dinámica de juego](#dinámica-del-juego)
- [Estructura de datos (recursos, procesos, edificios)](#estructura-de-datos-recursos-procesos-edificios)
- [Decisiones técnicas recomendadas](#decisiones-técnicas-recomendadas)
	- [Backend (Go)](#backend-go)
	- [Frontend (React / móvil)](#frontend-react--móvil)
	- [Base de datos y herramientas](#base-de-datos-y-herramientas)
- [Arquitectura y estructura de carpetas sugerida](#arquitectura-y-estructura-de-carpetas-sugerida)
- [DTO / DBO / Model: explicación y recomendaciones](#dto--dbo--model-explicación-y-recomendaciones)
- [Inyección de dependencias en Go (práctica recomendada)](#inyección-de-dependencias-en-go-práctica-recomendada)
- [API mínima (ejemplos)](#api-mínima-ejemplos)
- [Próximos pasos y decisiones pendientes](#próximos-pasos-y-decisiones-pendientes)


## Sobre el juego

Your Own Boss es un juego web y móvil idle de gestión y producción: producir recursos a partir de otros recursos, venderlos y con el dinero comprar edificios más avanzados. El foco de diseño es juego casual: sesiones cortas (p. ej. 5 minutos) cada varias horas.


## Dinámica del juego

- El jugador registra un usuario y crea una empresa con dinero inicial.
- Compra recursos y edificios de producción.
- Inicia procesos productivos en edificios; cuando el ciclo termina, el jugador pulsa "Obtener" para recoger la salida.
- Existen edificios de venta (venden a ritmo limitado) y un mercado invariable para compra/venta inmediata.


## Estructura de datos

- Usuarios: id, username, password, ...
- Empresas: id, id usuario, nombre, dinero, ...
- Recurso: id, nombre, categoría, precio_mercado (precio por unidad o por paquete).
- Edificio: id, nombre, coste_compra, tiempo_construcción, lista_de_procesos.
- Proceso: id, nombre, tiempo_ciclo (s), ventana_producción (horas o reglas), recursos_entrada, recursos_salida, output_por_ciclo.

Ejemplos prácticos están en el repositorio (migraciones y JSON de ejemplo). En la implementación, la tabla SQL puede reflejar DBOs (persistencia), los modelos en memoria pueden agregar estructuras anidadas (edificio -> lista procesos) y los DTOs exponen lo necesario a la API.

## Monetización

En la primera fase no habrá monetización. Cuando esté todo terminado, perfecto y haya una base de usuarios grande se plantea añadir anuncios no intrusivos. Si el juego consigue llamar la atención de marcas, la mejor forma de publicidad sería que se pudieran producir productos de esas marcas. Por ejemplo, en lugar de producir refresco que se pueda producir la nueva Fanta sabor melón.

## GDPR

Requisito legal/regulatorio: el proyecto debe cumplir GDPR desde el inicio (consentimiento para trackers, borrado de datos a petición, políticas de privacidad documentadas).


## Decisiones técnicas recomendadas

### Backend (Go)

- Lenguaje: Go.
- Router ligero: `chi` 
- Persistencia: SQLite usando `modernc.org/sqlite` para evitar cgo.
- Consultas generadas: `sqlc` para mantener consultas tipadas y seguras.
- Logger: `rs/zerolog`.
- Manejo de configuraciones: `envconfig` o simplemente `env` + `flag` con un pequeño loader.

Razonamiento: esta pila mantiene el binario puro (sin CGO), consultas claras y testables.


### Frontend (React / móvil)

React + PWA para ser usada tanto en web como en Android utilizando Web View. Quiero hacer primero la versión web.

El juego debería poderse jugar sin conexión. Luego se sincronizaría con el servidor y se comprobaría que el usuario no ha hecho trampas cambiando la hora para producir más rápido.

Primero desarrollar versión web para tener un MVP, posteriormente añadir funcionalidades PWA y se plantea en el futuro crear una app React Native.

### Base de datos y herramientas

- `modernc.org/sqlite` (evita C y facilita despliegues).
- `sqlc` para generar tipos/queries.
- Migraciones manuales

Backups diarios por la noche.

## Backend:

### Arquitectura y estructura de carpetas sugerida

Propongo organizar `internal/` por dominios (cada dominio es un paquete independiente) y dentro de cada dominio mantener las subcarpetas por responsabilidad: `http`, `service`, `repository`, `models` (y opcional `sql` para queries generadas). Esto facilita la navegación y el ownership del código.

Ejemplo de árbol mínimo:

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

Notas y buenas prácticas:

- Dependencias acíclicas: mantener el flujo `http -> service -> repository`. Nunca importe `http` desde `service` ni `repository` desde `http`.
- `models` dentro del dominio contienen tipos de dominio puros; los DTOs para la API pueden vivir en `http` o en `models/dto` según prefieras.
- Lugar de `sqlc`: puedes generar consultas por dominio (`internal/production/repository/queries.sql` -> package `repository`) para que el código generado quede junto al adaptador DB del dominio.
- Tipos compartidos (p. ej. errores, utilidades) en `internal/pkg` o `internal/shared` — evita meter reglas de negocio compartidas fuera de los dominios.
- Inyección/Contructor wiring: en `cmd/api/main.go` crea los repositorios, luego servicios y finalmente handlers. Ejemplo:

```
db := openDB(cfg)
usersRepo := users_repository.New(db)
usersSvc := users_service.New(usersRepo)
usersHandler := users_http.New(usersSvc)

router.Mount("/api/v1/users", usersHandler.Routes())
```

- Tests: cada dominio tiene sus propios tests y puedes mockear interfaces de `repository` para probar `service`.
- Migraciones y esquema global en `internal/db` para mantener orden; las migraciones pueden vivir en `internal/db/migrations` o en `simulation_server/migrations` según el componente que usen.

Beneficios de este enfoque:

- Escalabilidad: añadir un dominio nuevo es directo y evita mezclar responsabilidades.
- Ownership claro: cada dominio es una unidad de trabajo independiente.
- Facilitación de `sqlc` por paquete y tests más focalizados.


### DTO / DBO / Model: explicación y recomendaciones

- DBO (DB Object): estructuras exactamente mapeadas a tablas y columnas (usadas por sqlc/generated code). No exponer DBOs directamente a la API si contienen metadatos sensibles.
- Model: estructuras internas enriquecidas que combinan DBOs y reglas de negocio (p. ej. `Building` con su lista de `Process` ya cargada y métodos de validación).
- DTO (Data Transfer Object): estructuras para la API pública (request/response). Contienen sólo lo necesario para el cliente y validaciones.

Flujo típico:
DB <-> DBO (persistencia) -> convertir a Model (reglas) -> convertir a DTO (respuesta API)

Ejemplo: `BuildingDBO` (tabla), `Building` (modelo con procesos cargados), `BuildingDTO` (id, nombre, procesos: []ProcessDTO).


### Inyección de dependencias en Go (práctica recomendada)

Patrón común en Go: constructor injection con interfaces.

Ejemplo:

- Definir interfaz en la capa de servicios: `type UserRepo interface { GetByID(ctx context.Context, id int64) (*UserDBO, error) }`
- Implementación concreta en `repository/sql`: `type userRepoSQL struct { db *sql.DB }` que satisface `UserRepo`.
- En `cmd/api/main.go` construir cosas explícitamente:

	repo := repository.NewUserRepo(db)
	svc := service.NewUserService(repo)
	handler := http.NewUserHandler(svc)

Esto permite testear `service` inyectando un mock `UserRepo`.

Opcional: usar un DI container (fx, wire) añade complejidad; recomiendo empezar con constructor injection simple.


### API

Autenticación: JWT cookies httpOnly.

Si la respuesta tiene body, devolver un objeto result que tenga el valor o el error.

```
{
    value?: any,
    error?: {
        code: string,
        message: string	
    }
}
```


## Frontend

Propuesta de arquitectura frontend (React + TypeScript, Vite)

Objetivo: tener un frontend escalable, tipado y alineado por dominios con el backend (users, resources, production).

Estructura sugerida (dominios + responsabilidades):

```
web/                     # app web principal
	package.json
	vite.config.ts
	src/
		assets/
		layout/               # GameLayout, nav, footer
		pages/                # entradas de rutas (ProductionPage, ResourcesPage...)
		components/           # componentes UI genéricos (atoms, molecules)
		features/             # carpetas por dominio (users, resources, production)
			users/
				api.ts            # llamadas al backend (typed)
				hooks/            # hooks del dominio
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
	tests/                  # unit/integration

simulation_web/           # app de simulación si la quieres separada
packages/                 # (opcional monorepo) shared/ui, shared/types
```

Librerías y tooling recomendados

- `TypeScript` — mayor seguridad y mejores DX.  
- `Vite` — dev server rápido.  
- `react-router` — routing.  
- `react-query` (TanStack Query) — data fetching y caching.  
- `react-hook-form` — formularios.  
- `Vitest + Testing Library` — tests unitarios/integración.  
- `Playwright` o `Cypress` — e2e.  
- `ESLint + Prettier` — lint y formateo.  
- `pnpm` — gestionar monorepo.

State management

- Preferir `react-query` para estado asíncrono/cached.  
- Para estado global (usuario, UI) `React Context + reducer` o `zustand` si necesitas más flexibilidad.  

API client y tipos

- Centraliza el cliente en `services/api.ts` (baseURL, interceptors para JWT, manejo de errores).  
- Genera o mantiene tipos `types/api.d.ts` que reflejen los DTOs del backend para evitar desajustes.

PWA y móvil

- Habilita manifest + Service Worker (Workbox o la configuración de Vite) para PWA.  

CI / quality

- Scripts: `lint`, `test`, `build`.  
- Integrar checks en CI (lint, tests, build).  
- Revisión automática de cambios de tipos compartidos si usas monorepo.

Buenas prácticas

- Separar UI (components) de lógica (hooks/services).  
- Mantener pruebas para flujos críticos (inicio de producción, recoger producto, ventas).  
- Documentar contratos API en `docs/API.md` y mantener sincronía con `types/` del frontend.

## Simulaciones

Necesito un sistema de simulaciones que me permita probar diferentes valores de tiempo de producción, cantidad de recursos y precio de recursos para buscar un equilibrio entre todos los procesos productivos de manera que todos tengan un beneficio similar.

Había pensado tener un proyecto y una base de datos separada para todo esto. Aquí el frontend será muy sencillo, incluso se podría hacer con templ o algo así para no tener un frontend react. El simulador es completamente independiente del juego y correrá en otro servidor.

Las dos tablas de datos maestros serían recursos y procesos.

Luego una tabla de simulaciones con un id, un id de proceso, un beneficio calculado y un tiempo de fabricación. Y por último una tabla de simulacion_recursos donde se indique id, id de simulacion, id de recurso, si es de entrada o de salida, cantidad y precio.

El objetivo es por ejemplo calcular los beneficios que aporta el cultivo de tomates. Entonces indicaría que quiero simular para un tiempo de entre 1 y 30 segundos, entre 1 y 5 semillas con un precio de entre 1 y 10 y entre 1 y 3 tomates con un precio entre 2 y 200. El programa haría todas las posibilidades y calcularía su beneficio. Si el beneficio es mayor a un beneficio especificado, entonces se guarda en base de datos.

A su vez, el programa debe mostrar los datos de las simulaciones (para cada proceso) y ordenar los datos por diferentes parámetros.

### Estructura proyecto simulación

simulation_server/
	cmd/main.go — arranque y wiring (config, DB, logger).
	internal/api/ — handlers HTTP (endpoints para subir datos, lanzar simulación, consultar resultados).
	internal/service/ — orquestador de alto nivel (validación, preparar trabajos, políticas de invalidación).
	internal/worker/ — motor de simulación: worker pool, generación de combinaciones, control de concurrencia, checkpointing.
	internal/repository/ — adaptadores DB (insert batch, queries, transacciones).
	internal/types/ — DTOs y tipos compartidos (SimulationRequest, Combination, Result).
	internal/db/ — migraciones y helpers de conexión.
	internal/tools/ (opcional) — utilitarios (csv export, compress).

### Proceso de fondo

Añadir una tabla ligera simulation_jobs (id, process_id, status, started_at, finished_at, last_checkpoint_index, total_combinations) para:
- consultar progreso
- permitir cancelación controlada
- registrar logs/errors resumidos

Las combinaciones se generan en streaming para no saturar memoria. Se utilizan goroutines. Se guarda en base de datos cada X registros para no saturar memoria.

## Planificación de desarrollo del juego

### Resumen y objetivo

Objetivo: entregar un MVP jugable (web/PWA) que permita registro de usuario, creación de empresas, compra/venta de recursos, compra de edificios, iniciar procesos productivos y recoger producción. Paralelamente habilitar un servicio de simulaciones independiente para balance y tuning.

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

### Tareas técnicas detalladas (por dominio)

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

### Prioridades y criterios de aceptación (MVP)

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

### Estimaciones y roadmap tentativo

- Sprint 0 (setup): 1 semana — infra dev, DB, seeds, API docs.
- Sprint 1: 2 semanas — backend usuarios/empresas/recursos/producción básico.
- Sprint 2: 2 semanas — frontend MVP + integración básica.
- Sprint 3: 2 semanas — simulador inicial + pruebas de performance.
- Sprint 4: 1–2 semanas — CI/CD, Docker, backups y pulido.

Estos son estimados conservadores para una persona con experiencia en Go/React; ajustar según disponibilidad y prioridad.

### Riesgos y mitigaciones

- Riesgo: SQLite puede limitar concurrencia en producción.
	- Mitigación: diseñar capa de abstracción DB y plan de migración a PostgreSQL.
- Riesgo: generación exponencial de combinaciones en simulador.
	- Mitigación: parametrizar límites, sampling y validaciones previas.
- Riesgo: cheating offline/time manipulation.
	- Mitigación: sincronización y validaciones server-side en eventos críticos.

