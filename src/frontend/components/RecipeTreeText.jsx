"use client"

function RecipeTreeText({ node, level = 0 }) {
  if (!node) return null;

  const padding = level * 20;

  return (
    <div style={{ marginBottom: 8 }}>
      <div style={{ paddingLeft: padding, display: "flex", alignItems: "center" }}>
        {level > 0 && <span style={{ marginRight: 8, color: "#666" }}>{level === 1 ? "Made from:" : "└"}</span>}
        <span style={{ fontWeight: level === 0 ? "bold" : "normal" }}>{node.name}</span>
      </div>

      {node.children && node.children.map((child, i) => <RecipeTreeText key={`${child.name}-${i}`} node={child} level={level + 1} />)}
    </div>
  );
}

export default RecipeTreeText