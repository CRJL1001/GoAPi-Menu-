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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	Models "github.com/m/models"
)

type IngredientsHandler struct {
	DB *pgxpool.Pool
}

func (h *IngredientsHandler) Create(c *gin.Context) {
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

	var ingredient Models.Ingredient

	err := h.DB.QueryRow( //SQL request
		ctx,
		`INSERT INTO ingredients (name) VALUES ($1) RETURNING id, name`,
		name,
	).Scan(&ingredient.ID, &ingredient.Name) //save response in ingredient struct var

	if err != nil { //raise internal error
		log.Printf("Création d'un ingrédient %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de créer l'ingrédient",
		})
		return
	}

	c.Header("Location", fmt.Sprintf("/ingredients/%d", ingredient.ID))
	c.JSON(http.StatusCreated, ingredient) //return valid request
}

func (h *IngredientsHandler) Show(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64) //parse str id to int id base 10 on 64 bits

	if err != nil || id <= 0 { //ID Verif
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Identifiant invalide",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) //setting timeout params
	defer cancel()                                                         //defer context cancelling

	var ingredient Models.Ingredient //response struct var

	err = h.DB.QueryRow( //SQL queries
		ctx,
		`SELECT id, name FROM ingredients WHERE id = $1`,
		id,
	).Scan(&ingredient.ID, &ingredient.Name) //adding to struct var

	if errors.Is(err, pgx.ErrNoRows) { //checking if error is ID not found
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ingrédient introuvable",
		})
		return
	}
	if err != nil {
		log.Printf("Lecture d'un ingrédient : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de récupérer l'ingrédient",
		})
		return
	}
	c.JSON(http.StatusOK, ingredient) //returning ingredient
}

func (h *IngredientsHandler) Index(c *gin.Context) {

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5 * time.Second)
	defer cancel()

	rows, err2 := h.DB.Query(
		timed_ctx,
		`SELECT id, name FROM ingredients`,
	)

	if err2 != nil {
		log.Printf("Lecture des ingrédients impossible : %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Lecture des ingrédients impossible",
		})
		return
	}

	defer rows.Close()

	ingredients := make([]Models.Ingredient, 0)

	for rows.Next() {
		var ingredient Models.Ingredient

		if err := rows.Scan(
			&ingredient.ID,
			&ingredient.Name,
		); err != nil {
			log.Printf("Impossible de lire l'ingrédient : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error" : "Impossible de lire l'ingrédient",
			})
			return
		}
		ingredients = append(ingredients, ingredient)
	}

	if err := rows.Err(); err != nil { //récup les erreurs sur les rows
		log.Printf("Lecture des lignes : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Echec de la lecture des lignes",
		})
		return
	}
	c.JSON(http.StatusOK, ingredients)
}

func (h *IngredientsHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"),10, 64)

	if err != nil || id <= 0 {
		log.Printf(" Identifiant Invalide : %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error" : "Identifiant invalide",
		})
		return
	}

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5 * time.Second)
	defer cancel()

	result, err := h.DB.Exec(
		timed_ctx,
		`DELETE FROM ingredients WHERE id = $1`,
		id,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			log.Printf("Cet ingrédient est utilisé dans une recette")
			c.JSON(http.StatusConflict, gin.H{
				"error" : "Cet ingrédient est utilisé dans une recette",
			})
			return
		}
		log.Printf("Suppresion d'un ingredient : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Impossible de supprimer l'ingrédient",
		})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error" : "Ingrédient introuvable",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *IngredientsHandler) Update (c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		log.Printf("Identifiant invalide")
		c.JSON(http.StatusBadRequest, gin.H{
			"error" : "Identifiant invalide",
		})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		4096,
	)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error" : "Corps JSON invalide : name est obligatoire",
		})
		return
	}

	name := strings.TrimSpace(input.Name)

	if name == "" ||
		utf8.RuneCountInString(name) > 100 ||
		strings.ContainsRune(name, '\x00') {
			c.JSON(http.StatusBadRequest, gin.H {
				"error" : "le nom doit contenir entre 1 et 100 caractères, sans nul",
			})
			return
	}

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5 * time.Second)
	defer cancel()

	var ingredient Models.Ingredient

	err = h.DB.QueryRow(
		timed_ctx,
		`UPDATE ingredients SET name = $1 WHERE id = $2 RETURNING id, name`,
		name, id,
	).Scan(&ingredient.ID, &ingredient.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error" : "Ingrédient introuvable",
		})
		return
	}

	if err != nil {
		log.Printf("Modification d'un ingredient : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Impossible de modifier l'ingrédient",
		})
		return
	}
	c.JSON(http.StatusOK, ingredient)
}
