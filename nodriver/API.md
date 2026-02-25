# Mangakakalot Scraper API

A FastAPI-based web service that scrapes websites, solves Cloudflare challenges, and returns cookies.

## File Structure

- **main.py** - HTTP server startup (run this to start the API)
- **app.py** - FastAPI application with HTTP handlers
- **scraper.py** - Scraping logic and browser management
- **fetch_image.py** - Example script to fetch images using cookies
- **requirements.txt** - Python dependencies

## Features

- ✅ Automatic Cloudflare challenge solving
- ✅ Extract all cookies (including cf_clearance)
- ✅ RESTful API with JSON responses
- ✅ Headless browser for efficiency
- ✅ Scraping ONLY happens when `/getCfCookies` endpoint is called
- ✅ Clean separation of concerns (HTTP handlers vs scraping logic)

## Installation

1. Install dependencies:
```bash
pip install -r requirements.txt
```

## Running the API

Start the API server:
```bash
python main.py
```

The server will start on `http://localhost:8000`

**Note:** No scraping happens automatically. Scraping only occurs when you call the `/getCfCookies` endpoint.

## API Endpoints

### 1. Root

**Endpoint:** `GET /`

**Response:**
```json
{
  "message": "Mangakakalot Scraper API",
  "endpoints": {
    "/getCfCookies": "POST - Get Cloudflare cookies",
    "/health": "GET - Health check"
  }
}
```

### 2. Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "browser_active": true
}
```

### 3. Get Cloudflare Cookies

**Endpoint:** `POST /getCfCookies`

**Process:**
1. **First visit:** mangakakalot.gg (bypass Cloudflare, assign cookies)
2. **Second visit:** mangakakalot.gg/official (capture cookies)

**Request Body:**
```json
{
  "wait_time": 3
}
```

**Parameters:**

**Response:**
```json
{
  "url": "https://www.mangakakalot.gg/official",
  "success": true,
  "cookies": {
    "cf_clearance": "...",
    "UID": "...",
    "NID": "...",
    "_ga": "GA1.1.1136672048.1769341084",
    ...
  }
}
```

**Note:** Use the returned `cookies` dictionary for subsequent requests to mangakakalot.gg.

## Usage Examples

### Using Python

```python
import requests

# Get Cloudflare cookies
response = requests.post("http://localhost:8000/getCfCookies", json={
    "wait_time": 3
})

data = response.json()

if data['success']:
    # Get cookies (includes cf_clearance and session_id)
    cookies = data['cookies']
    print(f"Cookies: {cookies}")
    
    # Use cookies for subsequent requests to mangakakalot.gg
    response = requests.get("https://www.mangakakalot.gg/api/endpoint", 
                           cookies=cookies)
    print(f"API Response: {response.text}")
```

### Using cURL

```bash
curl -X POST http://localhost:8000/getCfCookies \
  -H "Content-Type: application/json" \
  -d '{
    "wait_time": 3
  }'
```

### Using JavaScript/Fetch

```javascript
fetch('http://localhost:8000/getCfCookies', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    wait_time: 3
  })
})
.then(response => response.json())
.then(data => {
  if (data.success) {
    console.log('Cookies:', data.cookies);
    
    // Use cookies for subsequent requests
    fetch('https://www.mangakakalot.gg/api/endpoint', {
      credentials: 'include',
      headers: {
        'Cookie': Object.entries(data.cookies)
          .map(([k, v]) => `${k}=${v}`)
          .join('; ')
      }
    });
  }
});
```

## Architecture

### File Responsibilities:

**main.py:**
- Startup and server configuration
- Imports app
- Runs uvicorn server

**app.py:**
- FastAPI application
- HTTP handlers (GET /, GET /health, POST /getCfCookies)
- Request/response handling
- Imports and calls scraper functions

**scraper.py:**
- Browser lifecycle management (start/stop)
- Web scraping logic
- Cloudflare challenge solving
- Cookie extraction

### Execution Flow:

1. **Startup** (python main.py)
   - uvicorn starts the FastAPI app
   - FastAPI triggers startup event
   - app.py calls scraper.start_browser()
   - scraper.py initializes headless browser
   - Browser stays idle, waiting for requests

2. **When /getCfCookies is called**
   - app.py receives request
   - app.py calls scraper.get_cookies()
   - scraper.py navigates to mangakakalot.gg
   - scraper.py solves Cloudflare
   - scraper.py visits mangakakalot.gg/official
   - scraper.py extracts cookies
   - scraper.py returns data
   - app.py sends JSON response

3. **Between requests**
   - Browser stays open and ready
   - No scraping happens automatically
   - No CPU usage when idle

## Important Cookies

- **cf_clearance**: Cloudflare clearance cookie - proves you solved the challenge
- **UID**: User identifier
- **NID**: Google/Cloudflare session ID
- **_ga**: Google Analytics session ID
- Other site-specific cookies

## Troubleshooting

**Issue:** "Browser not active" error
**Solution:** Restart the API server (Ctrl+C, then `python main.py`)

**Issue:** Timeout errors
**Solution:** Increase `wait_time` parameter in the request

**Issue:** Cloudflare challenge fails
**Solution:** Check if the website URL is accessible and not blocking automated traffic

**Issue:** Port 8000 already in use
**Solution:** Find and kill the process:
```bash
lsof -ti:8000 | xargs kill -9
```

**Issue:** Import errors
**Solution:** Clear Python cache: `find . -type d -name __pycache__ -exec rm -rf {} +`

## Notes

- First request may take 10-20 seconds due to Cloudflare challenge
- Browser is reused between requests for efficiency
- All cookies are captured, including those set by JavaScript
- The API runs headlessly (no visible browser window)
- Scraping ONLY happens when you explicitly call the `/getCfCookies` endpoint