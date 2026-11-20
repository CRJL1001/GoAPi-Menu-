package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	Database "github.com/m/Database"
	Routes "github.com/m/routes"

	"github.com/gin-gonic/gin"
)

func main(){
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //global context timeout set

	db, err := Database.Connect(ctx, os.Getenv("DATABASE_URL")) //testing db connexion
	cancel() //closing context wich was only use for testing db

	if err != nil { //verif DB connexion
		log.Fatalf("Connexion PostgreSQL impossible %v", err)
	}
	defer db.Close() //defer closing DB

	router := gin.Default() //creating router

	if err := router.SetTrustedProxies(nil); err != nil { //setting trusted proxy
		log.Fatal(err)
	}

	Routes.Register(router, db) //setting routes

	server := &http.Server{ //setting server params
		Addr: ":8080",
		Handler: router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 1 * time.Minute,
	}

	log.Println("API démarrée sur le port 8080")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { //verif & launching server
		log.Printf("Erreur du serveur : %v", err)
	}
}
