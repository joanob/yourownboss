# Project

Te explico aquí en qué consiste este proyecto. Estoy creando un juego de tipo idle. En este juego, se trata de fabricar recursos utilizando otros recursos y luego venderlos. Este proyecto es un simulador para balancear precios, tiempos de fabricación, cantidades y otros parámetros.

Para hacer las simulaciones y luego poder comparar datos para hacer balances, he pensado en que por una parte tengamos los datos maestros (recursos, edificios de producción y procesos de producción) y por otra las simulaciones. Para cada simulación, cada dato maestro tendrá datos variables y luego sobre esos datos variables se harán calculos que también se almacenarán.

Como medida de seguridad he pensado en que en todas las requests deba ir un Authentication: Bearer con un código especificado en .env. Ten en cuenta que ni es un proyecto profesional ni estará en la nube, estará en mi sistema local. No necesitas hacer cosas profesionales, pueden ser chapuzas. Por ejemplo, en la carga de datos o creación de simulaciones con simplemente recibir un 200 estoy feliz, no necesito que me devuelva nada

## Entidades maestras

Las entidades maestras tendrán ids numericos externos en lugar de ser autoincrementales.

Estos datos se gestionarán a través de una api, no habrá un seeder

### Recursos

Los recursos se utilizan para producir otros recursos o se venden. Por ahora he pensado que se vendan a precio fijo en el mercado. No es multijugador, el mercado tiene recursos infinitos para comprar y vender. 

Campos:
- id numérico no autogenerado
- nombre
- precio de mercado: variable

### Edificios de producción

Un edificio de producción puede tener uno o varios procesos de producción.

Campos:
- id numerico no autogenerado
- nombre
- coste: variable

### Procesos de producción

Un proceso de producción recibe 0 a N recursos de entrada y produce 1 a N recursos de salida. Las horas de inicio y fin de producción es porque hay algunos procesos que no tiene sentido que funcionen fuera de unas horas determinadas. Por ejemplo solo puedes generar electricidad utilizando paneles solares mientras haya sol. Para no complicarme, estas horas de inicio y fin serán fijas.

Campos:
- id numerico no autogenerado
- id edificio producción (FK)
- nombre
- tiempo: variable
- hora inicio produccion?
- hora fin producción?

### Recursos de producción

Un recurso de producción indica el recurso y cantidad de entrada o salida de un proceso de producción

Campos:
- id proceso (clave, FK)
- id recurso (clave, FK)
- is_output boolean (clave)
- cantidad numerico variable

Habrá una clave compuesta

## Simulación

Mi idea es tener una tabla simulaciones como enlace de todos los datos de una simulación. Por ejemplo, pruebo a hacer simulaciones para unos rangos de precios, costes de edificios, tiempos y cantidades. Pues para cada una de las combinaciones de ese rango se guardarán los datos y los resultados. Dime si hay una forma mejor, entiendo que así se generarán muchísimos recursos pero no se me ocurre forma mejor. Obviamente habrá cosas que no se guardarán, por ejemplo cuando el resultado sea que en vez de ganar dinero pierdas dinero.

Ejemplo: tengo panel solar que cada X milisegundos produce 1 unidad de electricidad (no hay recursos de entrada). Pues quiero probar con precios de 1 a 100 con saltos de 5, tiempos de 100 a 1000 con saltos de 100 y coste de la placa solar de 10000. El programa hará todas las combinaciones (precio de electricidad 1 y tiempo de 100, precio 1 y tiempo 200, y así todas) y guardará de cada combinación el beneficio de la produción por hora 

No me interesan las simulaciones en sí sino los resultados, por eso no quiero guardar combination_hash de simulación o params_json.

**Tablas**

Simulations: 
- id numérico autogenerado
- id proceso (FK)
- tiempo proceso (sacado del rango de tiempo)
- beneficio por hora (resultado de la simulación)

Simulation_resources
- id simulacion (clave, FK)
- id recurso (clave, FK)
- is_output (clave)
- precio (sacado del rango de precio del recurso)

## Plan de trabajo

1. Crear el esquema de base de datos en migrations/001_master_data.sql para entidades maestras:
	- `resources`, `production_buildings`, `production_processes`, `process_resources`.
	- Definir tipos, claves primarias y relaciones.

2. Crear el esquema para simulaciones en migrations/002_simulations.sql como especificado en ## Simulación:
	- `simulations`
	- `simulation_resources

3. Implementar endpoint para cargar/actualizar resources masivamente. Crear servicio y repositorio correspondiente. Como los ids con generados externamente, rechazar duplicados

4. Crear endpoints para cargar/actualizar y borrar edificio de producción, sus procesos y sus recursos (un mismo objeto con propiedades anidadas). Crear servicio y repositorio correspondiente. Como los ids con generados externamente, rechazar duplicados. Los datos que vengan deben ser los datos almacenados, si antes había un proceso que ya no está en el dto, borrar de base de datos.

5. Crear endpoint que reciba los datos de la simulación: id proceso, rango tiempos y lista de recursos con rangos de precios y cantidades. Este endpoint deberá generar un proceso en segundo plano que ejecute esta simulación. Hacer log cuando el proceso empiece y termine

6. Generar combinaciones según los rangos (respetando límites)

7. Realizar los calculos utilizando estas combinaciones y guardar los resultados de las simulaciones

8. Crear endpoint para leer todos los edificios con sus procesos y recursos (con los datos de los recursos también). Crear su servicio y repositorio

**Notas**

- Los ids maestros son externos siempre, nunca deben autogenerarse
- No quiero tener en cuenta que los procesos de fondo se puedan detener ni registrarlos en base de datos. No necesito una tabla simulation_jobs
- Implementa herramientas de concurrencia siempre que sea posible, me gustaría que se utilizaran varias goroutines para simular más rápido si es posible, pero sin estresar al ordenador. Prefiero que esté 1h procesando utilizando pocos recursos que 10 minutos a máxima potencia. Da igual si son 100 o 1000M de combinaciones, se hacen poco a poco y se consiguen todas
- No necesito tests por ahora. Si me gustaría hacer logs en consola o en un archivo .log si ocurre algún error, pero solo en ese caso. No me hacen falta métricas
- Buscar estrategia para guardar resultados cada X tiempo, así si algo falla que no se haya perdido el tiempo completamente. No necesito gestionar errores, es una herramienta privada y puedo volver a ejecutar manualmente lo que necesite

## DTOs (propuestos)

Abajo tienes propuestas de DTOs en Go y ejemplos JSON para los endpoints principales. Modifícalos como quieras.

1) Bulk upsert `resources` (Go)
```go
type ResourceDTO struct {
		ID          int64   `json:"id,omitempty"`
		Name        string  `json:"name"`
}

type BulkUpsertResourcesRequest struct {
		Resources []ResourceDTO `json:"resources"`
}
```

Ejemplo JSON request:
```json
{
	"resources": [
		{ "id": 1001, "name": "Electricidad" },
		{ "id": 200, "name": "Silicio" }
	]
}
```

2) Building / Processes composite (Go)
```go
type ProcessResourceDTO struct {
		ID         *int64  `json:"id,omitempty"`
		ResourceID int64   `json:"resource_id"`
		IsOutput   bool    `json:"is_output"`
}

type ProcessDTO struct {
		ID        *int64               `json:"id,omitempty"`
		Name      string               `json:"name"`
		StartHour *int                 `json:"start_hour,omitempty"`
		EndHour   *int                 `json:"end_hour,omitempty"`
		Resources []ProcessResourceDTO `json:"resources"`
}

type BuildingDTO struct {
		ID        *int64      `json:"id,omitempty"`
		Name      string      `json:"name"`
		Processes []ProcessDTO `json:"processes"`
}
```

Ejemplo JSON create:
```json
{
	"name": "Panel Solar",
	"processes": [
		{
			"name": "Generar electricidad",
			"start_hour": 6,
			"end_hour": 18,
			"resources": [ { "resource_id": 1, "quantity": 1, "is_output": true } ]
		}
	]
}
```

3) Solicitud de simulación (Go)
```go
type SimulationResourceRange struct {
		ResourceID int64   `json:"resource_id"`
		MinPrice   float64 `json:"min_price"`
		MaxPrice   float64 `json:"max_price"`
		Step       float64 `json:"step"`
}

type SimulationRequest struct {
		ProcessID       int64                   `json:"process_id"`
		TimeMinMs       int                     `json:"time_min_ms"`
		TimeMaxMs       int                     `json:"time_max_ms"`
		TimeStepMs      int                     `json:"time_step_ms"`
		ResourceRanges  []SimulationResourceRange `json:"resource_ranges"`
}
```

Ejemplo JSON solicitud:
```json
{
	"process_id": 10,
	"time_min_ms": 100,
	"time_max_ms": 1000,
	"time_step_ms": 100,
	"resource_ranges": [ { "resource_id": 1, "min_price": 1, "max_price": 100, "step": 5 } ]
}
```
