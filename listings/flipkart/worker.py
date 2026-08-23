import asyncio
import os

from . import config
from . import browser
from . import form_scripts as fs
from . import image_prep
from . import listing_form as lf

STEP_CREATE = 'create'
STEP_FILL = 'fill'
STEP_IMAGES = 'images'
STEP_SAVE = 'save'


JS_CREATE_CLICK = '''
    (() => {
        const b = Array.from(document.querySelectorAll('button')).find(x =>
            x.innerText.trim().toLowerCase() === 'create new listing');
        if (b) { b.click(); return true; }
        return false;
    })()
    '''

JS_CLOSE_BLOCKING = '''
    (() => {
        const m = Array.from(document.querySelectorAll('.styles__DialogElement-sc-1pm7qho-2, [role="dialog"]'));
        m.forEach(x => {
            const b = Array.from(x.querySelectorAll('button')).find(y => y.innerText.trim().toLowerCase() === 'close');
            if (b) b.click();
        });
    })()
    '''


async def _create_listing(page):
    """Navigate + click Create new listing; verify the fresh form really mounted."""
    for _ in range(3):
        await page.navigate(config.BASE_URL, wait=8)
        clicked = False
        for _ in range(5):
            clicked = await page.eval(JS_CREATE_CLICK)
            if clicked:
                break
            await asyncio.sleep(4)
        await asyncio.sleep(config.WAIT_AFTER_CREATE)
        url = await page.eval('window.location.href') or ''
        tabs = await page.eval(
            "Array.from(document.querySelectorAll('button[id*=\"tabitem\"]')).length") or 0
        if 'requestId=' in url and tabs:
            return True
        # A leftover overlay may be blocking the form; clear it and retry the cycle.
        await page.eval(JS_CLOSE_BLOCKING)
        await asyncio.sleep(2)
    return False


async def run_job(sku, product, page, screenshot_dir, status):
    step = STEP_CREATE
    log = print
    try:
        created = await _create_listing(page)
        log(f'[{sku}] create ok: {created}', flush=True)
        await status.set(sku, 'create_done', created=created)

        step = STEP_FILL
        log(f'[{sku}] fill selling', flush=True)
        await lf.switch_tab(page, 'selling_info')
        r1 = await page.eval(fs.build_fill_selling(product))
        await asyncio.sleep(3)
        log(f'[{sku}] fill product', flush=True)
        await lf.switch_tab(page, 'product_info')
        r2 = await page.eval(fs.build_fill_product(product))
        await asyncio.sleep(3)
        log(f'[{sku}] fill optional', flush=True)
        await lf.switch_tab(page, 'optional_product_info')
        r3 = await page.eval(fs.build_fill_optional(product))
        await asyncio.sleep(3)
        await status.set(sku, 'fill_done', selling=r1, product=r2, optional=r3)
        log(f'[{sku}] fill done', flush=True)

        step = STEP_IMAGES
        log(f'[{sku}] images ({len(product.get("images", []))})', flush=True)
        images = image_prep.prepare_images(product)
        await lf.switch_tab(page, 'product_images')
        imgs = await lf.upload_images(page, images)
        await asyncio.sleep(3)
        await status.set(sku, 'images_done', images=imgs)
        log(f'[{sku}] images done: {imgs}', flush=True)

        step = STEP_SAVE
        log(f'[{sku}] save draft', flush=True)
        outcome = await lf.save_draft(page)
        await status.set(sku, 'save_done', outcome=outcome)
        log(f'[{sku}] save outcome saved={outcome.get("saved")}', flush=True)

        if outcome.get('saved'):
            return {'status': 'completed', 'saved': True, 'step': step, 'outcome': outcome}
        errors = outcome.get('errors') or []
        return {'status': 'needs_review', 'saved': False, 'step': step,
                'outcome': outcome, 'errors': errors}

    except Exception as e:
        try:
            shot = os.path.join(screenshot_dir, f'{sku}_fail.png')
            await page.screenshot(shot)
            shot = shot
        except Exception:
            shot = None
        return {'status': 'failed', 'saved': False, 'step': step, 'error': str(e), 'screenshot': shot}


async def process_one(sku, product, cache_dir, screenshot_dir, status):
    target_id = None
    page = None
    try:
        target_id, page = await browser.open_page()
        return await run_job(sku, product, page, screenshot_dir, status)
    finally:
        if page is not None:
            await page.close()
        if target_id is not None:
            await browser.close_tab(target_id)