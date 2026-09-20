# Sezzle Calculator

A full-stack calculator application with a React + TypeScript frontend and a Go
REST API backend. All arithmetic is performed by the backend; the UI only
renders input and results.

## Design decisions

- The goal of the architecture at the root level is to showcase an AI workflow guided by SDD files.
- Each project can then pick a structure independently and contain their own AGENTS.md (or other) to tend to their needs.
- The folders front and back, self contained, allow the usage of parallel agents, speeding up the development.
- Since the focus lies on clean design, maintainable code, and testable architecture, building the calculator with a single input field that takes longer forms of arithmetic operations (a + b - c, a / (b * c), ...) would drastically reduce the number of endpoints, which might not be good to demonstrate the virtues required.
- With commonly seen architecture patterns for back and front parts, the reviewer may jump straight into code.
- The code is modular, no file is too big nor too small and I chose to use comments where a function is not super simple or have some behavior not quite represented by its name. 
- Endpoints pertinent to a microservice context, mentioned in the description, were added (/health and /ping) so that load balancers and orchestrators may interact with it.

## Architecture

- **`front/`** — React single-page application built with Vite, TypeScript, and
  Vitest. See [`front/README.md`](front/README.md) for frontend details.
- **`back/`** — Go REST microservice that validates input, performs operations,
  and returns JSON. See [`back/README.md`](back/README.md) for backend details.
- **`Dockerfile`** — Multi-stage build that produces one image running both the
- **`specs/`** — Source of truth to guide agents on what they should do.

## Requirements

The repository pins tool versions through `mise.toml`. Run `mise install` to
install them, or ensure you have:

- Go 1.27+
- Node.js 24+
- npm 11+
- Docker (optional, for the containerised deployment)

## Quick start

### Docker

Build and run a single image that serves both parts:

```console
$ docker build -t calculator .
$ docker run -p 8080:8080 calculator
```

Then open `http://localhost:8080` in your browser.

## Project structure

```
.
├── back/              # Go backend
│   ├── cmd/server/    # HTTP server entrypoint
│   └── internal/      # calc and api packages
├── front/             # React frontend
│   └── src/           # components, API client, validation, tests
├── specs/             # SDD spec files
├── Dockerfile         # Full-stack container build
└── README.md          # This file
```

## API

All requests and responses are JSON. Binary operations take
`{"a": number, "b": number}`; unary operations take `{"a": number}`. Success
responses are `{"result": number}`.

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
| 400 | `invalid_json` | malformed body, unknown field, or wrong value type |
| 400 | `invalid_request` | missing numeric operand |
| 404 | `not_found` | unknown path |
| 405 | `method_not_allowed` | known path reached with wrong method |
| 422 | `division_by_zero` | divisor is zero |
| 422 | `negative_sqrt` | square root of a negative number |
| 422 | `overflow` | non-finite result |
| 422 | `undefined` | no defined real result (e.g. `(-1)^0.5`) |
| 500 | `internal_error` | unexpected server failure |

## API examples

Health check:

```console
$ curl -s http://localhost:8080/health
{"status":"ok"}
```

Addition:

```console
$ curl -s http://localhost:8080/api/add -d '{"a": 2, "b": 3}'
{"result":5}
```

Square root (unary):

```console
$ curl -s http://localhost:8080/api/sqrt -d '{"a": 16}'
{"result":4}
```

Division by zero:

```console
$ curl -s http://localhost:8080/api/divide -d '{"a": 1, "b": 0}'
{"error":{"code":"division_by_zero","message":"division by zero"}}
```

Missing operand:

```console
$ curl -s http://localhost:8080/api/add -d '{"a": 1}'
{"error":{"code":"invalid_request","message":"addition requires numeric fields \"a\" and \"b\""}}
```

## Testing

Backend:

```console
$ cd back
$ go test ./... -coverprofile=cover.out
$ go tool cover -html=cover.out
```

Frontend:

```console
$ cd front
$ npm test
```

Linting:

```console
$ cd back && go vet ./... && gofmt -l .
$ cd front && npm run lint
```
