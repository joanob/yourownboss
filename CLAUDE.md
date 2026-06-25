# Capychef Web - Development Guidelines

## DO's ✅

### Code Style & Standards

- **TypeScript Strict**: Usa tipos explícitos. Las funciones deben tener tipos de retorno declarados (`@typescript-eslint/explicit-function-return-type`)
- **Imports**: Mantén los imports organizados según el orden automático:
  1. Node modules y dependencias externas
  2. Archivos del proyecto (`@`, `components`, `hooks`, `utils`, `services`)
  3. Estilos CSS
- **Naming**: Usa PascalCase para componentes React, camelCase para variables/funciones
- **Component Structure**: Cada componente debe tener su propia carpeta con `ComponentName.tsx` e idealmente `ComponentName.module.css`

### Component Development

- **Functional Components**: Utiliza siempre componentes funcionales con hooks
- **Custom Hooks**: Extrae lógica reutilizable en hooks personalizados
- **React Hooks Rules**: Respeta las reglas de hooks (el linter lo verifica)
  - Los hooks solo se llaman al nivel superior del componente o en custom hooks
  - `exhaustive-deps` es warning, revísalo pero puede haber excepciones documentadas
- **Minimal Comments**: Solo comenta el "por qué" no obvio, no el "qué"
- **No Premature Abstractions**: No crees componentes genéricos hasta que tengas al menos 2-3 usos reales

### Styling

- **CSS Modules**: Usa `ComponentName.module.css` en la misma carpeta
- **Naming Convention**: Los class names deben ser claros y utilizar camelCase
- **No Inline Styles**: Evita estilos inline, usa siempre CSS Modules

### Git & Commits

No hagas commits ni interactues con Git, lo hará el desarrollador

---

## DON'Ts ❌

### Code Patterns to Avoid

- **No `any` Type**: No uses `any` en TypeScript sin documentar el porqué
  - Excepción: Usa `no-explicit-any: off` solo si es absolutamente necesario y documenta
- **No Inline Logic**: No hagas lógica compleja dentro de JSX, extrae a funciones/hooks
- **No Prop Drilling**: Si necesitas pasar props por 3+ componentes, usa Context o Redux
- **No Magic Numbers/Strings**: Extrae constantes
- **No Commented-Out Code**: Si no se usa, elimina
- **No Console Logs en Producción**: Usa un sistema de logging proper
- **No Mutable Defaults**: Evita mutation en reducers - Redux Toolkit maneja esto pero sé consciente

### React Anti-Patterns

- **No Component Declarations Inside Components**: Cada componente en su archivo
- **No Direct DOM Manipulation**: No uses `document.querySelector`, `getElementById`, etc. en componentes
- **No useEffect Chains**: Si tienes múltiples useEffect interdependientes, refactoriza
- **No Unnecessary Re-renders**: Memoiza componentes/callbacks solo si hay problema de rendimiento detectado
- **No Missing Dependencies**: El linter de hooks lo detectará, no ignores estas warnings

### Styling Mistakes

- **No Inline Styles**: Jamás `style={{ color: 'red' }}`
- **No Hardcoded Colors/Spacing**: Usa variables CSS o constantes
- **No `!important`**: Resuelve specificity correctamente

### Redux Mistakes

- **No Mutable State Mutations**: Aunque RTK lo permita bajo el hood, sigue siendo bad practice
- **No Async Logic sin Thunk**: Usa `createAsyncThunk` para API calls
- **No Selectors Inline**: Define selectores en un archivo separado
- **No Store Logic in Components**: La lógica de estado va en reducers/thunks

### Testing Mistakes

- **No Snapshot Tests para Components**: Snapshots se rompen con cambios CSS triviales
- **No Mocking Internals**: Testa el comportamiento observable, no detalles de implementación
- **No Tests sin Plan**: No escribas tests sin entender qué estás testando

### Build & Deployment

- **No Secrets en Código**: Usa `.env` y variables de entorno
- **No Broken Builds**: Siempre verifica `npm run build` antes de pushear
- **No Breaking Changes sin Coordinación**: Avisa si cambias APIs públicas

### Performance

- **No Bundle Bloat**: Importa solo lo que necesitas
- **No Unnecessary re-renders**: Usa React DevTools Profiler para detectar
- **No N+1 API Calls**: Agrupa requests cuando sea posible

---

## Development Workflow

No ejecutes comandos, genera el código y ya lo probaré yo. Si necesitas ejecutar algún comando para instalar dependencias o resolver configuraciones dímelo y lo ejecutaré yo

---

## Architecture Decisions

### Why Redux Toolkit?

- Estado global compartido entre features
- Devtools para debugging
- RTK simplifica boilerplate comparado con Redux puro

### Why React Hook Form?

- Mejor performance que Formik para formas grandes
- Menos renders innecesarios
- Integración simple con TypeScript

### Why CSS Modules over Tailwind?

- Componentes más encapsulados
- Explícitamente nombrados (mejor para debugging)
- Migración más fácil en futuro si es necesario

### Why Vite?

- Hot Module Replacement rápido
- Build más veloz que Webpack
- ESM nativo

---

## Common Pitfalls

1. **Olvidar tipos en Redux**: Siempre tipifica `AppRootState` y `AppDispatch`
2. **useEffect sin dependencias claras**: Revisar siempre `exhaustive-deps`
3. **No manejar loading/error states**: Siempre muestra estados intermedios al usuario
4. **Sobre-componentizar**: Mejor 3 líneas de JSX repetidas que un componente con 20 props
5. **Ignorar ESLint warnings**: El linter existe por una razón

---

## Resources

- React 19 Docs: https://react.dev
- TypeScript: https://www.typescriptlang.org
- Redux Toolkit: https://redux-toolkit.js.org
- Vite: https://vitejs.dev
- ESLint Config: ver `eslint.config.js` en root
