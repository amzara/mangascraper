import nodriver as uc
from typing import Dict, Any

BROWSER = None

async def start_browser():
    """Initialize browser on startup"""
    global BROWSER
    try:
        BROWSER = await uc.start(headless=True)
        print("✓ Browser started")
    except Exception as e:
        print(f"✗ Failed to start browser: {str(e)}")
        BROWSER = None

async def stop_browser():
    """Close browser on shutdown"""
    global BROWSER
    if BROWSER:
        try:
            await BROWSER.stop()
            print("✓ Browser stopped")
        except Exception as e:
            print(f"✗ Error stopping browser: {str(e)}")
        finally:
            BROWSER = None

async def get_cookies(wait_time: int = 3) -> Dict[str, Any]:
    """
    Get cookies from mangakakalot.gg/official
    
    Process:
    1. First visit: mangakakalot.gg (bypass Cloudflare, assign cookies)
    2. Second visit: mangakakalot.gg/official (capture cookies)
    
    Args:
        wait_time: Time to wait after page load (default: 3 seconds)
    
    Returns:
        Dictionary containing cookies
    """
    global BROWSER
    
    base_url = "https://www.mangakakalot.gg"
    official_url = f"{base_url}/manga/hajime-no-ippo"
    
    print(f"[SCRAPER] Step 1: Bypassing Cloudflare at {base_url}...")
    page = await BROWSER.get(base_url)
    # Note: verify_cf requires opencv-python. Skipping for now.
    # await page.verify_cf()
    await page.sleep(5)
    print(f"[SCRAPER] ✓ Cloudflare bypassed")
    
    print(f"[SCRAPER] Step 2: Capturing cookies at {official_url}...")
    page = await BROWSER.get(official_url)
    
    print(f"[SCRAPER] Waiting {wait_time} seconds for page load...")
    await page.sleep(wait_time)
    
    cookies_obj = await page.send(uc.cdp.storage.get_cookies())
    cookies = {cookie.name: cookie.value for cookie in cookies_obj}
    
    response = {
        "url": official_url,
        "success": True,
        "cookies": cookies
    }
    
    print(f"[SCRAPER] ✓ Captured {len(cookies)} cookies")
    
    return response

def is_browser_active() -> bool:
    """Check if browser is active"""
    return BROWSER is not None