# Web

En la carpeta simulation_web estará el frontend del simulador (el backend está en simulation_server).

Mi idea es que haya una pestaña resources y otra pestaña production. En resources estará la lista de recursos (id y nombre) y se puedan crear nuevos y modificar existentes. Con el botón de guardar se envian todos los recursos a la vez para hacer un bulk upsert.

La pestaña de production tendrá los edificios de produccion y permitirá crear nuevos. Al hacer click se navegará a una pantalla con ese edificio de producción y todos sus procesos. Se debe poder cambiar el id, el nombre, añadir o modificar procesos. Y en la misma pantalla dentro de los procesos se podrán añadir o quitar recursos tanto de entrada como de salida.

Al lado de cada proceso habrá un botón para navegar a una pantalla de simulación de ese proceso. En esta página de simulación se podrán crear nuevas simulaciones y revisar los datos de simulaciones pasadas.

## Stack

El stack quiero que sea React con Vite y Typescript. Imagino que necesitaremos axios y wouter. Dime los comandos que tengo que ejecutar para tenerlo todo listo. 
No necesito que la web tenga estado, el unico estado será la contraseña que se pide al principio de la sesión y se envia en Authorization: Bearer 

En este mismo archivo, añade una planificación para poder ir punto por punto diciendote que vayas generando la aplicación. Preguntame lo que necesites y dime los comandos que necesitas que ejecute. Ten en cuenta que solamente necesito una app funcional, no necesito que sea profesional, es para uso propio y privado

## Plan y comandos recomendados

1. Crear proyecto en la carpeta simulation_web e instalar dependencias necesarias

2. Autenticación ligera: modal para pedir password al abrir la app y guardar en memoria durante la sesión. Utilizar password en todas las requests `Authorization: Bearer <password>`

3. Pestaña `Resources`
	- Listado de recursos (id, name).
	- Edición en línea y creación de nuevas filas.
	- Botón `Guardar` que haga bulk upsert a `POST /api/resources/upsert`.

4. Pestaña `Production`
	- Listado de edificios con creación rápida.
	- Al hacer click en un edificio: pantalla detalle con procesos.
	- Permitir editar id, nombre; añadir/editar procesos y recursos (entrada/salida).
	- Al guardar, enviar upsert al endpoint existente (`/api/production` o `/api/production/{id}` según convengamos).

5. Página de simulación por proceso
	- Desde cada proceso abrir la pantalla de simulación.
	- Formulario para `time_min_ms`, `time_max_ms`, `time_step_ms` y rangos de precio por recurso.
	- Botón `Iniciar simulación` que POST a `/api/simulations` y devuelve 202.
	- Página para listar simulaciones pasadas (leer de `simulations` y `simulation_resources`).
