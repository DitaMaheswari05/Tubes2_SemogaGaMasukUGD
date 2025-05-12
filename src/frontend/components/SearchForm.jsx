"use client";

function SearchForm({
  algorithm,
  setAlgorithm,
  multiMode,
  setMultiMode,
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
        <label htmlFor="algorithm">Choose Algorithm:</label>
        <div className="algorithm-options">
          <button
            className={`algorithm-btn ${algorithm === "bfs" ? "active" : ""}`}
            onClick={() => setAlgorithm("bfs")}
            disabled={isLoading}>
            Breadth-First Search (BFS)
          </button>
          <button
            className={`algorithm-btn ${algorithm === "dfs" ? "active" : ""}`}
            onClick={() => setAlgorithm("dfs")}
            disabled={isLoading}>
            Depth-First Search (DFS)
          </button>
          <button
            className={`algorithm-btn ${algorithm === "bidirectional" ? "active" : ""}`}
            onClick={() => setAlgorithm("bidirectional")}
            disabled={isLoading}>
            Bidirectional Search
          </button>
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="searchMode">Recipe Search Mode:</label>
        <div className="toggle-container">
          <button
            className={`toggle-btn ${multiMode === false ? "active" : ""}`}
            onClick={() => setMultiMode(false)}
            disabled={isLoading}>
            Shortest Recipe
          </button>
          <button
            className={`toggle-btn ${multiMode === true ? "active" : ""}`}
            onClick={() => setMultiMode(true)}
            disabled={isLoading}>
            Multiple Recipes
          </button>
        </div>
      </div>

      {multiMode === true && (
        <div className="form-group">
          <label htmlFor="maxRecipes">Maximum Recipes:</label>
          <input
            type="number"
            id="maxRecipes"
            value={maxRecipes}
            onChange={(e) => setMaxRecipes(Math.max(1, Number.parseInt(e.target.value) || 1))}
            min="1"
            max="100"
            disabled={isLoading}
          />
        </div>
      )}

      <div className="form-group">
        <label htmlFor="targetElement">Target Element:</label>
        <div className="input-group">
          <input
            type="text"
            id="targetElement"
            value={targetElement}
            onChange={(e) => setTargetElement(e.target.value)}
            placeholder="Masukkan nama elemen"
            disabled={isLoading}
          />
        </div>
      </div>

      <button className="search-btn" onClick={handleSearch} disabled={isLoading || !targetElement}>
        {isLoading ? "Searching..." : "Find Recipe(s)"}
      </button>
    </div>
  );
}

export default SearchForm;
