package Routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMenuAndRecipeDeleteErrors(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, tc := range []struct{ resource, want string }{
		{"menus", `{"error":"Impossible de supprimer le menu"}`},
		{"recipies", `{"error":"Impossible de supprimer la recette"}`},
	} {
		t.Run(tc.resource, func(t *testing.T) {
			for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
				t.Run("invalid id "+id, func(t *testing.T) {
					checkAssociationResponse(t, router, http.MethodDelete, "/"+tc.resource+"/"+id, "", http.StatusBadRequest,
						`{"error":"Identifiant invalide"}`)
				})
			}
			t.Run("database unavailable", func(t *testing.T) {
				checkAssociationResponse(t, router, http.MethodDelete, "/"+tc.resource+"/1", "", http.StatusInternalServerError, tc.want)
			})
		})
	}
}

func TestMenuDeleteDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `
		INSERT INTO menus (id, name) OVERRIDING SYSTEM VALUE
		VALUES (4000000001, 'Vide'), (4000000002, 'Midi'), (4000000003, 'Soir');
		INSERT INTO recipies (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Soupe'), (3000000002, 'Salade');
		INSERT INTO include_recipies (id_menu, id_recipe, count)
		VALUES (4000000002, 3000000001, 2), (4000000002, 3000000002, 1), (4000000003, 3000000001, 3);
	`); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	Register(router, db)

	checkAssociationResponse(t, router, http.MethodDelete, "/menus/1", "", http.StatusNotFound,
		`{"error":"menu introuvable"}`)
	for _, menuID := range []int64{4000000001, 4000000002} {
		id := strconv.FormatInt(menuID, 10)
		t.Run("delete menu "+id, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/menus/"+id, nil))
			if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
				t.Fatalf("response = %d %q, want 204 with empty body", response.Code, response.Body.String())
			}
			var exists bool
			if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM menus WHERE id = $1)`, menuID).Scan(&exists); err != nil {
				t.Fatal(err)
			}
			if exists {
				t.Fatal("deleted menu still exists")
			}
			checkAssociationResponse(t, router, http.MethodDelete, "/menus/"+id, "", http.StatusNotFound,
				`{"error":"menu introuvable"}`)
		})
	}

	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM include_recipies WHERE id_menu = 4000000002`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("deleted menu associations = %d, want 0", count)
	}
	checkAssociationResponse(t, router, http.MethodGet, "/recipies/3000000001", "", http.StatusOK,
		`{"id":3000000001,"name":"Soupe"}`)
	checkAssociationResponse(t, router, http.MethodGet, "/recipies/3000000002", "", http.StatusOK,
		`{"id":3000000002,"name":"Salade"}`)
	checkAssociationResponse(t, router, http.MethodGet, "/menus/4000000003/recipies", "", http.StatusOK,
		`[{"id_menu":4000000003,"id_recipe":3000000001,"count":3}]`)
}

func TestRecipeDeleteDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `
		INSERT INTO recipies (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Libre'), (3000000002, 'Au menu'), (3000000003, 'Avec ingredient');
		INSERT INTO menus (id, name) OVERRIDING SYSTEM VALUE VALUES (4000000001, 'Midi');
		INSERT INTO ingredients (id, name) OVERRIDING SYSTEM VALUE VALUES (5000000001, 'Tomate');
		INSERT INTO include_recipies (id_menu, id_recipe, count) VALUES (4000000001, 3000000002, 2);
		INSERT INTO include_ingredients (id_recipe, id_ingredient, count) VALUES (3000000003, 5000000001, 3);
	`); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	Register(router, db)

	checkAssociationResponse(t, router, http.MethodDelete, "/recipies/1", "", http.StatusNotFound,
		`{"error":"Recette introuvable"}`)
	t.Run("delete unused recipe", func(t *testing.T) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/recipies/3000000001", nil))
		if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
			t.Fatalf("response = %d %q, want 204 with empty body", response.Code, response.Body.String())
		}
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM recipies WHERE id = 3000000001)`).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("deleted recipe still exists")
		}
		checkAssociationResponse(t, router, http.MethodDelete, "/recipies/3000000001", "", http.StatusNotFound,
			`{"error":"Recette introuvable"}`)
	})
	for _, tc := range []struct{ id, want string }{
		{"3000000002", `{"id":3000000002,"name":"Au menu"}`},
		{"3000000003", `{"id":3000000003,"name":"Avec ingredient"}`},
	} {
		t.Run("associated recipe "+tc.id, func(t *testing.T) {
			checkAssociationResponse(t, router, http.MethodDelete, "/recipies/"+tc.id, "", http.StatusConflict,
				`{"error":"Cette recette contient des associations"}`)
			checkAssociationResponse(t, router, http.MethodGet, "/recipies/"+tc.id, "", http.StatusOK, tc.want)
		})
	}
	checkAssociationResponse(t, router, http.MethodGet, "/menus/4000000001/recipies", "", http.StatusOK,
		`[{"id_menu":4000000001,"id_recipe":3000000002,"count":2}]`)
	checkAssociationResponse(t, router, http.MethodGet, "/recipies/3000000003/ingredients", "", http.StatusOK,
		`[{"id_recipe":3000000003,"id_ingredient":5000000001,"count":3}]`)
}
