package Handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	Models "github.com/m/models"
)

type RecipiesHandler struct {
	DB *pgxpool.Pool
}

func (h *RecipiesHandler) Create(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		4096,
	)

	if err := c.ShouldBindJSON(&input); err != nil { //verif json binds
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Corps JSON invalide : name est obligatoire",
		})
		return
	}

	name := strings.TrimSpace(input.Name) //trim space

	if name == "" || //vérif format
		utf8.RuneCountInString(name) > 100 ||
		strings.ContainsRune(name, '\x00') {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Le nom doit contenir entre 1 et 100 caractères, sans nul",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) //timeout def
	defer cancel()                                                         //defer close context

	var recipe Models.Recipe

	err := h.DB.QueryRow( //SQL request
		ctx,
		`INSERT INTO recipies (name) VALUES ($1) RETURNING id, name`,
		name,
	).Scan(&recipe.ID, &recipe.Name) //save response in menu struct var

	if err != nil { //raise internal error
		log.Printf("Création d'une recette %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de créer la recette",
		})
		return
	}

	c.Header("Location", fmt.Sprintf("/menus/%d", recipe.ID))
	c.JSON(http.StatusCreated, recipe) //return valid request
}

func (h *RecipiesHandler) Show(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64) //parse str id to int id base 10 on 64 bits

	if err != nil || id <= 0 { //ID Verif
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Identifiant invalide",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) //setting timeout params
	defer cancel()                                                         //defer context cancelling

	var recipe Models.Recipe //response struct var

	err = h.DB.QueryRow( //SQL queries
		ctx,
		`SELECT id, name FROM recipies WHERE id = $1`,
		id,
	).Scan(&recipe.ID, &recipe.Name) //adding to struct var

	if errors.Is(err, pgx.ErrNoRows) { //checking if error is ID not found
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recette introuvable",
		})
		return
	}
	if err != nil {
		log.Printf("Lecture d'une recette : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de récupérer la recette",
		})
		return
	}
	c.JSON(http.StatusOK, recipe) //returning recipe
}

func (h *RecipiesHandler) ShowIngredients(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Identifiant invalide",
		})
		return
	}

	helper := HandlersHelper{DB: h.DB}

	//verif recipe exists
	if !helper.RecipeExists(c, id) {
		return
	}

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(
		timed_ctx,
		`SELECT id_recipe, id_ingredient, count
			FROM include_ingredients
			WHERE id_recipe = $1
			ORDER BY id_ingredient
		`,
		id,
	)

	if err != nil {
		log.Printf("Lecture des associations : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de récupérer les ingrédients de la recette",
		})
		return
	}
	defer rows.Close()

	ingredients := make([]Models.Include_ingredient, 0)

	for rows.Next() {
		var ingredient Models.Include_ingredient

		if err := rows.Scan(
			&ingredient.Id_recipe,
			&ingredient.Id_ingredient,
			&ingredient.Count,
		); err != nil {
			log.Printf("Lecture d'une association : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Impossible de lire les ingrédients",
			})
			return
		}
		ingredients = append(ingredients, ingredient)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Lecture des association : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Lecture des associations a échoué",
		})
		return
	}
	c.JSON(http.StatusOK, ingredients)
}
