# Tubes2_SemogaGaMasukUGD

## BE & FE
`src/backend`: Go (serves recipe.json & SVGs)  
`src/frontend`: React.js

## Prereq
- Go ≥ 1.18 (for non-Docker installation)
- Node.js ≥ 16 & npm (for non-Docker installation)
- Docker and Docker Compose (for Docker installation)
- (On Windows) Use WSL or a Linux shell so that paths & exec bits work

# App Installation & Usage

## Online Access
You can access the deployed application online at [https://tubes2-semoga-ga-masuk-ugd.vercel.app/](https://tubes2-semoga-ga-masuk-ugd.vercel.app/).

## Local Installation with Docker
Once Docker is installed and running:

1. Clone this repository:
   ```shell
   git clone https://github.com/wiwekaputera/Tubes2_SemogaGaMasukUGD
   ```

2. Navigate to the source directory of the program by running the following command in the terminal:
   ```shell
   cd ./src
   ```

3. Ensure Docker Desktop is running.

4. Build the Docker images:
   ```shell
   docker-compose build
   ```

5. Start the services:
   ```shell
   docker-compose up
   ```

6. To access the website, go to the following link in your web browser: [http://localhost:3000](http://localhost:3000)

## Local Installation without Docker

### Backend Setup
1. Open first terminal
2. Run:
   ```shell
   cd src/backend
   go mod tidy
   go run .
   ```
3. `go run .` will scrape the Little Alchemy 2 website and generate local `src/backend/svgs/{Tier}/{Element}.svg` and `src/backend/json/recipe.json`

### Frontend Setup
1. Open second terminal
2. Run:
   ```shell
   cd src/frontend
   npm install
   npm run dev
   ```
3. Open [http://localhost:3000](http://localhost:3000) in your browser

# How to use the app?
1. Access the app (home page)
   ![image](https://github.com/user-attachments/assets/c574cf2b-71d4-464c-9491-93af44f80c10)

3. Click Scrape All
   ![Screenshot 2025-05-13 213927](https://github.com/user-attachments/assets/edfce231-4c70-45fe-9460-57248956b5d7)

   ![image](https://github.com/user-attachments/assets/d9c43dd2-7161-421c-b972-04f468881312)

5. Enjoy!
   ![image](https://github.com/user-attachments/assets/5904c33e-e8fc-45ef-9eb2-16748cfa9e51)



## Team Information

| Name                   | ID       |
|------------------------|----------|
| Dita Maheswari         | 13523125 |
| I Made Wiweka Putera   | 13523160 |
| Asybel B.P. Sianipar   | 15223011 |
