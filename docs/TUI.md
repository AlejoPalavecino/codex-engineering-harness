# Chuli Manager 1.0

El menú se abre con `chuli` o `chuli.exe`. La acción de release obtiene el paquete del repositorio oficial, muestra el plan y espera confirmación para instalarlo. Una actualización del perfil puede incorporar skills, recursos o modelos. Para cambiar de versión del ejecutable descargá el nuevo Manager.

| Operación | Uso |
| --- | --- |
| Ver la versión del Manager | `chuli version` |
| Ver el destino, modelo registrado e inventario de agentes y skills | `chuli status` |
| Previsualizar / aplicar instalación | `chuli install` / `chuli install --apply` |
| Previsualizar / aplicar actualización | `chuli update` / `chuli update --apply` |
| Comprobar la instalación | `chuli doctor` |
| Revisar / confirmar la reversión | `chuli rollback` / `chuli rollback --apply` |
| Usar un checkout de confianza para desarrollo o recuperación | `chuli --source RUTA doctor` |

La distribución oficial conserva una copia verificable del paquete para diagnóstico y recuperación. Para una instalación histórica que todavía carece de ese paquete, especificá un checkout compatible con `--source`. El indicador de versión del perfil y la versión del Manager se refieren a productos instalables distintos.

El evaluador continúa disponible en el checkout con `chuli --source RUTA eval check`. Los snapshots y comparaciones locales se describen en [Evaluación](EVALUATION.md).
