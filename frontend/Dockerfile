# Development stage
FROM node:22.16.0-alpine

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache git

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application
COPY . .

# Expose the application port
EXPOSE 3000

# Start development server
CMD ["npm", "run", "dev"]
