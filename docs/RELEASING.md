# Preparar y publicar una versión de Chuli

El CI prepara un candidato descargable. La publicación requiere despacho manual desde `main` con la versión aprobada y los identificadores del run y del artefacto correspondientes al mismo commit.

## Preparar el candidato

1. Actualizá `VERSION`, la entrada de `CHANGELOG.md` y las notas en `docs/releases/v<VERSION>.md`. Documentá cambios de configuración, compatibilidad del perfil y Manager, migraciones y pasos de reversión aplicables.
2. Corré [los checks del repositorio](TESTING.md). Para una migración, instalá la versión anterior en un perfil temporal con personalizaciones, aplicá la candidata, verificá el resultado y comprobá que el rollback restaura el estado esperado. Si cambia el formato de datos, registrá qué versiones pueden leer cada estado.
3. Confirmá que [Validate harness](../.github/workflows/validate.yml) pasó para el candidato. Su resumen informa los IDs, el SHA del commit y el SHA-256 del artefacto `release-candidate`. El paquete reúne el perfil, los tres Managers nativos, el instalador de Windows y `SHA256SUMS.txt`. Desde el mismo checkout, comprobalo con `python scripts/verify_release.py <VERSION> <directorio> --check`.

## Publicar con autorización

Confirmá la versión, el SHA de `main`, el run exitoso de tipo push, el ID del artefacto y el destino con la persona responsable. Despachá [Publish validated Chuli profile](../.github/workflows/release.yml) sobre `main`. El workflow comprueba que el artefacto aprobado pertenece al run y al SHA y sigue disponible, descarga ese ID exacto y verifica los paquetes, sus sumas y el tag. Configurá la revisión y las protecciones de ramas y tags desde GitHub según los permisos efectivos del repositorio.

| Situación | Qué hacer |
| --- | --- |
| Artefacto expirado, faltante o modificado | Prepará un candidato completo en un run exitoso. Volvé a comprobar el SHA y solicitá aprobación del nuevo ID antes de usarlo. |
| Cambió `main` antes de publicar | Volvé a preparar y aprobar el candidato actual. |
| Falló la publicación después de crear el tag | Comprobá que el tag sigue apuntando al SHA aprobado y que `main` coincide. Retomá el mismo despacho después de revisar el fallo. Si existe una release, el reintento comprueba el SHA-256 de todos sus archivos, incluido el manifiesto. Si hay un borrador, falta un hash o difiere del candidato, la resolución requiere revisión humana. |
| El tag apunta a otro commit | Detené la entrega y consultá a la persona responsable del repositorio. |
| Hay que revertir una actualización de perfil | Revisá la vista previa de `chuli rollback` en el cliente afectado, aplicá el rollback aprobado y verificá personalizaciones y datos restaurados. Para el ejecutable, usá la versión publicada anterior compatible. |

Los hashes comprueban que los bytes coinciden con el candidato preparado. Su procedencia depende de la integridad del workflow, del run aprobado y de quienes tienen permisos para modificarlos. Cualquier modificación de una release ya publicada requiere autorización específica.
