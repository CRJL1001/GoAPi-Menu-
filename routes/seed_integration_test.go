package Routes

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSeedDatabase(t *testing.T) {
	db := associationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sql, err := os.ReadFile("../cmd/seed/seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	var existingID int64
	if err := db.QueryRow(ctx, `INSERT INTO ingredients (name) VALUES ('Existing') RETURNING id`).Scan(&existingID); err != nil {
		t.Fatal(err)
	}
	for run := 1; run <= 2; run++ {
		if _, err := db.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("seed run %d: %v", run, err)
		}
		for _, tc := range []struct {
			table string
			want  int
		}{
			{"ingredients", 4}, {"recipies", 3}, {"menus", 2},
			{"include_ingredients", 3}, {"include_recipies", 1},
		} {
			var got int
			if err := db.QueryRow(ctx, "SELECT count(*) FROM "+tc.table).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("run %d: %s count = %d, want %d", run, tc.table, got, tc.want)
			}
		}
	}
	var name string
	if err := db.QueryRow(ctx, `SELECT name FROM ingredients WHERE id = $1`, existingID).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Existing" {
		t.Fatalf("existing ingredient changed: %q", name)
	}
}
