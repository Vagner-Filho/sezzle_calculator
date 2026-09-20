# Frontend spec (scope: `front/`)

React + TypeScript SPA. All arithmetic results come from the backend REST API — never compute independently in the UI (see root `AGENTS.md`).

## Functional requirements (FR)
- Calculator UI for entering input and displaying results; intuitive, clean, responsive design (basic mobile support).
- Operations: addition, subtraction, multiplication, division (required); exponentiation, square root, percentage (optional).
- Input validation and error handling: validate input client-side; surface invalid data and backend errors (e.g. division by zero) without crashing the UI.
- Handle errors returned by the backend via a toast message.

## Non-functional requirements (NFR)
- Idiomatic, readable TypeScript.
- Unit tests covering key functionality: input validation and error handling.

## API consumption
- Consume the Go backend REST API defined in `specs/back.md`; JSON in/out.
- Do not duplicate operation endpoints or assume response shapes not in the backend spec.
- Base URL in dev: `http://localhost:8080`. Binary operations POST `{"a": number, "b": number}` to `/api/add`, `/api/subtract`, `/api/multiply`, `/api/divide`, `/api/pow`, `/api/percentage`; unary `/api/sqrt` takes `{"a": number}`. Success is `{"result": number}`; errors are `{"error":{"code":"...","message":"..."}}` (e.g. `422 division_by_zero`, `422 negative_sqrt`, `422 overflow`, `400 invalid_json`, `400 invalid_request`) and are surfaced via toast.
- Readiness probe: `GET /health` → `{"status":"ok"}`; `GET /ping` → `{"message":"pong"}`.

## Acceptance criteria
- Test runner: **Vitest** — record exact commands here (`npm run ...`) when chosen.
- Exact commands (from `front/package.json`):
  - `npm run dev` — start Vite dev server
  - `npm run build` — TypeScript compile + Vite production build
  - `npm run preview` — preview production build
  - `npm run lint` — run Oxlint
  - `npm test` — run Vitest tests
