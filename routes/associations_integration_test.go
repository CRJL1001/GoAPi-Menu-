package Routes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func associationTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close(context.Background()) })
	schema := fmt.Sprintf("association_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("clean up test schema: %v", err)
		}
	})
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	migrations, err := filepath.Glob("../migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append([]string{"../init.sql"}, migrations...) {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("apply %s: %v", path, err)
		}
	}
	return db
}

func TestAssociationDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := db.Exec(ctx, `
		INSERT INTO ingredients (id, name) OVERRIDING SYSTEM VALUE
		VALUES (3000000001, 'Tomate'), (3000000002, 'Oignon');
		INSERT INTO recipies (id, name) OVERRIDING SYSTEM VALUE
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
		{"missing recipe", "POST", "/recipies/add_ingredient_to_recipe", `{"id_recipe":1,"id_ingredient":3000000001,"count":2}`, 404, `{"error":"La recette renseignée n'existe pas"}`},
		{"missing ingredient", "POST", "/recipies/add_ingredient_to_recipe", `{"id_recipe":4000000001,"id_ingredient":1,"count":2}`, 404, `{"error":"L'ingrédient renseigné n'existe pas"}`},
		{"create onion", "POST", "/recipies/add_ingredient_to_recipe", `{"id_recipe":4000000001,"id_ingredient":3000000002,"count":3}`, 201, `{"id_recipe":4000000001,"id_ingredient":3000000002,"count":3}`},
		{"create tomato", "POST", "/recipies/add_ingredient_to_recipe", `{"id_recipe":4000000001,"id_ingredient":3000000001,"count":2}`, 201, `{"id_recipe":4000000001,"id_ingredient":3000000001,"count":2}`},
		{"duplicate", "POST", "/recipies/add_ingredient_to_recipe", `{"id_recipe":4000000001,"id_ingredient":3000000001,"count":7}`, 409, `{"error":"Cet ingrédient est déjà associé à la recette"}`},
		{"other recipe", "POST", "/recipies/add_ingredient_to_recipe", `{"id_recipe":4000000002,"id_ingredient":3000000001,"count":2147483647}`, 201, `{"id_recipe":4000000002,"id_ingredient":3000000001,"count":2147483647}`},
		{"sorted and filtered", "GET", "/recipies/4000000001/ingredients", "", 200, `[{"id_recipe":4000000001,"id_ingredient":3000000001,"count":2},{"id_recipe":4000000001,"id_ingredient":3000000002,"count":3}]`},
		{"empty recipe", "GET", "/recipies/4000000003/ingredients", "", 200, `[]`},
		{"missing recipe list", "GET", "/recipies/1/ingredients", "", 404, `{"error":"La recette renseignée n'existe pas"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkAssociationResponse(t, router, tc.method, tc.path, tc.body, tc.status, tc.want)
		})
	}
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM include_ingredients`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("stored associations = %d, want 3", count)
	}
	if _, err := db.Exec(ctx, `DROP TABLE include_ingredients`); err != nil {
		t.Fatal(err)
	}
	checkAssociationResponse(t, router, "GET", "/recipies/4000000001/ingredients", "", 500,
		`{"error":"Impossible de récupérer les ingrédients de la recette"}`)
	checkAssociationResponse(t, router, "POST", "/recipies/add_ingredient_to_recipe",
		`{"id_recipe":4000000001,"id_ingredient":3000000001,"count":1}`, 500,
		`{"error":"Impossible de créer la relation"}`)
}
