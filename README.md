# Your Own Boss

Un juego web idle de gestión y producción de recursos.

## Stack Tecnológico

- **Frontend**: React + TypeScript + Vite
- **Backend**: Go + Chi Router
- **Base de datos**: SQLite
- **Autenticación**: JWT (Access + Refresh tokens en httpOnly cookies)
- **Arquitectura**: Clean Architecture con capas (Handlers → Services → Repositories)

## Estructura del proyecto

```
yourownboss/
├── server/              # Backend en Go
│   ├── cmd/
│   │   └── api/        # Punto de entrada del servidor
│   ├── internal/
│   │   ├── auth/       # JWT y middleware de autenticación
│   │   ├── db/         # Conexión y modelos de BD
│   │   ├── repository/ # Capa de acceso a datos
│   │   ├── service/    # Capa de lógica de negocio
│   │   └── http/       # Handlers/Controllers HTTP
│   └── public/         # Build del frontend (generado)
├── web/                # Frontend en React
│   └── src/
│       ├── components/ # Componentes reutilizables
│       ├── contexts/   # React Context (estado global)
│       ├── lib/        # Utilidades (axios, etc)
│       └── pages/      # Páginas de la aplicación
└── docs/               # Documentación
    ├── YOUROWNBOSS.md  # Especificación del juego
    └── ARCHITECTURE.md # Arquitectura del backend
```

## Arquitectura

El backend sigue una arquitectura en capas con inyección de dependencias:

- **Handler Layer**: Maneja HTTP (validación, serialización, cookies)
- **Service Layer**: Lógica de negocio y orquestación
- **Repository Layer**: Acceso a datos y queries SQL

Ver [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) para más detalles.

## Requisitos

- Go 1.25+
- Node.js 18+
- pnpm (o npm)

## Instalación

### 1. Instalar dependencias del backend

```bash
cd server
go mod tidy
```

### 2. Instalar dependencias del frontend

```bash
cd web
pnpm install  # o npm install
```

## Desarrollo

### Opción 1: Ejecutar frontend y backend por separado (recomendado para desarrollo)

Terminal 1 - Backend:

```bash
cd server
go run cmd/api/main.go
```

Terminal 2 - Frontend:

```bash
cd web
pnpm dev  # o npm run dev
```

El frontend estará en `http://localhost:5173` y proxeará las peticiones `/api` al backend en `http://localhost:8080`.

### Opción 2: Build y servir todo desde Go

```bash
# Build del frontend
cd web
pnpm build  # o npm run build

# Ejecutar servidor (sirve el frontend + API)
cd ../server
go run cmd/api/main.go
```

Abre `http://localhost:8080` en tu navegador.

## Linting y Formateo

El proyecto usa ESLint para linting y Prettier para formateo de código.

### Comandos disponibles

```bash
cd web

# Linting
pnpm lint              # Revisar errores de ESLint
pnpm lint:fix          # Corregir errores automáticamente

# Formateo
pnpm format            # Formatear código con Prettier
pnpm format:check      # Verificar formato sin modificar
```

### Configuración del editor

Para VSCode, instala las extensiones:

- ESLint (`dbaeumer.vscode-eslint`)
- Prettier (`esbenp.prettier-vscode`)

El formateo automático al guardar se puede habilitar en la configuración del editor.

### Alias de rutas

El proyecto está configurado con el alias `@` que apunta a la carpeta `src/`. Puedes usarlo en tus imports:

```typescript
// En lugar de
import { useAuth } from "../../../contexts/AuthContext";

// Puedes usar
import { useAuth } from "@/contexts/AuthContext";
```

## Endpoints de la API

### Autenticación (públicos)

- `POST /api/auth/register` - Registrar nuevo usuario
- `POST /api/auth/login` - Iniciar sesión
- `POST /api/auth/refresh` - Renovar access token
- `POST /api/auth/logout` - Cerrar sesión

### Protegidos (requieren autenticación)

- `GET /api/auth/me` - Obtener usuario actual

## Flujo de autenticación

1. **Registro/Login**: El servidor genera un access token (15 min) y un refresh token (7 días), ambos como httpOnly cookies.
2. **Requests autenticados**: El frontend envía automáticamente las cookies. El backend valida el access token.
3. **Token expirado**: Si el access token expira (401), el interceptor de axios automáticamente llama a `/api/auth/refresh` usando el refresh token.
4. **Bloqueo de usuarios**: Revocar el refresh token en la BD bloquea al usuario.

## Variables de entorno

El servidor acepta los siguientes flags:

```bash
go run cmd/api/main.go -port 8080 -db yourownboss.db -jwt-secret "tu-secreto-aqui"
```

- `-port`: Puerto del servidor (default: 8080)
- `-db`: Ruta al archivo de base de datos SQLite (default: yourownboss.db)
- `-jwt-secret`: Clave secreta para firmar JWT (default: usa una clave por defecto)
- `-static`: Directorio de archivos estáticos (default: ../public)

**IMPORTANTE**: En producción, usa siempre `-jwt-secret` con una clave segura y aleatoria.

## Próximos pasos

- [ ] Crear modelo de empresas
- [ ] Implementar sistema de dinero
- [ ] Crear sistema de recursos y mercado
- [ ] Implementar edificios de producción
- [ ] Crear sistema de procesos de producción
- [ ] Panel de administración

## Licencia

MIT
