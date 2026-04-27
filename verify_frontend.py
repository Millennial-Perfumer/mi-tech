import time
from playwright.sync_api import sync_playwright

def verify_frontend():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()

        # Go to the app
        try:
            page.goto("http://localhost:5173")
        except Exception as e:
            print(f"Error connecting to dev server: {e}")
            # Try to read dev_server.log to see if it's still starting
            with open("frontend/dev_server.log", "r") as f:
                print(f"Dev server log:\n{f.read()}")
            return

        # Bypass login by injecting token
        # Mock token payload: {"role": "admin", "exp": 2524608000} (year 2050)
        # Token header: {"alg": "HS256", "typ": "JWT"}
        # Base64 encoded: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjogImFkbWluIiwgImV4cCI6IDI1MjQ2MDgwMDB9.mock_signature
        mock_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjogImFkbWluIiwgImV4cCI6IDI1MjQ2MDgwMDB9.mock_signature"

        page.evaluate(f"localStorage.setItem('token', '{mock_token}')")
        page.reload()

        # Navigate to Social Media tab
        page.click("text=Social Media")
        time.sleep(2) # Wait for tab to load

        # Open Compose modal
        page.click("text=Compose")
        time.sleep(1)

        # Take screenshot of the improved modal
        page.screenshot(path="improved_composer_modal.png")
        print("Screenshot saved to improved_composer_modal.png")

        browser.close()

if __name__ == "__main__":
    verify_frontend()
