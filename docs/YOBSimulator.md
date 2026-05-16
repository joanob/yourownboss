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