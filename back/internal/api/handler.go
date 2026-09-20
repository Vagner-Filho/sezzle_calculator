// Package api exposes the calculator operations over HTTP as a JSON REST API.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

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
	mux.HandleFunc("/", handleNotFound)

	return withCORS(mux)
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
