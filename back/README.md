# Calculator backend

Go REST microservice that performs the calculator's arithmetic. The React
frontend in `front/` is the only expected client; all results are computed
here, never in the UI.

## Requirements

- Go 1.27+ (the repository pins Go through `mise.toml`; run `mise install`).

## Commands

Run every command from `back/`:

| Task | Command |
| --- | --- |
| Run (listens on `:8080`) | `go run ./cmd/server` |
| Run on a custom port | `PORT=3000 go run ./cmd/server` |
| Build | `go build ./...` |
| Test | `go test ./...` |
| Vet | `go vet ./...` |
| Format check | `gofmt -l .` |

## API

All requests and responses are JSON. A binary operation takes
`{"a": number, "b": number}` and a unary operation takes `{"a": number}`.
Success responses are `{"result": number}`.

| Method | Path | Operation |
| --- | --- | --- |
| GET | `/health` | liveness, returns `{"status":"ok"}` |
| GET | `/ping` | readiness, returns `{"message":"pong"}` |
| POST | `/api/add` | `a + b` |
| POST | `/api/subtract` | `a - b` |
| POST | `/api/multiply` | `a * b` |
| POST | `/api/divide` | `a / b` |
| POST | `/api/pow` | `a` raised to `b` |
| POST | `/api/percentage` | `a` percent of `b` (`a*b/100`) |
| POST | `/api/sqrt` | square root of `a` |

Errors use `{"error":{"code":"...","message":"..."}}`:

| Status | Code | Cause |
| --- | --- | --- |
| 400 | `invalid_json` | body is not a single JSON object, unknown field, wrong value type |
| 400 | `invalid_request` | required numeric field `a` or `b` missing |
| 404 | `not_found` | unknown path |
| 405 | `method_not_allowed` | known path reached with the wrong method |
| 422 | `division_by_zero` | divisor is zero, or zero raised to a negative power |
| 422 | `negative_sqrt` | square root of a negative number |
| 422 | `overflow` | result is not a finite float64 |
| 422 | `undefined` | operation has no defined real result (e.g. `(-1)^0.5`) |
| 500 | `internal_error` | unexpected server failure |

Example:

```console
$ curl -s localhost:8080/api/divide -d '{"a":1,"b":0}'
{"error":{"code":"division_by_zero","message":"division by zero"}}
```

CORS is enabled (`Access-Control-Allow-Origin: *`) so the Vite dev server can
call the API from a different origin.

## Design rationale

- `internal/calc` holds pure operations returning `(float64, error)` with
  sentinel errors, so arithmetic is unit-tested in isolation from HTTP.
- `internal/api` is the transport layer: it strictly decodes one JSON object
  with known fields, uses pointer fields to detect missing operands, and maps
  each sentinel error to a status code and error code.
- `float64` matches JavaScript's number type, so values survive the round trip
  to the frontend without precision surprises beyond IEEE-754 itself.
- Results are rejected when NaN or infinite, because those cannot be encoded
  as valid JSON numbers; this is what backs `undefined` and `overflow`.
- The server exposes `PORT` and shuts down gracefully on `SIGINT`/`SIGTERM`.
