"use client";

import { useEffect, useState, useRef } from "react";
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
  const [searchTerm, setSearchTerm] = useState("");
  const elementRefs = useRef({});

  const handleSearch = (e) => {
    setSearchTerm(e.target.value);

    // Find the first matching element
    if (e.target.value.trim()) {
      const searchLower = e.target.value.toLowerCase();

      // Look through all tiers and elements
      for (const tier of tiersList) {
        for (const element of tier.elements || []) {
          if (element.name.toLowerCase().includes(searchLower)) {
            // We found a match, scroll to it
            const refKey = `element-${element.name}`;
            if (elementRefs.current[refKey]) {
              elementRefs.current[refKey].scrollIntoView({
                behavior: "smooth",
                block: "center",
              });
              return; // Stop after finding first match
            }
          }
        }
      }
    }
  };

  useEffect(() => {
    setLoading(true);
    fetch("/api/recipes")
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
  const tiersList = Array.isArray(tiersData) ? tiersData : Array.isArray(catalog) ? catalog : [];

  // Populate the nameMap
  tiersList.forEach((tier) => {
    const elements = tier.elements || [];
    elements.forEach((el) => {
      if (el.local_svg_path) {
        nameMap[el.name] = el.local_svg_path;
      }
    });
  });

  const getElementTier = (elementName) => {
    // Handle base elements (Air, Earth, Fire, Water)
    if (["Air", "Earth", "Fire", "Water"].includes(elementName)) {
      return "0";
    }

    // Search through all tiers to find this element
    for (let i = 0; i < tiersList.length; i++) {
      const tier = tiersList[i];
      const found = (tier.elements || []).find((el) => el.name === elementName);

      if (found) {
        return tier.Name || tier.name || (i + 1).toString();
      }
    }

    return "?"; // Unknown tier
  };

  return (
    <>
      <Head>
        <title>Little Alchemy 2 Elements Catalog</title>
        <meta name="description" content="Complete catalog of Little Alchemy 2 elements and recipes" />
      </Head>

      <div style={{ padding: 20, fontFamily: "Arial, sans-serif" }}>
        <h1 style={{ textAlign: "center" }}>Little Alchemy 2 Elements</h1>

        {/* Navigation bar - sticky at the top */}
        <div
          style={{
            position: "sticky",
            top: 0,
            backgroundColor: "white",
            padding: "15px",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            zIndex: 100,
            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
            marginBottom: 20,
          }}>
          {/* Nav back to finder */}
          <Link href="/">
            <button
              style={{
                height: "42px",
                padding: "0 16px",
                backgroundColor: "#1a7dc5",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
                fontSize: "16px",
                fontWeight: 500,
              }}>
              Go to Recipe Finder
            </button>
          </Link>

          {/* Search box */}
          <div style={{ flexGrow: 1, marginLeft: 20 }}>
            <input
              type="text"
              placeholder="Search for an element..."
              value={searchTerm}
              onChange={handleSearch}
              style={{
                height: "42px",
                padding: "0 16px",
                margin: "0 16px",
                width: "80%",
                maxWidth: "300px",
                borderRadius: "4px",
                border: "2px solid #1a7dc5",
                fontSize: "16px",
                boxSizing: "border-box",
              }}
            />
          </div>
        </div>

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
                {(tier.elements || []).map((el, elementIndex) => {
                  // Create a unique ref key for this element
                  const refKey = `element-${el.name}`;

                  // Check if this element matches the search
                  const isMatch = searchTerm && el.name.toLowerCase().includes(searchTerm.toLowerCase());

                  // Get the tier name for display
                  const tierName = tier.Name || tier.name || `Tier ${index + 1}`;

                  return (
                    <tr
                      key={el.name || `element-${elementIndex}`}
                      ref={(el) => (elementRefs.current[refKey] = el)}
                      style={isMatch ? { backgroundColor: "#fff3cd" } : {}}>
                      <td style={cell}>
                        <div style={{ display: "flex", alignItems: "center" }}>
                          {el.local_svg_path && (
                            <img
                              src={`/api/svgs/${el.local_svg_path}`}
                              alt={el.name}
                              width={40}
                              height={40}
                              style={{ marginRight: 10 }}
                            />
                          )}
                          <div>
                            <div>{el.name}</div>
                            <div
                              style={{
                                fontSize: "0.8em",
                                color: "#666",
                                backgroundColor: "#f0f0f0",
                                display: "inline-block",
                                padding: "2px 6px",
                                borderRadius: "4px",
                                marginTop: "2px",
                              }}>
                              {tierName}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td style={{ ...cell, whiteSpace: "nowrap" }}>
                        {(el.recipes || []).map((pair, recipeIndex) => (
                          <div key={recipeIndex} style={{ marginBottom: 4 }}>
                            {pair.map((name, nameIndex) => (
                              <span key={nameIndex}>
                                {nameMap[name] && (
                                  <img
                                    src={`/api/svgs/${nameMap[name]}`}
                                    alt={name}
                                    width={24}
                                    height={24}
                                    style={{ marginRight: 4 }}
                                  />
                                )}
                                {name}
                                <span style={{ color: "#888", fontSize: "0.85em" }}> ({getElementTier(name)})</span>
                                {nameIndex === 0 ? " + " : ""}
                              </span>
                            ))}
                          </div>
                        ))}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </section>
        ))}
      </div>
    </>
  );
}
