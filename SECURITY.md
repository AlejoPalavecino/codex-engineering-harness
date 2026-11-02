# Seguridad

## Reportar vulnerabilidades

No publiques secretos, exploits activos ni datos personales en issues públicas. Para fallos con impacto real, contactá al mantenedor mediante el mecanismo privado de reporte de vulnerabilidades de GitHub si está habilitado. De lo contrario, abrí un issue **sin detalles explotables** solicitando un canal privado.

## Límites conocidos

- Los subagentes nativos pueden compartir permisos, filesystem y MCP con el proceso principal. Instrucciones como «no escribas» **no garantizan aislamiento**.
- Los MCP pueden realizar acciones externas no cubiertas por el sandbox de filesystem.
- El instalador modifica archivos del home de Codex solo después de `-Apply`, guarda un respaldo local y no toca memorias Engram.
- Los perfiles ASEP estrictos necesitan una implementación y prueba de aislamiento separadas; **no están garantizados en v0.1**.
- Los identificadores de modelos pueden no estar disponibles en todas las cuentas/versiones; comprobar el modelo ejecutado.

## Seguridad al contribuir

No añadir configuraciones privadas, archivos de clientes, prompts con datos sensibles, datos de salud, sesiones de navegador, tokens, `.env`, copias de `auth.json`, bases SQLite de Engram ni backups. Usar datos ficticios.
