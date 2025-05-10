"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Head from "next/head";

const cell = {
  border: "1px solid #ccc",
  padding: "0.5em",
};

export default function RecipesPage() {
  const [catalog, setCatalog] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    fetch("http://localhost:8080/api/recipes")
      .then((r) => {
        if (!r.ok) {
          throw new Error(`HTTP error! status: ${r.status}`);
        }
        return r.json();
      })
      .then((data) => {
        console.log("API response:", data); // Debug: log the structure of the response
        setCatalog(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Error fetching recipes:", err);
        setError(err.message);
        setLoading(false);
      });
  }, []);

  // Loading state
  if (loading) {
    return (
      <div style={{ padding: 20, fontFamily: "Arial, sans-serif", textAlign: "center" }}>
        <h1>Loading Elements...</h1>
        <p>Please wait while we fetch the catalog.</p>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div style={{ padding: 20, fontFamily: "Arial, sans-serif", textAlign: "center" }}>
        <h1>Error Loading Elements</h1>
        <p>There was a problem loading the element catalog: {error}</p>
        <Link href="/">
          <button style={{ marginTop: 16 }}>Go to Finder</button>
        </Link>
      </div>
    );
  }

  // No data state
  if (!catalog) {
    return (
      <div style={{ padding: 20, fontFamily: "Arial, sans-serif", textAlign: "center" }}>
        <h1>No Data Available</h1>
        <p>The element catalog is empty or could not be loaded.</p>
        <Link href="/">
          <button style={{ marginTop: 16 }}>Go to Finder</button>
        </Link>
      </div>
    );
  }

  // Build name→svg map - adapt to the actual structure of your data
  const nameMap = {};
  
  // Check which property contains the tiers data
  const tiersData = catalog.tiers || catalog;
  
  // Handle the case where the response is an array directly
  const tiersList = Array.isArray(tiersData) ? tiersData : 
                   Array.isArray(catalog) ? catalog : [];
  
  // Populate the nameMap
  tiersList.forEach((tier) => {
    const elements = tier.elements || [];
    elements.forEach((el) => {
      if (el.local_svg_path) {
        nameMap[el.name] = el.local_svg_path;
      }
    });
  });

  return (
    <>
      <Head>
        <title>Little Alchemy 2 Elements Catalog</title>
        <meta name="description" content="Complete catalog of Little Alchemy 2 elements and recipes" />
      </Head>
      
      <div style={{ padding: 20, fontFamily: "Arial, sans-serif" }}>
        <h1 style={{ textAlign: "center" }}>Little Alchemy 2 Elements</h1>

        {/* Nav back to finder */}
        <Link href="/">
          <button style={{ 
            marginBottom: 16, 
            padding: "8px 16px",
            backgroundColor: "#1a7dc5",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer"
          }}>
            Go to Recipe Finder
          </button>
        </Link>

        {tiersList.map((tier, index) => (
          <section key={tier.Name || tier.name || `tier-${index}`}>
            <h2
              style={{
                marginTop: "2em",
                borderBottom: "1px solid #666",
                paddingBottom: "0.2em",
              }}>
              {tier.Name || tier.name || `Tier ${index + 1}`}
            </h2>

            <table
              style={{
                width: "100%",
                borderCollapse: "separate",
                borderSpacing: "0 1em",
                tableLayout: "fixed",
              }}>
              <thead>
                <tr>
                  <th style={{ ...cell, width: "280px" }}>Element</th>
                  <th style={cell}>Recipes</th>
                </tr>
              </thead>
              <tbody>
                {(tier.elements || []).map((el, elementIndex) => (
                  <tr key={el.name || `element-${elementIndex}`}>
                    <td style={cell}>
                      <div style={{ display: "flex", alignItems: "center" }}>
                        {el.local_svg_path && (
                          <img
                            src={`http://localhost:8080/svgs/${el.local_svg_path}`}
                            alt={el.name}
                            width={40}
                            height={40}
                            style={{ marginRight: 10 }}
                          />
                        )}
                        {el.name}
                      </div>
                    </td>
                    <td style={{ ...cell, whiteSpace: "nowrap" }}>
                      {(el.recipes || []).map((pair, recipeIndex) => (
                        <div key={recipeIndex} style={{ marginBottom: 4 }}>
                          {pair.map((name, nameIndex) => (
                            <span key={nameIndex}>
                              {nameMap[name] && (
                                <img
                                  src={`http://localhost:8080/svgs/${nameMap[name]}`}
                                  alt={name}
                                  width={24}
                                  height={24}
                                  style={{ marginRight: 4 }}
                                />
                              )}
                              {name}
                              {nameIndex === 0 ? " + " : ""}
                            </span>
                          ))}
                        </div>
                      ))}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        ))}
      </div>
    </>
  );
}