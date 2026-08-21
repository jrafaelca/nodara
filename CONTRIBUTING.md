# Contribuir a Nodara

La primera entrega define estructura y documentación. Los cambios posteriores deben respetar los límites de cada directorio y mantener las decisiones de arquitectura documentadas.

## Reglas

- Mantener el monorepo compilable cuando existan binarios.
- No guardar secretos, tokens, certificados privados, bases locales ni datos de hosts.
- Añadir contratos antes de implementar consumidores que dependan de ellos.
- Mantener migraciones versionadas y reversibles cuando sea posible.
- Agregar pruebas junto con cada comportamiento nuevo.
- Registrar decisiones estructurales en `docs/adr/`.
- Usar commits pequeños y descriptivos.

## Primera entrega

Esta entrega no incluye código ejecutable, dependencias, Compose ni infraestructura de producción.
