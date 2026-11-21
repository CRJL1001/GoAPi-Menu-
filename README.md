# GoAPi-Menu-
a Go API to handle menu, recipies and ingredients in Postgres Database

# to init database inside container

'docker compose exec db psql -U app -d appdb -v ON_ERROR_STOP=1 -f /docker-entrypoint-initdb.d/init.sql'
