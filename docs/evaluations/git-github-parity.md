# Verificación del flujo Git y GitHub en tres clientes

Se evaluó [#64](https://github.com/AlejoPalavecino/codex-engineering-harness/issues/64) sobre el head público `d4b4e784778983b601634cb837efe00aca161527` del [PR #65](https://github.com/AlejoPalavecino/codex-engineering-harness/pull/65). El PR sigue en borrador con `Refs #64`. El [CI del PR](https://github.com/AlejoPalavecino/codex-engineering-harness/actions/runs/44620447559) pasó para ese head. La revisión humana y las etiquetas del issue y del PR siguen pendientes.

## Batería local

Se instalaron los tres perfiles en directorios temporales. Las instrucciones de entrada resolvieron el mismo protocolo en los tres. El [guion local](../../tests/git_workflow_fixtures.py) ejecutó 18 casos con Git, seis por perfil, en repositorios sintéticos nuevos sin remotos configurados. Pasale las tres rutas de los perfiles instalados. El entorno de Git excluye la configuración del sistema y del usuario.

| Caso | Evidencia local en los tres perfiles |
| --- | --- |
| R0 con edición ajena | Diagnóstico por inspección. Rama, HEAD, refs y diff idénticos antes y después. |
| R1 sin issue | Rama `fix/widget-copy` desde `trunk`, diff de un archivo y comprobación focalizada. |
| R2 con convención propia | Se usó #7 desde una instantánea con un duplicado cerrado. Rama `topic/007-widget` y commit `TASK 007` con código, prueba y documentación. |
| Presupuesto de revisión | Git midió 120 líneas de autoría y 700 de lockfile. Se mantuvo un corte cohesivo y se informó el total de 820 para la política del destino. El aviso se cubrió por separado con la API simulada. |
| PRs dependientes sin GitHub | El borrador local dejó dos `Refs #7` parciales, el orden y las bases inmediatas. La integración quedó pendiente. |
| Continuidad con HEAD nuevo | Se cotejó la rama y el resultado previo de CI contra el HEAD nuevo. El resultado anterior quedó obsoleto y el CI actual quedó pendiente. |

Repetí la batería con Python y Git. Cada argumento apunta al directorio temporal donde instalaste el perfil correspondiente con el instalador del repo.

```text
python tests/git_workflow_fixtures.py --codex-home <perfil-codex> --claude-home <perfil-claude> --agy-home <perfil-agy>
```

## Sesiones de cliente

Se intentó R0 con el mismo texto y repositorio sintético, sin inyectar credenciales ni permitir operaciones remotas de Git. Los resultados restantes de cada cliente quedaron sin ejecutar por el bloqueo de su preflight.

| Cliente | Resultado observado |
| --- | --- |
| Codex CLI 0.170.0 | El perfil aislado recibió HTTP 401 de Responses. Git quedó idéntico. |
| Claude Code | El binario no estaba en PATH ni en las rutas de instalación habituales comprobadas. Sí se instaló y verificó el perfil en disco. |
| AGY CLI 1.5.10 | El cliente inició un turno. El modo no interactivo en plan denegó el comando, la respuesta quedó vacía y Git quedó idéntico. La fuente de autenticación del cliente quedó sin verificar. |

## Corrección y verificación local

Se alinearon la skill `chained-pr` y su referencia con el protocolo común. La decisión de corte usa líneas de autoría, cada PR apunta a su base inmediata y la política del destino gobierna los límites sobre el diff total y las excepciones. La prueba de contrato falló para los tres perfiles con la guía anterior y pasó tras el ajuste.

- `python -m unittest discover -s tests -p test_git_workflow.py` verifica la distribución del protocolo, las cuatro skills y la guía de cortes en los tres perfiles.
- `python -m unittest discover -s tests -p test_review_budget.py` ejecuta el check con la API simulada para 400 y 401 líneas y para datos incompletos o de otro head.
- `pwsh -NoProfile -File tests/windows-multiclient.ps1` comprueba instalación, verificación y rollback en los tres perfiles.

La batería local establece resultados esperados reproducibles con Git y la comprobación de recursos acredita su distribución. La evaluación de conducta queda como no disponible en este entorno, según el criterio de #64. Para aportar esa evidencia adicional, hacen falta binarios, acceso de prueba y permisos acotados. El CI publicado corresponde al head remoto indicado, previo a esta corrección local.
