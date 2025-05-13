import React, { useEffect, useRef } from "react";
import {drawEntireRecipe} from "./RecipeAtlas"; // adjust path as needed

function mergeTrees(treeList) {
  const mergeNode = (a, b) => {
    if (a.name !== b.name) return null;

    const merged = { name: a.name, children: [] };

    const allChildren = [...(a.children || []), ...(b.children || [])];

    for (const child of allChildren) {
      const existing = merged.children.find((c) => c.name === child.name);
      if (existing) {
        const mergedChild = mergeNode(existing, child);
        if (mergedChild) {
          merged.children = merged.children.map((c) => (c.name === child.name ? mergedChild : c));
        }
      } else {
        merged.children.push(child);
      }
    }

    return merged;
  };

  let unifiedTree = treeList[0];
  for (let i = 1; i < treeList.length; i++) {
    unifiedTree = mergeNode(unifiedTree, treeList[i]);
  }

  return unifiedTree;
}

function RecipeAtlasUnified({ recipes, elementName }) {
  const canvasRef = useRef(null);
  const scale = 1.0;
  const position = { x: window.innerWidth / 2, y: 100 }; // tweak as needed

  useEffect(() => {
    if (!recipes || recipes.length === 0) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");

    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    ctx.translate(position.x, position.y);
    ctx.scale(scale, scale);

    ctx.fillStyle = "#faf9f4";
    ctx.fillRect(-10000, -10000, 20000, 20000);

    const unifiedTree = mergeTrees(recipes);

    drawEntireRecipe(ctx, unifiedTree, 0, 100, 0);
  }, [recipes]);

  return (
    <div>
      <h2 style={{ textAlign: "center" }}>
        Unified Recipe Tree for <strong>{elementName}</strong>
      </h2>
      <canvas ref={canvasRef} style={{ border: "1px solid #ccc", width: "100%", height: "100%" }} />
    </div>
  );
}

export default RecipeAtlasUnified;
