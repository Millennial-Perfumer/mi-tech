# Build stage
FROM node:20-alpine AS build

WORKDIR /app
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy all frontend files
COPY . .

RUN ls -la

# Build the app for production
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built assets from build stage to Nginx web root
COPY --from=build /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Run nginx in foreground
CMD ["nginx", "-g", "daemon off;"]
