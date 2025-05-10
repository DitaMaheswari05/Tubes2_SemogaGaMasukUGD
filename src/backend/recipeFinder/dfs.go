package recipeFinder

import (
	"hash/fnv"
	"sort"
)

/* -------------------------------------------------------------------------
   Reverse-index (read-only)                                                */
type pair struct{ a, b int }
type revIndex map[int][]pair

var revIdx revIndex // productID → []pair

func BuildReverseIndex(g IndexedGraph) { // call once at start-up
	idx := make(revIndex)
	for a, nbrs := range g.Edges {
		for _, e := range nbrs {
			p := pair{a: min(a, e.PartnerID), b: max(a, e.PartnerID)}
			idx[e.ProductID] = append(idx[e.ProductID], p)
		}
	}
	for _, list := range idx {
		sort.Slice(list, func(i, j int) bool {
			ti := getElementTier(g.IDToName[list[i].a]) +
				getElementTier(g.IDToName[list[i].b])
			tj := getElementTier(g.IDToName[list[j].a]) +
				getElementTier(g.IDToName[list[j].b])
			return ti < tj
		})
	}
	revIdx = idx
}

/* -------------------------------------------------------------------------
   Single-path DFS (recursive)                                              */
var canReachBaseCache map[int]bool

func findPathToBaseCnt(
	id, depth, maxDepth int,
	g IndexedGraph,
	recipes ProductToIngredients,
	visit map[int]bool,
	counter *int,
) bool {
	if depth > maxDepth {
		return false
	}
	if res, ok := canReachBaseCache[id]; ok {
		return res
	}
	if visit[id] {
		canReachBaseCache[id] = false
		return false
	}
	visit[id] = true
	defer func() { visit[id] = false }()
	*counter++

	name := g.IDToName[id]
	for _, b := range BaseElements {
		if name == b {
			canReachBaseCache[id] = true
			return true
		}
	}

	ing := findIngredientsFor(id, g)
	sort.Slice(ing, func(i, j int) bool {
		ti := getElementTier(g.IDToName[ing[i][0]]) + getElementTier(g.IDToName[ing[i][1]])
		tj := getElementTier(g.IDToName[ing[j][0]]) + getElementTier(g.IDToName[ing[j][1]])
		return ti < tj
	})

	for _, pr := range ing {
		a, b := pr[0], pr[1]
		if findPathToBaseCnt(a, depth+1, maxDepth, g, recipes, visit, counter) &&
			findPathToBaseCnt(b, depth+1, maxDepth, g, recipes, visit, counter) {
			recipes[name] = RecipeStep{
				Combo: IngredientCombo{
					A: g.IDToName[a],
					B: g.IDToName[b],
				},
			}
			canReachBaseCache[id] = true
			return true
		}
	}
	canReachBaseCache[id] = false
	return false
}

func DFSBuildTargetToBase(target string, g IndexedGraph) (ProductToIngredients, int) {
	targetID := g.NameToID[target]
	recipes  := make(ProductToIngredients)
	visited  := make(map[int]bool)

	canReachBaseCache = map[int]bool{}
	for _, b := range BaseElements {
		canReachBaseCache[g.NameToID[b]] = true
	}

	nodes := 0
	if !findPathToBaseCnt(targetID, 0, 1000, g, recipes, visited, &nodes) {
		visited = map[int]bool{}
		findPathToBaseCnt(targetID, 0, 10000, g, recipes, visited, &nodes)
	}
	return recipes, nodes
}

/* -------------------------------------------------------------------------
   Multi-path DFS (iterative)                                               */
func hashPath(p [][]int) uint64 {
	h := fnv.New64a()
	var buf [4]byte
	put := func(v int) {
		buf[0] = byte(v)
		buf[1] = byte(v >> 8)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 24)
		_, _ = h.Write(buf[:])
	}
	for _, t := range p {
		put(t[0]); put(t[1]); put(t[2])
	}
	return h.Sum64()
}

func RangeDFSPaths(target string, maxPaths int, g IndexedGraph) ([]RecipeStep, int) {
	type elem struct{ id, childPos int }

	targetID := g.NameToID[target]
	stack    := []elem{{id: targetID}}
	path     := make([][]int, 0, 64)
	visited  := make(map[int]bool)
	seenSig  := make(map[uint64]struct{})
	out      := make([]RecipeStep, 0, maxPaths)
	nodes    := 0

	for len(stack) > 0 && len(out) < maxPaths {
		top := &stack[len(stack)-1]
		id  := top.id
		nodes++

		// base?
		isBase := false
		for _, b := range BaseElements {
			if id == g.NameToID[b] {
				isBase = true
				break
			}
		}
		if isBase {
			sig := hashPath(path)
			if _, ok := seenSig[sig]; !ok {
				seenSig[sig] = struct{}{}
				out = append(out, buildRecipeStepFromPath(path, targetID, g))
			}
			stack = stack[:len(stack)-1]
			continue
		}

		// cycle guard
		if visited[id] {
			stack = stack[:len(stack)-1]
			continue
		}
		visited[id] = true

		if top.childPos >= len(revIdx[id]) {
			visited[id] = false
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			stack = stack[:len(stack)-1]
			continue
		}

		// explore next child
		p := revIdx[id][top.childPos]
		top.childPos++
		stack = append(stack, elem{id: p.a})
		stack = append(stack, elem{id: p.b})
		path = append(path, []int{p.a, p.b, id})
	}
	return out, nodes
}

/* -------------------------------------------------------------------------
   Tiny helper used by recursion                                            */
func findIngredientsFor(productID int, g IndexedGraph) [][]int {
	var res [][]int
	for aID, nbrs := range g.Edges {
		for _, e := range nbrs {
			if e.ProductID == productID {
				res = append(res, []int{aID, e.PartnerID})
			}
		}
	}
	return res
}
