# Recursos TTP instalados y tableros canónicos

**Estado:** diseño e implementación autorizados por el usuario el 2026-11-05. Alcance: distribución de plantilla/guía, referencias instaladas y verificación de generación, instalación y rollback. No cambiar modelos, esquema/estados de TTP ni instalación personal; sin release, merge o deploy.

## Problema y diseño aprobado

En `r2-implementation-mxc`, el agente no encontró la plantilla/protocolo en el home de laboratorio y generó un tablero legible con formato rechazado por `taskboard.py`. La propuesta aprobada es entregar los recursos canónicos existentes junto al validador y hacerlos localizables antes de crear/cerrar un tablero R2.

| Fuente única | Destino relativo a CODEX_HOME |
| --- | --- |
| `scripts/taskboard.py` | `harness/bin/taskboard.py` (existente) |
| `templates/feature-tasks.md` | `harness/templates/feature-tasks.md` |
| `docs/protocols/TTP.md` | `harness/docs/protocols/TTP.md` |

La guía conserva su enlace relativo a la plantilla; no se crean copias mantenidas manualmente en el repositorio. `config/AGENTS.md` y la skill de especificación indicarán las rutas instaladas, el formato canónico y la validación previa al cierre. R0/R1 conservan su proporcionalidad. Los recursos usan el manifiesto, hashes, backup, preservación y rollback actuales; no se agrega otro gestor ni cambia el esquema del manifiesto.

## Criterios de aceptación

1. Dry-run no escribe el home sintético. Apply instala recursos idénticos a sus fuentes y registra archivos creados/modificados con hashes.
2. La plantilla instalada permite crear un tablero válido para el validador instalado. Las instrucciones enlazan rutas presentes y la guía distingue uso en repo/home.
3. Reinstalación no duplica ni altera recursos sin cambios. Personalizaciones de plantilla, guía e instrucciones se conservan por defecto; reemplazo explícito crea backup y rollback las restaura.
4. Rollback retira recursos nuevos/restaura anteriores, preserva archivos ajenos y rechaza una reversión insegura antes de mutar si cambió un recurso administrado.
5. Una nueva sesión real con el mismo fixture/prompt/rúbrica de implementación y el paquete candidato genera un tablero canónico aceptado por `taskboard.py`, con IDs, dependencias y evidencia. Se conservan el registro anterior no canónico y las limitaciones de la comparación.
6. Pilotos R0/R1 mantienen trabajo proporcional y contratos; Go/Python/instalador/rollback/CI no introducen regresiones verificables. Revisión independiente de candidato identificable.

## Verificación y límites

Pruebas deterministas test-first: referencias/estructura portable y contratos PowerShell en homes descartables. Para la generación conductual se instala el candidato en otro home, con permisos limitados ya comprobados y credencial temporal autorizada; no se corrige el tablero generado para aprobarlo. Una muestra no garantiza comportamiento universal. Si no se dispone de un cliente operable, generación queda NOT_TESTED, sin sustituirla por una plantilla rellenada por el observador.

Evidencia completa privada fuera de Git; resumen sanitizado y tareas en [registro TTP](../../odd/tasks/chuli-ttp-resources.md). Reversión del código: retirar únicamente entradas de recursos, referencias y sección de uso instalada; el rollback de instalaciones se prueba con su manifiesto real.
