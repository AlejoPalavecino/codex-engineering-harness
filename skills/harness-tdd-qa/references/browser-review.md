# Comprobar una interfaz por tareas

Leer [harness-tdd-qa](../SKILL.md) y el [guion de diseño](../../harness-ux-ui/references/task-design.md). Trabajar sobre criterios aprobados y un candidato identificado; registrar entorno, ruta y datos sintéticos.

## Herramienta y alcance

Preferir ego-browser/egolite cuando esté disponible y autorizado. Leer su SKILL.md y ejecutar `ego-browser skill`; seguir la guía completa de esa versión. No copiar APIs de Playwright ni asumir que existen. No instalar, actualizar, elegir perfiles personales o acceder a producción por esta guía. Una alternativa disponible y autorizada puede cubrir el mismo guion; registrar el cambio. Si no hay navegador, continuar verificaciones disponibles y marcar las dependientes como NOT_TESTED.

Seguir el ciclo de vida del navegador: observar el estado antes de elegir objetivos, esperar resultados observables, respetar una toma de control del usuario y cerrar el espacio propio al completar. No reutilizar capturas o referencias de elementos de otro estado.

## Ciclo de comprobación

1. Precisar qué se quiere comprobar y qué evidencia lo demostraría. Elegir tamaños por uso esperado y puntos de ruptura afectados, sin exigir siete resoluciones o modo oscuro si no pertenece al producto.
2. **Apariencia:** capturas acotadas para revisar jerarquía, legibilidad, foco visible, recortes, solapamientos y adaptación. Inspeccionar las imágenes; guardarlas sin mirarlas no acredita QA visual. Medir lo que requiera medición; una impresión visual no certifica contraste.
3. **Interacción:** usar controles reales y teclado para completar la tarea, provocar errores relevantes y recuperarse. Revisar orden/foco, etiquetas, feedback y navegación; en superposiciones, cierre y retorno de foco cuando correspondan.
4. **Resultado:** comprobar el efecto esperado y su estado posterior. Si debe persistir, verificarlo con el mecanismo del proyecto; texto de éxito o clic despachado no prueban persistencia. No inferir autorización del backend desde un botón oculto.
5. Registrar hallazgo reproducible, impacto, evidencia y criterio afectado. Corregir solo dentro del alcance y volver a comprobar lo afectado sobre el candidato actualizado. Si una iteración no aporta progreso observable, revisar causa y restricciones antes de seguir; consultar solo decisiones pendientes.

## Informe

Devolver candidato/entorno, tarea y tamaño, acciones, esperado frente a observado, evidencia de apariencia/interacción/resultado, defectos y límites. Usar PASS/FAIL/PARTIAL/NOT_TESTED por criterio; no cerrar con bloqueantes ni asignar una nota estética arbitraria. Entregar evidencia al orquestador para TTP; no editar su tablero desde una revisión delegada.

La inspección del agente no demuestra comprensión humana ni una auditoría completa de accesibilidad. Indicar pruebas de tecnología asistiva y comprobaciones omitidas; nunca declarar cumplimiento completo por una captura o un análisis automático.
