package recipeFinder

type Neighbor struct {
	Partner string
	Product string
}

type Graph map[string][]Neighbor

func BuildGraph(combinationMap CombinationMap) Graph {
	graph := make(Graph, len(combinationMap)*2)
	for pair, products := range combinationMap {
		for _, prod := range products {
			graph[pair.A] = append(graph[pair.A], Neighbor{Partner: pair.B, Product: prod})
			graph[pair.B] = append(graph[pair.B], Neighbor{Partner: pair.A, Product: prod})
		}
	}
	return graph
}
