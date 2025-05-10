"use client"

import { useRef, useEffect, useState } from "react"
// import "./RecipeTree.css"

function RecipeTree({ path, index }) {
  const canvasRef = useRef(null)
  const [scale, setScale] = useState(1)
  let nextX = 0

  useEffect(() => {
    if (!path) return
    const canvas = canvasRef.current
    const ctx = canvas.getContext("2d")

    const width = Math.max(1200, getTreeWidth(path) * 120)
    const height = Math.max(800, getTreeDepth(path) * 120)

    canvas.width = width
    canvas.height = height

    // Reset dan apply zoom scale
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, width, height)
    ctx.scale(scale, scale)

    ctx.fillStyle = "#faf9f4"
    ctx.fillRect(0, 0, width, height)

    nextX = 0
    drawRecipeTree(ctx, path, 0, 50)
  }, [path, scale])

  const getTreeDepth = (node) => {
    if (!node.ingredients) return 1
    return 1 + Math.max(getTreeDepth(node.ingredients[0]), getTreeDepth(node.ingredients[1]))
  }

  const getTreeWidth = (node) => {
    if (!node.ingredients) return 1
    return getTreeWidth(node.ingredients[0]) + getTreeWidth(node.ingredients[1])
  }

  const drawRecipeTree = (ctx, node, depth, y) => {
    const xSpacing = 120
    let x

    if (!node.ingredients) {
      x = nextX * xSpacing + 100
      nextX++
    } else {
      const leftX = drawRecipeTree(ctx, node.ingredients[0], depth + 1, y + 100)
      const rightX = drawRecipeTree(ctx, node.ingredients[1], depth + 1, y + 100)
      x = (leftX + rightX) / 2

      ctx.beginPath()
      ctx.moveTo(x, y + 40)
      ctx.lineTo(leftX, y + 100)
      ctx.stroke()

      ctx.beginPath()
      ctx.moveTo(x, y + 40)
      ctx.lineTo(rightX, y + 100)
      ctx.stroke()

      ctx.font = "20px Arial"
      ctx.fillStyle = "#000"
      ctx.textAlign = "center"
      ctx.fillText("+", (leftX + rightX) / 2, y + 90)
    }

    drawElementBox(ctx, node.element, x, y)
    return x
  }

  const drawElementBox = (ctx, element, x, y) => {
    const boxWidth = 105
    const boxHeight = 40

    const colorPalette = [
      "#a8d5a8", "#a8c5d5", "#d5a8a8", "#d5d5d5", "#f4c2c2",
      "#f0e68c", "#dda0dd", "#add8e6", "#90ee90", "#ffcccb"
    ]

    let hash = 0
    for (let i = 0; i < element.length; i++) {
      hash = element.charCodeAt(i) + ((hash << 5) - hash)
    }
    let color = colorPalette[Math.abs(hash) % colorPalette.length]

    if (element === "Earth") color = "#228B22"
    if (element === "Water") color = "#1E90FF"
    if (element === "Fire") color = "#FF6347"
    if (element === "Air") color = "#D3D3D3"

    ctx.fillStyle = color
    ctx.strokeStyle = "#333"
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.rect(x - boxWidth / 2, y, boxWidth, boxHeight)
    ctx.fill()
    ctx.stroke()

    ctx.font = "bold 13px Arial"
    ctx.fillStyle = "#000"
    ctx.textAlign = "center"
    ctx.textBaseline = "middle"
    ctx.fillText(element, x, y + boxHeight / 2)
  }

  const zoomIn = () => setScale((prev) => Math.min(prev + 0.1, 3))
  const zoomOut = () => setScale((prev) => Math.max(prev - 0.1, 0.2))

  return (
    <div className="recipe-tree">
      <h3>Recipe {index + 1}</h3>
      <div style={{ marginBottom: 8 }}>
        <button onClick={zoomIn} style={{ marginRight: 8 }}>➕ </button>
        <button onClick={zoomOut}>➖</button>
      </div>
      <div className="canvas-container" style={{ overflow: "auto" }}>
        <canvas ref={canvasRef}></canvas>
      </div>
    </div>
  )
}

export default RecipeTree
