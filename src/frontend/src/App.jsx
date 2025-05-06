"use client"

import { useState, useEffect } from "react"
import "./App.css"
import SearchForm from "./components/SearchForm"
import RecipeTree from "./components/RecipeTree"
import { searchRecipes } from "./utils/recipeSearch"
import { recipes, baseElements } from "./data/recipes"

function App() {
  const [algorithm, setAlgorithm] = useState("bfs")
  const [searchMode, setSearchMode] = useState("shortest")
  const [maxRecipes, setMaxRecipes] = useState(5)
  const [targetElement, setTargetElement] = useState("")
  const [results, setResults] = useState(null)
  const [isLoading, setIsLoading] = useState(false)
  const [searchStats, setSearchStats] = useState({ time: 0, nodesVisited: 0 })
  const [availableElements, setAvailableElements] = useState([])

  useEffect(() => {
    // ekstrak semua elemen unik dari resep
    const elements = new Set()

    // tambah starting element
    baseElements.forEach((element) => elements.add(element))

    // tambah semua elemen dari recipes
    Object.keys(recipes).forEach((result) => {
      elements.add(result)
      recipes[result].forEach((combo) => {
        elements.add(combo[0])
        elements.add(combo[1])
      })
    })

    setAvailableElements(Array.from(elements).sort())
  }, [])

  const handleSearch = () => {
    if (!targetElement) return

    setIsLoading(true)
    setResults(null)

    const startTime = performance.now()

    // buat hitung waktu pencarian
    setTimeout(() => {
      const { paths, nodesVisited } = searchRecipes({
        targetElement,
        algorithm,
        searchMode,
        maxRecipes: Number.parseInt(maxRecipes),
        recipes,
        baseElements,
      })

      const endTime = performance.now()
      const searchTime = (endTime - startTime).toFixed(2) // ini yang akan ditampilin

      setResults(paths)
      setSearchStats({
        time: searchTime,
        nodesVisited,
      })
      setIsLoading(false)
    }, 100)
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Little Alchemy 2 Recipe Finder</h1>
      </header>

      <main className="app-content">
        <SearchForm
          algorithm={algorithm}
          setAlgorithm={setAlgorithm}
          searchMode={searchMode}
          setSearchMode={setSearchMode}
          maxRecipes={maxRecipes}
          setMaxRecipes={setMaxRecipes}
          targetElement={targetElement}
          setTargetElement={setTargetElement}
          handleSearch={handleSearch}
          isLoading={isLoading}
          availableElements={availableElements}
        />

        {isLoading && (
          <div className="loading">
            <p>Mencari resep.....</p>
          </div>
        )}

        {results && (
          <div className="results-container">
            <div className="search-stats">
              <p>
                Waktu pencarian : <strong>{searchStats.time} ms</strong>
              </p>
              <p>
                Node yang dikunjungi : <strong>{searchStats.nodesVisited}</strong>
              </p>
            </div>

            {results.length > 0 ? (
              <div className="recipe-trees">
                <h2>
                  Found {results.length} recipe{results.length !== 1 ? "s" : ""} for {targetElement}
                </h2>
                {results.map((path, index) => (
                  <RecipeTree key={index} path={path} index={index} />
                ))}
              </div>
            ) : (
              <div className="no-results">
                <p>No recipes found for {targetElement}</p>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  )
}

export default App