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

type MenusHandler struct {
	DB *pgxpool.Pool
}

func (h *MenusHandler) Create(c *gin.Context) {
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

	var menu Models.Menu

	err := h.DB.QueryRow( //SQL request
		ctx,
		`INSERT INTO menus (name) VALUES ($1) RETURNING id, name`,
		name,
	).Scan(&menu.ID, &menu.Name) //save response in menu struct var

	if err != nil { //raise internal error
		log.Printf("Création d'un menu %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de créer le menu",
		})
		return
	}

	c.Header("Location", fmt.Sprintf("/menus/%d", menu.ID))
	c.JSON(http.StatusCreated, menu) //return valid request
}

func (h *MenusHandler) Show(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64) //parse str id to int id base 10 on 64 bits

	if err != nil || id <= 0 { //ID Verif
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Identifiant invalide",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) //setting timeout params
	defer cancel()                                                         //defer context cancelling

	var menu Models.Menu //response struct var

	err = h.DB.QueryRow( //SQL queries
		ctx,
		`SELECT id, name FROM menus WHERE id = $1`,
		id,
	).Scan(&menu.ID, &menu.Name) //adding to struct var

	if errors.Is(err, pgx.ErrNoRows) { //checking if error is ID not found
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Menu introuvable",
		})
		return
	}
	if err != nil {
		log.Printf("Lecture d'un menu : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de récupérer le menu",
		})
		return
	}
	c.JSON(http.StatusOK, menu) //returning menu
}

func (h *MenusHandler) ShowRecipies(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Identifiant invalide",
		})
		return
	}

	helper := HandlersHelper{DB: h.DB}

	//verif recipe exists
	if !helper.MenuExists(c, id) {
		return
	}

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(
		timed_ctx,
		`SELECT id_recipe, id_menu, count
			FROM include_recipies
			WHERE id_menu = $1
			ORDER BY id_recipe
		`,
		id,
	)

	if err != nil {
		log.Printf("Lecture des associations : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Impossible de récupérer les recettes du menu",
		})
		return
	}
	defer rows.Close()

	recipies := make([]Models.Include_recipe, 0)

	for rows.Next() {
		var recipe Models.Include_recipe

		if err := rows.Scan(
			&recipe.Id_recipe,
			&recipe.Id_menu,
			&recipe.Count,
		); err != nil {
			log.Printf("Lecture d'une association : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Impossible de lire les recettes",
			})
			return
		}
		recipies = append(recipies, recipe)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Lecture des association : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Lecture des associations a échoué",
		})
		return
	}
	c.JSON(http.StatusOK, recipies)
}

func (h *MenusHandler) Index(c *gin.Context) {

	timed_ctx, cancel := context.WithTimeout(c.Request.Context(), 5 * time.Second)
	defer cancel()

	rows, err2 := h.DB.Query(
		timed_ctx,
		`SELECT id, name FROM menus`,
	)

	if err2 != nil {
		log.Printf("Lecture des menus impossible : %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Lecture des menus impossible",
		})
		return
	}

	defer rows.Close()

	menus := make([]Models.Menu, 0)

	for rows.Next() {
		var menu Models.Menu

		if err := rows.Scan(
			&menu.ID,
			&menu.Name,
		); err != nil {
			log.Printf("Impossible de lire le menu : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error" : "Impossible de lire le menu",
			})
			return
		}
		menus = append(menus, menu)
	}

	if err := rows.Err(); err != nil { //récup les erreurs sur les rows
		log.Printf("Lecture des lignes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Echec de la lecture des lignes",
		})
		return
	}

	c.JSON(http.StatusOK, menus)
}

func (h *MenusHandler) Delete(c *gin.Context) {
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
		`DELETE FROM menus WHERE id = $1`,
		id,
	)
	if err != nil {
		log.Printf("Suppresion d'un menu : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Impossible de supprimer le menu",
		})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error" : "menu introuvable",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MenusHandler) Update (c *gin.Context) {
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

	var menu Models.Menu

	err = h.DB.QueryRow(
		timed_ctx,
		`UPDATE menus SET name = $1 WHERE id = $2 RETURNING id, name`,
		name, id,
	).Scan(&menu.ID, &menu.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error" : "Menu introuvable",
		})
		return
	}

	if err != nil {
		log.Printf("Modification d'un menu : %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error" : "Impossible de modifier le menu",
		})
		return
	}
	c.JSON(http.StatusOK, menu)
}
