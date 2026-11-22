INSERT INTO ingredients (name)
VALUES ('[seed] Tomate'), ('[seed] Oignon'), ('[seed] Ingredient libre')
ON CONFLICT (name) DO NOTHING;

INSERT INTO recipies (name)
SELECT name FROM (VALUES
    ('[seed] Soupe'), ('[seed] Salade'), ('[seed] Recette libre')
) AS seed(name)
WHERE NOT EXISTS (SELECT 1 FROM recipies WHERE recipies.name = seed.name);

INSERT INTO menus (name)
SELECT name FROM (VALUES ('[seed] Dejeuner'), ('[seed] Menu vide')) AS seed(name)
WHERE NOT EXISTS (SELECT 1 FROM menus WHERE menus.name = seed.name);

INSERT INTO include_ingredients (id_recipe, id_ingredient, count)
SELECT r.id, i.id, seed.count
FROM (VALUES
    ('[seed] Soupe', '[seed] Tomate', 2),
    ('[seed] Soupe', '[seed] Oignon', 1),
    ('[seed] Salade', '[seed] Tomate', 3)
) AS seed(recipe, ingredient, count)
JOIN recipies r ON r.name = seed.recipe
JOIN ingredients i ON i.name = seed.ingredient
ON CONFLICT (id_ingredient, id_recipe) DO NOTHING;

INSERT INTO include_recipies (id_menu, id_recipe, count)
SELECT m.id, r.id, 2
FROM menus m
JOIN recipies r ON r.name = '[seed] Soupe'
WHERE m.name = '[seed] Dejeuner'
ON CONFLICT (id_menu, id_recipe) DO NOTHING;
