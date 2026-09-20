# Multi-stage Dockerfile that builds the React frontend and the Go backend,
# then produces a single image running both from one process.

# -----------------------------------------------------------------------------
# Stage 1: Build the frontend
# -----------------------------------------------------------------------------
FROM node:24-alpine AS front-builder

WORKDIR /build/front

COPY front/package*.json ./
RUN npm ci

COPY front/ ./
RUN VITE_API_URL=/ npm run build

# -----------------------------------------------------------------------------
# Stage 2: Build the backend
# -----------------------------------------------------------------------------
FROM golang:1.27-alpine AS back-builder

WORKDIR /build/back

COPY back/go.mod ./
RUN go mod download

COPY back/ ./
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

# -----------------------------------------------------------------------------
# Stage 3: Final runtime image
# -----------------------------------------------------------------------------
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=back-builder /bin/server /app/server
COPY --from=front-builder /build/front/dist /app/static

ENV STATIC_DIR=/app/static
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/server"]
