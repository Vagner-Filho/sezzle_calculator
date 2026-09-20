package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"calculator/back/internal/api"
)

type resultBody struct {
	Result float64 `json:"result"`
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func doRequest(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	api.NewHandler().ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("decoding response %q: %v", rec.Body.String(), err)
	}
}

func TestOperations(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantResult float64
		wantCode   string
	}{
		{name: "add", path: "/api/add", body: `{"a":2,"b":3}`, wantStatus: http.StatusOK, wantResult: 5},
		{name: "add floating point", path: "/api/add", body: `{"a":0.1,"b":0.2}`, wantStatus: http.StatusOK, wantResult: 0.30000000000000004},
		{name: "subtract", path: "/api/subtract", body: `{"a":5,"b":3}`, wantStatus: http.StatusOK, wantResult: 2},
		{name: "multiply", path: "/api/multiply", body: `{"a":4,"b":2.5}`, wantStatus: http.StatusOK, wantResult: 10},
		{name: "divide", path: "/api/divide", body: `{"a":10,"b":4}`, wantStatus: http.StatusOK, wantResult: 2.5},
		{name: "divide by zero", path: "/api/divide", body: `{"a":1,"b":0}`, wantStatus: http.StatusUnprocessableEntity, wantCode: "division_by_zero"},
		{name: "divide by negative zero", path: "/api/divide", body: `{"a":1,"b":-0}`, wantStatus: http.StatusUnprocessableEntity, wantCode: "division_by_zero"},
		{name: "add overflow", path: "/api/add", body: `{"a":1e308,"b":1e308}`, wantStatus: http.StatusUnprocessableEntity, wantCode: "overflow"},
		{name: "pow", path: "/api/pow", body: `{"a":2,"b":10}`, wantStatus: http.StatusOK, wantResult: 1024},
		{name: "pow negative exponent", path: "/api/pow", body: `{"a":2,"b":-1}`, wantStatus: http.StatusOK, wantResult: 0.5},
		{name: "pow zero to negative", path: "/api/pow", body: `{"a":0,"b":-1}`, wantStatus: http.StatusUnprocessableEntity, wantCode: "division_by_zero"},
		{name: "pow undefined", path: "/api/pow", body: `{"a":-1,"b":0.5}`, wantStatus: http.StatusUnprocessableEntity, wantCode: "undefined"},
		{name: "percentage", path: "/api/percentage", body: `{"a":10,"b":45}`, wantStatus: http.StatusOK, wantResult: 4.5},
		{name: "sqrt", path: "/api/sqrt", body: `{"a":9}`, wantStatus: http.StatusOK, wantResult: 3},
		{name: "sqrt negative", path: "/api/sqrt", body: `{"a":-4}`, wantStatus: http.StatusUnprocessableEntity, wantCode: "negative_sqrt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, http.MethodPost, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}

			if tt.wantCode != "" {
				var got errorBody
				decodeBody(t, rec, &got)
				if got.Error.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
				}
				if got.Error.Message == "" {
					t.Error("error message is empty")
				}
				return
			}

			var got resultBody
			decodeBody(t, rec, &got)
			if got.Result != tt.wantResult {
				t.Errorf("result = %v, want %v", got.Result, tt.wantResult)
			}
		})
	}
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{name: "empty body", path: "/api/add", body: "", wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "malformed json", path: "/api/add", body: `{"a":`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "json array", path: "/api/add", body: `[1,2]`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "missing operand", path: "/api/add", body: `{"a":1}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
		{name: "null operand", path: "/api/add", body: `{"a":null,"b":2}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
		{name: "wrong operand type", path: "/api/add", body: `{"a":"1","b":2}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "unknown field", path: "/api/add", body: `{"a":1,"b":2,"c":3}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "trailing data", path: "/api/add", body: `{"a":1,"b":2}{"a":3,"b":4}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "number out of range", path: "/api/add", body: `{"a":1e400,"b":1}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "sqrt unknown field", path: "/api/sqrt", body: `{"a":1,"b":2}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "sqrt missing operand", path: "/api/sqrt", body: `{}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, http.MethodPost, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var got errorBody
			decodeBody(t, rec, &got)
			if got.Error.Code != tt.wantCode {
				t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
			}
		})
	}
}

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/health", want: `{"status":"ok"}`},
		{path: "/ping", want: `{"message":"pong"}`},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := doRequest(t, http.MethodGet, tt.path, "")

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tt.want {
				t.Errorf("body = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantAllow  string
		wantStatus int
	}{
		{name: "get on operation", method: http.MethodGet, path: "/api/add", wantAllow: "POST", wantStatus: http.StatusMethodNotAllowed},
		{name: "post on health", method: http.MethodPost, path: "/health", wantAllow: "GET", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, tt.method, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
			var got errorBody
			decodeBody(t, rec, &got)
			if got.Error.Code != "method_not_allowed" {
				t.Errorf("error code = %q, want method_not_allowed", got.Error.Code)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	rec := doRequest(t, http.MethodGet, "/api/unknown", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var got errorBody
	decodeBody(t, rec, &got)
	if got.Error.Code != "not_found" {
		t.Errorf("error code = %q, want not_found", got.Error.Code)
	}
}

func TestCORS(t *testing.T) {
	t.Run("preflight", func(t *testing.T) {
		rec := doRequest(t, http.MethodOptions, "/api/add", "")

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q, want *", got)
		}
	})

	t.Run("actual request", func(t *testing.T) {
		rec := doRequest(t, http.MethodPost, "/api/add", `{"a":1,"b":2}`)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q, want *", got)
		}
	})
}

func TestStaticFiles(t *testing.T) {
	tmpDir := t.TempDir()
	indexHTML := "<!doctype html><html><body>Calculator SPA</body></html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexHTML), 0o644); err != nil {
		t.Fatalf("writing index.html: %v", err)
	}

	handler := api.NewHandlerWithStatic(tmpDir)

	t.Run("serves index.html at root", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Body.String(); got != indexHTML {
			t.Errorf("body = %q, want %q", got, indexHTML)
		}
	})

	t.Run("falls back to index.html for unknown paths", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/some/route", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Body.String(); got != indexHTML {
			t.Errorf("body = %q, want %q", got, indexHTML)
		}
	})

	t.Run("unknown api paths still return json 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		var got errorBody
		decodeBody(t, rec, &got)
		if got.Error.Code != "not_found" {
			t.Errorf("error code = %q, want not_found", got.Error.Code)
		}
	})

	t.Run("api routes still work", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(`{"a":1,"b":2}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var got resultBody
		decodeBody(t, rec, &got)
		if got.Result != 3 {
			t.Errorf("result = %v, want 3", got.Result)
		}
	})
}
