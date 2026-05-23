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