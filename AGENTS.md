# AGENTS — Reglas y Instrucciones Persistentes para Agentes

Este archivo contiene reglas que los agentes (humanos o automatizados) deberán seguir siempre al trabajar en este repositorio. Está escrito en español para su comprensión por parte del equipo.

Reglas obligatorias (siempre aplicar):

- NO ejecutar comandos del sistema ni lanzar procesos por su cuenta. Cualquier ejecución de comandos (p. ej. `go build`, `npm install`, `docker`, `bash`, etc.) debe ser ejecutada por el usuario. Los agentes notificarán al usuario de qué comando debe ejecutar
- NO modificar archivos fuera de los cambios solicitados explícitamente por el usuario.

Instrucciones persistentes que el agente debe comprobar antes de responder:

1. Buscar en este archivo `AGENTS.md` instrucciones adicionales del usuario y obedecerlas.
2. Priorizar seguridad y no hacer suposiciones que impliquen ejecución remota o local de comandos.
3. Si el usuario solicita crear o editar archivos, realizar los cambios en el repositorio pero no ejecutar builds ni tests a menos que el usuario lo pida explícitamente.

Formato para añadir instrucciones extra (por el usuario):

El usuario puede agregar instrucciones debajo de la sección `INSTRUCCIONES-EXTRA` con texto claro. Ejemplo:

INSTRUCCIONES-EXTRA:
- Nunca ejecutar `npm` o `go` sin confirmación.
- Prefiere usar SQLite en modo local.
- El archivo YOUROWNBOSS.md es un cajón desastre de ideas. Cuando el archivo esté en el contexto, solo se debe hacer caso a la informacíon de ese archivo y a nada más de todo el proyecto. Todos los cambios se harán sobre ese archivo. Es un archivo que contendrá la idea del proyecto, la planificación para desarrollarlo y todas sus especificaciones.

Cuando un agente lea este archivo, debe parsear la sección `INSTRUCCIONES-EXTRA` (si existe) y actuar conforme a ella.

Notas para desarrolladores / agentes:

- Este archivo es informativo y tiene prioridad sobre instrucciones por defecto del entorno cuando el usuario lo ha creado.
- Mantener el lenguaje en español salvo que el usuario indique lo contrario.

Fecha de creación: 2026-04-09
