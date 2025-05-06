package recipeFinder

// IndexedNeighbor uses integer IDs instead of strings for faster lookups
type IndexedNeighbor struct {
    PartnerID int
    ProductID int
}

// IndexedGraph uses integer IDs for faster lookups and less memory
type IndexedGraph struct {
    NameToID  map[string]int      // Maps element names to their ID
    IDToName  map[int]string      // Reverse mapping for reconstruction
    Edges     map[int][]IndexedNeighbor // Adjacency list using IDs
}

// BuildIndexedGraph creates an optimized graph representation from the catalog
func BuildIndexedGraph(cat Catalog) IndexedGraph {
    // First pass: assign IDs to all element names
    nameToID := make(map[string]int)
    idToName := make(map[int]string)
    
    // Start assigning IDs
    nextID := 0
    
    // Ensure base elements get IDs first
    baseIDs := make([]int, len(baseElements))
    for i, name := range baseElements {
        nameToID[name] = nextID
        idToName[nextID] = name
        baseIDs[i] = nextID
        nextID++
    }
    
    // Then assign IDs to all other elements
    for _, tier := range cat.Tiers {
        for _, el := range tier.Elements {
            if _, exists := nameToID[el.Name]; !exists {
                nameToID[el.Name] = nextID
                idToName[nextID] = el.Name
                nextID++
            }
            
            // Also assign IDs to all ingredient names
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
    
    // Second pass: build the graph edges
    edges := make(map[int][]IndexedNeighbor)
    
    for _, tier := range cat.Tiers {
        for _, el := range tier.Elements {
            productID := nameToID[el.Name]
            
            for _, rec := range el.Recipes {
                if len(rec) != 2 {
                    continue
                }
                
                aID := nameToID[rec[0]]
                bID := nameToID[rec[1]]
                
                // Add both directions
                edges[aID] = append(edges[aID], IndexedNeighbor{
                    PartnerID: bID,
                    ProductID: productID,
                })
                
                edges[bID] = append(edges[bID], IndexedNeighbor{
                    PartnerID: aID,
                    ProductID: productID,
                })
            }
        }
    }
    
    return IndexedGraph{
        NameToID:  nameToID,
        IDToName:  idToName,
        Edges:     edges,
    }
}

func (g *IndexedGraph) GetBaseElementIDs() []int {
    ids := make([]int, len(baseElements))
    for i, name := range baseElements {
        ids[i] = g.NameToID[name]
    }
    return ids
}