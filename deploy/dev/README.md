# Desarrollo local

Este Compose levanta el vertical slice de heartbeat con PostgreSQL, Core, Agent y la consola React.

```bash
docker compose -f deploy/dev/compose.yaml up --build
```

La consola queda disponible en `http://localhost:5173`. Core expone `http://localhost:8080/healthz` y `http://localhost:8080/readyz`.

Al iniciar una base de datos vacía, Core crea el usuario local `admin@nodara.dev` con contraseña temporal `password`. La consola exigirá cambiarla antes de mostrar el dashboard. En desarrollo, el enlace de recuperación de contraseña se registra en los logs de `nodara-core` porque `PASSWORD_RESET_DELIVERY=log`.

Para producción, configura `PASSWORD_RESET_DELIVERY=smtp`, `SMTP_URL`, `SMTP_FROM`, `PUBLIC_URL` y `COOKIE_SECURE=true`.

Los certificados son efímeros de desarrollo y se guardan en el volumen Docker `dev-certs`. No se deben copiar al repositorio.
