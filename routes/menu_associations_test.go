package Routes

import (
	"context"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMenuAssociationValidation(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
		t.Run("invalid menu "+id, func(t *testing.T) {
			checkAssociationResponse(t, router, "GET", "/menus/"+id+"/recipies", "", 400, `{"error":"Identifiant invalide"}`)
		})
	}
	for _, body := range []string{"", "{", `{"id_menu":"1"}`, strings.Repeat(" ", 4097) + "{}"} {
		t.Run("invalid JSON", func(t *testing.T) {
			checkAssociationResponse(t, router, "POST", "/menus/add_recipe_to_menu", body, 400,
				`{"error":"Corps JSON invalide : count, id_menu et id_recipe obligatoire"}`)
		})
	}
	for _, body := range []string{
		`{}`, `null`,
		`{"id_menu":0,"id_recipe":1,"count":1}`,
		`{"id_menu":1,"id_recipe":-1,"count":1}`,
		`{"id_menu":1,"id_recipe":1,"count":0}`,
		`{"id_menu":1,"id_recipe":1,"count":-1}`,
	} {
		t.Run("invalid values", func(t *testing.T) {
			checkAssociationResponse(t, router, "POST", "/menus/add_recipe_to_menu", body, 400,
				`{"error":"Les ids et le compte ne peuvent pas être négatif ou null"}`)
		})
	}
	checkAssociationResponse(t, router, "POST", "/menus/add_recipe_to_menu",
		`{"id_menu":1,"id_recipe":1,"count":2147483648}`, 400,
		`{"error":"Le compte dépasse la valeur maximale autorisée"}`)
	checkAssociationResponse(t, router, "POST", "/menus/add_recipe_to_menu",
		`{"id_menu":1,"id_recipe":1,"count":1}`, 500,
		`{"error":"Impossible de retrouver la recette"}`)
	checkAssociationResponse(t, router, "GET", "/menus/1/recipies", "", 500,
		`{"error":"Impossible de retrouver le menu"}`)
}
