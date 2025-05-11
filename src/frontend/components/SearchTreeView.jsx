"use client";

import { useRef, useEffect, useState } from "react";

function SearchTreeView({ discovered, currentNode, queueNodes, seenNodes }) {
  const canvasRef = useRef(null);
  const [scale, setScale] = useState(1);

  // Build adjacency list from discovered edges
  const buildGraphStructure = () => {
    // Map of element name -> array of connected elements
    const graph = {};
    const allNodes = new Set();

    // First add base elements
    const baseElements = ["Air", "Earth", "Fire", "Water"];
    baseElements.forEach((el) => {
      graph[el] = [];
      allNodes.add(el);
    });

    // Add all nodes and connections
    Object.entries(discovered).forEach(([product, { A, B }]) => {
      allNodes.add(product);
      allNodes.add(A);
      allNodes.add(B);

      // Only add directional connections (base→product)
      if (!graph[A]) graph[A] = [];
      if (!graph[B]) graph[B] = [];

      graph[A].push({ target: product, partner: B });
      graph[B].push({ target: product, partner: A });
    });

    return {
      graph,
      allNodes: Array.from(allNodes),
    };
  };

  // Position calculation for nodes using layered approach
  const calculateNodePositions = (graph, allNodes) => {
    const positions = {};
    const layerMap = {}; // element → layer
    const baseElements = ["Air", "Earth", "Fire", "Water"];

    // Place base elements at the top in a row
    const topY = 70;
    const width = 900;
    const spacing = width / (baseElements.length + 1);

    baseElements.forEach((el, i) => {
      positions[el] = {
        x: spacing * (i + 1),
        y: topY,
      };
      layerMap[el] = 0;
    });

    // Track nodes in each layer for better sizing
    const layerCounts = [baseElements.length];
    let maxNodesPerLayer = baseElements.length;

    // Breadth-first layout for other elements
    const queue = [...baseElements];
    const visited = new Set(baseElements);
    let currentLayer = 1;

    while (queue.length > 0) {
      const levelSize = queue.length;
      let nodesInThisLayer = [];

      for (let i = 0; i < levelSize; i++) {
        const node = queue.shift();

        // Find all products created using this node
        const neighbors = (graph[node] || []).map((n) => n.target).filter((n) => !visited.has(n));

        // Add unique neighbors to the next layer
        neighbors.forEach((neighbor) => {
          if (!visited.has(neighbor)) {
            visited.add(neighbor);
            queue.push(neighbor);
            nodesInThisLayer.push(neighbor);
            layerMap[neighbor] = currentLayer;
          }
        });
      }

      // Position all nodes in this layer horizontally
      layerCounts[currentLayer] = nodesInThisLayer.length;
      if (nodesInThisLayer.length > maxNodesPerLayer) {
        maxNodesPerLayer = nodesInThisLayer.length;
      }

      const layerWidth = Math.max(width, nodesInThisLayer.length * 150);
      const layerSpacing = layerWidth / (nodesInThisLayer.length + 1);
      const layerY = topY + currentLayer * 400;

      nodesInThisLayer.forEach((node, i) => {
        positions[node] = {
          x: layerSpacing * (i + 1),
          y: layerY,
        };
      });

      currentLayer++;
    }

    // Return positions and layout info for canvas sizing
    return {
      positions,
      maxDepth: currentLayer,
      maxWidth: maxNodesPerLayer,
    };
  };

  useEffect(() => {
    if (!discovered) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");

    // Build graph structure
    const { graph, allNodes } = buildGraphStructure();

    // Calculate node positions using layered approach
    const { positions, maxDepth, maxWidth } = calculateNodePositions(graph, allNodes);

    // Size canvas appropriately based on actual layout
    const nodeWidth = 150; // Width needed per node
    const nodeHeight = 400; // Height needed per layer
    const horizontalPadding = 200;
    const verticalPadding = 100;

    // Calculate dimensions based on actual layout needs
    const canvasWidth = Math.max(900, (maxWidth + 1) * nodeWidth);
    const canvasHeight = Math.max(600, (maxDepth + 1) * nodeHeight);

    // Set canvas size
    const xs = allNodes.map((n) => positions[n].x);
    const minX = Math.min(...xs);
    const maxX = Math.max(...xs);
    const padding = 100;
    const width = Math.max(maxX - minX + padding * 2, allNodes.length * 120);
    canvas.width = canvasWidth + horizontalPadding;
    canvas.height = canvasHeight + verticalPadding;

    // Standard canvas setup
    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.scale(scale, scale);

    // Background
    ctx.fillStyle = "#faf9f4";
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // Draw connections
    ctx.strokeStyle = "#666";
    ctx.lineWidth = 1;

    Object.entries(discovered).forEach(([product, { A, B }]) => {
      if (positions[product] && positions[A]) {
        // Draw product ← A line
        ctx.beginPath();
        ctx.moveTo(positions[product].x, positions[product].y - 20);
        ctx.lineTo(positions[A].x, positions[A].y + 20);
        ctx.stroke();

        // Label with partner element
        const midX = (positions[product].x + positions[A].x) / 2;
        const midY = (positions[product].y + positions[A].y) / 2;
        ctx.font = "13px Arial";
        ctx.fillStyle = "#444";
        ctx.textAlign = "center";
        ctx.fillText(`+ ${B}`, midX, midY);
      }

      if (positions[product] && positions[B]) {
        // Draw product ← B line
        ctx.beginPath();
        ctx.moveTo(positions[product].x, positions[product].y - 20);
        ctx.lineTo(positions[B].x, positions[B].y + 20);
        ctx.stroke();

        // Label with partner element
        const midX = (positions[product].x + positions[B].x) / 2;
        const midY = (positions[product].y + positions[B].y) / 2;
        ctx.font = "13px Arial";
        ctx.fillStyle = "#444";
        ctx.textAlign = "center";
        ctx.fillText(`+ ${A}`, midX, midY);
      }
    });

    // Draw nodes
    allNodes.forEach((element) => {
      if (!positions[element]) return;
      const pos = positions[element];

      // Determine node appearance based on search state
      let color,
        strokeColor = "#333",
        strokeWidth = 1;

      // Base element colors
      if (element === "Earth") color = "#228B22";
      else if (element === "Water") color = "#1E90FF";
      else if (element === "Fire") color = "#FF6347";
      else if (element === "Air") color = "#E0E0E0";
      // Other elements by hash
      else {
        const colorPalette = [
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
        for (let i = 0; i < element.length; i++) {
          hash = element.charCodeAt(i) + ((hash << 5) - hash);
        }
        color = colorPalette[Math.abs(hash) % colorPalette.length];
      }

      // Highlight current node
      if (element === currentNode) {
        strokeColor = "#ff0000";
        strokeWidth = 3;
      }
      // Highlight queue nodes
      else if (queueNodes.includes(element)) {
        strokeColor = "#0000ff";
        strokeWidth = 2;
      }

      // Draw node box
      const boxWidth = 105;
      const boxHeight = 40;

      ctx.fillStyle = color;
      ctx.strokeStyle = strokeColor;
      ctx.lineWidth = strokeWidth;
      ctx.beginPath();
      ctx.rect(pos.x - boxWidth / 2, pos.y - boxHeight / 2, boxWidth, boxHeight);
      ctx.fill();
      ctx.stroke();

      // Node label
      ctx.font = "bold 13px Arial";
      ctx.fillStyle = "#000";
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText(element, pos.x, pos.y);
    });
  }, [discovered, currentNode, queueNodes, seenNodes, scale]);

  const zoomIn = () => setScale((prev) => Math.min(prev + 0.1, 3));
  const zoomOut = () => setScale((prev) => Math.max(prev - 0.1, 0.2));

  return (
    <div className="search-tree-view">
      <div style={{ marginBottom: 8 }}>
        <button onClick={zoomIn} style={{ marginRight: 8 }}>
          ➕
        </button>
        <button onClick={zoomOut}>➖</button>
      </div>
      <div className="canvas-container">
        <canvas ref={canvasRef} style={{ display: "block", border: "1px solid #ddd", overflowX: "auto", width: "100%" }}></canvas>
      </div>
      <div className="legend" style={{ marginTop: 8, fontSize: 12 }}>
        <div>
          <span style={{ color: "#ff0000", fontWeight: "bold" }}>●</span> Current Node
        </div>
        <div>
          <span style={{ color: "#0000ff", fontWeight: "bold" }}>●</span> Queue Nodes
        </div>
        <div>
          <span style={{ color: "#228B22", fontWeight: "bold" }}>▮</span> Earth
        </div>
        <div>
          <span style={{ color: "#1E90FF", fontWeight: "bold" }}>▮</span> Water
        </div>
        <div>
          <span style={{ color: "#FF6347", fontWeight: "bold" }}>▮</span> Fire
        </div>
        <div>
          <span style={{ color: "#E0E0E0", fontWeight: "bold" }}>▮</span> Air
        </div>
      </div>
    </div>
  );
}

export default SearchTreeView;
