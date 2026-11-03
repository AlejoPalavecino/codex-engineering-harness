# Pruebas y validación

## Checks automáticos

```bash
python tests/validate_repository.py
python -m unittest discover -s tests -p 'test_*.py'
```

GitHub Actions ejecuta validaciones estructurales y tests Python de TOML. Windows realiza smoke y regresiones sobre un CODEX_HOME descartable.

### Regresiones específicas de v0.1.2

- P1-1: `[tools.update_plan] # comentario` con `enabled = false` no debe duplicarse ni destruir una preferencia local.
- P1-2: falta de `_previous_state.json` debe impedir TODO el rollback, con hashes intactos.
- P1-3: personalizaciones en AGENTS se conservan salvo autorización `-ReplaceCustomized`.
- P2-1: cambios válidos ajenos al harness no deben confundirse con TOML inválido; TOML inválido debe fallar aun con hash coincidente.

La validación conductual de agentes, modelos y aislamiento permanece en v0.2/v0.3.

## Pruebas de Chuli Manager v0.1.3

```bash
go test ./...
go build ./cmd/chuli
```

En Windows probar una instalación con CODEX_HOME temporal: navegación, vista previa, cancelación sin cambios, confirmación, verificación y rollback. Las releases anteriores a v0.1.2 deben rechazarse para instalaciones automáticas. El Manager no comprueba la calidad de los agentes ni las garantías de su sandbox.

## Pruebas manuales dentro de Codex

1. **R0:** pedir análisis de un repo sin autorizar cambios; comprobar ausencia de modificaciones.
2. **R1:** cambiar una etiqueta trivial; no debe generar specs ni subagentes.
3. **R2:** describir un sistema clínico; debe preguntar actores, datos, permisos, alcance y UX antes de implementar.
4. **Modelos:** comprobar los modelos efectivamente usados por los subagentes, no solo su TOML.
5. **TDD:** crear un caso comprobable; validar RED/GREEN y protección de tests cuando exista aislamiento.
6. **Visual:** revisar flujos y pantallas, estados, responsive y accesibilidad con herramienta adecuada.
7. **Memoria:** recuperar una decisión del proyecto correcto sin modificar memorias por accidente.
8. **Regresión:** comprobar que MCP, plugins y hooks existentes siguen registrados.
9. **TTP:** R1 no genera tablero; R2 aprobada genera tareas con dependencias, estados y evidencia.
10. **Progreso nativo:** en ejecución R2 fuera de Plan Mode, comprobar que `update_plan` está disponible y se refleja en el cliente real (Codex Desktop y/o CLI). Registrar si no se muestra.
11. **Recuperación:** interrumpir una tarea IN_PROGRESS, reiniciar la sesión y corroborar estado real contra Git antes de continuar.
12. **Integridad TTP:** probar DONE sin evidencia, dependencias pendientes y bloqueo sin causa; el validador debe rechazarlos.

## Limitaciones

Los checks de sintaxis no prueban el razonamiento del modelo. El entorno de GitHub Actions no reproduce todas las versiones de Codex Desktop ni prueba su sandbox. Un resultado `PASS` está limitado a la comprobación informada.
