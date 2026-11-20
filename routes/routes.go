package Routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	Handlers "github.com/m/handlers"
)

func Register(router *gin.Engine, db *pgxpool.Pool) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route introuvable"})
	})
	ingredient_handler := &Handlers.IngredientsHandler{DB: db}
	menu_handler := &Handlers.MenusHandler{DB: db}
	recipe_handler := &Handlers.RecipiesHandler{DB: db}

	ingredients := router.Group("/ingredients")
	ingredients.POST("", ingredient_handler.Create)
	ingredients.GET("/:id", ingredient_handler.Show)
	menus := router.Group("/menus")
	menus.POST("", menu_handler.Create)
	menus.GET("/:id", menu_handler.Show)
	recipies := router.Group("/recipies")
	recipies.POST("", recipe_handler.Create)
	recipies.GET("/:id", recipe_handler.Show)
}
