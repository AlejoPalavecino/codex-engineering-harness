# Importador SQLite offline (fixture sintético)

El sistema recibe CSV de inventario y almacena filas en una base SQLite local. Actualmente acumula todo el CSV en memoria antes de insertar. Un requisito futuro indica un volumen **diez veces mayor**, sin números reales ni mediciones provistas.

El caso es **arquitectónico R2**: antes de implementar, comparar conservar el enfoque, usar lotes/streaming o reescribir. No inventar tiempos, consumo de RAM ni datos de producción. Proponer cómo instrumentar un benchmark representativo, límites, recuperación de errores y persistencia transaccional.

Solo trabajar en una copia temporal del fixture.
