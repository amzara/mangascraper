import uvicorn

if __name__ == "__main__":
    print("=" * 80)
    print("Starting Mangakakalot Scraper API Server")
    print("=" * 80)
    print("Server will start on http://localhost:8000")
    print("Scraping will ONLY happen when /scrape endpoint is called")
    print("=" * 80)
    print()
    
    # Import and run the FastAPI app
    from app import app
    uvicorn.run(app, host="0.0.0.0", port=8000)