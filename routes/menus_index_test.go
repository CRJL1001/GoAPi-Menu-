package Routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	Handlers "github.com/m/handlers"
	Models "github.com/m/models"
)

func TestMenusIndexErrors(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	handler := &Handlers.MenusHandler{DB: db}
	router := gin.New()
	router.GET("/menus", handler.Index)
	router.GET("/menus/:id", handler.Index)

	for _, path := range []string{"/menus", "/menus/abc", "/menus/0", "/menus/-1", "/menus/9223372036854775808"} {
		t.Run(path, func(t *testing.T) {
			checkAssociationResponse(t, router, http.MethodGet, path, "", http.StatusBadRequest,
				`{"error":"Identifiant invalide"}`)
		})
	}
	t.Run("database unavailable", func(t *testing.T) {
		checkAssociationResponse(t, router, http.MethodGet, "/menus/1", "", http.StatusInternalServerError,
			`{"error":"Lecture des menus impossible"}`)
	})
}

func TestMenusIndexDatabase(t *testing.T) {
	db := associationTestDB(t)
	handler := &Handlers.MenusHandler{DB: db}
	router := gin.New()
	router.GET("/menus/:id", handler.Index)

	t.Run("empty list", func(t *testing.T) {
		checkAssociationResponse(t, router, http.MethodGet, "/menus/1", "", http.StatusOK, `[]`)
	})

	t.Run("all menus", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := db.Exec(ctx, `
			INSERT INTO menus (id, name) OVERRIDING SYSTEM VALUE
			VALUES (4000000002, 'Salade'), (4000000001, 'Soupe')
		`); err != nil {
			t.Fatal(err)
		}

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/menus/1", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", response.Code, response.Body)
		}

		var got []Models.Menu
		if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })
		want := []Models.Menu{{ID: 4000000001, Name: "Soupe"}, {ID: 4000000002, Name: "Salade"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("menus = %+v, want %+v", got, want)
		}
	})
}
