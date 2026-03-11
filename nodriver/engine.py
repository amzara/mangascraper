import nodriver as uc
from typing import Dict, Any
import asyncio
import os

try:
    import cv2
    import numpy as np
    CV2_AVAILABLE = True
except ImportError:
    CV2_AVAILABLE = False
    print("[WARNING] OpenCV not available, checkbox detection disabled")

BROWSER = None

# Stealth Chrome flags
STEALTH_ARGS = [
    '--no-sandbox',
    '--disable-setuid-sandbox',
    '--disable-dev-shm-usage',
    '--disable-accelerated-2d-canvas',
    '--no-first-run',
    '--no-zygote',
    '--disable-gpu',
    '--disable-features=IsolateOrigins,site-per-process',
    '--disable-blink-features=AutomationControlled',
    '--window-size=1920,1080',
    '--start-maximized'
]

async def start_browser():
    """Initialize browser on startup"""
    global BROWSER
    try:
        browser_config_args = {
            "headless": False,
            "browser_args": STEALTH_ARGS,
            "lang": "en-US",
            "no_sandbox": True
        }
        BROWSER = await uc.start(**browser_config_args)
        print(f"✓ Browser started: {BROWSER.info.get('User-Agent')}")
    except Exception as e:
        print(f"✗ Failed to start browser: {str(e)}")
        BROWSER = None

async def stop_browser():
    """Close browser on shutdown"""
    global BROWSER
    if BROWSER:
        try:
            BROWSER.stop()
            print("✓ Browser stopped")
        except Exception as e:
            print(f"✗ Error stopping browser: {str(e)}")
        finally:
            BROWSER = None

async def click_checkbox_at(page, x, y):
    """Click at specific coordinates"""
    print(f"[SCRAPER] Clicking at ({x}, {y})")
    await page.send(
        uc.cdp.input_.dispatch_mouse_event(
            "mousePressed",
            x=x,
            y=y,
            button=uc.cdp.input_.MouseButton("left"),
            click_count=1,
        )
    )
    await page.send(
        uc.cdp.input_.dispatch_mouse_event(
            "mouseReleased",
            x=x,
            y=y,
            button=uc.cdp.input_.MouseButton("left"),
            click_count=1,
        )
    )

async def find_and_click_checkbox(page):
    """Find CF checkbox by looking for specific elements or patterns"""
    if not CV2_AVAILABLE:
        print("[SCRAPER] OpenCV not available, using fallback click")
        # Just click center screen
        await click_checkbox_at(page, 960, 490)
        return True
        
    try:
        await page.save_screenshot("screen.jpg")
        await asyncio.sleep(0.5)
        
        im = cv2.imread("screen.jpg")
        if im is None:
            return False
            
        im_gray = cv2.cvtColor(im, cv2.COLOR_BGR2GRAY)
        height, width = im_gray.shape
        
        # CF checkbox is typically a white/light square in the center
        # Look for square-ish white regions in center area
        center_x, center_y = width // 2, height // 2
        
        # Define search area (center 600x400 region)
        x1 = max(0, center_x - 300)
        y1 = max(0, center_y - 200)
        x2 = min(width, center_x + 300)
        y2 = min(height, center_y + 200)
        
        roi = im_gray[y1:y2, x1:x2]
        
        # Template: square checkbox (CF uses ~48x48 checkbox)
        # Try multiple sizes
        for size in [40, 48, 56]:
            template = np.zeros((size, size), dtype=np.uint8)
            cv2.rectangle(template, (0, 0), (size-1, size-1), 255, 2)
            
            match = cv2.matchTemplate(roi, template, cv2.TM_CCOEFF_NORMED)
            (_, max_v, _, max_l) = cv2.minMaxLoc(match)
            
            if max_v > 0.4:
                cx = x1 + max_l[0] + size // 2
                cy = y1 + max_l[1] + size // 2
                print(f"[SCRAPER] Found checkbox at ({cx}, {cy}) with score {max_v:.2f}")
                await click_checkbox_at(page, cx, cy)
                return True
                
        # Fallback: try clicking common CF checkbox location
        # CF usually places it at center screen
        fallback_x = center_x
        fallback_y = center_y - 50
        print(f"[SCRAPER] Using fallback click at ({fallback_x}, {fallback_y})")
        await click_checkbox_at(page, fallback_x, fallback_y)
        return True
        
    except Exception as e:
        print(f"[SCRAPER] Error finding checkbox: {e}")
        return False

async def get_cookies(wait_time: int = 3) -> Dict[str, Any]:
    """Get cookies from mangakakalot.gg with CF bypass"""
    global BROWSER
    
    if BROWSER is None:
        return {"success": False, "error": "Browser not started"}
    
    chapter_url = "https://www.nelomanga.net/manga/hajime-no-ippo/chapter-1"
    
    try:
        # Step 1: Visit page
        print(f"[SCRAPER] Visiting {chapter_url}...")
        page = await BROWSER.get(chapter_url, new_tab=True)
        
        # Step 2: Wait for CF challenge to appear (longer wait)
        print(f"[SCRAPER] Waiting for CF challenge (15s)...")
        await page.sleep(15)
        
        # Step 3: Look for and click checkbox
        print(f"[SCRAPER] Looking for CF checkbox...")
        clicked = await find_and_click_checkbox(page)
        
        if clicked:
            # Step 4: Wait longer for CF to process (CRITICAL!)
            print(f"[SCRAPER] Waiting for CF verification (20s)...")
            await page.sleep(20)
        
        # Step 5: Get cookies
        print(f"[SCRAPER] Getting cookies...")
        cookie_obj_list = await BROWSER.cookies.get_all()
        
        cookies_dict = {}
        for cookie in cookie_obj_list:
            cookies_dict[cookie.name] = cookie.value
        
        print(f"[SCRAPER] Cookies: {list(cookies_dict.keys())}")
        if 'cf_clearance' in cookies_dict:
            print(f"[SCRAPER] ✓ cf_clearance found: {cookies_dict['cf_clearance'][:50]}...")
        else:
            print(f"[SCRAPER] ⚠ cf_clearance not found")
        
        # Cleanup
        for f in ["screen.jpg", "screen2.jpg"]:
            try:
                os.remove(f)
            except:
                pass
        
        return {
            "url": chapter_url,
            "success": True,
            "cookies": cookies_dict,
            "userAgent": BROWSER.info.get('User-Agent')
        }
        
    except Exception as e:
        print(f"[SCRAPER] ✗ Error: {str(e)}")
        import traceback
        traceback.print_exc()
        return {
            "success": False,
            "error": str(e)
        }

def is_browser_active() -> bool:
    """Check if browser is active"""
    return BROWSER is not None
