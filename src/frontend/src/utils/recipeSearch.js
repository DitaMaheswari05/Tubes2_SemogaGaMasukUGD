// gua minta gpt buat ini
// ini cuman ngetes web nya aja

// Check if an element is a base element
const isBaseElement = (element, baseElements) => {
    return baseElements.includes(element);
  };
  
  // Build a recipe path from the search result
  const buildRecipePath = (element, parentMap, baseElements) => {
    if (isBaseElement(element, baseElements)) {
      return { element, ingredients: null };
    }
  
    const parents = parentMap.get(element);
    if (!parents) return { element, ingredients: null };
  
    return {
      element,
      ingredients: [
        buildRecipePath(parents[0], parentMap, baseElements),
        buildRecipePath(parents[1], parentMap, baseElements),
      ],
    };
  };
  
  // Search algorithms
  const searchAlgorithms = {
    // Breadth-First Search
    bfs: (targetElement, recipes, baseElements, maxRecipes = 1) => {
      const queue = [];
      const visited = new Set();
      const parentMap = new Map();
      let nodesVisited = 0;
      const paths = [];
  
      baseElements.forEach((element) => {
        queue.push(element);
        visited.add(element);
      });
  
      while (queue.length > 0 && paths.length < maxRecipes) {
        const current = queue.shift();
        nodesVisited++;
  
        for (const [result, combinations] of Object.entries(recipes)) {
          if (paths.length >= maxRecipes) break;
  
          for (const [elem1, elem2] of combinations) {
            if ((current === elem1 || current === elem2) && !visited.has(result)) {
              const otherElem = current === elem1 ? elem2 : elem1;
  
              if (visited.has(otherElem)) {
                parentMap.set(result, [elem1, elem2]);
                visited.add(result);
                queue.push(result);
  
                if (result === targetElement) {
                  paths.push(buildRecipePath(result, parentMap, baseElements));
  
                  if (paths.length >= maxRecipes || maxRecipes === 1) {
                    return { paths, nodesVisited };
                  }
                }
              }
            }
          }
        }
      }
  
      return { paths, nodesVisited };
    },
  
    // Depth-First Search
    dfs: (targetElement, recipes, baseElements, maxRecipes = 1) => {
      const stack = [];
      const visited = new Set();
      const parentMap = new Map();
      let nodesVisited = 0;
      const paths = [];
  
      baseElements.forEach((element) => {
        stack.push(element);
        visited.add(element);
      });
  
      while (stack.length > 0 && paths.length < maxRecipes) {
        const current = stack.pop();
        nodesVisited++;
  
        for (const [result, combinations] of Object.entries(recipes)) {
          if (paths.length >= maxRecipes) break;
  
          for (const [elem1, elem2] of combinations) {
            if ((current === elem1 || current === elem2) && !visited.has(result)) {
              const otherElem = current === elem1 ? elem2 : elem1;
  
              if (visited.has(otherElem)) {
                parentMap.set(result, [elem1, elem2]);
                visited.add(result);
                stack.push(result);
  
                if (result === targetElement) {
                  paths.push(buildRecipePath(result, parentMap, baseElements));
  
                  if (paths.length >= maxRecipes || maxRecipes === 1) {
                    return { paths, nodesVisited };
                  }
                }
              }
            }
          }
        }
      }
  
      return { paths, nodesVisited };
    },
  
    // Bidirectional Search
    bidirectional: (targetElement, recipes, baseElements, maxRecipes = 1) => {
      const forwardQueue = [];
      const forwardVisited = new Set();
      const forwardParentMap = new Map();
  
      const backwardQueue = [targetElement];
      const backwardVisited = new Set([targetElement]);
      const backwardParentMap = new Map();
  
      let nodesVisited = 0;
      const paths = [];
  
      baseElements.forEach((element) => {
        forwardQueue.push(element);
        forwardVisited.add(element);
      });
  
      while (forwardQueue.length > 0 && backwardQueue.length > 0 && paths.length < maxRecipes) {
        const current = forwardQueue.shift();
        nodesVisited++;
  
        if (backwardVisited.has(current)) {
          const path = { element: current, ingredients: null };
  
          if (!isBaseElement(current, baseElements)) {
            const parents = forwardParentMap.get(current);
            if (parents) {
              path.ingredients = [
                buildRecipePath(parents[0], forwardParentMap, baseElements),
                buildRecipePath(parents[1], forwardParentMap, baseElements),
              ];
            }
          }
  
          paths.push(path);
  
          if (paths.length >= maxRecipes || maxRecipes === 1) {
            return { paths, nodesVisited };
          }
  
          continue;
        }
  
        for (const [result, combinations] of Object.entries(recipes)) {
          for (const [elem1, elem2] of combinations) {
            if ((current === elem1 || current === elem2) && !forwardVisited.has(result)) {
              const otherElem = current === elem1 ? elem2 : elem1;
  
              if (forwardVisited.has(otherElem)) {
                forwardParentMap.set(result, [elem1, elem2]);
                forwardVisited.add(result);
                forwardQueue.push(result);
              }
            }
          }
        }
  
        const currentBack = backwardQueue.shift();
        nodesVisited++;
  
        if (forwardVisited.has(currentBack)) {
          const path = { element: currentBack, ingredients: null };
  
          if (!isBaseElement(currentBack, baseElements)) {
            const parents = forwardParentMap.get(currentBack);
            if (parents) {
              path.ingredients = [
                buildRecipePath(parents[0], forwardParentMap, baseElements),
                buildRecipePath(parents[1], forwardParentMap, baseElements),
              ];
            }
          }
  
          paths.push(path);
  
          if (paths.length >= maxRecipes || maxRecipes === 1) {
            return { paths, nodesVisited };
          }
  
          continue;
        }
  
        for (const [result, combinations] of Object.entries(recipes)) {
          if (result === currentBack) {
            for (const [elem1, elem2] of combinations) {
              if (!backwardVisited.has(elem1)) {
                backwardParentMap.set(elem1, [result, elem2]);
                backwardVisited.add(elem1);
                backwardQueue.push(elem1);
              }
  
              if (!backwardVisited.has(elem2)) {
                backwardParentMap.set(elem2, [result, elem1]);
                backwardVisited.add(elem2);
                backwardQueue.push(elem2);
              }
            }
          }
        }
      }
  
      return { paths, nodesVisited };
    }
  };
  
  export const searchRecipes = ({ targetElement, algorithm, searchMode, maxRecipes, recipes, baseElements }) => {
    return searchAlgorithms[algorithm](
      targetElement,
      recipes,
      baseElements,
      searchMode === "shortest" ? 1 : maxRecipes
    );
  };  