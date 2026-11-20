package Handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HandlersHelper struct {
	DB *pgxpool.Pool
}

func (h *HandlersHelper) RecipeExists(c *gin.Context, recipe_id int64) bool {
	var recipe_exist bool

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.DB.QueryRow(
		timed_ctx,
		`SELECT EXISTS (SELECT 1 FROM recipies WHERE id = $1)`,
		recipe_id,
	).Scan(&recipe_exist)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de retrouver la recette",
		})
		return false
	}

	if !recipe_exist {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "La recette renseignée n'existe pas",
		})
		return false
	}

	return recipe_exist
}

func (h *HandlersHelper) IngredientExists(c *gin.Context, ingredient_id int64) bool {
	var ingredient_exist bool

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.DB.QueryRow(
		timed_ctx,
		`SELECT EXISTS (SELECT 1 FROM ingredients WHERE id = $1)`,
		ingredient_id,
	).Scan(&ingredient_exist)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de retrouver l'ingrédient",
		})
		return false
	}

	if !ingredient_exist {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "L'ingrédient renseigné n'existe pas",
		})
		return false
	}

	return ingredient_exist
}
