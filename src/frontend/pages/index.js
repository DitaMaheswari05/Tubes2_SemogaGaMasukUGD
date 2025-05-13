"use client";

import { useState, useEffect } from "react";
import Head from "next/head";
import Link from "next/link";
import SearchForm from "../components/SearchForm";
import RecipeTreeText from "../components/RecipeTreeText";
import RecipeTree from "../components/RecipeTree";
import LiveSearchVisualizer from "../components/LiveSearchVisualizer";
import RecipeAtlas from "../components/RecipeAtlas";
import RecipeAtlasUnified from "../components/RecipeAtlasUnified";

// tambahan function buat convert tree
const convertTree = (node) => {
  if (!node) return null;

  const newNode = {
    element: node.name,
  };

  if (node.children && node.children.length === 2) {
    newNode.ingredients = [convertTree(node.children[0]), convertTree(node.children[1])];
  }

  return newNode;
};

export default function Index() {
  const [algorithm, setAlgorithm] = useState("bfs");
  const [multiMode, setMultiMode] = useState(false);
  const [maxRecipes, setMaxRecipes] = useState(5);
  const [targetElement, setTargetElement] = useState("");
  const [submittedTarget, setSubmittedTarget] = useState("");
  const [results, setResults] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [searchStats, setSearchStats] = useState({ time: 0, nodesVisited: 0 });
  const [availableElements, setAvailableElements] = useState([]);
  const [error, setError] = useState(null);
  const [response, setResponse] = useState(null);
  const [viewMode, setViewMode] = useState("result");
  const [showAtlas, setShowAtlas] = useState(false);

  //   useEffect(() => {
  //     fetch("/api/elements")
  //       .then((res) => res.json())
  //       .then((data) => {
  //         setAvailableElements(data.elements || []);
  //       })
  //       .catch((err) => {
  //         console.error("Failed to load elements:", err);
  //         setAvailableElements([]);
  //       });
  //   }, []);

  const handleSearch = async () => {
    if (!targetElement) return;

    setIsLoading(true);
    setError(null);
    setResponse(null);
    setResults(null);
    setSubmittedTarget(targetElement);

    try {
      const url = `/api/find?target=${encodeURIComponent(targetElement)}&multi=${
        multiMode ? "true" : "false"
      }&maxPaths=${maxRecipes}&algorithm=${algorithm}`;
      const res = await fetch(url);

      if (!res.ok) throw new Error(`HTTP ${res.status}`);

      const data = await res.json();
      setResponse(data);

      setSearchStats({
        time: data.duration_ms,
        nodesVisited: data.nodes_visited,
      });

      if (Array.isArray(data.tree)) {
        setResults(data.tree.slice(0, maxRecipes)); // agar jumlah resep sesuai dengan input
      } else if (data.tree) {
        setResults([data.tree].slice(0, maxRecipes));
      } else {
        setResults([]);
      }

      setAlgorithm(data.algorithm || "bfs");
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleScrape = async () => {
    if (!confirm("This will scrape all recipes from the wiki. Continue?")) {
      return;
    }

    try {
      setIsLoading(true);
      const response = await fetch("/api/scrape", {
        method: "POST",
      });

      if (!response.ok) {
        throw new Error(`HTTP error ${response.status}`);
      }

      const data = await response.json();
      alert(`Scraping completed! ${data.message}`);

      // Refresh available elements to get the new data
      //   fetch("/api/elements")
      //     .then((res) => res.json())
      //     .then((data) => {
      //       setAvailableElements(data.elements || []);
      //     });
    } catch (error) {
      alert("Error scraping data: " + error.message);
      console.error("Scrape error:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <Head>
        <title>Little Alchemy 2 Recipe Finder</title>
        <meta name="description" content="Find recipes for Little Alchemy 2 elements" />
      </Head>

      <div className="app">
        <header className="app-header">
          <h1>Little Alchemy 2 Recipe Finder</h1>
          <div className="nav-links">
            <Link href="/recipes" className="nav-link">
              <img src="/icons/catalog.png" alt="Catalog" className="nav-icon" width="24" height="24" />
            </Link>
          </div>
          <div className="scrape-button">
            <button onClick={handleScrape} className="scrape-button-inner" disabled={isLoading}>
              {isLoading ? "Scraping..." : "Scrape All"}
            </button>
          </div>
        </header>

        <main className="app-content">
          <SearchForm
            algorithm={algorithm}
            setAlgorithm={setAlgorithm}
            multiMode={multiMode}
            setMultiMode={setMultiMode}
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
              <p>Find recipe(s).....</p>
            </div>
          )}

          {error && (
            <div className="error">
              <p>Error: {error}</p>
            </div>
          )}

          {results && (
            <div className="results-container">
              <div className="search-stats">
                <p>
                  Search Time : <strong>{searchStats.time} ms</strong>
                </p>
                <p>
                  Visited Node: <strong>{searchStats.nodesVisited}</strong>
                </p>
              </div>

              {/* Add view mode toggle buttons */}
              {response && response.search_steps && (
                <div className="view-toggle" style={{ marginBottom: 20, display: "flex", justifyContent: "center" }}>
                  <button
                    onClick={() => setViewMode("result")}
                    style={{
                      backgroundColor: viewMode === "result" ? "#1a7dc5" : "#e0e0e0",
                      color: viewMode === "result" ? "white" : "black",
                      padding: "8px 16px",
                      border: "none",
                      borderRadius: "4px 0 0 0",
                      cursor: "pointer",
                    }}>
                    Final Results
                  </button>
                  <button
                    onClick={() => setViewMode("process")}
                    style={{
                      backgroundColor: viewMode === "process" ? "#1a7dc5" : "#e0e0e0",
                      color: viewMode === "process" ? "white" : "black",
                      padding: "8px 16px",
                      border: "none",
                      cursor: "pointer",
                    }}>
                    Search Process
                  </button>
                  <button
                    onClick={() => setViewMode("atlas")}
                    style={{
                      backgroundColor: viewMode === "atlas" ? "#1a7dc5" : "#e0e0e0",
                      color: viewMode === "atlas" ? "white" : "black",
                      padding: "8px 16px",
                      border: "none",
                      cursor: "pointer",
                    }}>
                    Atlas
                  </button>
                  <button
                    onClick={() => setViewMode("unified")}
                    style={{
                      backgroundColor: viewMode === "unified" ? "#1a7dc5" : "#e0e0e0",
                      color: viewMode === "unified" ? "white" : "black",
                      padding: "8px 16px",
                      border: "none",
                      borderRadius: "0 0 0 4px",
                      cursor: "pointer",
                    }}>
                    Unified
                  </button>
                </div>
              )}

              {/* choose view */}
              {viewMode === "unified" ? (
                <RecipeAtlasUnified recipes={results} elementName={submittedTarget} />
              ) : viewMode === "atlas" ? (
                <RecipeAtlas recipes={results} elementName={submittedTarget} />
              ) : viewMode === "process" ? (
                <LiveSearchVisualizer searchSteps={response.search_steps} targetElement={submittedTarget} />
              ) : results.length > 0 ? (
                <div className="recipe-trees">
                  <h2>
                    Found {results.length} recipe{results.length !== 1 ? "s" : ""} for {submittedTarget}
                  </h2>
                  {results.map((tree, i) => (
                    <RecipeTree key={i} path={convertTree(tree)} index={i} />
                  ))}
                </div>
              ) : (
                <div className="no-results">
                  <p>No recipes found for {submittedTarget}</p>
                </div>
              )}
            </div>
          )}
        </main>
      </div>
    </>
  );
}
