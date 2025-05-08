// pages/index.js
import { useState } from "react";
import Link from "next/link";

// Recursive component to render the recipe tree
function RecipeTree({ node, level = 0 }) {
  if (!node) return null;

  const padding = level * 20;

  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ paddingLeft: padding, display: "flex", alignItems: "center" }}>
        {level > 0 && <span style={{ marginRight: 8, color: "#666" }}>{level === 1 ? "Made from:" : "└"}</span>}
        <span style={{ fontWeight: level === 0 ? "bold" : "normal" }}>{node.name}</span>
      </div>

      {node.children && node.children.map((child, i) => <RecipeTree key={`${child.name}-${i}`} node={child} level={level + 1} />)}
    </div>
  );
}

export default function FinderPage() {
  const [query, setQuery] = useState("");
  const [response, setResponse] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [multiMode, setMultiMode] = useState(true); // Add this line
  const [maxPaths, setMaxPaths] = useState(5); // Add this line

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    setError(null);
    setResponse(null);

    try {
      // Add multiMode and maxPaths to the query string
      const url = `/api/find?target=${encodeURIComponent(query)}&multi=${multiMode ? "true" : "false"}&maxPaths=${maxPaths}`;
      const res = await fetch(url);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setResponse(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main style={{ padding: 20, fontFamily: "sans-serif" }}>
      <h1>Little Alchemy 2 – Recipe Finder</h1>

      {/* nav button */}
      <Link href="/recipes">
        <button style={{ marginBottom: 16 }}>View All Elements</button>
      </Link>

      <form onSubmit={handleSubmit} style={{ marginBottom: 16 }}>
        <div style={{ marginBottom: 12 }}>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Enter element name…"
            style={{ padding: 8, fontSize: 16, width: 200 }}
          />
          <button type="submit" style={{ marginLeft: 8, padding: "8px 12px", fontSize: 16 }}>
            Find
          </button>
        </div>

        <div style={{ display: "flex", alignItems: "center", marginBottom: 12 }}>
          <div style={{ marginRight: 16 }}>
            <label style={{ display: "flex", alignItems: "center" }}>
              <input type="radio" checked={multiMode} onChange={() => setMultiMode(true)} style={{ marginRight: 6 }} />
              Show multiple recipes
            </label>
          </div>
          <div>
            <label style={{ display: "flex", alignItems: "center" }}>
              <input type="radio" checked={!multiMode} onChange={() => setMultiMode(false)} style={{ marginRight: 6 }} />
              Show single recipe
            </label>
          </div>
        </div>

        {multiMode && (
          <div style={{ marginBottom: 8 }}>
            <label style={{ display: "flex", alignItems: "center" }}>
              Max recipes:
              <input
                type="number"
                min="1"
                max="20"
                value={maxPaths}
                onChange={(e) => setMaxPaths(Math.max(1, Math.min(20, parseInt(e.target.value) || 1)))}
                style={{ width: 60, marginLeft: 8, padding: 4 }}
              />
            </label>
          </div>
        )}
      </form>

      {loading && <p>Loading…</p>}
      {error && <p style={{ color: "crimson" }}>Error: {error}</p>}

      {response && (
        <>
          <div style={{ marginBottom: 24 }}>
            <h2>Recipe for "{Array.isArray(response.tree) ? response.tree[0]?.name : response.tree?.name}"</h2>
            <div style={{ marginBottom: 16 }}>
              <strong>Algorithm:</strong> {response.algorithm}
              <br />
              <strong>Search time:</strong> {response.duration_ms.toFixed(2)} ms
            </div>

            <div
              style={{
                background: "#f0f0f0",
                padding: 16,
                borderRadius: 4,
                marginBottom: 24,
              }}>
              {Array.isArray(response.tree) ? (
                response.tree.map((tree, index) => (
                  <div key={index} style={{ marginBottom: 20 }}>
                    <h3>Recipe {index + 1}</h3>
                    <RecipeTree node={tree} />
                  </div>
                ))
              ) : (
                <RecipeTree node={response.tree} />
              )}
            </div>
          </div>

          <h3>Raw JSON Response</h3>
          <pre
            style={{
              background: "#f8f8f8",
              padding: 16,
              overflowX: "auto",
              maxHeight: "40vh",
              fontSize: "12px",
              borderRadius: 4,
            }}>
            {JSON.stringify(response, null, 2)}
          </pre>
        </>
      )}

      {!loading && !response && !error && <p>Type something above and press "Find" to discover its recipe!</p>}
    </main>
  );
}
