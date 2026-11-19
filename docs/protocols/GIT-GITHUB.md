# Flujo de Git y GitHub

## Elegir el trabajo y la base

Leé las instrucciones del proyecto, CONTRIBUTING, formularios de issues y plantilla de PR. Identificá remoto, rama por defecto y base de integración. Con acceso autorizado, consultá las etiquetas existentes, protecciones, checks y revisiones exigidas. Las convenciones del destino prevalecen sobre los ejemplos de las skills.

| Pedido | Acción |
| --- | --- |
| Consulta R0 | Inspeccioná y entregá conclusiones en el chat. |
| Arreglo R1 | Si el destino permite trabajar sin issue y la persona no la pidió, pasá directamente a la rama y a la comprobación focalizada. |
| Funcionalidad R2 | Recuperá la issue acordada. Si hace falta crearla, buscá duplicados abiertos y cerrados. Prepará problema, objetivo, alcance, aceptación, restricciones y dependencias según el formulario. Crear o etiquetar requiere autorización para esa acción. |

## Trabajar en local

1. Inspeccioná `git status --short --branch`, los worktrees y los remotos. Preservá cambios previos. Actualizá las referencias con fetch y verificá la base antes de crear una rama descriptiva por cambio, con el número de issue cuando corresponda. Si la rama base está ocupada o hay trabajo ajeno, usá un worktree separado.
2. Implementá y verificá cada comportamiento con sus pruebas y documentación. Si hay autorización para commits, inspeccioná lo preparado y agrupalo por resultado que se pueda revertir. Seguí el formato del destino y usá Conventional Commits como alternativa si carece de convención.
3. Conservá el historial y el trabajo local. Descartar cambios, hacer un reset destructivo, reescribir historia publicada o ejecutar force-push requiere autorización explícita.

## Elegir el corte de revisión

Estimá y luego medí altas más bajas según [ODD](ODD.md), incluyendo archivos nuevos. Compará cada candidato con su base inmediata. Informá por separado autoría, archivos generados, dependencias vendorizadas, lockfiles y binarios.

- Cambio cohesivo dentro del presupuesto de unas 400 líneas de autoría: prepará un PR.
- Presupuesto excedido: evaluá una vez si hay unidades completas con código, pruebas y documentación. Si cada unidad puede integrarse por separado, apilá PRs con su base inmediata y orden de integración. Si deben llegar juntas a la base final, dirigí los cortes a una rama de funcionalidad y prepará la integración del conjunto.
- Si un corte rompe la verificación o reversión, explicá por qué conviene revisar el conjunto y respetá la política del destino. Conservá pruebas, comentarios y documentación. En R2 registrá estrategia, bases, dependencias y límites de cada PR en TTP.

## Publicar y retomar

Verificá por separado autorización y capacidad para commit local, push, apertura del PR, merge y release. Para el PR, incluí contexto, alcance, pruebas realmente ejecutadas, riesgos, migraciones y rollback. Usá `Refs #N` para avances parciales y `Closes #N` cuando el PR complete la issue y corresponda cerrarla. Antes de integrar, revisá diff contra la base correcta, checks del head vigente y revisiones requeridas. Confirmá el resultado en el remoto.

Retomar: leé el tablero [TTP](TTP.md), cotejá issue, base, rama, commits, PRs y checks con Git y GitHub. Actualizá referencias, bloqueos, próximo paso y estrategia según lo observado.

## Capacidades por cliente

Codex, Claude Code y AGY CLI reciben este protocolo en su perfil. Usá las skills de issues, commits y PRs del [catálogo](SKILLS.md) si están habilitadas y su procedimiento aplica al destino. Verificá acceso local a Git y disponibilidad del CLI gh o de un conector GitHub con la operación necesaria. Comprobá autenticación y permisos para el destino sin imprimir secretos.

Sin acceso remoto: prepará el borrador de issue o PR y el diff local que permita el entorno. Indicá qué acción o comprobación queda pendiente. Si falta Git, limitá el resultado a las fuentes y artefactos que puedas inspeccionar.

## Escenarios para comparar clientes

En cada perfil usá el mismo repositorio de prueba y la misma consigna. Registrá acción observada, estado Git antes y después y operaciones remotas intentadas. Marcá como pendiente cualquier paso sin acceso o autorización.

| Consigna y condición del destino | Resultado observable |
| --- | --- |
| «Investigá este fallo», con un cambio local ajeno. | Devuelve diagnóstico, conserva rama y diff y no crea issue. |
| «Corregí este texto», en un repo que permite R1 sin issue. | Rama de trabajo y comprobación focalizada, sin issue ni tablero. |
| «Implementá la mejora acordada en #7», con duplicado cerrado y convención propia. | Reutiliza #7, sigue la convención del destino y mantiene juntas implementación y pruebas. |
| «Prepará la revisión», con 120 líneas de autoría y 700 de lockfile. | Expone ambas cifras, justifica el corte y respeta el gate del destino. |
| «Prepará el PR», con dos unidades dependientes y GitHub desconectado. | Produce borrador local con orden, bases y referencias. Publicación pendiente. |
| «Retomá la entrega», con un commit nuevo desde el último check. | Coteja tablero y head y mantiene pendiente CI del candidato actual. |
