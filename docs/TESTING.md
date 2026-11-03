# Pruebas y validación

## Checks automáticos

```bash
python tests/validate_repository.py
```

GitHub Actions ejecuta validaciones estructurales. Windows también ejecuta pruebas de instalación, verificación y rollback sobre un home descartable.

## Pruebas manuales dentro de Codex

1. **R0:** pedir análisis de un repo sin autorizar cambios; comprobar ausencia de modificaciones.
2. **R1:** cambiar una etiqueta trivial; no debe generar specs ni subagentes.
3. **R2:** describir un sistema clínico; debe preguntar actores, datos, permisos, alcance y UX antes de implementar.
4. **Modelos:** comprobar los modelos efectivamente usados por los subagentes, no solo su TOML.
5. **TDD:** crear un caso comprobable; validar RED/GREEN y protección de tests cuando exista aislamiento.
6. **Visual:** revisar flujos y pantallas, estados, responsive y accesibilidad con herramienta adecuada.
7. **Memoria:** recuperar una decisión del proyecto correcto sin modificar memorias por accidente.
8. **Regresión:** comprobar que MCP, plugins y hooks existentes siguen registrados.

## Limitaciones

Los checks de sintaxis no prueban el razonamiento del modelo. El entorno de GitHub Actions no reproduce todas las versiones de Codex Desktop ni prueba su sandbox. Un resultado `PASS` está limitado a la comprobación informada.
