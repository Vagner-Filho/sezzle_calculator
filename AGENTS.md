# AGENTS.md

Calculator app: React + TypeScript frontend (`front/`) consuming a Go backend REST API (`back/`). Development is spec-driven (SDD): one spec file per folder under specs/ holding scope-limited instructions; this file holds global rules.

## SDD layout
- `specs/front.md` — frontend-only instructions; scope is that folder.
- `specs/back.md` — backend-only instructions; scope is that folder.
- Keep folder-specific detail in the SDD files, cross-cutting facts here. Update both when a fact's scope changes.
- When implementing a spec, editing or creating files outside the folder the spec defines is prohibited. Reading in the same condition is allowed.

## Architecture constraints
- Frontend must get arithmetic results from the backend API only — no independent calculation logic in the UI.
- Backend REST API: validate input and edge cases (division by zero, invalid data), return JSON.
- Operations: addition, subtraction, multiplication, division (required); exponentiation, square root, percentage (optional).

## Quality bar
- Unit tests for key functionality on both layers; frontend: input validation and error handling; backend: edge cases.
- Test coverage for any automated test run.
- Idiomatic code; responsive frontend (basic mobile support).
- Documentation must cover setup, API usage, and design rationale.
- Dockerfile (`Dockerfile`) for full-stack deployment: builds the React frontend and Go backend into one image; run with `docker build -t calculator . && docker run -p 8080:8080 calculator`.
- Functions whose name may not clearly define what they do will need commenting.

## Conventions
- Toolchain not scaffolded yet — verify commands against manifests (`front/package.json`, `back/go.mod`) once they exist, and record the exact commands in the relevant SDD file.
