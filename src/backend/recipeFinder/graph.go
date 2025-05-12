package recipeFinder

func BuildGraphFromCatalog(cat Catalog) Graph {
	graph := make(Graph)
	for _, tier := range cat.Tiers {
		for _, el := range tier.Elements {
			for _, rec := range el.Recipes {
				if len(rec) != 2 {
					continue
				}
				a, b := rec[0], rec[1]
				// Add both directions
				graph[a] = append(graph[a], Neighbor{Partner: b, Product: el.Name})
				graph[b] = append(graph[b], Neighbor{Partner: a, Product: el.Name})
			}
		}
	}
	return graph
}

// BuildIndexedGraph creates an optimized graph representation from the catalog.
// This function converts the Catalog structure into a more efficient IndexedGraph
// using integer IDs to speed up searches and reduce memory usage.
//
// The conversion is done in two phases:
// 1. First phase: Assign IDs to each element name, prioritizing base elements
// 2. Second phase: Build graph edges connecting elements based on recipes
//
// Parameters:
//   - cat: Catalog structure containing all element data and recipes
//
// Returns:
//   - IndexedGraph: Optimized graph representation with integer IDs
func BuildIndexedGraph(cat Catalog) IndexedGraph {
	// First phase: assign IDs to all element names
	nameToID := make(map[string]int) // Maps element names to integer IDs
	idToName := make(map[int]string) // Maps integer IDs back to element names

	// Start assigning IDs sequentially
	nextID := 0

	// Ensure base elements (BaseElements) get IDs first
	// This is important because BFS and DFS algorithms will start from base elements
	baseIDs := make([]int, len(BaseElements))
	for i, name := range BaseElements {
		nameToID[name] = nextID
		idToName[nextID] = name
		baseIDs[i] = nextID
		nextID++
	}

	// Then assign IDs for all other elements in the catalog
	// We traverse each tier and the elements within it
	for _, tier := range cat.Tiers {
		for _, el := range tier.Elements {
			// Check if element already has an ID
			if _, exists := nameToID[el.Name]; !exists {
				nameToID[el.Name] = nextID
				idToName[nextID] = el.Name
				nextID++
			}

			// Also assign IDs for all ingredient names in recipes
			// This ensures all elements appearing in recipes have IDs
			for _, rec := range el.Recipes {
				for _, ingredient := range rec {
					if _, exists := nameToID[ingredient]; !exists {
						nameToID[ingredient] = nextID
						idToName[nextID] = ingredient
						nextID++
					}
				}
			}
		}
	}

	// Second phase: building graph edges based on recipes
	// Each recipe (A+B->C) will be added as an edge from A to B with result C
	// and also from B to A with the same result
	edges := make(map[int][]IndexedNeighbor)

	for _, tier := range cat.Tiers {
		for _, el := range tier.Elements {
			productID := nameToID[el.Name] // ID of the product (combination result)
			productTier := getElementTier(el.Name)

			for _, rec := range el.Recipes {
				// Ensure recipe consists of 2 ingredients
				if len(rec) != 2 {
					continue
				}

				aID := nameToID[rec[0]] // First ingredient ID
				bID := nameToID[rec[1]] // Second ingredient ID

				aTier := getElementTier(rec[0])
				bTier := getElementTier(rec[1])

				// Check if this is a valid recipe (ingredients not higher tier than product)
				// Either ingredient tier should not exceed product tier
				if aTier > productTier || bTier > productTier {
					// Skip this recipe - it doesn't make logical sense
					continue
				}

				// Add edges to the graph in both directions
				// Because A+B=C and B+A=C are the same
				edges[aID] = append(edges[aID], IndexedNeighbor{
					PartnerID: bID,       // Second ingredient
					ProductID: productID, // Resulting product
				})

				edges[bID] = append(edges[bID], IndexedNeighbor{
					PartnerID: aID,       // First ingredient
					ProductID: productID, // Resulting product
				})
			}
		}
	}

	// Return the complete IndexedGraph structure
	return IndexedGraph{
		NameToID: nameToID, // Name to ID mapping
		IDToName: idToName, // ID to name mapping
		Edges:    edges,    // Graph edges with IDs
	}
}

// GetBaseElementIDs returns a list of integer IDs for all base elements.
// This function is useful for accessing base elements (Air, Earth, Fire, Water)
// in their ID form, which is needed for search algorithms.
//
// Returns:
//   - []int: Slice containing integer IDs for all base elements
func (g *IndexedGraph) GetBaseElementIDs() []int {
	ids := make([]int, len(BaseElements))
	for i, name := range BaseElements {
		ids[i] = g.NameToID[name]
	}
	return ids
}

// Global variable to cache element tiers
var elementTierCache map[string]int

// InitElementTiers builds a mapping of elements to their tier levels
func InitElementTiers(cat Catalog) {
	elementTierCache = make(map[string]int)

	// Add base elements with tier 0
	for _, base := range BaseElements {
		elementTierCache[base] = 0
	}

	// Process catalog tiers
	for tierIndex, tier := range cat.Tiers {
		tierLevel := tierIndex + 1 // Tier levels start at 1 (tier 0 is base elements)

		// Add all elements in this tier to the cache
		for _, element := range tier.Elements {
			elementTierCache[element.Name] = tierLevel
		}
	}
}

// getElementTier returns the tier level of an element
// Base elements have tier 0, and higher tiers increase from there
func getElementTier(element string) int {
	// Check if element exists in cache
	if tier, exists := elementTierCache[element]; exists {
		return tier
	}

	// If not in cache, check if it's a base element
	for _, base := range BaseElements {
		if element == base {
			return 0
		}
	}

	// Unknown element, return a high value to avoid using it
	return 999
}
