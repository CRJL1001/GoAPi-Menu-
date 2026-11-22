 CREATE TABLE IF NOT EXISTS include_recipies(
            id_menu BIGINT NOT NULL,
            id_recipe BIGINT NOT NULL,
            count INT NOT NULL DEFAULT 1,
            FOREIGN KEY (id_menu) REFERENCES menus(id) ON DELETE CASCADE,
            FOREIGN KEY (id_recipe) REFERENCES recipies(id),
            PRIMARY KEY (id_menu, id_recipe)
        );
