"use client";

import { useRef, useEffect, useState } from "react";
import Link from "next/link";

export function drawEntireRecipe(ctx, recipe, startX, startY, index) {
  ctx.save();
  ctx.translate(startX, startY);
  ctx.font = "bold 24px Arial";
  ctx.fillStyle = "#333";
  ctx.textAlign = "center";
  ctx.fillText(`Recipe ${index + 1}`, 0, -20);

  let nextX = 0;

  const drawRecipeTree = (node, depth, y) => {
    if (!node) return 0;
    
    const xSpacing = 120;
    let x;

    if (!node.children || node.children.length === 0) {
      x = nextX * xSpacing;
      nextX++;
    } else {
      const leftX = drawRecipeTree(node.children[0], depth + 1, y + 100);
      const rightX = drawRecipeTree(node.children[1], depth + 1, y + 100);
      x = (leftX + rightX) / 2;

      ctx.beginPath();
      ctx.moveTo(x, y + 40);
      ctx.lineTo(leftX, y + 100);
      ctx.stroke();

      ctx.beginPath();
      ctx.moveTo(x, y + 40);
      ctx.lineTo(rightX, y + 100);
      ctx.stroke();

      ctx.font = "20px Arial";
      ctx.fillStyle = "#000";
      ctx.textAlign = "center";
      ctx.fillText("+", x, y + 60);
    }

    drawElementBox(node.name, x, y);
    return x;
  };

  const drawElementBox = (name, x, y) => {
    const boxWidth = 105,
      boxHeight = 40;
    const palette = [
      "#a8d5a8",
      "#a8c5d5",
      "#d5a8a8",
      "#808080",
      "#f4c2c2",
      "#f0e68c",
      "#dda0dd",
      "#add8e6",
      "#90ee90",
      "#ffcccb",
    ];
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    let color = palette[Math.abs(hash) % palette.length];
    if (name === "Earth") color = "#228B22";
    if (name === "Water") color = "#1E90FF";
    if (name === "Fire") color = "#FF6347";
    if (name === "Air") color = "#E0E0E0";

    ctx.fillStyle = color;
    ctx.strokeStyle = "#333";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.rect(x - boxWidth / 2, y, boxWidth, boxHeight);
    ctx.fill();
    ctx.stroke();

    ctx.font = "bold 13px Arial";
    ctx.fillStyle = "#000";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(name, x, y + boxHeight / 2);
  };

  drawRecipeTree(recipe, 0, 0);
  ctx.restore();
}

function RecipeAtlas({ recipes, elementName }) {
  const canvasRef = useRef(null);
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });

  useEffect(() => {
    if (!recipes || recipes.length === 0) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");

    // Set canvas to full window size
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    // Clear and prepare canvas
    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Apply transformations (pan and zoom)
    ctx.translate(position.x, position.y);
    ctx.scale(scale, scale);

    // Draw background
    ctx.fillStyle = "#faf9f4";
    ctx.fillRect(-10000, -10000, 20000, 20000); // Large background

    // Calculate layout for multiple recipes
    const spacing = 800;
    let totalWidth = recipes.length * spacing;
    let startX = -totalWidth / 2 + spacing / 2;

    // Draw all recipes
    recipes.forEach((recipe, i) => {
      const x = startX + i * spacing;
      drawEntireRecipe(ctx, recipe, x, 100, i);
    });
  }, [recipes, scale, position]);

  // Mouse event handlers for panning
  const handleMouseDown = (e) => {
    setIsDragging(true);
    setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y });
  };

  const handleMouseMove = (e) => {
    if (isDragging) {
      setPosition({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y,
      });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  // Mouse wheel for zooming
  const handleWheel = (e) => {
    e.preventDefault();
    const delta = -Math.sign(e.deltaY) * 0.1;
    setScale((prevScale) => {
      const newScale = Math.max(0.1, Math.min(3, prevScale + delta));

      // Adjust position to zoom toward cursor
      const rect = canvasRef.current.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;
      const mouseY = e.clientY - rect.top;

      const newPosition = {
        x: position.x + (mouseX - position.x) * (1 - newScale / prevScale),
        y: position.y + (mouseY - position.y) * (1 - newScale / prevScale),
      };

      setPosition(newPosition);
      return newScale;
    });
  };

  // Drawing functions (similar to RecipeTree but adapted for multiple recipes)
  // Drop these into RecipeAtlas.jsx

  const drawEntireRecipe = (ctx, recipe, startX, startY, index) => {
    // Draw recipe title
    ctx.save();
    ctx.translate(startX, startY);
    ctx.font = "bold 24px Arial";
    ctx.fillStyle = "#333";
    ctx.textAlign = "center";
    ctx.fillText(`Recipe ${index + 1}`, 0, -20);

    let nextX = 0;

    const drawRecipeTree = (node, depth, y) => {
      const xSpacing = 120;
      let x;

      // leaf node
      if (!node.children || node.children.length === 0) {
        x = nextX * xSpacing;
        nextX++;
      } else {
        const leftX = drawRecipeTree(node.children[0], depth + 1, y + 100);
        const rightX = drawRecipeTree(node.children[1], depth + 1, y + 100);
        x = (leftX + rightX) / 2;

        ctx.beginPath();
        ctx.moveTo(x, y + 40);
        ctx.lineTo(leftX, y + 100);
        ctx.stroke();

        ctx.beginPath();
        ctx.moveTo(x, y + 40);
        ctx.lineTo(rightX, y + 100);
        ctx.stroke();

        ctx.font = "20px Arial";
        ctx.fillStyle = "#000";
        ctx.textAlign = "center";
        ctx.fillText("+", x, y + 60);
      }

      drawElementBox(node.name, x, y);
      return x;
    };

    const drawElementBox = (name, x, y) => {
      const boxWidth = 105,
        boxHeight = 40;
      const palette = [
        "#a8d5a8",
        "#a8c5d5",
        "#d5a8a8",
        "#808080",
        "#f4c2c2",
        "#f0e68c",
        "#dda0dd",
        "#add8e6",
        "#90ee90",
        "#ffcccb",
      ];
      let hash = 0;
      for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash);
      }
      let color = palette[Math.abs(hash) % palette.length];
      if (name === "Earth") color = "#228B22";
      if (name === "Water") color = "#1E90FF";
      if (name === "Fire") color = "#FF6347";
      if (name === "Air") color = "#E0E0E0";

      ctx.fillStyle = color;
      ctx.strokeStyle = "#333";
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.rect(x - boxWidth / 2, y, boxWidth, boxHeight);
      ctx.fill();
      ctx.stroke();

      ctx.font = "bold 13px Arial";
      ctx.fillStyle = "#000";
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText(name, x, y + boxHeight / 2);
    };

    drawRecipeTree(recipe, 0, 0);
    ctx.restore();
  };

  return (
    <div style={{ position: "relative", width: "100vw", height: "100vh", overflow: "hidden" }}>
      <div
        className="controls"
        style={{
          position: "absolute",
          top: 20,
          left: 20,
          zIndex: 1000,
          background: "rgba(255,255,255,0.8)",
          padding: 10,
          borderRadius: 5,
        }}>
        <h2>All Recipes for {elementName}</h2>
        <div>
          <button onClick={() => setScale((prev) => Math.min(prev + 0.1, 3))}>Zoom In</button>
          <button onClick={() => setScale((prev) => Math.max(prev - 0.1, 0.1))}>Zoom Out</button>
          <button
            onClick={() => {
              setScale(1);
              setPosition({ x: 0, y: 0 });
            }}>
            Reset View
          </button>
        </div>
        <Link href="/" style={{ display: "block", marginTop: 10 }}>
          Back to Finder
        </Link>
      </div>

      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onWheel={handleWheel}
        style={{
          display: "block",
          cursor: isDragging ? "grabbing" : "grab",
        }}
      />
    </div>
  );
}

export default RecipeAtlas;
