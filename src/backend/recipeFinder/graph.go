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