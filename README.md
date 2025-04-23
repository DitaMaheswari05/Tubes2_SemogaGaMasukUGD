## BE & FE
`src/backend`: Go (serves recipe.json & SVGs)
`src/frontend`: Next.js

## Prereq
- Go ≥ 1.18
- Node.js ≥ 16 & npm
- (On Windows) Use WSL or a Linux shell so that paths & exec bits work

## Backend Setup
1. Open first terminal
2. Run
    ```
    cd src/backend
    go mod tidy
    go run .
    ```
3. `go run .` will scrape Little Alchemy 2 website and generate local `src/backend/svgs/{Tier}/{Element}.svg` and `src/backend/json/recipe.json`

## Frontend Setup
1. Open second terminal
2. Run
    ```
    cd src/frontend
    npm install
    npm run dev
    ```
3. Open http://localhost:3000/ in browser
