package recipeFinder

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
)

func IndexedBFSBuild(targetName string, graph IndexedGraph) map[string]Info {
    targetID := graph.NameToID[targetName]
    
    queue := list.New()
    seen := make(map[int]bool)
    
    for _, baseName := range baseElements {
        baseID := graph.NameToID[baseName]
        queue.PushBack(baseID)
        seen[baseID] = true
    }
    
    // Track parents using integer IDs
    prevIDs := make(map[int]struct{ParentID, PartnerID int})
    
    for queue.Len() > 0 {
        curID := queue.Remove(queue.Front()).(int)
        
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
    
    // Convert integer results back to strings
    prev := make(map[string]Info)
    for productID, info := range prevIDs {
        productName := graph.IDToName[productID]
        parentName := graph.IDToName[info.ParentID]
        partnerName := graph.IDToName[info.PartnerID]
        
        prev[productName] = Info{
            Parent:  parentName,
            Partner: partnerName,
        }
    }
    
    return prev
}

// findKthPathIndexed returns the (skip+1)-th shortest recipe path to targetID.
// It keeps every distinct *path* (sequence of reactions), so the same
// reagent pair can appear again if the preceding chain is different.
func findKthPathIndexed(targetID, skip int, g IndexedGraph) Info {
    type state struct {
        elem  int
        path  [][]int        // [a,b,product] triples
        depth int
    }
    type edge struct{ PartnerID, ProductID int }

    reachable := make(map[int]bool)    // elements already dequeued
    waiting   := make(map[int][]edge)  // partnerID -> reactions awaiting it
    seenPath  := make(map[string]bool) // full‑recipe signature -> dedupe

    q := list.New()

    // seed BFS with base elements
    for _, b := range baseElements {
        id := g.NameToID[b]
        reachable[id] = true
        q.PushBack(state{elem: id, depth: 0})
    }

    // helper: push a reaction when both reagents are reachable
    enqueue := func(a, b, product int, curPath [][]int, curDepth int) {
        newTriple := []int{min(a, b), max(a, b), product}
        newPath   := append(append([][]int{}, curPath...), newTriple)

        // build a unique signature for the *entire* path
        var sb strings.Builder
        for _, t := range newPath {
            fmt.Fprintf(&sb, "%d-%d-%d|", t[0], t[1], t[2])
        }
        sig := sb.String()
        if seenPath[sig] { // already generated this exact chain
            return
        }
        seenPath[sig] = true

        q.PushBack(state{elem: product, path: newPath, depth: curDepth + 1})
    }

    hits := 0
    const maxDepth = 10000 // raise if you ever need deeper combos

    for q.Len() > 0 {
        st := q.Remove(q.Front()).(state)

        if st.depth > maxDepth {
            continue
        }
        if st.elem == targetID {
            if hits == skip {
                return buildInfoFromPath(st.path, targetID, g)
            }
            hits++
            continue
        }

        // (a) reactions that use st.elem as first reagent
        for _, r := range g.Edges[st.elem] { // st.elem + partner → product
            if reachable[r.PartnerID] {
                enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
            } else {
                waiting[r.PartnerID] = append(waiting[r.PartnerID], edge{
                    PartnerID: st.elem, ProductID: r.ProductID})
            }
        }

        // (b) reactions waiting *for* st.elem
        if list, ok := waiting[st.elem]; ok {
            for _, r := range list {
                if reachable[r.PartnerID] {
                    enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
                }
            }
            delete(waiting, st.elem)
        }

        reachable[st.elem] = true
    }

    return Info{} // no further paths
}

// RangePathsIndexed returns up to `limit` distinct paths to `targetID`
// starting from global rank `start` (0-based). It uses findKthPathIndexed
// repeatedly, so it preserves breadth-first order.
func RangePathsIndexed(targetID int, start, limit int, g IndexedGraph) []Info {
    out := make([]Info, 0, limit)
    for k := 0; k < limit; k++ {
        info := findKthPathIndexed(targetID, start+k, g)
        if info.Path == nil {
            break // BFS exhausted
        }
        out = append(out, info)
    }
    return out
}


// findDistinctPathsIndexed returns up to maxPaths shortest *unique* paths
// to targetID in a single BFS sweep.
func findDistinctPathsIndexed(targetID int, maxPaths int64, g IndexedGraph) []Info {
    type state struct {
        elem  int
        path  [][]int   // sequence of [a,b,product]
        depth int
    }
    type edge struct{ PartnerID, ProductID int }

    reachable := map[int]bool{}
    waiting   := map[int][]edge{}
    seenPath  := map[string]bool{}   // dedupe while exploring
    doneSig   := map[string]bool{}   // dedupe final hits

    q := list.New()
    for _, b := range baseElements {
        id := g.NameToID[b]
        reachable[id] = true
        q.PushBack(state{elem: id, depth: 0})
    }

    // makeSig := func(p [][]int) string {
	// 	uniq := map[[3]int]struct{}{}
	// 	for _, t := range p {
	// 		// canonical triple
	// 		tr := [3]int{min(t[0], t[1]), max(t[0], t[1]), t[2]}
	// 		uniq[tr] = struct{}{}
	// 	}
	
	// 	list := make([]string, 0, len(uniq))
	// 	for tr := range uniq {
	// 		list = append(list,
	// 			fmt.Sprintf("%d-%d-%d", tr[0], tr[1], tr[2]))
	// 	}
	// 	sort.Strings(list)
	// 	return strings.Join(list, "|")      // set‑based signature
	// }
	

    enqueue := func(a, b, c int, cur [][]int, d int) {
        np := append(append([][]int{}, cur...), []int{min(a, b), max(a, b), c})
        sig := canonicalHash(np)
        if seenPath[sig] {
            return
        }
        seenPath[sig] = true
        q.PushBack(state{elem: c, path: np, depth: d + 1})
    }

    const maxDepth = 40
    out := make([]Info, 0, maxPaths)

    for q.Len() > 0 && int64(len(out)) < maxPaths {
        st := q.Remove(q.Front()).(state)
        if st.depth > maxDepth {
            continue
        }

        if st.elem == targetID {
            sig := canonicalHash(st.path)
            if doneSig[sig] {
                continue
            }
            doneSig[sig] = true
            out = append(out, buildInfoFromPath(st.path, targetID, g))
            continue
        }

        // (a) reactions that use st.elem as first reagent
        for _, r := range g.Edges[st.elem] {
            if reachable[r.PartnerID] {
                enqueue(st.elem, r.PartnerID, r.ProductID, st.path, st.depth)
            } else {
                waiting[r.PartnerID] = append(waiting[r.PartnerID], edge{
                    PartnerID: st.elem, ProductID: r.ProductID})
            }
        }

        // (b) reactions waiting for st.elem
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
    return out
}

// canonicalHash builds a key from the UNORDERED, DUPLICATE‑FREE set of
// reaction triples contained in `p`.
func canonicalHash(p [][]int) string {
    counts := map[[3]int]int{}
    for _, t := range p {
        tr := [3]int{min(t[0], t[1]), max(t[0], t[1]), t[2]}
        counts[tr]++
    }
    list := make([]string, 0, len(counts))
    for tr, c := range counts {
        list = append(list,
            fmt.Sprintf("%d-%d-%d:%d", tr[0], tr[1], tr[2], c))
    }
    sort.Strings(list)
    return strings.Join(list, "|")    // multiset signature
}




  
// In your multi recipe loop:
func IndexedBFSBuildMulti(target string, g IndexedGraph, max int64) map[string][]Info {
    targetID := g.NameToID[target]
    paths := findDistinctPathsIndexed(targetID, max, g)
    return map[string][]Info{target: paths}
}


// Helper function that wasn't shown in your code
func buildInfoFromPath(path [][]int, targetID int, graph IndexedGraph) Info {
    if path == nil {
        return Info{}
    }
    
    stringPath := make([][]string, len(path))
    for i, step := range path {
        stringPath[i] = []string{
            graph.IDToName[step[0]],
            graph.IDToName[step[1]],
            graph.IDToName[step[2]],
        }
    }
    
    info := Info{Path: stringPath}
    if len(path) > 0 {
        // Get parent/partner from the last step
        lastStep := path[len(path)-1]
        info.Parent = graph.IDToName[lastStep[0]]
        info.Partner = graph.IDToName[lastStep[1]]
    }
    
    return info
}

// Simple min/max helpers
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
