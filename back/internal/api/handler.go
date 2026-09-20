// Package api exposes the calculator operations over HTTP as a JSON REST API.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"calculator/back/internal/calc"
)

// maxBodyBytes caps request bodies; the expected payloads are tiny.
const maxBodyBytes = 1 << 10

type resultResponse struct {
	Result float64 `json:"result"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type binaryRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

type unaryRequest struct {
	A *float64 `json:"a"`
}

type binaryOp func(a, b float64) (float64, error)

type unaryOp func(a float64) (float64, error)

// NewHandler returns the handler serving every calculator route.
func NewHandler() http.Handler {
	return NewHandlerWithStatic("")
}

// NewHandlerWithStatic returns the API handler and serves static files from
// staticDir for all other requests. An empty staticDir is equivalent to
// NewHandler and returns JSON 404 responses for unmatched paths.
func NewHandlerWithStatic(staticDir string) http.Handler {
	return withCORS(buildAPIMux(fallbackHandler(staticDir)))
}

// fallbackHandler returns a handler for unmatched paths. When staticDir is set,
// it serves the SPA static files and falls back to index.html; unknown /api/
// paths still return the JSON 404 response.
func fallbackHandler(staticDir string) http.Handler {
	if staticDir == "" {
		return http.HandlerFunc(handleNotFound)
	}

	fs := http.FileServer(&spaFileSystem{root: http.Dir(staticDir)})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			handleNotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})
}

// spaFileSystem falls back to index.html so client-side routing works.
type spaFileSystem struct {
	root http.FileSystem
}

func (s *spaFileSystem) Open(name string) (http.File, error) {
	f, err := s.root.Open(name)
	if err != nil && os.IsNotExist(err) {
		return s.root.Open("/index.html")
	}
	return f, err
}

// buildAPIMux registers all calculator API routes and mounts the fallback
// handler for every other path.
func buildAPIMux(fallback http.Handler) http.Handler {
	routes := []struct {
		method  string
		path    string
		handler http.Handler
	}{
		{http.MethodGet, "/health", http.HandlerFunc(handleHealth)},
		{http.MethodGet, "/ping", http.HandlerFunc(handlePing)},
		{http.MethodPost, "/api/add", handleBinary("addition", calc.Add)},
		{http.MethodPost, "/api/subtract", handleBinary("subtraction", calc.Subtract)},
		{http.MethodPost, "/api/multiply", handleBinary("multiplication", calc.Multiply)},
		{http.MethodPost, "/api/divide", handleBinary("division", calc.Divide)},
		{http.MethodPost, "/api/pow", handleBinary("exponentiation", calc.Pow)},
		{http.MethodPost, "/api/percentage", handleBinary("percentage", calc.Percentage)},
		{http.MethodPost, "/api/sqrt", handleUnary("square root", calc.Sqrt)},
	}

	mux := http.NewServeMux()
	configured := make(map[string]bool, len(routes))
	for _, route := range routes {
		mux.Handle(route.method+" "+route.path, route.handler)
		if !configured[route.path] {
			configured[route.path] = true
			mux.HandleFunc(route.path, methodNotAllowed(route.method))
		}
	}
	mux.Handle("/{path...}", fallback)

	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handlePing(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

// handleBinary decodes {a, b} and applies op, writing the result as JSON.
func handleBinary(description string, op binaryOp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req binaryRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.A == nil || req.B == nil {
			writeError(w, http.StatusBadRequest, "invalid_request",
				fmt.Sprintf("%s requires numeric fields %q and %q", description, "a", "b"))
			return
		}
		result, err := op(*req.A, *req.B)
		if err != nil {
			writeCalcError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resultResponse{Result: result})
	})
}

// handleUnary decodes {a} and applies op, writing the result as JSON.
func handleUnary(description string, op unaryOp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req unaryRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if req.A == nil {
			writeError(w, http.StatusBadRequest, "invalid_request",
				fmt.Sprintf("%s requires a numeric field %q", description, "a"))
			return
		}
		result, err := op(*req.A)
		if err != nil {
			writeCalcError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resultResponse{Result: result})
	})
}

// decodeJSON strictly decodes a single JSON object from the request body and
// reports a client error when it fails.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be a JSON object with numeric fields")
		return false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must contain a single JSON object")
		return false
	}
	return true
}

// writeCalcError maps a domain error to its HTTP status and error code.
func writeCalcError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, calc.ErrDivisionByZero):
		writeError(w, http.StatusUnprocessableEntity, "division_by_zero", err.Error())
	case errors.Is(err, calc.ErrNegativeSqrt):
		writeError(w, http.StatusUnprocessableEntity, "negative_sqrt", err.Error())
	case errors.Is(err, calc.ErrOverflow):
		writeError(w, http.StatusUnprocessableEntity, "overflow", err.Error())
	case errors.Is(err, calc.ErrUndefined):
		writeError(w, http.StatusUnprocessableEntity, "undefined", err.Error())
	default:
		slog.Error("unhandled calculation error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "resource not found")
}

// methodNotAllowed responds to a known path reached with the wrong method.
func methodNotAllowed(allowed string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allowed)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed",
			fmt.Sprintf("method %s is not allowed", r.Method))
	}
}

// withCORS allows the browser-based frontend to call the API from its own
// origin, including preflight requests.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("encoding response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message}})
}
