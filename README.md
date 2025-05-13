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
   ![image](https://github.com/user-attachments/assets/3532e3b0-5519-466a-ae5b-6ea44bc96b29)


3. Click Scrape All
   ![Screenshot 2025-05-13 214330](https://github.com/user-attachments/assets/e74d0c7c-0ec2-4f72-848c-9e8dcdb3c557)

   ![image](https://github.com/user-attachments/assets/2a41caf1-0bfe-41de-be87-331d66e08537)


5. Enjoy!
   ![image](https://github.com/user-attachments/assets/9d3d5105-2f19-49f7-a343-9702d82ed999)




## Team Information

| Name                   | ID       |
|------------------------|----------|
| Dita Maheswari         | 13523125 |
| I Made Wiweka Putera   | 13523160 |
| Asybel B.P. Sianipar   | 15223011 |
