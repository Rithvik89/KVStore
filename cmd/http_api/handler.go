package main

import "github.com/go-chi/chi"

func (app *App) InitializeHandler() *chi.Mux {
	R := app.Handler

	// Add your routes here
	R.Get("/api/v1/", app.ReadRecords)
	R.Post("/api/v1/", app.WriteRecord)

	return R
}
