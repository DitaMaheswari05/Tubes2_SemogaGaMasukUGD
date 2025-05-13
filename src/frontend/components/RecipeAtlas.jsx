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
    ctx.fillRect(-1000000, -1000000, 2000000, 2000000);

    // Calculate grid layout
    const count = recipes.length;
    const gridSize = Math.ceil(Math.sqrt(count));

    // Spacing between recipes (adjust based on complexity)
    const spacing = calculateSpacing();
    const horizontalSpacing = spacing.horizontal;
    const verticalSpacing = spacing.vertical;

    // Total grid dimensions
    const gridWidth = gridSize * horizontalSpacing;
    const gridHeight = gridSize * verticalSpacing;

    // Starting position (center the grid)
    const startX = -gridWidth / 2 + horizontalSpacing / 2;
    const startY = -gridHeight / 2 + verticalSpacing / 2;

    // Draw all recipes in a grid pattern
    recipes.forEach((recipe, i) => {
      const row = Math.floor(i / gridSize);
      const col = i % gridSize;

      const x = startX + col * horizontalSpacing;
      const y = startY + row * verticalSpacing;

      drawEntireRecipe(ctx, recipe, x, y, i);
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

    const delta = -Math.sign(e.deltaY) * 0.05;

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

  // Calculate dynamic spacing based on recipe complexity
  const calculateSpacing = () => {
    // Initialize with reasonable minimum values
    let maxWidth = 800;
    let maxHeight = 600;

    // Analyze each recipe tree
    recipes.forEach((recipe) => {
      // Calculate width by counting leaf nodes
      const countLeaves = (node) => {
        if (!node) return 0;
        if (!node.children || node.children.length === 0) return 1;
        return node.children.reduce((sum, child) => sum + countLeaves(child), 0);
      };

      // Calculate depth by finding longest path
      const findDepth = (node) => {
        if (!node) return 0;
        if (!node.children || node.children.length === 0) return 1;
        return 1 + Math.max(...node.children.map(findDepth));
      };

      const leaves = countLeaves(recipe);
      const depth = findDepth(recipe);

      // Each leaf needs ~120px, each level needs ~100px
      const treeWidth = Math.max(leaves * 150, 500);
      const treeHeight = Math.max(depth * 120, 400);

      maxWidth = Math.max(maxWidth, treeWidth);
      maxHeight = Math.max(maxHeight, treeHeight);
    });

    // Add padding between trees
    return {
      horizontal: maxWidth + 200, // Extra spacing between columns
      vertical: maxHeight + 200, // Extra spacing between rows
    };
  };

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
      const xSpacing = 125;
      let x;

      // leaf node
      if (!node.children || node.children.length === 0) {
        x = nextX * xSpacing;
        nextX++;
      } else {
        const leftX = drawRecipeTree(node.children[0], depth + 1, y + 75);
        const rightX = drawRecipeTree(node.children[1], depth + 1, y + 75);
        x = (leftX + rightX) / 2;

        ctx.beginPath();
        ctx.moveTo(x, y + 40);
        ctx.lineTo(leftX, y + 75);
        ctx.stroke();

        ctx.beginPath();
        ctx.moveTo(x, y + 40);
        ctx.lineTo(rightX, y + 75);
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
