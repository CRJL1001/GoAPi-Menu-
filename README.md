# GoAPi-Menu-
a Go API to handle menu, recipies and ingredients in Postgres Database

'docker compose exec db psql -U app -d appdb -v ON_ERROR_STOP=1 -f /docker-entrypoint-initdb.d/init.sql'

- Ingredients: `[seed] Tomate`, `[seed] Oignon`, `[seed] Ingredient libre`.
- Recipes: `[seed] Soupe` (tomato and onion), `[seed] Salade` (tomato),
  `[seed] Recette libre` (no associations).
- Menus: `[seed] Dejeuner` (soup), `[seed] Menu vide` (no recipes).

```bash
docker compose exec -T db psql -U app -d appdb -v ON_ERROR_STOP=1 --single-transaction < cmd/seed/seed.sql
```

