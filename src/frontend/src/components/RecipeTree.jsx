"use client"

import { useRef, useEffect } from "react"
import "./RecipeTree.css"

function RecipeTree({ path, index }) {
  const canvasRef = useRef(null)

  useEffect(() => {
    if (!path) return

    const canvas = canvasRef.current
    const ctx = canvas.getContext("2d")

    const depth = getTreeDepth(path)
    const width = Math.max(800, depth * 200)
    const height = Math.max(400, getTreeWidth(path) * 80)

    canvas.width = width
    canvas.height = height

    // clear canvas
    ctx.fillStyle = "#faf9f4"
    ctx.fillRect(0, 0, width, height)

    // buat recipe tree
    drawRecipeTree(ctx, path, width / 2, 50, width / 4)
  }, [path])

  // function buat dapetin kedalaman node
  const getTreeDepth = (node) => {
    if (!node.ingredients) return 1
    return 1 + Math.max(getTreeDepth(node.ingredients[0]), getTreeDepth(node.ingredients[1]))
  }

  // function buat dapetin lebar node
  const getTreeWidth = (node) => {
    if (!node.ingredients) return 1
    return getTreeWidth(node.ingredients[0]) + getTreeWidth(node.ingredients[1])
  }

  // function buat gambar tree
  const drawRecipeTree = (ctx, node, x, y, xOffset) => {
    // gambar kotak elemen
    drawElementBox(ctx, node.element, x, y)

    if (node.ingredients) {
      const leftChild = node.ingredients[0]
      const rightChild = node.ingredients[1]

      const childY = y + 80
      const leftX = x - xOffset
      const rightX = x + xOffset

      // garis ke child
      ctx.beginPath()
      ctx.moveTo(x, y + 50)
      ctx.lineTo(leftX, childY)
      ctx.stroke()

      ctx.beginPath()
      ctx.moveTo(x, y + 50)
      ctx.lineTo(rightX, childY)
      ctx.stroke()

      // tanda +
      ctx.font = "20px Arial"
      ctx.fillStyle = "#000"
      ctx.textAlign = "center"
      ctx.fillText("+", (leftX + rightX) / 2, childY - 10)

      // gambar node child secara rekursif
      drawRecipeTree(ctx, leftChild, leftX, childY, xOffset / 2)
      drawRecipeTree(ctx, rightChild, rightX, childY, xOffset / 2)
    }
  }

  // gmbar kotak elemen
  const drawElementBox = (ctx, element, x, y) => {
    const boxWidth = 100
    const boxHeight = 50

    let color = "#f0f0f0" // Default gray

    if (["Earth", "Water", "Fire", "Air"].includes(element)) {
      if (element === "Earth") color = "#a8d5a8" // Green
      if (element === "Water") color = "#a8c5d5" // Blue
      if (element === "Fire") color = "#d5a8a8" // Red
      if (element === "Air") color = "#d5d5d5" // Light gray
    }

    // buat kotak elemennya
    ctx.fillStyle = color
    ctx.strokeStyle = "#333"
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.rect(x - boxWidth / 2, y, boxWidth, boxHeight)
    ctx.fill()
    ctx.stroke()

    // buat nama elemennya
    ctx.font = "14px Arial"
    ctx.fillStyle = "#000"
    ctx.textAlign = "center"
    ctx.textBaseline = "middle"
    ctx.fillText(element, x, y + boxHeight / 2)
  }

  return (
    <div className="recipe-tree">
      <h3>Recipe {index + 1}</h3>
      <div className="canvas-container">
        <canvas ref={canvasRef}></canvas>
      </div>
    </div>
  )
}

export default RecipeTree
