## Backend
src/backend: Go (serves recipe.json & SVGs)

## Frontend
src/frontend: Next.js

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

## Frontend Setup
1. Open second terminal
2. Run
    ```
    cd src/frontend
    npm run dev
    ```
3. Open http://localhost:3000/ in browser
