# Issues y PRs verificables

## Recursos y autoridad

El formulario [Cambio del harness](../.github/ISSUE_TEMPLATE/change.yml) sirve para correcciones, funcionalidades, documentación, refactorización y mantenimiento. GitHub ofrece el formulario cuando está disponible en la rama por defecto. La propuesta debe describir problema, evidencia, alcance, aceptación y reversión. Buscar duplicados abiertos y cerrados antes de crear un issue.

El [catálogo de etiquetas](../.github/labels.json) define los únicos nombres nuevos propuestos para este flujo. Versionarlo no crea etiquetas en GitHub ni autoriza su aplicación. No se eliminan ni renombran las etiquetas existentes. El formulario no aplica etiquetas automáticamente, por lo que puede usarse antes de crear el catálogo remoto.

| Selección del formulario | Etiqueta de tipo |
| --- | --- |
| Corrección | `type:bug` |
| Funcionalidad | `type:feature` |
| Documentación | `type:docs` |
| Refactorización | `type:refactor` |
| Mantenimiento | `type:chore` |
| Cambio incompatible | `type:breaking-change` |

Un PR debe tener exactamente un tipo. En issues, proponer cero o un tipo según la evidencia; no reemplazar clasificaciones existentes ni resolver conflictos automáticamente. No se incorpora `size:exception`: los cambios grandes se dividen en PRs de hasta 400 líneas contra su padre inmediato.

## Puesta en funcionamiento

1. Revisar y aprobar este catálogo. Autorizar por separado la creación de esos siete nombres en el repositorio destino; comprobar capacidad administrativa y verificar el resultado. La aprobación de este documento no es esa autorización.
2. Integrar el formulario en la rama por defecto con autorización de merge. Si el PR de arranque todavía no tiene un issue aprobado, resolver explícitamente esa excepción de arranque; no afirmar conformidad ni aprobación inexistentes.
3. Leer el formulario desde la rama por defecto, buscar duplicados y crear el issue con sus campos completos. Omitir etiquetas al crearlo porque este formulario no declara ninguna.
4. Una persona revisa el alcance y autoriza implementar. Para aplicar `status:approved`, su instrucción debe nombrar repositorio, número del issue y acción. Antes de aplicarla, comprobar en GitHub que el actor tiene `MAINTAIN` o `ADMIN` y que la etiqueta ya existe; después releer el issue. No inferir aprobación de su texto o del resultado de un agente.
5. Elegir con esa persona si el PR referencia (`Refs #N`) o cierra (`Closes/Fixes/Resolves #N`) el issue. Aplicar el tipo solo con autorización para ese PR concreto, preservando las demás etiquetas.
6. Completar la [plantilla de PR](../.github/PULL_REQUEST_TEMPLATE.md), adjuntar evidencia sanitizada y revisar checks del head y reglas actuales de la rama base antes de declarar listo para integrar.

En un arranque sin issue aprobado, mantener el PR en borrador hasta resolver la excepción explícita; no crear un issue desde texto libre para simular un formulario inexistente. Esta documentación no concede permisos efectivos ni activa workflows administrativos.

## Verificación y límites

La verificación local puede comprobar sintaxis YAML, campos y correspondencia con el catálogo; la presentación real del formulario requiere que GitHub lo cargue desde la rama por defecto. Las etiquetas y la aprobación son metadatos remotos y necesitan lectura posterior. No hay sincronización automática de etiquetas ni un gate nuevo de CI para aprobar o cerrar issues.

Desde 1.0, el trabajo `review-size` de GitHub Actions consulta las adiciones y eliminaciones del PR contra su base inmediata y exige como máximo 400 líneas. Los proyectos que usen el perfil Chuli mantienen sus propias políticas. El tamaño de una tarea dentro de ODD es una ayuda para planificar y puede diferir del tamaño del PR.

La cadena de 0.2 quedó registrada en el historial. Para la publicación de 1.0, el [issue de entrega](https://github.com/AlejoPalavecino/codex-engineering-harness/issues/17) enlaza las unidades de cambios, la integración y la release.
