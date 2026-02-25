FROM python:3.10-alpine

ENV PYTHONUNBUFFERED=1

WORKDIR /app

# Install temporary build dependencies
RUN apk update && apk add --no-cache --virtual .build-deps \
    alpine-sdk \
    curl \
    wget \
    unzip \
    gnupg \
    gcc \
    musl-dev \
    libffi-dev \
    openssl-dev

# Install runtime dependencies (including Chromium)
RUN apk add --no-cache \
    xvfb \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    chromium

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Remove build dependencies to reduce image size
RUN apk del .build-deps

# Copy application code
COPY . .

# Tell nodriver where Chromium is installed on Alpine
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV DISPLAY=:99

EXPOSE 8000

CMD ["python", "main.py"]