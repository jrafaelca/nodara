# web

Consola React de Nodara. Usa Vite, Tailwind CSS v4 y componentes locales de shadcn/ui.

La estructura de la aplicación es plana:

- `src/components/`: componentes visuales y componentes UI de shadcn.
- `src/layouts/`: composiciones compartidas de página.
- `src/pages/`: páginas de la consola, actualmente `dashboard.jsx` y `login.jsx`.

Rutas disponibles:

- `/`: boilerplate principal de monitoreo.
- `/login`: boilerplate de autenticación.
- Las rutas desconocidas redirigen a `/`.

La configuración de shadcn está en [`components.json`](components.json). Los componentes viven en `src/components/ui/` y los aliases usan `@/*`. El dashboard inicial sigue la distribución literal del boilerplate `sidebar-07`.

Comandos principales:

```bash
npm run dev
npm run build
npx shadcn@latest add <component>
```

La configuración del MCP y la skill para asistentes está documentada en [`docs/architecture/shadcn-ui.md`](../docs/architecture/shadcn-ui.md).

Consola privada de Nodara. Contendrá la aplicación React, componentes visuales, recursos estáticos y cliente de la API.
