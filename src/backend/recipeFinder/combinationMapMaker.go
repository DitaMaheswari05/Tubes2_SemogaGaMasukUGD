package recipeFinder

// BuildCombinationMap builds a CombinationMap from the full Catalog.
func BuildCombinationMap(cat Catalog) CombinationMap {
  combos := make(CombinationMap)
  for _, tier := range cat.Tiers {
    for _, el := range tier.Elements {
      for _, rec := range el.Recipes {
        if len(rec) != 2 {
          continue
        }
        a, b := rec[0], rec[1]
        c1 := IngredientCombo{A: a, B: b}
        c2 := IngredientCombo{A: b, B: a}
        combos[c1] = append(combos[c1], el.Name)
        combos[c2] = append(combos[c2], el.Name)
      }
    }
  }
  return combos
}