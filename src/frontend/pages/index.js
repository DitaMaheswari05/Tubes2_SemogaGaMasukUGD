"use client";

import { useState, useEffect } from "react";
import Head from "next/head";
import Link from "next/link";
import SearchForm from "../components/SearchForm";
import RecipeTreeText from "../components/RecipeTreeText";
import RecipeTree from "../components/RecipeTree";


// tambahan function buat convert tree
const convertTree = (node) => {
  if (!node) return null;

  const newNode = {
    element: node.name,
  };

  if (node.children && node.children.length === 2) {
    newNode.ingredients = [
      convertTree(node.children[0]),
      convertTree(node.children[1]),
    ];
  }

  return newNode;
};


export default function Index() {
  const [algorithm, setAlgorithm] = useState("bfs");
  const [multiMode, setMultiMode] = useState(false);
  const [maxRecipes, setMaxRecipes] = useState(5);
  const [targetElement, setTargetElement] = useState("");
  const [results, setResults] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [searchStats, setSearchStats] = useState({ time: 0, nodesVisited: 0 });
  const [availableElements, setAvailableElements] = useState([]);
  const [error, setError] = useState(null);
  const [response, setResponse] = useState(null);

  useEffect(() => {
    // Load available elements
    fetch("/api/elements")
      .then((res) => res.json())
      .then((data) => {
        setAvailableElements(data.elements || []);
      })
      .catch((err) => {
        console.error("Failed to load elements:", err);
        setAvailableElements([]);
      });
  }, []);


  const handleSearch = async () => {
    if (!targetElement) return;

    setIsLoading(true);
    setError(null);
    setResponse(null);
    setResults(null);

    try {
      // Use Next.js API route
      const url = `/api/find?target=${encodeURIComponent(targetElement)}&multi=${
        multiMode ? "true" : "false"
      }&maxPaths=${maxRecipes}&algorithm=${algorithm}`;
      const res = await fetch(url);

      if (!res.ok) throw new Error(`HTTP ${res.status}`);

      const data = await res.json();
      setResponse(data);

      // Update stats
      setSearchStats({
        time: data.duration_ms,
        nodesVisited: data.nodes_visited,
      });

      // Format results for display
      if (Array.isArray(data.tree)) {
        setResults(data.tree);
      } else if (data.tree) {
        setResults([data.tree]); // Wrap single tree in array
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

  return (
    <>
      <Head>
        <title>Little Alchemy 2 Recipe Finder</title>
        <meta name="description" content="Find recipes for Little Alchemy 2 elements" />
      </Head>

      <div className="app">
        <header className="app-header">
          <h1>Little Alchemy 2 Recipe Finder</h1>

          {/* Next.js Link component */}
          <div className="nav-links">
            <Link href="/recipes" className="nav-link">
              <img src="/icons/catalog.png" alt="Catalog" className="nav-icon" width="24" height="24" />
            </Link>
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
              <p>Mencari resep.....</p>
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
                  Waktu pencarian: <strong>{searchStats.time} ms</strong>
                </p>
                <p>
                  Node yang dikunjungi: <strong>{searchStats.nodesVisited}</strong>
                </p>
              </div>

              {results.length > 0 ? (
                <div className="recipe-trees">
                  <h2>
                    Found {results.length} recipe{results.length !== 1 ? "s" : ""} for {targetElement}
                  </h2>
                  {results.map((tree, index) => (
                    <div key={index} className="recipe-tree">
                      {/* <h3>Recipe {index + 1}</h3> */}
                      <div style={{ display: "flex", flexDirection: "row", gap: "0px", alignItems: "flex-start" }}>
                        {/* <RecipeTreeText node={tree} /> */}
                        <RecipeTree path={convertTree(tree)} index={index} />
                      </div>
                    </div>
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
    </>
  );
}