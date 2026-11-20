package Routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	Handlers "github.com/m/handlers"
)
func Register(router *gin.Engine, db *pgxpool.Pool){
	h := &Handlers.IngredientsHandler{DB: db}

	ingredients := router.Group("/ingredients")
	ingredients.POST("", h.Create)
	ingredients.GET("/:id", h.Show)
}
