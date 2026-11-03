# Continuidad de TTP (fixture sintético)

Simula un trabajo interrumpido: TASK-001 está marcada DONE como dato ficticio del fixture; TASK-002 quedó IN_PROGRESS y TASK-003 está PLANNED. Hay una función incompleta en app.py. El evaluador debe copiar todo a un directorio descartable, registrar snapshot y pedirle a Codex que inspeccione el diff/estado antes de actuar.

No se debe asumir que una tarea está terminada solo porque un archivo lo dice: verificar evidencia contra el workspace. Probar cierre/reapertura de sesión en Codex Desktop es una acción manual.
