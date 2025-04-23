package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
  // read relative to cwd, so run from src/backend
  data, err := os.ReadFile("json/recipe.json")
  if err != nil {
    log.Fatalf("failed to read JSON: %v", err)
  }
  log.Printf("loaded %d bytes of recipe.json", len(data))

  http.HandleFunc("/api/recipes", func(w http.ResponseWriter, r *http.Request) {
    // allow cross‑origin from your React dev server
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")
    w.Write(data)
  })

  // serve SVG directory at /svgs/*
  fs := http.FileServer(http.Dir("svgs"))
  http.Handle("/svgs/", http.StripPrefix("/svgs/", fs))

  log.Println("Go API listening on :8080")
  log.Fatal(http.ListenAndServe(":8080", nil))
}