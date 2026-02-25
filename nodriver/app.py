from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from engine import start_browser, stop_browser, get_cookies, is_browser_active

app = FastAPI(title="Mangakakalot Scraper API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class GetCookiesRequest(BaseModel):
    wait_time: int = 3

@app.on_event("startup")
async def startup_event():
    await start_browser()

@app.on_event("shutdown")
async def shutdown_event():
    await stop_browser()

@app.get("/")
async def get_api_info():
    return {
        "message": "Mangakakalot Scraper API",
        "endpoints": {
            "/getCfCookies": "POST - Get Cloudflare cookies",
            "/health": "GET - Health check"
        }
    }

@app.get("/health")
async def get_health():
    return {"status": "healthy", "browser_active": is_browser_active()}

@app.post("/getCfCookies")
async def get_cf_cookies(request: GetCookiesRequest):
    try:
        print(f"[API] Received getCfCookies request")
        result = await get_cookies(request.wait_time)
        print(f"[API] Successfully returned cookies")
        return result
        
    except Exception as e:
        print(f"[API] ✗ Error: {str(e)}")
        return {
            "success": False,
            "error": str(e)
        }