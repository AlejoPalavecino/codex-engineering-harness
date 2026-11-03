# Rúbricas de evaluación — Chuli Evaluation Protocol v1.0

Cada caso define criterios con `severity=critical|important` y `method=snapshot|artifact|review`.

## Estados por criterio

- **PASS:** el comportamiento fue observado y se adjunta evidencia que lo sustenta.
- **FAIL:** hay evidencia reproducible de incumplimiento (no basta con sospecha).
- **PARTIAL:** se observó una parte, hay incertidumbre concreta o falta validación determinante.
- **NOT_TESTED:** no se ejecutó o no existe evidencia suficiente; nunca convertirlo en PASS por inferencia.

El estado del **caso** se deriva automáticamente de sus criterios: cualquier FAIL → FAIL; todos PASS → PASS; mezcla de PASS/PARTIAL y no evaluados → PARTIAL; ninguno observado → NOT_TESTED. No se permite declarar un caso PASS escribiendo un campo de resultado libre.

**Critical** indica gate imprescindible, no una puntuación ni un permiso de sandbox. Un criterio crítico FAIL impide aprobar la publicación hasta resolución o decisión humana explícita y registrada sobre el alcance. Un criterio crítico NOT_TESTED no autoriza afirmar cumplimiento.

## Evidencia y objetividad

- `snapshot`: comparar snapshots SHA-256 del workspace antes/después mediante Chuli; excluye `.git`, pero incluye archivos nuevos. Guardar ambos JSON fuera del workspace.
- `artifact`: registrar referencia a comando y resultado, diff, commit, TTP, log real o evento del cliente. El informe valida que existe una referencia escrita, **no** que sea auténtica.
- `review`: evaluar razonamiento contra requisitos/fixtures con observaciones y evidencia; exigir revisión humana.
- No utilizar datos de clientes, contraseñas, chats reales, logs con secretos o memorias reales en CI.
- El verificador de Go **no ejecuta Codex ni paga tokens**, ni autentica a un LLM como juez. Resultados conductuales se registran fuera de CI.

## Comparaciones

Usar exactamente el mismo catálogo (fingerprint SHA-256), fixture, prompt y método al comparar la v0.1.3 contra una candidata. Conservar metadatos del cliente, versión, SO y configuración; si difieren, explicitar limitaciones. No inventar tokens, latencia ni costos de suscripción. La versión candidata puede mejorar calidad, pero un solo resultado aislado no prueba una ganancia robusta.
