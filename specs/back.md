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

## Acceptance criteria
- No tooling scaffolded yet — verify against `back/go.mod` once created, and record exact commands here (build, test, lint, run) when chosen.
- Endpoint paths and request/response shapes are the contract with `specs/front.md`; keep the two specs in sync.
