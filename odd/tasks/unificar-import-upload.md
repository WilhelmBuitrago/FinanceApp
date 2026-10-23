# Unificar el import: un solo dueño del staging

Objetivo: POST /documents/upload es la unica ruta de import funcional y
delega el estacionado en ImportService (transaccional, con review items).
Branch: developer. Runner: go test ./... desde backend/.

## Specs

- S1: El upload estagea los movimientos del parsed.json en UNA transaccion
  (source_entry + financial_event + movement por movimiento): un fallo
  parcial no deja filas huerfanas ni el job en SUCCESS.
- S2: Los movimientos review_required del parsed.json quedan estacionados
  por la misma ruta (source entry REVIEW_REQUIRED + item kind EXTRACTION en
  review_items). Hoy el upload los pierde en silencio.
- S3: Re-subir el mismo PDF (mismo sha256) responde 409 Conflict con
  {error, document_id, job_id}: ni 500 ni duplicados.
- S4: El handler de upload no contiene logica de staging: delega en
  ImportService via interface inyectada. Se retiran POST /jobs (scaffold) y
  POST /jobs/{id}/confirm (sin cliente HTTP).
- S5: El job termina SUCCESS o FAILED segun el resultado del servicio; un
  fallo del servicio responde 500 con detalle y marca FAILED.

## Tasks

| ID | Specs | Ruta | Estado |
| --- | --- | --- | --- |
| T1 | S1-S5 | tests RED→GREEN (duplicado 409, delegacion end-to-end, fallo cierra FAILED) | hecho |
| T2 | S1,S2,S4,S5 | handler delega + retira staging | hecho |
| T3 | S4 | retiro rutas muertas + wiring | hecho |
| T4 | S1-S5 | GREEN + verificacion completa | hecho |
| T5 | S4 | docs (development.md) | hecho (sin cambios: no lista rutas HTTP) |

## Log

- L1 (pedido del usuario, verbatim): "Comienza con los Criticos, sin embargo,
  hay que revisar bien, si se tienen dos servicios pero entiendo que uno es a
  nivel de moviments (UploadHandler) y el otro a nivel de transacciones
  (ImportService) debemos revisar si hay diferencia o si son lo mismo en verdad
  (mismo input, mismo comportamiento, mismo output, misma ruta). SI es asi
  eliminar uno y acomodar el otro."
- L2 (evidencia de auditoria): UploadHandler.Handle (vivo, cliente
  ImportDocumentDrawer) estageaba sin transaccion, ignoraba review_required y
  fallaba con 500 en re-subida (docID determinista doc_+sha[:8]).
  ImportService.Import (transaccional, con review items) no tenia cliente
  HTTP: solo tests. Veredicto del usuario: "El upload delega al servicio".
- L3 (cierre): handler 312->232 lineas, staging en un solo dueno.
  Tests nuevos en tests/upload_test.go: TestUpload_DuplicatePdf_ReturnsConflict
  (409 con document_id, no 500), TestUpload_StagesThroughService (2 movimientos
  en TX + 1 item EXTRACTION en review_items + job SUCCESS),
  TestUpload_ServiceFailure_ClosesJobFailed (500 + job FAILED).
  Verificado: gofmt/goimports 0, build/vet 0, golangci-lint 0 findings,
  suite completa ok.
