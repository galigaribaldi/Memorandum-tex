# con-paquetes (fixture)

Manifiesto válido cuyo `requires.files` incluye `memorandum-paquete-inexistente.sty`. Con un `HasFiles` real
(`kpsewhich`) o simulado, `compat.Evaluate` debe devolver `missing-packages` y sugerir
`tlmgr install memorandum-paquete-inexistente`.
