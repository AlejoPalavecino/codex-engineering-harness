# Contribución al repositorio del harness

Este `AGENTS.md` corresponde al **repositorio fuente**. El archivo que se instala en el home de Codex es `config/AGENTS.md`; no los confundas.

- Mantener EWP R0/R1/R2 + C, MRP, ASEP, TKA y CES como contratos documentados.
- Cambios de agentes, modelos, permisos o activación de skills deben actualizar documentación y pruebas.
- No incluir valores de `config.toml` personales, tokens, credenciales, logs, backups ni memoria Engram.
- No modificar herramientas de terceros ni vender instrucciones de prompts como permisos efectivos.
- Preferir implementaciones sencillas y comprobables. Evitar duplicar skills.
- Todo cambio importante debe contar con criterios verificables y una ruta de reversión.
- El repositorio es público; antes de hacer commit, revisar secretos, paths personales y contenido con licencias de terceros.
- En los PR incluir alcance, riesgo, pruebas ejecutadas y límites conocidos. No declarar verificado lo que no se probó.
