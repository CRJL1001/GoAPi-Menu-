 CREATE TABLE IF NOT EXISTS include_ingredients(
            id_ingredient INT NOT NULL,
            id_recipie INT NOT NULL,
            count INT NOT NULL DEFAULT 1,
            PRIMARY KEY (id_ingredient, id_recipie),
            FOREIGN KEY (id_ingredient) REFERENCES ingredients(id_ingredient),
            FOREIGN KEY (id_recipie) REFERENCES recipies(id_recipie)
        );
