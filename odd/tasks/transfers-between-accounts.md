# Transferencias entre cuentas propias (RelationRepository, de scaffolding a producción)

- **Objetivo:** una transferencia entre cuentas propias es UNA decisión con dos
  piernas: el monto sale de la cuenta origen (-X) y entra en la destino (+X),
  bajo un evento financiero de tipo TRANSFER, ligadas por una relación
  `TRANSFER_COUNTERPART`. Borrar una pierna borra la transferencia.
- **Rama:** `developer`. **Estrategia:** unidades de commit por capa, tests
  primero, batería por unidad.
- **Runner backend:** `go test -count=1 ./...` desde `backend/` (+ gofmt, vet,
  golangci-lint, sqlc diff). **Runner desktop:** `npx vitest run` desde
  `apps/desktop/` (+ build, lint, prettier).

## Specs

- S1 — El usuario: "deja el sistema de trasferencias y construyelo
  correctamente, aunque actualmente solo sea la cuenta Nu, en el futuro se
  añadiran mas (para v0.2.Z) probablemnte quede en funcionamiento".
- S2 — Modelo elegido por el mantenedor: "Un evento, dos movientos es correcto
  como lo propones, sale de la cuenta de origen y enra a la cuenta de destino,
  ligados por la relacion TRANSFER_CONUTERPART. No cuenta como ingreso ni
  egreso en las distribuciones. el neto es cero. Pero como te menciono, el
  sistema debe estar listo y utilizable pero no implementando, ya que la
  unica cuneta existente actualmente es Nu, no existen mas cuentas asi que no
  se pueden hacer transferencias entre cuentas aun."
- S3 — Escritura atómica: evento + dos movimientos + relación en una sola
  transacción; un fallo en cualquier paso no deja pierna huérfana.
- S4 — Signos derivados: el monto se ingresa positivo; la pierna origen es
  negativa y la destino positiva (el front ya valida cantidad positiva).
- S5 — Una transferencia NUNCA se estaciona en Revisión: la cola modela
  correcciones de un movimiento, no de una decisión de dos piernas.
- S6 — Borrado consciente de la contraparte: borrar una pierna borra ambas y
  la relación; borrar un movimiento sin relación sigue el camino actual.
- S7 — Compatibilidad: un `TRANSFER` sin destino (cliente viejo) sigue
  creando una sola pierna por el camino manual actual.
- S8 — Sin código específico de institución: las cuentas ya la cargan; el
  modelo sirve para más instituciones sin reescribirse.

## Tasks

| ID | Unidad | Ruta | Estado |
|----|--------|------|--------|
| T1 | Núcleo backend: `TransferService` (escritura transaccional + borrado consciente), interfaz inyectada, rama en `createMovement`, `deleteMovement` transfer-aware | inline | done (a0df231) |
| T2 | Tests backend (servicio + HTTP) | inline | done (a0df231) |
| T3 | Frontend: campo de destino en el drawer + tests | inline | done (091f4fc) |
| T4 | Contraparte visible: join en la hidratación + KV en el drawer | inline | done (18b0812) |

## Log

- L1 (usuario, 2026-10-26): "toda funcionalidad de las tarjetas de credito,
  aun no las pondremos asi que sera dejarlo sattus quo y si esa condicion
  esta bien dejar de crecer hasta que verdaderamente se vaya a utilizar. Pero
  deja el sistema de trasferencias y construyelo correctamente..."
- L2: opción de modelo elegida tras la pregunta: un evento/dos
  movimientos/relación TRANSFER_COUNTERPART.

## Log

- L3 (2026-10-26): feature COMPLETO. Unidad backend a0df231 (TransferService
  + rama en createMovement + borrado transfer-aware, 11 tests), e194b87
  (documento + comentario de scaffolding actualizado), 091f4fc (drawer con
  cuenta destino, 6 tests, contrato de createMovement con counterpart_id),
  18b0812 (contraparte visible en las listings hidratadas y en el drawer,
  tests en ambos extremos). Con una sola cuenta la opción ni aparece en el
  drawer (accounts.length > 1), así que el sistema queda listo y dormido hasta
  que exista la segunda cuenta. Pendiente de decisión futura: si una
  transferencia debe poder dejarse sin clasificar en Revisión (hoy nunca se
  estaciona, por diseño S5).
