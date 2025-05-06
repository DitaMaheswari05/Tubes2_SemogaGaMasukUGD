## BE & FE
`src/backend`: Go (serves recipe.json & SVGs)
`src/frontend`: Next.js

## Prereq
- Go ≥ 1.18
- Node.js ≥ 16 & npm
- (On Windows) Use WSL or a Linux shell so that paths & exec bits work
Make sure you have Docker and Docker Compose installed:
- Docker https://www.docker.com/products/docker-desktop/ (Docker Compose comes bundled with Docker Desktop on Windows and macOS.
On Linux, follow the official Docker Compose installation guide.)

## Installation with Docker
Once Docker is installed and running:
1. Clone this repository:
```shell
git clone https://github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD
```
2. Navigate to the source directory of the program by running the following command in the terminal:

```shell
cd ./src
```

3. Ensure Docker Desktop is running. Once the user is in the root directory, run the following command in the terminal:

```shell
docker compose up --build
```
6. To access the website, go to the following link in your web browser: [http://localhost:3000](http://localhost:3000)

## Installation without Docker
### Backend Setup
1. Open first terminal
2. Run
    ```
    cd src/backend
    go mod tidy
    go run .
    ```
3. `go run .` will scrape Little Alchemy 2 website and generate local `src/backend/svgs/{Tier}/{Element}.svg` and `src/backend/json/recipe.json`

### Frontend Setup
1. Open second terminal
2. Run
    ```
    cd src/frontend
    npm install
    npm run dev
    ```
3. Open http://localhost:3000/ in browser
