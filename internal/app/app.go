package app

import (
	"database/sql"
	"fmt"
	"kabancount/internal/api"
	"kabancount/internal/config"
	"kabancount/internal/middleware"
	"kabancount/internal/store"
	"kabancount/migrations"
	"log"
	"net/http"
	"os"
)

type Application struct {
	Logger              *log.Logger
	UserHandler         *api.UserHandler
	OrganizationHandler *api.OrganizationHandler
	TokenHandler        *api.TokenHandler
	AuthHandler         *api.AuthHandler
	ItemHandler         *api.ItemHandler
	CategoryHandler     *api.CategoryHandler
	MiddlewareHandler   middleware.UserMiddleware
	DB                  *sql.DB
}

func NewApplication() (*Application, error) {
	_, err := config.Load()
	if err != nil {
		panic(err)
	}

	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// our stores will go here
	userStore := store.NewPostgresUserStore(pgDB)
	tokenStore := store.NewPostgresTokenStore(pgDB)
	organizationStore := store.NewPostgresOrganizationStore(pgDB)
	itemStore := store.NewPostgresItemStore(pgDB)
	categoryStore := store.NewPostgresCategoryStore(pgDB)

	// our handlers will go here
	userHandler := api.NewUserHandler(userStore, logger)
	organizationHandler := api.NewOrganizationHandler(organizationStore, logger)
	tokenHandler := api.NewTokenHandler(tokenStore, userStore, logger)
	authHandler := api.NewAuthHandler(organizationStore, userStore, logger)
	middlewareHandler := middleware.UserMiddleware{UserStore: userStore}
	itemHandler := api.NewItemHandler(itemStore, logger)
	categoryHandler := api.NewCategoryHandler(categoryStore, logger)

	app := &Application{
		Logger:              logger,
		UserHandler:         userHandler,
		OrganizationHandler: organizationHandler,
		TokenHandler:        tokenHandler,
		AuthHandler:         authHandler,
		ItemHandler:         itemHandler,
		CategoryHandler:     categoryHandler,
		MiddlewareHandler:   middlewareHandler,
		DB:                  pgDB,
	}

	return app, nil
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status is available\n")
}
