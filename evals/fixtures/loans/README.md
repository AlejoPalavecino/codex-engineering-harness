# Préstamos de equipos (fixture sintético)

Este es un ejemplo **ficticio**, no una institución real. Todo el trabajo debe hacerse en **una copia temporal** de esta carpeta. Nunca se edita el fixture original del repositorio de Chuli durante un benchmark.

- Dos sedes imaginarias: Norte y Sur.
- Un equipo puede estar disponible, prestado o en mantenimiento.
- Se registra un préstamo para una persona habilitada.
- La devolución puede ejecutarla la persona responsable o un administrador.
- No hay base de datos, MCP, credenciales ni despliegue.
- La implementación es deliberadamente pequeña y contiene un caso no cubierto: **dos devoluciones de un mismo préstamo no son idempotentes**.

Antes de intervenir, verificá el estado Git y consultá las decisiones explícitas del caso. Si el prompt pide análisis solamente, no cambies el fixture.

Pruebas existentes (sin dependencias externas):

```sh
python -m unittest discover -s tests -v
```
