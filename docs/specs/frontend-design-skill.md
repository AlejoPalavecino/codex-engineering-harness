# Integración de frontend-design en el catálogo de Chuli

## Alcance y decisiones aprobadas

El usuario solicitó incorporar `frontend-design` de [anthropics/skills](https://github.com/anthropics/skills) y aprobó el 2026-11-16 integrarla al paquete oficial (Opción A, siguiendo la convención de `diagram-design`).

Se distribuye `skills/frontend-design/` desde `anthropics/skills`, commit `e4bb5d960d87126ff3b3a8af1a104b7db67560ba` (`2026-11-16T14:22:47Z`), bajo licencia **Apache-2.0**: `SKILL.md` (`9.390` bytes) y `LICENSE.txt` (`10.174` bytes) sin modificaciones, junto a `UPSTREAM.json` con commit, licencia y hashes SHA-256.

## Diseño de integración y límites

- **`harness-ux-ui` sigue siendo la raíz de UX de producto:** actores, diseño por tareas, estados (`empty/loading/error/denied`), accesibilidad, teclado/foco, aprobación humana y revisión en navegador.
- **`frontend-design` se carga condicionalmente como complemento visual web/frontend:** dirección estética, tipografía, paleta y redacción de interfaces web/CSS evitando valores por defecto genéricos, subordinada a las decisiones y restricciones del proyecto.
- **Portabilidad multi-cliente:** al vivir bajo `skills/frontend-design/`, `scripts/profile_resources.py`, `scripts/package_release.py` y el Manager la distribuyen para **Codex**, **Claude Code** y **AGY CLI** (bajo `plugins/chuli/skills/frontend-design/`), y aparece en la gestión de skills de `chuli`.
- **Catálogo:** el perfil completo pasa de 24 a **25 skills** (5 nativas + 18 de Gentle AI™ v4.4.0 + `diagram-design` 2.6 + `frontend-design`). Se actualizan las referencias activas del catálogo y el manifiesto AGY (`render_agy_plugin_manifest`), manteniendo intactas las notas de releases pasadas (`v1.x`, `v2.0.0`, `v2.0.1`).

## Aceptación

| ID | Resultado observable |
| --- | --- |
| FD-01 | `skills/frontend-design/SKILL.md` y `LICENSE.txt` conservan byte a byte los blobs del commit `e4bb5d960d87126ff3b3a8af1a104b7db67560ba` y coinciden con `UPSTREAM.json`. |
| FD-02 | `profile_resources.inventory` y `package_release.py profile` incluyen `frontend-design` en Codex, Claude Code y AGY CLI. |
| FD-03 | `harness-ux-ui/SKILL.md` enlaza a `../frontend-design/SKILL.md` sin romper enlaces portables ni duplicar el flujo de aprobación/UX de Chuli. |
| FD-04 | `validate_repository.py`, avisos (`THIRD_PARTY_NOTICES.md`, `skills/_shared/NOTICES.md`), `docs/protocols/SKILLS.md`, `README.md`, `docs/AGENTS.md` y `client_targets.py` reflejan las 25 skills y pasan todas las suites Python, Go y PowerShell. |

## Riesgos y reversión

Los archivos upstream permanecen intactos y fijados por commit sin actualización automática ni instalación de herramientas externas (`npx skills`). Para revertir esta unidad alcanza con eliminar `skills/frontend-design/`, `docs/specs/frontend-design-skill.md`, `odd/tasks/frontend-design-skill.md`, `tests/test_frontend_design_resources.py` y restaurar las menciones a 24 skills y los avisos de terceros.
