---
name: skill-registry
description: "Trigger: update skills, skill registry, actualizar skills, after skill changes. Index available skills by trigger and path."
license: MIT
metadata:
  author: gentleman-programming
  version: "1.0"
  upstream: "Gentle AI v4.4.0"
  chuli-runtime: "Codex"
  chuli-dependencies: "Metadatos locales; CLI Gentle AI y Engram opcionales"
---

> Adaptación Chuli 1.0. Seguí la política y los permisos del repositorio de destino.

## Resolución local

Cuando necesites seleccionar skills, leé nombre, descripción y ruta en el inventario del cliente o en los directorios de skills activos del proyecto y del perfil. Priorizá la instrucción específica del proyecto. Si dos implementaciones con igual nombre difieren, señalá las rutas y comprobá cuál cargó realmente el cliente. Excluí `_shared` y esta propia skill del índice.

Entregá las rutas exactas de las skills elegidas para que cada delegado lea sólo las que necesita. Podés responder con un índice en memoria. Si el usuario pide persistirlo, escribí `.atl/skill-registry.md` en el proyecto autorizado, con origen, alcance, nombre, descripción y ruta. Verificá su contenido después de escribirlo. La integración externa de Gentle AI admite `gentle-ai skill-registry refresh --cwd <project>` cuando ese binario está instalado. Engram es opcional y cualquier registro de memoria queda acotado al proyecto y a los datos autorizados.

## Referencias

- [Resolución de skills](../_shared/skill-resolver.md)
- [Guía de formato](../skill-creator/references/skill-style-guide.md)
