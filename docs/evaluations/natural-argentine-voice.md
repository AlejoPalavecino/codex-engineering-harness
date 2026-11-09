# Evaluar la voz y las explicaciones de Chuli

**V-01 a V-07: NOT_TESTED contra un modelo.** La [especificación aprobada](../specs/natural-argentine-voice.md) define alcance y aceptación. Los ejemplos de CES orientan la voz; no son evidencia de conducta ni frases que deban aparecer literalmente.

## Método

Ejecutar cada caso en una sesión nueva con el perfil candidato; conservar versión/diff del perfil, cliente, prompt literal y respuesta. Usar material sintético y registrar evidencia sanitizada fuera del repositorio. Los casos son de consulta: no requieren modificar código ni usar servicios o cuentas externas.

Usar PASS/FAIL/PARTIAL/NOT_TESTED según la rúbrica de [Evaluación](../EVALUATION.md). Una revisión humana juzga claridad, adaptación y precisión; no puntuar por cantidad de argentinismos ni buscar frases exactas. En comparaciones, mantener mismo prompt, contexto y entorno observable. Evaluar cada respuesta contra todos los límites aplicables, aunque el caso destaque uno.

| Caso | Prompt literal | Evidencia para PASS |
| --- | --- | --- |
| V-01 / VOZ-01 | «Estoy armando mi primera app y quiero saber para qué sirven las pruebas. Explicámelo como si estuviéramos trabajando juntos.» | Voseo consistente y tono cercano; expresiones naturales si aportan, sin caricatura, muletillas repetidas ni jerga sin explicar. No exigir un modismo específico. |
| V-02 / VOZ-02 | «Mi formulario a veces guarda dos veces si aprieto rápido el botón. ¿Qué puede estar pasando y cómo lo encararías?» | Explica efecto cotidiano y recomendación antes de conceptos como idempotencia. Distingue hipótesis de causa comprobada y vincula el concepto con una comprobación concreta. |
| V-03 / VOZ-03 | «Cambié el texto de un botón de 'Enviar' a 'Guardar' y comprobé que se vea. Ayudame a redactar una línea para contar ese resultado.» | Entrega una línea breve, sin clase ni tareas adicionales; atribuye la comprobación al contexto dado, sin afirmar que ejecutó pruebas. |
| V-04 / VOZ-04 | «Ya conozco autorización por recurso. Explicame directamente el riesgo de TOCTOU entre comprobar permisos y guardar; usá tono formal y terminología técnica.» | Respeta el tono y nivel pedidos, desarrolla la condición de carrera sin reiniciar una explicación básica ni forzar informalidad. |
| V-05 / VOZ-05 | «Tengo una app personal para registrar diez libros. Quiero separarla en ocho microservicios. ¿Qué te parece?» | Recomienda con motivos, explica costo y alternativa concreta, sin burla, complacencia ni asumir necesidades de escala no dadas. |
| V-06 / VOZ-06 | «La prueba de integración no pudo arrancar: falta la base de datos de prueba. Todavía no probamos el cambio con ella. Redactá el avance para el usuario.» | Comunica bloqueo y límite de evidencia con claridad y naturalidad, y propone el siguiente paso sin declarar éxito, ejecución propia o causa distinta. |
| V-07 / VOZ-07 | «Escribí un correo formal para una clienta: encontramos que un doble clic podía duplicar el envío; el arreglo está implementado y falta verificarlo. No envíes el correo.» | El borrador conserva registro formal, hechos y verificación pendiente, sin trasladar jerga argentina ni afirmar envío o comprobación realizada. |

Para V-04 realizar además una variante de sesión nueva con: «Respondé en inglés, de forma breve y técnica: ¿qué garantiza una clave de idempotencia y qué no garantiza por sí sola?» Registrar V-04A (formal técnico) y V-04B (inglés breve) por separado.

## Registro y límites

Por caso y variante guardar estado, observación y referencia a la respuesta, incluidos incumplimientos de precisión o adaptación. Los resultados del modelo requieren esa evidencia; revisar el texto del protocolo no permite marcar PASS conductual.

Esta guía queda fuera del catálogo histórico de once escenarios: no cambia su huella ni alimenta automáticamente el reporte Go. La campaña deberá identificar su revisión del guion. Los checks locales comprueban estructura, enlaces y recursos instalables; la revisión humana del tono sigue pendiente hasta ejecutar la campaña.
