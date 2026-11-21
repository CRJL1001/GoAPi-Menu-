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
)

func TestIndexErrors(t *testing.T) {
	db, err := pgxpool.New(context.Background(), "postgres://localhost/test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	router := gin.New()
	Register(router, db)

	for _, tc := range []struct {
		path, want string
	}{
		{"/menus", `{"error":"Lecture des menus impossible"}`},
		{"/recipies", `{"error":"Lecture des recettes impossible"}`},
		{"/ingredients", `{"error":"Lecture des ingrédients impossible"}`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			checkAssociationResponse(t, router, http.MethodGet, tc.path, "", http.StatusInternalServerError, tc.want)
		})
	}
}

func TestIndexDatabase(t *testing.T) {
	for _, resource := range []string{"menus", "recipies", "ingredients"} {
		t.Run(resource, func(t *testing.T) {
			db := associationTestDB(t)
			router := gin.New()
			Register(router, db)
			path := "/" + resource

			t.Run("empty list", func(t *testing.T) {
				checkAssociationResponse(t, router, http.MethodGet, path, "", http.StatusOK, `[]`)
			})

			t.Run("all items", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if _, err := db.Exec(ctx, `
					INSERT INTO `+resource+` (id, name) OVERRIDING SYSTEM VALUE
					VALUES (4000000002, 'Salade'), (4000000001, 'Soupe')
				`); err != nil {
					t.Fatal(err)
				}

				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
				if response.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200: %s", response.Code, response.Body)
				}

				type item struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				}
				var got []item
				if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })
				want := []item{{ID: 4000000001, Name: "Soupe"}, {ID: 4000000002, Name: "Salade"}}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("items = %+v, want %+v", got, want)
				}
			})
		})
	}
}
