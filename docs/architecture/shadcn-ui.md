# shadcn/ui en Nodara

La consola vive en `web/` y usa shadcn/ui como una colección de componentes locales. No se importa un paquete monolítico de componentes: cada pieza que se use debe existir bajo `web/src/components/ui/` y poder personalizarse junto al producto.

## Configuración

- `web/components.json` define el estilo `new-york`, JavaScript, Tailwind v4, el alias `@/*` y Lucide como biblioteca de iconos.
- `web/vite.config.js` habilita React, Tailwind y el alias `@`.
- `web/src/lib/utils.js` centraliza `cn`, usando `clsx` y `tailwind-merge`.
- `web/src/components/theme-provider.jsx` y `mode-toggle.jsx` implementan el cambio persistente entre claro, oscuro y sistema.
- Los componentes generados por shadcn deben permanecer en `web/src/components/ui/`.

Desde `web/`:

```bash
npx shadcn@latest add <component>
npm run build
```

El proveedor se monta en `web/src/main.jsx` y el header expone el selector de tema. Los tokens visuales viven en `web/src/styles.css` y los componentes consumen nombres semánticos como `bg-background`, `text-foreground` y `border-border`.

## Skill para asistentes

La skill versionada está en `.agents/skills/shadcn`. La skill auxiliar `migrate-radix-to-base` también se conserva porque forma parte de la instalación oficial y ayuda a mantener componentes compatibles con la base elegida.

## MCP de shadcn

El MCP es una herramienta del entorno de desarrollo, no un contenedor de Nodara. Para habilitarlo en Codex, añadir la siguiente configuración al archivo global `~/.codex/config.toml` y reiniciar Codex:

```toml
[mcp_servers.shadcn]
command = "npx"
args = ["shadcn@latest", "mcp"]
```

La configuración del proyecto (`components.json`) es la fuente de contexto que el MCP utiliza para resolver framework, aliases, iconos y componentes instalados. No se guardan credenciales ni configuraciones personales del asistente en Git.
