package Handlers

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	Models "github.com/m/models"
)

type IncludeIngredientHandler struct {
	DB *pgxpool.Pool
}

func (h *IncludeIngredientHandler) Create(c *gin.Context) {
	var input struct { //input
		Id_ingredient int64 `json:"id_ingredient"`
		Id_recipe     int64 `json:"id_recipe"`
		Count         int   `json:"count"`
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		4096,
	)

	if err := c.ShouldBindJSON(&input); err != nil { //verif binding json
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Corps JSON invalide : count, id_ingredient et id_recipe obligatoire",
		})
		return
	}

	if input.Id_ingredient <= 0 || input.Id_recipe <= 0 || input.Count <= 0 { //verif id et count
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Les ids et le compte ne peuvent pas être négatif ou null",
		})
		return
	}
	if input.Count > math.MaxInt32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le compte dépasse la valeur maximale autorisée"})
		return
	}

	//verif recipe existe
	helper := HandlersHelper{DB: h.DB}

	if !helper.RecipeExists(c, input.Id_recipe) {
		return
	}

	//verif ingredient exist

	if !helper.IngredientExists(c, input.Id_ingredient) {
		return
	}

	var include_ingredient Models.Include_ingredient

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.DB.QueryRow(
		timed_ctx,
		`INSERT INTO include_ingredients (id_ingredient, id_recipe, count) VALUES ($1, $2, $3)
		 RETURNING id_ingredient, id_recipe, count`,
		input.Id_ingredient, input.Id_recipe, input.Count,
	).Scan(&include_ingredient.Id_ingredient, &include_ingredient.Id_recipe, &include_ingredient.Count)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				c.JSON(http.StatusConflict, gin.H{"error": "Cet ingrédient est déjà associé à la recette"})
				return
			case "23503":
				c.JSON(http.StatusNotFound, gin.H{"error": "La recette ou l'ingrédient n'existe plus"})
				return
			}
		}
		log.Printf("Création d'une association : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de créer la relation",
		})
		return
	}

	c.JSON(http.StatusCreated, include_ingredient)
}
