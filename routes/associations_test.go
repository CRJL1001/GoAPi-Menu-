package Routes

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkAssociationResponse(t *testing.T, router *gin.Engine, method, path, body string, status int, want string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: status = %d, want %d: %s", method, path, response.Code, status, response.Body)
	}
	if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("expected JSON response, got %q", response.Header().Get("Content-Type"))
	}
	var gotJSON, wantJSON any
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&gotJSON); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("handler wrote more than one response: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Fatalf("response = %v, want %v", gotJSON, wantJSON)
	}
}

func TestAssociationValidation(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
		t.Run("invalid recipe "+id, func(t *testing.T) {
			checkAssociationResponse(t, router, "GET", "/recipies/"+id+"/ingredients", "", 400, `{"error":"Identifiant invalide"}`)
		})
	}
	for _, body := range []string{"", "{", `{"id_recipe":"1"}`, strings.Repeat(" ", 4097) + "{}"} {
		t.Run("invalid JSON", func(t *testing.T) {
			checkAssociationResponse(t, router, "POST", "/recipies/add_ingredient_to_recipe", body, 400,
				`{"error":"Corps JSON invalide : count, id_ingredient et id_recipe obligatoire"}`)
		})
	}
	for _, body := range []string{
		`{}`, `null`,
		`{"id_recipe":0,"id_ingredient":1,"count":1}`,
		`{"id_recipe":1,"id_ingredient":-1,"count":1}`,
		`{"id_recipe":1,"id_ingredient":1,"count":0}`,
		`{"id_recipe":1,"id_ingredient":1,"count":-1}`,
	} {
		t.Run("invalid values", func(t *testing.T) {
			checkAssociationResponse(t, router, "POST", "/recipies/add_ingredient_to_recipe", body, 400,
				`{"error":"Les ids et le compte ne peuvent pas être négatif ou null"}`)
		})
	}
	checkAssociationResponse(t, router, "POST", "/recipies/add_ingredient_to_recipe",
		`{"id_recipe":1,"id_ingredient":1,"count":2147483648}`, 400,
		`{"error":"Le compte dépasse la valeur maximale autorisée"}`)
	checkAssociationResponse(t, router, "POST", "/recipies/add_ingredient_to_recipe",
		`{"id_recipe":1,"id_ingredient":1,"count":1}`, 500,
		`{"error":"Impossible de retrouver la recette"}`)
	checkAssociationResponse(t, router, "GET", "/recipies/1/ingredients", "", 500,
		`{"error":"Impossible de retrouver la recette"}`)
}
