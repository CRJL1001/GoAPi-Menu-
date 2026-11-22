package Routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIngredientDeleteErrors(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
		t.Run("invalid id "+id, func(t *testing.T) {
			checkAssociationResponse(t, router, http.MethodDelete, "/ingredients/"+id, "", http.StatusBadRequest,
				`{"error":"Identifiant invalide"}`)
		})
	}
	t.Run("database unavailable", func(t *testing.T) {
		checkAssociationResponse(t, router, http.MethodDelete, "/ingredients/1", "", http.StatusInternalServerError,
			`{"error":"Impossible de supprimer l'ingrédient"}`)
	})
}

func TestIngredientDeleteDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `
		INSERT INTO ingredients (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Tomate'), (3000000002, 'Oignon');
		INSERT INTO recipies (id, name) OVERRIDING SYSTEM VALUE
		VALUES (4000000001, 'Soupe');
		INSERT INTO include_ingredients (id_ingredient, id_recipe, count)
		VALUES (3000000002, 4000000001, 2);
	`); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	Register(router, db)

	t.Run("missing ingredient", func(t *testing.T) {
		checkAssociationResponse(t, router, http.MethodDelete, "/ingredients/1", "", http.StatusNotFound,
			`{"error":"Ingrédient introuvable"}`)
	})

	t.Run("delete unused ingredient", func(t *testing.T) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/ingredients/3000000001", nil))
		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204: %s", response.Code, response.Body)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("expected empty response, got %q", response.Body.String())
		}
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ingredients WHERE id = 3000000001)`).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("deleted ingredient still exists")
		}
		checkAssociationResponse(t, router, http.MethodDelete, "/ingredients/3000000001", "", http.StatusNotFound,
			`{"error":"Ingrédient introuvable"}`)
	})

	t.Run("ingredient used in recipe", func(t *testing.T) {
		checkAssociationResponse(t, router, http.MethodDelete, "/ingredients/3000000002", "", http.StatusConflict,
			`{"error":"Cet ingrédient est utilisé dans une recette"}`)
		checkAssociationResponse(t, router, http.MethodGet, "/ingredients/3000000002", "", http.StatusOK,
			`{"id":3000000002,"name":"Oignon"}`)
		checkAssociationResponse(t, router, http.MethodGet, "/recipies/4000000001/ingredients", "", http.StatusOK,
			`[{"id_recipe":4000000001,"id_ingredient":3000000002,"count":2}]`)
	})
}
