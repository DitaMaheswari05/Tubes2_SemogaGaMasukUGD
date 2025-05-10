package recipeFinder

import (
	"container/list"
	"fmt"
	"strings"
)

/* -------------------------------------------------------------------------
   1.  Satu jalur tercepat (product → ingredients)                           */
// func IndexedBFSBuild(target string, g IndexedGraph) (ProductToIngredients, int) {
// 	targetID := g.NameToID[target]

// 	q    := list.New()
// 	seen := make(map[int]bool)

// 	for _, b := range BaseElements {
// 		id := g.NameToID[b]
// 		q.PushBack(id)
// 		seen[id] = true
// 	}

// 	prev  := make(map[int]struct{ ParentID, PartnerID int })
// 	nodes := 0

// 	for q.Len() > 0 {
// 		cur := q.Remove(q.Front()).(int)
// 		nodes++
// 		if cur == targetID {
// 			break
// 		}
// 		for _, e := range g.Edges[cur] { // cur + partner → product
// 			if seen[e.PartnerID] && !seen[e.ProductID] {
// 				seen[e.ProductID] = true
// 				prev[e.ProductID] = struct{ ParentID, PartnerID int }{
// 					ParentID: cur, PartnerID: e.PartnerID}
// 				q.PushBack(e.ProductID)
// 			}
// 		}
// 	}

// 	out := make(ProductToIngredients)
// 	for prod, p := range prev {
// 		out[g.IDToName[prod]] = RecipeStep{
// 			Combo: IngredientCombo{
// 				A: g.IDToName[p.ParentID],
// 				B: g.IDToName[p.PartnerID],
// 			},
// 		}
// 	}
// 	return out, nodes
// }

func IndexedBFSBuild(targetName string, graph IndexedGraph) (ProductToIngredients, int) {
	targetID := graph.NameToID[targetName]

	queue := list.New()
	seen := make(map[int]bool)

	for _, baseName := range BaseElements {
		baseID := graph.NameToID[baseName]
		queue.PushBack(baseID)
		seen[baseID] = true
	}

	// Track parents using integer IDs
	prevIDs := make(map[int]struct{ParentID, PartnerID int})
	nodes := 0
	for queue.Len() > 0 {
		curID := queue.Remove(queue.Front()).(int)
		nodes++
		
		if curID == targetID {
			break
		}
		
		for _, neighbor := range graph.Edges[curID] {
			partnerID := neighbor.PartnerID
			productID := neighbor.ProductID
			
			if seen[partnerID] && !seen[productID] {
				seen[productID] = true
				prevIDs[productID] = struct{ParentID, PartnerID int}{
					ParentID:  curID,
					PartnerID: partnerID,
				}
				queue.PushBack(productID)
			}
		}
	}

	// Convert integer results to ProductToIngredients
	recipes := make(ProductToIngredients)
	for productID, info := range prevIDs {
		productName := graph.IDToName[productID]
		parentName := graph.IDToName[info.ParentID]
		partnerName := graph.IDToName[info.PartnerID]
		
		recipes[productName] = RecipeStep{
			Combo: IngredientCombo{
				A: parentName,
				B: partnerName,
			},
			// Path is nil here since we're not tracking full paths in this function
		}
	}

	return recipes, nodes
}

/* -------------------------------------------------------------------------
   2.  Jalur ke‑(skip+1) (distinct)                                          */
func findKthPathIndexed(targetID, skip int, g IndexedGraph) (RecipeStep, int) {
	type state struct {
		elem  int
		path  [][]int // sequence of [a,b,product]
		depth int
	}
	type edge struct{ PartnerID, ProductID int }

	reachable := make(map[int]bool)
	waiting   := make(map[int][]edge)
	seenPath  := make(map[string]bool)

	q := list.New()
	for _, b := range BaseElements {
		id := g.NameToID[b]
		reachable[id] = true
		q.PushBack(state{elem: id})
	}

	nodes, hits := 0, 0
	const maxDepth = 40

	enqueue := func(a, b, c int, cur [][]int, d int) {
		np := append(append([][]int(nil), cur...), []int{min(a, b), max(a, b), c})
		sig := canonicalHash(np)
		if seenPath[sig] {
			return
		}
		seenPath[sig] = true
		q.PushBack(state{elem: c, path: np, depth: d + 1})
	}

	for q.Len() > 0 {
		st := q.Remove(q.Front()).(state)
		nodes++
		if st.depth > maxDepth {
			continue
		}
		if st.elem == targetID {
			if hits == skip {
				return buildRecipeStepFromPath(st.path, targetID, g), nodes
			}
			hits++
			continue
		}

		for _, r := range g.Edges[st.elem] {
			if reachable[r.PartnerID] {
				enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
			} else {
				waiting[r.PartnerID] = append(waiting[r.PartnerID], edge{
					PartnerID: st.elem, ProductID: r.ProductID})
			}
		}
		if lst, ok := waiting[st.elem]; ok {
			for _, r := range lst {
				if reachable[r.PartnerID] {
					enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
				}
			}
			delete(waiting, st.elem)
		}
		reachable[st.elem] = true
	}
	return RecipeStep{}, nodes
}

/* -------------------------------------------------------------------------
   3.  Beberapa jalur berurutan                                              */
func RangePathsIndexed(targetID, start, limit int, g IndexedGraph) ([]RecipeStep, int) {
	out   := make([]RecipeStep, 0, limit)
	total := 0
	for k := 0; k < limit; k++ {
		step, visited := findKthPathIndexed(targetID, start+k, g)
		total += visited
		if step.Path == nil {
			break
		}
		out = append(out, step)
	}
	return out, total
}

/* -------------------------------------------------------------------------
   4.  Sapu sekali dapat N jalur unik                                        */
func findDistinctPathsIndexed(targetID int, maxPaths int64, g IndexedGraph) ([]RecipeStep, int) {
	type state struct {
		elem  int
		path  [][]int
		depth int
	}
	type edge struct{ PartnerID, ProductID int }

	reachable := make(map[int]bool)
	waiting   := make(map[int][]edge)
	seenPath  := make(map[string]bool)
	doneSig   := make(map[string]bool)

	q := list.New()
	for _, b := range BaseElements {
		id := g.NameToID[b]
		reachable[id] = true
		q.PushBack(state{elem: id})
	}

	var out []RecipeStep
	nodes := 0
	const maxDepth = 40

	enqueue := func(a, b, c int, cur [][]int, d int) {
		np := append(append([][]int(nil), cur...), []int{min(a, b), max(a, b), c})
		sig := canonicalHash(np)
		if seenPath[sig] {
			return
		}
		seenPath[sig] = true
		q.PushBack(state{elem: c, path: np, depth: d + 1})
	}

	for q.Len() > 0 && int64(len(out)) < maxPaths {
		st := q.Remove(q.Front()).(state)
		nodes++
		if st.depth > maxDepth {
			continue
		}
		if st.elem == targetID {
			sig := canonicalHash(st.path)
			if !doneSig[sig] {
				doneSig[sig] = true
				out = append(out, buildRecipeStepFromPath(st.path, targetID, g))
			}
			continue
		}

		for _, r := range g.Edges[st.elem] {
			if reachable[r.PartnerID] {
				enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
			} else {
				waiting[r.PartnerID] = append(waiting[r.PartnerID], edge{
					PartnerID: st.elem, ProductID: r.ProductID})
			}
		}
		if lst, ok := waiting[st.elem]; ok {
			for _, r := range lst {
				if reachable[r.PartnerID] {
					enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
				}
			}
			delete(waiting, st.elem)
		}
		reachable[st.elem] = true
	}
	return out, nodes
}

/* -------------------------------------------------------------------------
   Helper & util                                                            */
func canonicalHash(p [][]int) string {
	var steps []string
	for _, t := range p {
		a, b := min(t[0], t[1]), max(t[0], t[1])
		steps = append(steps, fmt.Sprintf("%d-%d-%d", a, b, t[2]))
	}
	return strings.Join(steps, "|")
}

func buildRecipeStepFromPath(path [][]int, targetID int, g IndexedGraph) RecipeStep {
	if len(path) == 0 {
		return RecipeStep{}
	}
	strPath := make([][]string, len(path))
	for i, t := range path {
		strPath[i] = []string{
			g.IDToName[t[0]],
			g.IDToName[t[1]],
			g.IDToName[t[2]],
		}
	}
	last := path[len(path)-1]
	return RecipeStep{
		Combo: IngredientCombo{
			A: g.IDToName[last[0]],
			B: g.IDToName[last[1]],
		},
		Path: strPath,
	}
}

func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }
