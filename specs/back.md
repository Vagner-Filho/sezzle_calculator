# Backend spec (scope: `back/`)

Go REST microservice that performs the calculator's arithmetic; the frontend is the only expected client. Never trust client input.

## Functional requirements (FR)
- Expose REST endpoints for: addition, subtraction, multiplication, division (required); exponentiation, square root, percentage (optional).
- Health/readiness endpoints: `GET /health` and `GET /ping` (microservice convention).
- Return results as JSON.
- Handle errors and edge cases gracefully: validate input (division by zero, invalid/malformed data, numeric overflow where relevant), return structured JSON errors with appropriate status codes in the 399–499 range — never crashes or 200-with-error-body.

## Non-functional requirements (NFR)
- Idiomatic, readable Go; layer logic so operations are unit-testable in isolation from HTTP handlers.
- Unit tests covering key functionality, especially edge cases (division by zero, invalid data).

## API contract
All requests and responses are JSON; binary operations take `{"a": number, "b": number}`, unary operations take `{"a": number}`, and success responses are `{"result": number}`.

| Method | Path | Operation / response |
| --- | --- | --- |
| GET | `/health` | `200 {"status":"ok"}` |
| GET | `/ping` | `200 {"message":"pong"}` |
| POST | `/api/add` | `a + b` |
| POST | `/api/subtract` | `a - b` |
| POST | `/api/multiply` | `a * b` |
| POST | `/api/divide` | `a / b` |
| POST | `/api/pow` | `a` raised to `b` |
| POST | `/api/percentage` | `a` percent of `b` (`a*b/100`) |
| POST | `/api/sqrt` | square root of `a` |

Errors are `{"error":{"code":"...","message":"..."}}` with status/code: `400 invalid_json` (malformed body, unknown field, wrong type), `400 invalid_request` (missing numeric `a`/`b`), `404 not_found`, `405 method_not_allowed`, `422 division_by_zero`, `422 negative_sqrt`, `422 overflow` (non-finite result), `422 undefined` (no defined real result), `500 internal_error`. CORS is enabled for the browser frontend.

## Acceptance criteria
- Commands (Go 1.27, from `back/`): build `go build ./...`; test `go test ./...`; lint/vet `gofmt -l .` and `go vet ./...`; run `go run ./cmd/server` (listens on `:8080`, override with `PORT`).
- Endpoint paths and request/response shapes are the contract with `specs/front.md`; keep the two specs in sync.
