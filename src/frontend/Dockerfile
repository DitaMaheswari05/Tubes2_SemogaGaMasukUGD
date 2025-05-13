FROM node:23-alpine

# Create app user
RUN addgroup -S app && adduser -S -G app app

# Set working directory
WORKDIR /app

# Copy only package files first (for better caching)
COPY package*.json ./

# Set environment variable for production
ARG BACKEND_URL
ENV BACKEND_URL=$BACKEND_URL
ENV NODE_ENV=production

# Install dependencies
RUN npm ci

# Copy app source code and assign ownership
COPY --chown=app:app . .

# Build the Next.js app
RUN npm run build

USER app

EXPOSE 3000

# Start the app
CMD ["npm", "start"]
