"use client"
import "./SearchForm.css"

function SearchForm({
  algorithm,
  setAlgorithm,
  searchMode,
  setSearchMode,
  maxRecipes,
  setMaxRecipes,
  targetElement,
  setTargetElement,
  handleSearch,
  isLoading,
  availableElements,
}) {
  return (
    <div className="search-form">
      <div className="form-group">
        <label htmlFor="algorithm">Pilih Algoritma:</label>
        <div className="algorithm-options">
          <button
            className={`algorithm-btn ${algorithm === "bfs" ? "active" : ""}`}
            onClick={() => setAlgorithm("bfs")}
            disabled={isLoading}
          >
            Breadth-First Search (BFS)
          </button>
          <button
            className={`algorithm-btn ${algorithm === "dfs" ? "active" : ""}`}
            onClick={() => setAlgorithm("dfs")}
            disabled={isLoading}
          >
            Depth-First Search (DFS)
          </button>
          <button
            className={`algorithm-btn ${algorithm === "bidirectional" ? "active" : ""}`}
            onClick={() => setAlgorithm("bidirectional")}
            disabled={isLoading}
          >
            Bidirectional Search
          </button>
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="searchMode">Mode Pencarian Resep:</label>
        <div className="toggle-container">
          <button
            className={`toggle-btn ${searchMode === "shortest" ? "active" : ""}`}
            onClick={() => setSearchMode("shortest")}
            disabled={isLoading}
          >
            Shortest Recipe
          </button>
          <button
            className={`toggle-btn ${searchMode === "multiple" ? "active" : ""}`}
            onClick={() => setSearchMode("multiple")}
            disabled={isLoading}
          >
            Multiple Recipes
          </button>
        </div>
      </div>

      {searchMode === "multiple" && (
        <div className="form-group">
          <label htmlFor="maxRecipes">Maksimum Resep yang Dicari:</label>
          <input
            type="number"
            id="maxRecipes"
            value={maxRecipes}
            onChange={(e) => setMaxRecipes(Math.max(1, Number.parseInt(e.target.value) || 1))}
            min="1"
            max="20"
            disabled={isLoading}
          />
        </div>
      )}

      <div className="form-group">
        <label htmlFor="targetElement">Target Elemen (in english) :</label>
        <select
          id="targetElement"
          value={targetElement}
          onChange={(e) => setTargetElement(e.target.value)}
          disabled={isLoading}
        >
          <option value="">Pilih satu elemen</option>
          {availableElements.map((element) => (
            <option key={element} value={element}>
              {element}
            </option>
          ))}
        </select>
      </div>

      <button className="search-btn" onClick={handleSearch} disabled={isLoading || !targetElement}>
        {isLoading ? "Searching..." : "Find Recipes"}
      </button>
    </div>
  )
}

export default SearchForm
