package routes

import (
	"kabancount/internal/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/healthcheck", app.HealthCheck)

	r.Post("/auth/signin", app.TokenHandler.HandleCreateToken)
	r.Post("/auth/signup", app.AuthHandler.HandleRegister)

	r.Group(func(r chi.Router) {
		r.Use(app.MiddlewareHandler.Authenticate)
		r.Use(app.MiddlewareHandler.RequireAuthenticatedUser)

		r.Group(func(r chi.Router) {
			r.Use(app.MiddlewareHandler.RequireAdminUser)

			r.Post("/organizations", app.OrganizationHandler.HandleCreateOrganization)
			r.Get("/organizations/{id}", app.OrganizationHandler.HandleGetOrganizationByID)
			r.Put("/organizations/{id}", app.OrganizationHandler.HandleUpdateOrganization)
			r.Delete("/organizations/{id}", app.OrganizationHandler.HandleDeleteOrganization)
		})

		r.Post("/users", app.UserHandler.HandleCreateUser)

		r.Post("/locations", app.LocationHandler.HandleCreateLocation)
		r.Get("/locations", app.LocationHandler.HandleGetLocationsByOrganization)

		r.Post("/items", app.ItemHandler.HandleCreateItem)
		r.Get("/items", app.ItemHandler.HandleGetItemsByOrganization)
		r.Get("/items/{id}", app.ItemHandler.HandleGetItemByID)
		r.Put("/items/{id}", app.ItemHandler.HandleUpdateItem)
		r.Delete("/items/{id}", app.ItemHandler.HandleDeleteItem)

		r.Post("/categories", app.CategoryHandler.HandleCreateCategory)
		r.Get("/categories", app.CategoryHandler.HandleGetCategoriesByOrganization)
		r.Get("/categories/{id}", app.CategoryHandler.HandleGetCategoryByID)
		r.Put("/categories/{id}", app.CategoryHandler.HandleUpdateCategory)
		r.Delete("/categories/{id}", app.CategoryHandler.HandleDeleteCategory)

	})

	return r
}
