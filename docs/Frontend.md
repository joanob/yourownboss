## Frontend

### Stack y decisiones

- React + TypeScript + Vite.
- Estilos: componentes custom con CSS modules.
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

---

## Preguntas para aclarar antes de empezar desarrollo

### Validaciones en formularios

1. **Username**: ¿Longitud mínima/máxima? ¿Caracteres permitidos? (ej. alfanuméricos + guion)
2. **Password**: ¿Requisitos mínimos de seguridad? (ej. mínimo 8 caracteres, mayúscula, número, etc.)
3. **Email**: ¿Validación simple de formato o algo más estricto?
4. **Nombre de empresa**: ¿Longitud máxima? ¿Caracteres especiales permitidos?

Respuesta a todas las preguntas: este es un proyecto personal que no será llevado a producción, no necesito esas medidas de seguridad ni validacion

### Persistencia offline - Estructura IndexedDB

5. ¿Debería diseñar un esquema específico para IndexedDB, o la estructura es flexible? Si existe un esquema recomendado, ¿cuál es?
6. ¿Qué información se debe cachear localmente además de los datos maestros? (ej. inventario, dinero, estado de edificios)
7. Para la cola de acciones offline, ¿cuál es la estructura exacta que debería usar? Ej:
   ```typescript
   {
     id: string;
     action: 'BUY_RESOURCE' | 'START_PRODUCTION' | ...;
     payload: any;
     timestamp: number;
     status: 'pending' | 'synced' | 'failed';
   }
   ```
Respuesta a todas las preguntas: estas cosas debes decidirlas tú, quiero probar tu capacidad para desarrollar la aplicación completa

### Renovación de sesión JWT

8. Cuando el session token expire, ¿debería intentar renovarlo automáticamente de forma transparente, o redirigir al login?
9. ¿Qué debería hacer si el refresh token también expira o es inválido?

El refresh se hace automáticamente en la siguiente llamada, cuando pasa por el middleware de autenticación y el jwt session ha caducado, utiliza el jwt refresh para comprobar el estado de la sesión y devolve el jwt de sesion nuevo

### Manejo de tasas de límite (Rate Limiting)

10. ¿El cliente debería mostrar al usuario cuando está siendo rate-limited, o solo confiar en las respuestas 429 del servidor?
11. ¿Debería mostrar countdowns o restricciones predictivas (ej. "Espera 10 segundos antes de intentar de nuevo")?

R: solo 429, sin notificación de ningun tipo

### Internacionalización (i18n)

12. ¿El juego debe soportar múltiples idiomas desde el inicio, o comenzar solo en español/inglés?
13. ¿Hay un sistema de i18n específico que debería usar, o es flexible?

R: por ahora solo un idioma

### Tests

14. ¿Hay un objetivo de cobertura mínimo de tests? (ej. 80%, solo rutas críticas, etc.)
15. ¿Debería incluir tests e2e desde el inicio, o enfocarse en tests unitarios/integración primero?

R: por ahora sin tests

### Otros

16. ¿El diseño visual ya existe (figma, mockups) o debo crear una interfaz funcional básica?
17. ¿Debería considerar temas oscuro/claro desde el inicio, o comenzar con uno solo?

R: no existe, lo tienes que crear tú. Tema claro, quiero que sea un juego clicker así que será muy parecido a cualquier web de gestión de inventario empresarial

---

### Preguntas de seguimiento

18. **Idioma de la UI**: ¿En qué idioma deben estar los textos de la interfaz (botones, mensajes, labels)? ¿Español o inglés? R: español

19. **Dinero inicial de empresa**: En YOUROWNBOSS.md se menciona "Al crear la empresa empieza con una cantidad de dinero inicial" pero no se especifica cuánto. ¿Cuál es el valor inicial? R: eso lo define el servidor

20. **Endpoint de sincronización offline**: La especificación menciona que al reconectar "el servidor procesa la cola validando las acciones", pero no hay ningún endpoint de sync en la lista de endpoints de YOUROWNBOSS.md. ¿Existe un endpoint `POST /api/v1/sync` o similar? Si no existe, ¿debo implementar la sincronización reenviando cada acción en cola individualmente a sus endpoints respectivos? R: es probable que no exista aún y se vaya a hacer en el futuro. En ese caso, dejamos esta implementación para más tarde, ahora que sea todo online

21. **Páginas y rutas**: ¿Las rutas deben seguir esta estructura o tienes preferencia?
    - `/login`, `/register` — autenticación
    - `/` — dashboard con resumen de empresa (dinero, inventario resumido)
    - `/market` — compra/venta de recursos
    - `/production` — edificios de producción
    - `/sale` — edificios de venta
    - `/inventory` — inventario completo
    - `/profile` — perfil y timezone del usuario

	R: tu decides rutas y su contenido, igual que los guards. Quiero hacer vibe coding puro, yo te digo cómo quiero que funcione y tú lo haces funcionar.

22. **URL del backend en desarrollo**: ¿En qué puerto corre el backend localmente? (para configurar la `baseURL` del API client en desarrollo)

R: Haz que lea la baseURL de un archivo .env