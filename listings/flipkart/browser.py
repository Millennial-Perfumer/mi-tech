import base64
import asyncio
import json
import urllib.request

import websockets

from . import config


def browser_ws_url():
    try:
        v = json.loads(urllib.request.urlopen(
            f'http://localhost:{config.DEBUG_PORT}/json/version', timeout=3).read())
        return v['webSocketDebuggerUrl']
    except Exception as e:
        raise RuntimeError(
            f'Chrome DevTools not reachable on port {config.DEBUG_PORT}. '
            f'Start Chrome with --remote-debugging-port={config.DEBUG_PORT} '
            f'and log in to seller.flipkart.com. ({e})')


async def create_tab(url='about:blank'):
    bw = browser_ws_url()
    async with websockets.connect(bw, max_size=100_000_000) as ws:
        await ws.send(json.dumps({'id': 1, 'method': 'Target.createTarget', 'params': {'url': url}}))
        while True:
            res = json.loads(await asyncio.wait_for(ws.recv(), timeout=15))
            if res.get('id') == 1:
                if 'error' in res:
                    raise RuntimeError(res['error'])
                return res['result']['targetId']


def tab_ws_url(target_id):
    tabs = json.loads(urllib.request.urlopen(
        f'http://localhost:{config.DEBUG_PORT}/json/list', timeout=3).read())
    t = next((x for x in tabs if x.get('id') == target_id), None)
    if not t:
        raise RuntimeError(f'target {target_id} not found')
    return t['webSocketDebuggerUrl']


async def close_tab(target_id):
    bw = browser_ws_url()
    try:
        async with websockets.connect(bw, max_size=100_000_000) as ws:
            await ws.send(json.dumps({'id': 2, 'method': 'Target.closeTarget', 'params': {'targetId': target_id}}))
            try:
                while True:
                    res = json.loads(await asyncio.wait_for(ws.recv(), timeout=5))
                    if res.get('id') == 2:
                        break
            except asyncio.TimeoutError:
                pass
    except Exception:
        pass


class Page:
    def __init__(self, ws):
        self.ws = ws
        self._mid = 1000

    async def send(self, method, params=None, timeout=120):
        self._mid += 1
        mid = self._mid
        await self.ws.send(json.dumps({'id': mid, 'method': method, 'params': params or {}}))
        while True:
            res = json.loads(await asyncio.wait_for(self.ws.recv(), timeout=timeout))
            if res.get('id') == mid:
                return res

    async def eval(self, expr, timeout=120):
        res = await self.send('Runtime.evaluate', {
            'expression': expr, 'returnByValue': True, 'awaitPromise': True}, timeout=timeout)
        r = res.get('result', {})
        if 'exceptionDetails' in r:
            return {'__error': r['exceptionDetails'].get('exception', {}).get('description', '')}
        return r.get('result', {}).get('value')

    async def navigate(self, url, wait=8):
        await self.send('Page.navigate', {'url': url})
        await asyncio.sleep(wait)

    async def screenshot(self, path):
        try:
            res = await self.send('Page.captureScreenshot', {'format': 'png'}, timeout=30)
            data = res.get('result', {}).get('data')
            if data:
                with open(path, 'wb') as f:
                    f.write(base64.b64decode(data))
                return True
        except Exception:
            pass
        return False

    async def close(self):
        try:
            await self.ws.close()
        except Exception:
            pass


async def open_page():
    target_id = await create_tab()
    ws = await websockets.connect(tab_ws_url(target_id), max_size=100_000_000)
    page = Page(ws)
    await page.send('Page.enable')
    await page.send('DOM.enable')
    return target_id, page