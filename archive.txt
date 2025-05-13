package search

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"bfs/graph"
)

// Menyimpan tiap kombinasi resep
type levelTask struct {
	Product string
	Recipe  []string
	Left	*graph.TreeNode
	Right	*graph.TreeNode
}

func MultiBFS(target string, g graph.Graph, maxRecipes int, tierMap map[string]int) (graph.MultiTreeResult, error) {
	start := time.Now()

	startingElements := []string{"Air", "Fire", "Water", "Earth"}

	existing := map[string][]*graph.TreeNode{} // Menyimpan node yang sudah dibuat
	var visitedCombo sync.Map // Nanti digunain untuk cek duplikat
	visitedElem := map[string]bool{}

	// Inisialisasi node dasar
	for _, e := range startingElements {
		node := &graph.TreeNode{Name: e, Children: []*graph.TreeNode{}}
		existing[e] = []*graph.TreeNode{node}
		visitedElem[e] = true
	}

	// kalau target elemen dasar, lgsg return aja
	if contains(startingElements, target) {
		n := &graph.TreeNode{
			Name: target, 
			NodeDiscovered: 0, 
			Children: []*graph.TreeNode{},
		}
		return graph.MultiTreeResult{
			Trees:        []*graph.TreeNode{n},
			Algorithm:    "Multi_BFS",
			DurationMS:   float64(time.Since(start).Microseconds()) / 1000.0,
			VisitedNodes: 1,
		}, nil
	}

	foundRecipes := []*graph.TreeNode{}
	tier := 0
	numWorkers := runtime.NumCPU() * 2
	targetTier := tierMap[target]
	maxCombinationPerPair := 100 // Batas kombinasi per pasang node
	var recipeMutex sync.Mutex

	queue := make([]levelTask, 0)

	// Inisialsiasi kombinasi dari awal
	for product, recipes := range g {
		for _, recipe := range recipes {
			if len(recipe) != 2 {
				continue
			}
			if _, ok1 := existing[recipe[0]]; !ok1 {
				continue
			}
			if _, ok2 := existing[recipe[1]]; !ok2 {
				continue
			}
			for _, l := range existing[recipe[0]] {
				for _, r := range existing[recipe[1]] {
					queue = append(queue, levelTask{
						Product: product,
						Recipe:  recipe,
						Left:    l,
						Right:   r,
					})
				}
			}
		}
	}

	for len(queue) > 0 && (maxRecipes <= 0 || len(foundRecipes) < maxRecipes) {
		fmt.Printf("\nTier %d: Memproses %d kombinasi\n", tier, len(queue)) // debugging

		tasks := make(chan levelTask, len(queue))
		results := make(chan *graph.TreeNode, len(queue))
		var wg sync.WaitGroup

		// Worker BFS
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range tasks {
					product := task.Product
					recipe := task.Recipe
					productTier := tierMap[product]

					// Elemen pembentuk harus lebih rendah tiernya dari elemen yang dibetuk
					if tierMap[recipe[0]] >= productTier || tierMap[recipe[1]] >= productTier {
						continue
					}
					
					// Key dibuat untuk mendeteksi kombinasi duplikat
					key := fmt.Sprintf("%s+%s>%s#%p|%p", recipe[0], recipe[1], product, task.Left, task.Right)
					if _, exists := visitedCombo.LoadOrStore(key, true); exists {
						continue
					}

					n := &graph.TreeNode{Name: product, Children: []*graph.TreeNode{task.Left, task.Right}}
					results <- n
				}
			}()
		}

		for _, t := range queue {
			tasks <- t
		}
		close(tasks)
		wg.Wait()
		close(results)

		nextQueue := []levelTask{}
		newNodes := map[string][]*graph.TreeNode{}

		for res := range results {
			newNodes[res.Name] = append(newNodes[res.Name], res)
			visitedElem[res.Name] = true
			if res.Name == target {
				clone := deepCopyTree(res)
				recipeMutex.Lock()
				if maxRecipes <= 0 || len(foundRecipes) < maxRecipes {
					foundRecipes = append(foundRecipes, clone)
				}
				recipeMutex.Unlock()
			}
		}

		// Menambahkan node yang ditemukan ke existing
		for product, list := range newNodes {
			existing[product] = append(existing[product], list...)
		}

		// Menyiapkan kombinasi untuk tier berikutnya
		for product, recipes := range g {
			if tierMap[product] != tier+1 {
				continue
			}
			for _, recipe := range recipes {
				if len(recipe) != 2 {
					continue
				}
				lefts, ok1 := existing[recipe[0]]
				rights, ok2 := existing[recipe[1]]
				if !ok1 || !ok2 {
					continue
				}
				count := 0
				for _, l := range lefts {
					for _, r := range rights {
						if count >= maxCombinationPerPair {
							break
						}
						nextQueue = append(nextQueue, levelTask{Product: product, Recipe: recipe, Left: l, Right: r})
						count++
					}
					if count >= maxCombinationPerPair {
						break
					}
				}
			}
		}
		queue = nextQueue
		tier++

		// Cek tier biar ga semua tier di kombinasiin (tewas goroutinenya)
		if tier > targetTier {
			break
		}
	}

	// Ambil rute pertama sebagai primary rute sebagai live update
	for i, t := range foundRecipes {
		if i == 0 {
			idx := 0
			setDiscoveredIndexMultiple(t, &idx)
		} else {
			markTreeAsAlternative(t)
		}
	}

	// Cek apakah resep sudah cukup atau belum
	if maxRecipes > 0 && len(foundRecipes) > maxRecipes {
		foundRecipes = foundRecipes[:maxRecipes]
	}

	return graph.MultiTreeResult{
		Trees:        foundRecipes,
		Algorithm:    "Multi_BFS",
		DurationMS:   float64(time.Since(start).Microseconds()) / 1000.0,
		VisitedNodes: len(visitedElem),
	}, nil
}

// Membuat salinan dari pohon node
func deepCopyTree(node *graph.TreeNode) *graph.TreeNode {
	if node == nil {
		return nil
	}

	copy := &graph.TreeNode{
		Name: node.Name,
		NodeDiscovered: node.NodeDiscovered,
		Children: []*graph.TreeNode{},
	}
	for _, child := range node.Children {
		copy.Children = append(copy.Children, deepCopyTree(child))
	}
	return copy
}

func setDiscoveredIndexMultiple(node *graph.TreeNode, counter *int) {
if node == nil {
		return
	}

	queue := []*graph.TreeNode{node}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		curr.NodeDiscovered = *counter
		*counter++

		queue = append(queue, curr.Children...)
	}
}

// yang alternative nilai node discoverednya adalah -1
func markTreeAsAlternative(node *graph.TreeNode) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		markTreeAsAlternative(child)
	}
	node.NodeDiscovered = -1
}

// Contains untuk memeriksa apakah elemen ada di dalam list rating atau tidak
func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}