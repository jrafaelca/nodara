# Desarrollo local

Este Compose levanta el vertical slice de heartbeat con PostgreSQL, Core, Agent y la consola React.

```bash
docker compose -f deploy/dev/compose.yaml up --build
```

La consola queda disponible en `http://localhost:5173`. Core expone `http://localhost:8080/healthz` y `http://localhost:8080/readyz`.

Los certificados son efímeros de desarrollo y se guardan en el volumen Docker `dev-certs`. No se deben copiar al repositorio.
