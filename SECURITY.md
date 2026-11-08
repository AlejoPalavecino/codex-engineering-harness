# Seguridad

## Reportes

Usá el canal privado de vulnerabilidades de GitHub si está habilitado. En su defecto, abrí un issue solicitando un canal privado. Conservá los secretos, memorias, perfiles reales de Codex y datos de clientes fuera del repositorio.

## Límites de confianza

El Manager obtiene los paquetes de publicación de este repositorio por HTTPS y contrasta su SHA-256 con el manifiesto de la release. Una suma emitida por el mismo origen comprueba integridad de transporte y caché. La autenticidad sigue dependiendo de GitHub y de quienes tienen permiso de publicación. Usá un checkout local únicamente cuando conozcas su procedencia.

El instalador hace una vista previa, preserva personalizaciones, verifica recursos y registra las operaciones reversibles. Diagnóstico y rollback rechazan recursos incompatibles o respaldos dañados. Las skills de terceros son instrucciones y recursos de procedencia revisable. Algunas precisan herramientas o permisos adicionales, detallados en el catálogo.

Los subagentes pueden heredar el sandbox, los MCP y las autorizaciones del proceso principal. Para una tarea crítica configurá controles efectivos en el cliente y verificá sus límites. El perfil da instrucciones de trabajo al agente. La separación de tareas y la revisión de un diff no constituyen aislamiento del sistema operativo.
