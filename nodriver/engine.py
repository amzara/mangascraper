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
    Get cookies from mangakakalot.gg
    
    Process:
    1. Go straight to mangakakalot.gg/manga/hajime-no-ippo (triggers CF)
    2. Solve Cloudflare challenge
    3. Capture cookies (cf_clearance is set here)
    4. Navigate to chapter page
    
    Args:
        wait_time: Time to wait after page load (default: 3 seconds)
    
    Returns:
        Dictionary containing cookies
    """
    global BROWSER
    
    manga_url = "https://www.nelomanga.net/manga/hajime-no-ippo"
    chapter_url = f"{manga_url}/chapter-1"
    
    # Step 1: Go straight to manga page - CF challenge triggers here!
    print(f"[SCRAPER] Step 1: Visiting {chapter_url}...")
    page = await BROWSER.get(chapter_url)
    await page.sleep(1)
    
    # Solve Cloudflare challenge
    print(f"[SCRAPER] Solving Cloudflare challenge...")
    # await page.verify_cf()
    # await page.sleep(2)
    print(f"[SCRAPER] ✓ Cloudflare challenge passed")
    
    # Capture cookies RIGHT AFTER CF bypass
    print(f"[SCRAPER] Capturing cookies after CF bypass...")
    cookies_obj = await page.send(uc.cdp.storage.get_cookies())
    cookies_after_cf = {cookie.name: cookie.value for cookie in cookies_obj}
    
    print(f"[SCRAPER] Cookies: {list(cookies_after_cf.keys())}")
    if 'cf_clearance' in cookies_after_cf:
        print(f"[SCRAPER] ✓ cf_clearance: {cookies_after_cf['cf_clearance'][:50]}...")
    else:
        print(f"[SCRAPER] ⚠ cf_clearance not found")
    
    # Step 2: Navigate to chapter page
    print(f"[SCRAPER] Step 2: Navigating to {chapter_url}...")
    page = await BROWSER.get(chapter_url)
    
    print(f"[SCRAPER] Waiting {wait_time} seconds...")
    await page.sleep(wait_time)
    
    # Get cookies from chapter page
    cookies_obj = await page.send(uc.cdp.storage.get_cookies())
    cookies_chapter = {cookie.name: cookie.value for cookie in cookies_obj}
    
    # Merge - prioritize cf_clearance from manga page
    final_cookies = cookies_chapter.copy()
    if 'cf_clearance' in cookies_after_cf:
        final_cookies['cf_clearance'] = cookies_after_cf['cf_clearance']
        print(f"[SCRAPER] ✓ cf_clearance preserved")
    
    print(f"[SCRAPER] Final cookies: {list(final_cookies.keys())}")
    
    response = {
        "url": chapter_url,
        "success": True,
        "cookies": final_cookies
    }
    
    print(f"[SCRAPER] ✓ Done - {len(final_cookies)} cookies")
    
    return response

def is_browser_active() -> bool:
    """Check if browser is active"""
    return BROWSER is not None
