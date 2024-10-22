package handlers

import (
	"html/template"
	"net/http"

	"github.com/gobuffalo/plush"
)

func Custom404Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	ctx := plush.NewContext()

	// Load the base layout
	baseLayout, err := template.ParseFiles("templates/layouts/base.plush.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Load the 404 template
	notFoundTemplate, err := plush.Parse("templates/404.plush.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the 404 template
	notFoundContent, err := notFoundTemplate.Exec(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the 404 content in the base layout context
	ctx.Set("yield", template.HTML(notFoundContent))

	// Execute the base layout
	err = baseLayout.Execute(w, ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
