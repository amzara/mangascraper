# Build stage
FROM ghcr.io/cirruslabs/flutter:3.24.0 AS builder

WORKDIR /app

# Copy pubspec files
COPY pubspec.yaml pubspec.lock ./
RUN flutter pub get

# Copy source code
COPY . .

# Inject host values from docker-compose build args (overrides copied .env)
ARG API_HOST
ARG IMAGE_HOST
RUN if [ -n "$API_HOST" ]; then \
      printf "API_HOST=${API_HOST}\nIMAGE_HOST=${IMAGE_HOST}\n" > .env; \
    fi

# Build for web
RUN flutter build web --release

# Production stage - serve with nginx
FROM nginx:alpine

# Copy built web app
COPY --from=builder /app/build/web /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
