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

**Límite del harness:** no se suministra un sandbox propio ni un mecanismo portable de aislamiento de rutas. La validación y los permisos de Codex deben comprobarse en cada entorno.
