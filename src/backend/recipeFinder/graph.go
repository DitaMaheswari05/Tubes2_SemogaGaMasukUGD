package recipeFinder

type Neighbor struct {
	Partner string
	Product string
}

type Graph map[string][]Neighbor

func BuildGraph(byPair map[Pair][]string) Graph {
	graph := make(Graph, len(byPair)*2)
	for pair, products := range byPair {
		for _, prod := range products {
			graph[pair.A] = append(graph[pair.A], Neighbor{Partner: pair.B, Product: prod})
			graph[pair.B] = append(graph[pair.B], Neighbor{Partner: pair.A, Product: prod})
		}
	}
	return graph
}
