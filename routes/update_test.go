package Routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUpdateErrors(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, resource := range []struct{ path, databaseError string }{
		{"/ingredients", `{"error":"Impossible de modifier l'ingrédient"}`},
		{"/menus", `{"error":"Impossible de modifier le menu"}`},
		{"/recipies", `{"error":"Impossible de modifier la recette"}`},
	} {
		t.Run(resource.path, func(t *testing.T) {
			for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
				t.Run("invalid id "+id, func(t *testing.T) {
					checkAssociationResponse(t, router, http.MethodPatch, resource.path+"/"+id, `{"name":"Nouveau"}`, http.StatusBadRequest,
						`{"error":"Identifiant invalide"}`)
				})
			}
			for _, tc := range []struct{ name, body string }{
				{"empty body", ""},
				{"malformed JSON", "{"},
				{"missing name", `{}`},
				{"null body", `null`},
				{"null name", `{"name":null}`},
				{"empty name", `{"name":""}`},
				{"wrong type", `{"name":123}`},
				{"oversized body", strings.Repeat(" ", 4097) + `{"name":"Nouveau"}`},
			} {
				t.Run(tc.name, func(t *testing.T) {
					checkAssociationResponse(t, router, http.MethodPatch, resource.path+"/1", tc.body, http.StatusBadRequest,
						`{"error":"Corps JSON invalide : name est obligatoire"}`)
				})
			}
			for _, tc := range []struct{ name, value string }{
				{"blank name", " \t\n "},
				{"too long", strings.Repeat("a", 101)},
				{"too many Unicode characters", strings.Repeat("é", 101)},
				{"nul character", "Tom\x00ate"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					body, err := json.Marshal(map[string]string{"name": tc.value})
					if err != nil {
						t.Fatal(err)
					}
					checkAssociationResponse(t, router, http.MethodPatch, resource.path+"/1", string(body), http.StatusBadRequest,
						`{"error":"le nom doit contenir entre 1 et 100 caractères, sans nul"}`)
				})
			}
			t.Run("database unavailable", func(t *testing.T) {
				checkAssociationResponse(t, router, http.MethodPatch, resource.path+"/1", `{"name":"Nouveau"}`, http.StatusInternalServerError, resource.databaseError)
			})
		})
	}
}

func TestUpdateDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `
		INSERT INTO ingredients (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Tomate'), (3000000002, 'Oignon');
		INSERT INTO recipies (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Soupe'), (3000000002, 'Salade');
		INSERT INTO menus (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Midi'), (3000000002, 'Soir');
		INSERT INTO include_ingredients (id_recipe, id_ingredient, count)
		VALUES (3000000001, 3000000001, 2);
		INSERT INTO include_recipies (id_menu, id_recipe, count)
		VALUES (3000000001, 3000000001, 3);
	`); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	Register(router, db)

	for _, resource := range []struct{ table, missingError, otherName string }{
		{"ingredients", `{"error":"Ingrédient introuvable"}`, "Oignon"},
		{"menus", `{"error":"Menu introuvable"}`, "Soir"},
		{"recipies", `{"error":"recette introuvable"}`, "Salade"},
	} {
		t.Run(resource.table, func(t *testing.T) {
			path := "/" + resource.table
			checkAssociationResponse(t, router, http.MethodPatch, path+"/1", `{"name":"Nouveau"}`, http.StatusNotFound, resource.missingError)
			for _, tc := range []struct{ name, input, want string }{
				{"trim spaces", "  Nouveau  ", "Nouveau"},
				{"same name", "Nouveau", "Nouveau"},
				{"one character", "A", "A"},
				{"100 Unicode characters", strings.Repeat("é", 100), strings.Repeat("é", 100)},
			} {
				t.Run(tc.name, func(t *testing.T) {
					body, err := json.Marshal(map[string]string{"name": tc.input})
					if err != nil {
						t.Fatal(err)
					}
					want := `{"id":3000000001,"name":"` + tc.want + `"}`
					checkAssociationResponse(t, router, http.MethodPatch, path+"/3000000001", string(body), http.StatusOK, want)
					var stored string
					if err := db.QueryRow(ctx, "SELECT name FROM "+resource.table+" WHERE id = 3000000001").Scan(&stored); err != nil {
						t.Fatal(err)
					}
					if stored != tc.want {
						t.Fatalf("stored name = %q, want %q", stored, tc.want)
					}
				})
			}
			checkAssociationResponse(t, router, http.MethodGet, path+"/3000000002", "", http.StatusOK,
				`{"id":3000000002,"name":"`+resource.otherName+`"}`)
		})
	}

	t.Run("duplicate ingredient name preserves original", func(t *testing.T) {
		var existingName string
		if err := db.QueryRow(ctx, `SELECT name FROM ingredients WHERE id = 3000000001`).Scan(&existingName); err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(map[string]string{"name": existingName})
		if err != nil {
			t.Fatal(err)
		}
		// The current handler returns 500 for unique constraint violations.
		checkAssociationResponse(t, router, http.MethodPatch, "/ingredients/3000000002", string(body), http.StatusInternalServerError,
			`{"error":"Impossible de modifier l'ingrédient"}`)
		checkAssociationResponse(t, router, http.MethodGet, "/ingredients/3000000002", "", http.StatusOK,
			`{"id":3000000002,"name":"Oignon"}`)
	})
	checkAssociationResponse(t, router, http.MethodGet, "/menus/3000000001/recipies", "", http.StatusOK,
		`[{"id_menu":3000000001,"id_recipe":3000000001,"count":3}]`)
	checkAssociationResponse(t, router, http.MethodGet, "/recipies/3000000001/ingredients", "", http.StatusOK,
		`[{"id_recipe":3000000001,"id_ingredient":3000000001,"count":2}]`)
}
