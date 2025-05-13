// pages/atlas/[element].js
export default function AtlasPage({ params }) {
  const { element } = params;
  const [recipes, setRecipes] = useState([]);

  useEffect(() => {
    // Fetch recipes for this element
    fetch(`/api/find?target=${element}&multi=true&maxPaths=100`)
      .then((res) => res.json())
      .then((data) => {
        setRecipes(data.tree || []);
      });
  }, [element]);

  return <RecipeAtlas recipes={recipes} elementName={element} />;
}
