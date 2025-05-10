package recipeFinder

import (
	"container/list"
)

/* -------------------------------------------------------------------------
   Hash util
---------------------------------------------------------------------------*/

const fnvOffset = 14695981039346656037 // FNV-1a 64-bit offset basis
const fnvPrime  = 1099511628211

// extendHash – update running FNV-1a hash with one triple.
func extendHash(h uint64, t [3]int) uint64 {
	put := func(v int) {
		for i := 0; i < 4; i++ {
			h ^= uint64(byte(v >> (8 * i)))
			h *= fnvPrime
		}
	}
	put(t[0]); put(t[1]); put(t[2])
	return h
}

/* -------------------------------------------------------------------------
   Node type (digunakan oleh BFS jalur‐ganda)
---------------------------------------------------------------------------*/
type bfsNode struct {
	id       int
	depth    int8
	edge     [3]int  // {a,b,product} yang membawa ke node ini; kosong untuk base
	pathHash uint64  // hash FNV-1a penuh hingga node ini
	prev     *bfsNode
}

/* -------------------------------------------------------------------------
   IndexedBFSBuild – satu jalur tercepat
---------------------------------------------------------------------------*/

func IndexedBFSBuild(targetName string, g IndexedGraph) (ProductToIngredients, int) {
	targetID := g.NameToID[targetName]

	q    := list.New()
	seen := make(map[int]bool)

	for _, base := range BaseElements {
		id := g.NameToID[base]
		q.PushBack(id)
		seen[id] = true
	}

	prev  := make(map[int]struct{ ParentID, PartnerID int })
	nodes := 0

	for q.Len() > 0 {
		cur := q.Remove(q.Front()).(int)
		nodes++
		if cur == targetID {
			break
		}
		for _, e := range g.Edges[cur] { // cur + partner → product
			if seen[e.PartnerID] && !seen[e.ProductID] {
				seen[e.ProductID] = true
				prev[e.ProductID] = struct{ ParentID, PartnerID int }{
					ParentID: cur, PartnerID: e.PartnerID,
				}
				q.PushBack(e.ProductID)
			}
		}
	}

	recipes := make(ProductToIngredients)
	for prodID, p := range prev {
		recipes[g.IDToName[prodID]] = RecipeStep{
			Combo: IngredientCombo{
				A: g.IDToName[p.ParentID],
				B: g.IDToName[p.PartnerID],
			},
		}
	}
	return recipes, nodes
}

/* -------------------------------------------------------------------------
   findKthPathIndexed – jalur ke-(skip+1) unik (BFS)
---------------------------------------------------------------------------*/

func findKthPathIndexed(targetID, skip int, g IndexedGraph) (RecipeStep, int) {
	type edge struct{ PartnerID, ProductID int }

	reachable := make(map[int]bool)  // sudah didequeue
	waiting   := make(map[int][]edge)
	seenHash  := make(map[uint64]struct{}) // dedupe path

	q := list.New()
	for _, b := range BaseElements {
		id := g.NameToID[b]
		q.PushBack(&bfsNode{
			id:       id,
			depth:    0,
			pathHash: fnvOffset,
		})
		reachable[id] = true
	}

	hits, nodes := 0, 0
	const maxDepth = 40

	for q.Len() > 0 {
		cur := q.Remove(q.Front()).(*bfsNode)
		nodes++

		if cur.depth > maxDepth {
			continue
		}
		if cur.id == targetID {
			if _, dup := seenHash[cur.pathHash]; dup {
				continue // jalur identik sudah pernah
			}
			if hits == skip {
				// Rekonstruksi path mundur
				path := make([][]int, 0, cur.depth)
				for n := cur; n.prev != nil; n = n.prev {
					path = append(path, n.edge[:])
				}
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				return buildRecipeStepFromPath(path, targetID, g), nodes
			}
			seenHash[cur.pathHash] = struct{}{}
			hits++
			continue
		}

		// (a) reaksi dengan cur.id sebagai reagent pertama
		for _, e := range g.Edges[cur.id] {
			if reachable[e.PartnerID] {
				t := [3]int{min(cur.id, e.PartnerID), max(cur.id, e.PartnerID), e.ProductID}
				newHash := extendHash(cur.pathHash, t)
				q.PushBack(&bfsNode{
					id:       e.ProductID,
					depth:    cur.depth + 1,
					edge:     t,
					pathHash: newHash,
					prev:     cur,
				})
			} else {
				waiting[e.PartnerID] = append(waiting[e.PartnerID], edge{
					PartnerID: cur.id, ProductID: e.ProductID})
			}
		}

		// (b) reaksi yang menunggu cur.id
		if lst, ok := waiting[cur.id]; ok {
			for _, e := range lst {
				if reachable[e.PartnerID] {
					t := [3]int{min(cur.id, e.PartnerID), max(cur.id, e.PartnerID), e.ProductID}
					newHash := extendHash(cur.pathHash, t)
					q.PushBack(&bfsNode{
						id:       e.ProductID,
						depth:    cur.depth + 1,
						edge:     t,
						pathHash: newHash,
						prev:     cur,
					})
				}
			}
			delete(waiting, cur.id)
		}
		reachable[cur.id] = true
	}
	return RecipeStep{}, nodes
}

/* -------------------------------------------------------------------------
   RangePathsIndexed – ambil banyak jalur (preserve BFS order)
---------------------------------------------------------------------------*/

func RangePathsIndexed(targetID, start, limit int, g IndexedGraph) ([]RecipeStep, int) {
	out   := make([]RecipeStep, 0, limit)
	total := 0
	for k := 0; k < limit; k++ {
		step, visited := findKthPathIndexed(targetID, start+k, g)
		total += visited
		if step.Path == nil || len(step.Path) == 0 {
			break // BFS habis
		}
		out = append(out, step)
	}
	return out, total
}

/* -------------------------------------------------------------------------
   Helpers
---------------------------------------------------------------------------*/

func buildRecipeStepFromPath(path [][]int, targetID int, g IndexedGraph) RecipeStep {
	if len(path) == 0 {
		return RecipeStep{}
	}
	str := make([][]string, len(path))
	for i, t := range path {
		str[i] = []string{
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
		Path: str,
	}
}
func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }
