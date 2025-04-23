// pages/index.js
import { useEffect, useState } from "react";

export default function Home() {
  const [data, setData] = useState(null);
  useEffect(() => {
    fetch("http://localhost:8080/api/recipes")
      .then((r) => r.json())
      .then(setData)
      .catch(console.error);
  }, []);
  if (!data) return <p>Loading…</p>;

  // build name→svg map
  const nameMap = {};
  Object.values(data)
    .flat()
    .forEach((el) => {
      if (el.local_svg_path) nameMap[el.name] = el.local_svg_path;
    });

  return (
    <div style={{ padding: 20, fontFamily: "Arial,sans-serif" }}>
      <h1 style={{ textAlign: "center" }}>Little Alchemy Elements</h1>
      {Object.entries(data).map(([section, elems]) => (
        <section key={section}>
          <h2
            style={{
              marginTop: "2em",
              borderBottom: "1px solid #666",
              paddingBottom: "0.2em",
            }}>
            {section}
          </h2>
          <table
            style={{
              width: "100%",
              borderCollapse: "separate",
              borderSpacing: "0 1em",
            }}>
            <thead>
              <tr>
                <th style={cell}>Element</th>
                <th style={{ ...cell, whiteSpace: "nowrap" }}>Recipes</th>
              </tr>
            </thead>
            <tbody>
              {elems.map((el) => (
                <tr key={el.name}>
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
                    {el.recipes.map((pair, i) => (
                      <div key={i} style={{ marginBottom: 4 }}>
                        {pair.map((name, j) => (
                          <span key={j}>
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
                            {j === 0 ? " + " : ""}
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
  );
}

const cell = {
  border: "1px solid #ccc",
  padding: "0.5em",
};
