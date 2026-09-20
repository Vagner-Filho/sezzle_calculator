# Calculator frontend

React + TypeScript single-page application for the calculator UI. It consumes the
Go backend REST API in `back/`; all arithmetic results come from the backend,
never computed independently in the UI.

## Requirements

- Node.js v24.21.0 (the repository pins Node through `mise.toml`; run `mise install`).
- npm v11.19.0 (installed with Node).

## Commands

Run every command from `front/`:

| Task | Command |
| --- | --- |
| Dev server (Vite) | `npm run dev` |
| Type-check and production build | `npm run build` |
| Preview production build | `npm run preview` |
| Run tests | `npm test` |
| Serve Test Coverage View | `npx vite preview --outDir coverage` |
| Lint | `npm run lint` |

## API

The frontend talks to the backend at `http://localhost:8080` by default. Set the
`VITE_API_URL` environment variable to override it, for example:

```console
VITE_API_URL=http://api.example.com npm run dev
```

All requests and responses are JSON. A binary operation sends
`{"a": number, "b": number}` and a unary operation sends `{"a": number}`.
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

Errors use `{"error":{"code":"...","message":"..."}}` and are surfaced to the
user as a toast notification:

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

## Design rationale

- `src/api/client.ts` is the single place where the UI calls the backend. It
  centralises fetch logic, JSON decoding, and error handling so components stay
  focused on presentation.
- `src/validation.ts` performs client-side validation of user input (required
  numbers, finite values, non-zero divisors) to fail fast before any network
  request.
- `src/components/Calculator.tsx` is the main UI surface. It delegates every
  calculation to the API and shows either the returned result or a validation
  error.
- `src/components/Toast.tsx` displays backend errors without crashing the UI.
- The layout stays stable: the second input field is disabled rather than hidden
  when an operation only needs one operand, preventing unexpected element
  movement.
- Accessibility is built in with semantic HTML, WAI-ARIA live regions for the
  result, and `aria-pressed` on operation buttons.
- Styling uses CSS custom properties and mixes soft neumorphism (raised/inset
  shadows) with liquid-glass effects (`backdrop-filter` blur, translucent
  surfaces) while keeping text contrast high enough for readability.
- Tests run with Vitest and cover input validation and component-level
  accessibility/layout behaviour.

## Example

Start the backend from `back/`, then start the frontend:

```console
$ cd front
$ npm run dev
```

Open the URL printed by Vite, enter two numbers, choose an operation, and click
**Calculate**. The result is fetched from the backend and shown in the display.
