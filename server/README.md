# Your Own Boss - Backend Server

Backend del juego "Your Own Boss" escrito en Go.

## Stack Tecnológico

- **Lenguaje**: Go
- **Router**: chi v5
- **Base de datos**: SQLite (modernc.org/sqlite, sin CGO)
- **Generador de queries**: sqlc
- **Logger**: zerolog
- **Validación**: go-playground/validator
- **Hash de contraseñas**: argon2id
- **Manejo de configuración**: godotenv

## Requisitos

- Go 1.21 o superior
- sqlc 1.25 o superior

## Setup Inicial

### 1. Configurar variables de entorno

Copiar `.env.example` a `.env` y ajustar los valores:

```bash
cp .env.example .env
```

### 2. Instalar sqlc (si no está instalado)

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 3. Generar código desde SQL

```bash
sqlc generate
```

### 4. Inicializar base de datos

La base de datos se crea automáticamente al iniciar el servidor. Las tablas se crean ejecutando `schema.sql` en la primera ejecución.

### 5. Cargar datos maestros (gamedata)

Colocar archivo `config/gamedata.json` con los datos maestros. El servidor no iniciará sin estos datos.

## Ejecutar el servidor

```bash
go run ./cmd/api
```

El servidor escuchará en el puerto configurado en `.env` (default: 8080).

## Estructura de Carpetas

```
server/
├── cmd/
│   └── api/              # Punto de entrada de la aplicación
├── internal/
│   ├── auth/             # Autenticación y JWT
│   ├── users/            # Gestión de usuarios
│   ├── company/          # Gestión de empresas
│   ├── resources/        # Recursos del juego
│   ├── production/       # Edificios y procesos de producción
│   ├── sale/             # Edificios de venta
│   ├── market/           # Mercado (compra/venta)
│   ├── gamedata/         # Carga de datos maestros
│   ├── pkg/              # Utilidades compartidas
│   │   ├── logger.go     # Configuración de zerolog
│   │   ├── validator.go  # Wrapper de validación
│   │   └── cache/        # Caches en memoria
│   └── db/
│       └── migrations/   # Esquema de base de datos
├── .env.example          # Variables de entorno (ejemplo)
├── go.mod               # Dependencias de Go
└── README.md            # Este archivo
```

## Desarrollo

### Ejecutar tests

```bash
go test ./...
```

### Ejecutar tests con race detector

```bash
go test -race ./...
```

### Build para producción

```bash
go build -o yourownboss-api ./cmd/api
```

## Documentación

Ver `../../docs/YOUROWNBOSS.md` para especificaciones completas del proyecto.

## Fases de Desarrollo

El desarrollo sigue el plan de fases definido en YOUROWNBOSS.md:

- **Fase 0**: Inicialización (Infraestructura Base) ← Aquí estamos
- **Fase 1**: Autenticación y Usuarios
- **Fase 2**: Empresas y Gestión de Dinero
- **Fase 3**: Sistema de Recursos Maestros
- **Fase 4**: Mercado (Compra/Venta)
- **Fase 5**: Edificios de Producción
- **Fase 6**: Edificios de Venta
- **Fase 7**: Optimizaciones y Tests Finales
