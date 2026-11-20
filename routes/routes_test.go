package Routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIngredientErrors(t *testing.T) {
	// A closed pool exercises database failures without a running PostgreSQL server.
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, tc := range []struct {
		name, method, path, body, message string
		status                            int
	}{
		{"unknown route", "GET", "/ingredient/1", "", "Route introuvable", http.StatusNotFound},
		{"invalid id", "GET", "/ingredients/abc", "", "Identifiant invalide", http.StatusBadRequest},
		{"read database failure", "GET", "/ingredients/1", "", "Impossible de récupérer l'ingrédient", http.StatusInternalServerError},
		{"invalid JSON", "POST", "/ingredients", "{}", "Corps JSON invalide : name est obligatoire", http.StatusBadRequest},
		{"blank name", "POST", "/ingredients", `{"name":"  "}`, "Le nom doit contenir entre 1 et 100 caractères, sans nul", http.StatusBadRequest},
		{"create database failure", "POST", "/ingredients", `{"name":"Tomate"}`, "Impossible de créer l'ingrédient", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body)
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["error"] != tc.message {
				t.Fatalf("error = %q, want %q", body["error"], tc.message)
			}
		})
	}
}

func TestMenusErrors(t *testing.T) {
	// A closed pool exercises database failures without a running PostgreSQL server.
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, tc := range []struct {
		name, method, path, body, message string
		status                            int
	}{
		{"unknown route", "GET", "/menu/1", "", "Route introuvable", http.StatusNotFound},
		{"invalid id", "GET", "/menus/abc", "", "Identifiant invalide", http.StatusBadRequest},
		{"read database failure", "GET", "/menus/1", "", "Impossible de récupérer le menu", http.StatusInternalServerError},
		{"invalid JSON", "POST", "/menus", "{}", "Corps JSON invalide : name est obligatoire", http.StatusBadRequest},
		{"blank name", "POST", "/menus", `{"name":"  "}`, "Le nom doit contenir entre 1 et 100 caractères, sans nul", http.StatusBadRequest},
		{"create database failure", "POST", "/menus", `{"name":"Tomate"}`, "Impossible de créer le menu", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body)
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["error"] != tc.message {
				t.Fatalf("error = %q, want %q", body["error"], tc.message)
			}
		})
	}
}

func TestRecipiesErrors(t *testing.T) {
	// A closed pool exercises database failures without a running PostgreSQL server.
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, tc := range []struct {
		name, method, path, body, message string
		status                            int
	}{
		{"unknown route", "GET", "/recipe/1", "", "Route introuvable", http.StatusNotFound},
		{"invalid id", "GET", "/recipies/abc", "", "Identifiant invalide", http.StatusBadRequest},
		{"read database failure", "GET", "/recipies/1", "", "Impossible de récupérer la recette", http.StatusInternalServerError},
		{"invalid JSON", "POST", "/recipies", "{}", "Corps JSON invalide : name est obligatoire", http.StatusBadRequest},
		{"blank name", "POST", "/recipies", `{"name":"  "}`, "Le nom doit contenir entre 1 et 100 caractères, sans nul", http.StatusBadRequest},
		{"create database failure", "POST", "/recipies", `{"name":"Tomate"}`, "Impossible de créer la recette", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body)
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["error"] != tc.message {
				t.Fatalf("error = %q, want %q", body["error"], tc.message)
			}
		})
	}
}
