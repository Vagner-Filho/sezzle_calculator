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

## Acceptance criteria
- Test runner TBD — verify against `front/package.json` once scaffolded; record exact commands here (`npm run ...`) when chosen.
