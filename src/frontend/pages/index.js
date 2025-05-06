// pages/index.js
import { useState } from "react";
import Link from "next/link";

export default function FinderPage() {
  const [query, setQuery] = useState("");
  const [json, setJson] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    setError(null);
    setJson(null);

    try {
      const res = await fetch(`/api/find?target=${encodeURIComponent(query)}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setJson(data);
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
      </form>

      {loading && <p>Loading…</p>}
      {error && <p style={{ color: "crimson" }}>Error: {error}</p>}

      {json && (
        <>
          <h2>Raw JSON Response</h2>
          <pre
            style={{
              background: "#f0f0f0",
              padding: 16,
              overflowX: "auto",
              maxHeight: "60vh",
            }}>
            {JSON.stringify(json, null, 2)}
          </pre>
        </>
      )}

      {!loading && !json && !error && <p>Type something above and press “Find” to see the raw recipe JSON.</p>}
    </main>
  );
}
