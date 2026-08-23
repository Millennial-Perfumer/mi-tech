"""Python orchestration of the Flipkart single-listing form.

JavaScript that runs in the page lives in `form_scripts.py`; this module only
drives the CDP page: tab switching, image upload, validation and draft save.
"""

import asyncio

from . import config
from . import form_scripts as fs


async def switch_tab(page, part, wait=config.WAIT_TAB_MOUNT):
    await page.eval(fs.JS_CLICK_TAB.format(part=part))
    await asyncio.sleep(wait)


async def upload_images(page, image_paths, wait=config.WAIT_IMAGE_COMMIT):
    results = []
    for idx, img in enumerate(image_paths):
        opened = await page.eval(fs.upload_icon(idx))
        if opened is not True:
            results.append(opened)
            continue
        await asyncio.sleep(1.8)
        doc = await page.send('DOM.getDocument')
        root_id = doc['result']['root']['nodeId']
        fn = await page.send('DOM.querySelector', {'nodeId': root_id, 'selector': 'input[type="file"]'})
        node_id = fn.get('result', {}).get('nodeId')
        if not node_id:
            results.append('no-file-input')
            continue
        await page.send('DOM.setFileInputFiles', {'files': [img], 'nodeId': node_id})
        await asyncio.sleep(wait)
        results.append('committed')
    return results


async def read_errors(page):
    return await page.eval(fs.JS_READ_ERRORS) or []


async def close_modal(page):
    await page.eval(fs.JS_CLOSE_MODALS)
    await asyncio.sleep(0.8)


async def save_draft(page, wait=config.WAIT_SAVE):
    """Click Save & Go Back; confirm modal if needed; return outcome dict."""
    outcome = {}
    clicked = await page.eval(fs.JS_SAVE_CLICK)
    outcome['save_clicked'] = clicked
    await asyncio.sleep(wait)

    url = await page.eval('window.location.href')
    outcome['url'] = url
    outcome['saved'] = bool(url and ('listingsInProgress' in url or 'addListings/single' not in url))

    if not outcome['saved']:
        confirmed = await page.eval(fs.JS_SAVE_CONFIRM)
        outcome['modal_confirm'] = confirmed
        await asyncio.sleep(wait)
        url2 = await page.eval('window.location.href')
        outcome['url_after'] = url2
        outcome['saved'] = bool(url2 and ('listingsInProgress' in url2 or 'addListings/single' not in url2))
        outcome['errors'] = await read_errors(page)
    return outcome