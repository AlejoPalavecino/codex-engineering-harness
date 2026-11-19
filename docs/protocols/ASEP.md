# ASEP v1.0 — Agent Security & Execution Policy

Los permisos declarados en prompts **no son** controles de seguridad del filesystem ni del sistema externo.

## Perfiles conceptuales

| Perfil | Acceso previsto |
| --- | --- |
| observe | Lectura, sin mutaciones |
| test-author | Escritura sobre pruebas/fixtures aprobados |
| code-author | Escritura sobre implementación aprobada; tests protegidos |
| workspace-edit | Cambios acotados en el workspace |
| delivery | Operaciones de entrega explícitamente autorizadas |

## Modos

**Colaborativo:** subagentes nativos y comprobación de diff, para riesgo normal. No garantiza separación de escritura real.

**Aislado:** entornos/sesiones separadas con sandbox y herramientas MCP restringidas, para TDD estricto o cambios críticos. Se deben verificar los permisos efectivos antes de afirmar aislamiento.

## Controles mínimos

- Proteger las pruebas de aceptación aprobadas frente al implementador; si una prueba es incorrecta debe revisarla su autor.
- El verificador no debe modificar el candidato; si ejecutar pruebas genera artefactos, usar entorno aparte.
- MCP de bases de datos, GitHub, navegador y memoria tienen permisos propios que no heredan necesariamente el sandbox.
- Comprobar integridad de tests protegidos mediante hashes/diff fuera del implementador.
- Prohibir acceso productivo, push, merge o deploy sin autorización humana explícita.
- En repositorios importantes usar revisión de PR y checks requeridos; limitar permisos de tokens CI.

## Seguridad y resiliencia en el diseño

1. Identificar datos sensibles, integraciones, operaciones irreversibles y fronteras donde cambia quién controla la entrada o puede escribir. Confirmar permisos mínimos, credenciales necesarias y estados que deben preservarse ante un rechazo.
2. Para cambios C, modelar las amenazas antes de implementar. Por cada escenario relevante, registrar activo, actor, frontera o entrada, efecto, control, prueba y riesgo residual. Incluir usos indebidos de funciones válidas, como ejecutar un rollback con un respaldo que ya no coincide.
3. Revisar dependencias y su procedencia, rutas de exposición de secretos, permisos efectivos y configuración crítica. Combinar análisis automatizado con revisión humana del uso real, los hallazgos y su alcance. Un análisis sin hallazgos indica lo que cubrió la herramienta en esa ejecución.
4. Probar con datos sintéticos los rechazos, fallos parciales, reintentos y recuperaciones aplicables. Ante concurrencia, comprobar la protección elegida y el estado tras dos operaciones superpuestas. Ante una dependencia externa, comprobar qué pasa si tarda, falla o responde con datos inesperados. Documentar cómo reconocer y recuperar un estado incompleto.
5. Para C, pedir verificación independiente del candidato, incluyendo límites efectivos de sus permisos y resultados negativos. Cualquier excepción a un control exige decisión humana, alcance registrado y evidencia del riesgo residual.

Para un cambio rutinario, el resultado puede quedar en el diff y la evidencia del PR. En un cambio crítico, vinculá las decisiones y pruebas al registro de trabajo existente para que se puedan revisar y repetir.

**Límite del harness:** no se suministra un sandbox propio ni un mecanismo portable de aislamiento de rutas. La validación y los permisos de Codex deben comprobarse en cada entorno.
