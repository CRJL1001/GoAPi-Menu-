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
	include_ingredient_handler := &Handlers.IncludeIngredientHandler{DB: db}
	include_recipe_handler := &Handlers.IncludeRecipeHandler{DB: db}

	ingredients := router.Group("/ingredients")
	ingredients.POST("/create", ingredient_handler.Create)
	ingredients.GET("", ingredient_handler.Index)
	ingredients.GET("/:id", ingredient_handler.Show)
	ingredients.DELETE("/:id", ingredient_handler.Delete)
	ingredients.PATCH("/:id", ingredient_handler.Update)
	menus := router.Group("/menus")
	menus.POST("/create", menu_handler.Create)
	menus.GET("", menu_handler.Index)
	menus.GET("/:id", menu_handler.Show)
	menus.POST("/add_recipe_to_menu", include_recipe_handler.Create)
	menus.DELETE("/:id", menu_handler.Delete)
	menus.GET("/:id/recipies", menu_handler.ShowRecipies)
	menus.PATCH("/:id", menu_handler.Update)
	menus.GET("/:id/ingredients", menu_handler.ShowIngredients)
	recipies := router.Group("/recipies")
	recipies.POST("/create", recipe_handler.Create)
	recipies.GET("", recipe_handler.Index)
	recipies.GET("/:id", recipe_handler.Show)
	recipies.POST("/add_ingredient_to_recipe", include_ingredient_handler.Create)
	recipies.GET("/:id/ingredients", recipe_handler.ShowIngredients)
	recipies.DELETE("/:id", recipe_handler.Delete)
	recipies.PATCH("/:id", recipe_handler.Update)
}
