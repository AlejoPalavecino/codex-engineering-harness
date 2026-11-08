# ODD · Organic Driven Development en Chuli

La unidad de trabajo es un cambio de comportamiento que se puede verificar y revertir. Chuli adapta el enfoque distribuido con Gentle AI™ v4.4.0 al orquestador de Codex. EWP decide el nivel de proceso, TTP conserva estado en el repositorio y ASEP define el aislamiento requerido. La versión fijada figura en el aviso de terceros del catálogo.

| Entrada | Trabajo mínimo | Evidencia para cerrar |
| --- | --- | --- |
| R0 · consulta | Formular la pregunta e inspeccionar sólo las fuentes que la resuelven | Ruta, símbolo o dato observado y límite de la respuesta |
| R1 · corrección | Precisar comportamiento actual y esperado, reproducir y corregir | Prueba focalizada y regresión relevante |
| R2 · funcionalidad | Acordar alcance y aceptación, localizar contratos y dividir por resultados | Tablero TTP con evidencia por tarea y comprobación de integración |
| R1+C o R2+C | Identificar el dato, permiso, contrato o efecto irreversible expuesto | Camino válido, rechazo sin efectos no deseados y verificación independiente cuando corresponda |

## Del objetivo al candidato

1. Expresá el resultado como algo que otro pueda observar. Para «corregir devolución de préstamos», un criterio podría ser «la segunda devolución informa que el préstamo está cerrado y conserva el saldo anterior».
2. Identificá el punto de entrada, el contrato consumido y la prueba más cercana. Si falta una decisión de producto, preguntá únicamente por esa decisión. Si existe una definición aprobada, guardá su ruta.
3. Escribí primero el caso que falla por el motivo esperado cuando exista un runner. Implementá el mínimo para que pase y verificá el comportamiento anterior. Al refactorizar, mantené las pruebas verdes. Para texto pasivo, verificá lectura, estructura y enlaces. Declarar «RED» requiere haber observado el fallo.
4. En trabajo sustancial, mantené la especificación estable y el progreso en un único tablero TTP. El autor de una unidad informa el comando, el resultado y el riesgo pendiente. El orquestador verifica y actualiza el estado. Si cambia el contrato, actualizá sólo la decisión y los criterios afectados.
5. Identificá el candidato por commit o por diff reproducible. Terminá los formateadores y generadores antes de verificar. Si luego cambia el candidato, repetí la comprobación afectada.

## Revisión y tamaño

Usá unas **400 líneas de autoría** por entrega como presupuesto de atención. Sumá altas y bajas por unidad, incluyendo código, tests y documentación escritos para el cambio. Por ejemplo, 180 líneas nuevas más 70 eliminadas suman 250. Contá aparte código vendorizado, lockfiles y salidas generadas, e informá su impacto. Para cambios preparados usá `git diff --cached --numstat`; para un candidato versionado, `git diff --numstat BASE...HEAD`. Un diff omite archivos nuevos sin agregar: inventarialos al proyectar.

Si la proyección supera el presupuesto, intentá separar un comportamiento completo con sus pruebas y documentación. Si no hay un corte que conserve el contrato, explicá el motivo y el tamaño al revisor. No sacrifiques regresiones, legibilidad ni líneas de contexto para cumplir una cifra. Las reglas de publicación y las excepciones formales son las del proyecto de destino.

## Uso del contexto y recuperación

Leé metadatos antes de abrir skills, pasá rutas exactas al delegado y solicitá la mínima salida verificable. Dos investigadores que leen lo mismo duplican trabajo. Separá escritores por superficies independientes y usá revisión externa al escritor para riesgos identificados. Dos respuestas del mismo contexto no acreditan dos revisiones ciegas.

Al retomar, empezá por el tablero, el diff y los criterios. Volvé a ejecutar una prueba si la evidencia no corresponde al candidato actual. Engram es una ayuda opcional y debe resolverse por proyecto. Un cambio en el presupuesto o en el modelo necesita evidencia de costo o calidad en ese entorno. La evaluación histórica de v0.1.3 sigue siendo su propia línea de base.
