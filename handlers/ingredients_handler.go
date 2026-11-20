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

type IngredientsHandler struct {
	DB *pgxpool.Pool
}

func (h *IngredientsHandler) Create (c *gin.Context){
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
		strings.ContainsRune(name, '\x00'){
			c.JSON(http.StatusBadRequest, gin.H{
				"error" : "Le nom doit contenir entre 1 et 100 caractères, sans nul",
			})
			return
		}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) //timeout def
	defer cancel() //defer close context

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

func (h *IngredientsHandler) Show (c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64) //parse str id to int id base 10 on 64 bits

	if err != nil || id <= 0 { //ID Verif
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Identifiant invalide",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) //setting timeout params
	defer cancel() //defer context cancelling

	var ingredient Models.Ingredient //response struct var

	err = h.DB.QueryRow( //SQL queries
		ctx,
		`SELECT id, name FROM ingredients WHERE id = $1`,
		id,
	).Scan(&ingredient.ID, &ingredient.Name) //adding to struct var

	if errors.Is(err, pgx.ErrNoRows) { //checking if error is ID not found
		c.JSON(http.StatusNotFound, gin.H{
			"error" : "Ingrédient introuvable",
		})
		return
	}
	c.JSON(http.StatusOK, ingredient) //returning ingredient
}
