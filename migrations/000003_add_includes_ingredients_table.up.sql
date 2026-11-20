 CREATE TABLE IF NOT EXISTS include_ingredients(
            id_ingredient BIGINT NOT NULL,
            id_recipe BIGINT NOT NULL,
            count INT NOT NULL DEFAULT 1,
            FOREIGN KEY (id_ingredient) REFERENCES ingredients(id),
            FOREIGN KEY (id_recipe) REFERENCES recipies(id),
            PRIMARY KEY (id_ingredient, id_recipe)
        );
