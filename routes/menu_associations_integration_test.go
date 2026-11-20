package Routes

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMenuAssociationDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := db.Exec(ctx, `
		INSERT INTO recipies (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Tomate'), (3000000002, 'Oignon');
		INSERT INTO menus (id, name) OVERRIDING SYSTEM VALUE
		VALUES (4000000001, 'Soupe'), (4000000002, 'Salade'), (4000000003, 'Vide');
	`)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	Register(router, db)
	for _, tc := range []struct {
		name, method, path, body string
		status                   int
		want                     string
	}{
		{"missing menu", "POST", "/menus/add_recipe_to_menu", `{"id_menu":1,"id_recipe":3000000001,"count":2}`, 404, `{"error":"Le menu renseigné n'existe pas"}`},
		{"missing recipe", "POST", "/menus/add_recipe_to_menu", `{"id_menu":4000000001,"id_recipe":1,"count":2}`, 404, `{"error":"La recette renseignée n'existe pas"}`},
		{"create second recipe", "POST", "/menus/add_recipe_to_menu", `{"id_menu":4000000001,"id_recipe":3000000002,"count":3}`, 201, `{"id_menu":4000000001,"id_recipe":3000000002,"count":3}`},
		{"create first recipe", "POST", "/menus/add_recipe_to_menu", `{"id_menu":4000000001,"id_recipe":3000000001,"count":2}`, 201, `{"id_menu":4000000001,"id_recipe":3000000001,"count":2}`},
		{"duplicate", "POST", "/menus/add_recipe_to_menu", `{"id_menu":4000000001,"id_recipe":3000000001,"count":7}`, 409, `{"error":"Cette recette est déjà associée au menu"}`},
		{"other menu", "POST", "/menus/add_recipe_to_menu", `{"id_menu":4000000002,"id_recipe":3000000001,"count":2147483647}`, 201, `{"id_menu":4000000002,"id_recipe":3000000001,"count":2147483647}`},
		{"sorted and filtered", "GET", "/menus/4000000001/recipies", "", 200, `[{"id_menu":4000000001,"id_recipe":3000000001,"count":2},{"id_menu":4000000001,"id_recipe":3000000002,"count":3}]`},
		{"empty menu", "GET", "/menus/4000000003/recipies", "", 200, `[]`},
		{"missing menu list", "GET", "/menus/1/recipies", "", 404, `{"error":"Le menu renseigné n'existe pas"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkAssociationResponse(t, router, tc.method, tc.path, tc.body, tc.status, tc.want)
		})
	}
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM include_recipies`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("stored associations = %d, want 3", count)
	}
	if _, err := db.Exec(ctx, `DROP TABLE include_recipies`); err != nil {
		t.Fatal(err)
	}
	checkAssociationResponse(t, router, "GET", "/menus/4000000001/recipies", "", 500,
		`{"error":"Impossible de récupérer les recettes du menu"}`)
	checkAssociationResponse(t, router, "POST", "/menus/add_recipe_to_menu",
		`{"id_menu":4000000001,"id_recipe":3000000001,"count":1}`, 500,
		`{"error":"Impossible de créer la relation"}`)
}
