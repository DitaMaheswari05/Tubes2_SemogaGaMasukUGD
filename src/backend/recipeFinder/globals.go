// globals/globals.go
package recipeFinder

// ==================== BASE ELEMENTS ====================
// Starting elements
var BaseElements = []string{"Air", "Earth", "Fire", "Water"}

// ==================== GRAPH TYPES ====================
// Regular graph representations
type Neighbor struct {
    Partner string
    Product string
}

type Graph map[string][]Neighbor

// Optimized indexed graph for faster lookups
type IndexedNeighbor struct {
    PartnerID int
    ProductID int
}

type IndexedGraph struct {
    NameToID map[string]int          	// Maps element names to their ID
    IDToName map[int]string          	// Reverse mapping for reconstruction
    Edges    map[int][]IndexedNeighbor 	// Adjacency list using IDs
}

// ==================== RECIPE TYPES ====================
// Ingredient pair
type IngredientCombo struct {
    A string `json:"a"`
    B string `json:"b"`
}

// Recipe info with parent/partner format
type Info struct {
    Parent, Partner string
    Path            [][]string
}

// Recipe step using IngredientCombo
type RecipeStep struct {
    Combo IngredientCombo `json:"combo"`          // the two ingredients
    Path  [][]string      `json:"path,omitempty"` // full sequence: each is [parent,partner,product]
}

// ==================== MAPPING TYPES ====================
// Maps from product to its ingredients (single recipe)
type ProductToIngredients map[string]RecipeStep

// Maps from product to multiple possible recipes
type ProductToMultipleIngredients map[string][]RecipeStep

// ==================== GLOBALS ====================
// Global indexed graph accessible throughout the package
var GlobalIndexedGraph IndexedGraph

// Global variable to store the catalog
var GlobalCatalog Catalog