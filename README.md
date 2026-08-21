# Nodara

Nodara es una plataforma privada para observar la salud operativa de hosts y aplicaciones, coordinar incidentes y notificar al equipo.

Este repositorio es un monorepo. Los componentes comparten contratos y decisiones de arquitectura, pero mantienen límites claros para poder evolucionar y desplegarse de forma independiente cuando sea necesario.

## Estructura

- `cmd/`: puntos de entrada de los binarios de Nodara.
- `internal/`: módulos privados del servidor y del agente.
- `api/`: contratos de comunicación y API pública.
- `db/`: migraciones de la base de datos.
- `web/`: consola privada.
- `deploy/`: material de despliegue futuro.
- `config/`: ejemplos de configuración portable.
- `docs/`: arquitectura, decisiones y operación.
- `scripts/`: automatizaciones de desarrollo y mantenimiento.
- `test/`: pruebas de integración, extremo a extremo y fixtures.

La especificación de producto se mantiene localmente en `SPEC.md` y no forma parte del repositorio.

## Principios iniciales

- El servidor y el agente se desarrollan como binarios Go independientes.
- La consola web se integrará con el servidor cuando exista la primera implementación funcional.
- La configuración operativa será validada, versionable y auditable.
- Los datos locales, secretos, volúmenes y exportaciones generadas no se versionan.

Consulta [`CONTRIBUTING.md`](CONTRIBUTING.md) antes de realizar cambios.
