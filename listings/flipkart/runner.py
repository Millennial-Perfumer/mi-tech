import argparse
import asyncio
import json
import os
import sys
import traceback

# Support both supported ways of starting the runner:
#   python3 -m listings.flipkart.runner ...
#   python3 listings/flipkart/runner.py ...
# The latter has no package context by default, so expose the repository root
# and import the module through its package name.
if __package__ in (None, ''):
    sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
    from listings.flipkart import browser, config, tracker, worker
else:
    from . import browser, config, tracker, worker


def load_catalog():
    with open(config.CATALOG) as f:
        return json.load(f)


def build_manifest(catalog):
    return [p['sku_code'] for p in catalog if p['sku_code'] not in config.ALREADY_LISTED]


def write_manifest(path, skus):
    with open(path, 'w') as f:
        json.dump(skus, f, indent=2)


class StatusReporter:
    def __init__(self, queue):
        self.queue = queue
        self.lock = asyncio.Lock()

    async def set(self, sku, step=None, **fields):
        async with self.lock:
            self.queue.mark(sku, step=step, **fields)


async def run(args):
    if not os.path.exists(config.MANIFEST):
        write_manifest(config.MANIFEST, build_manifest(load_catalog()))
    manifest = json.load(open(config.MANIFEST))
    catalog = {p['sku_code']: p for p in load_catalog()}

    q = tracker.Tracker(manifest, config.RESULTS)
    pending = q.pending(include_review=args.retry_failed, include_failed=args.retry_failed)
    selected_skus = set(args.sku + args.skus)
    if selected_skus:
        pending = [s for s in pending if s in selected_skus]
    if args.limit:
        pending = pending[:args.limit]

    print(f'Manifest: {len(manifest)} products | Pending: {len(pending)} | Workers: {args.workers}')

    if args.dry_run:
        for sku in pending:
            p = catalog.get(sku)
            imgs = sum(1 for i in (p or {}).get('images', []) if os.path.exists(i))
            print(f'  {sku}: {p.get("model_name") if p else "?"} | images={imgs}/8')
        print('Dry run complete. Nothing was touched in Chrome.')
        return

    os.makedirs(config.IMG_CACHE, exist_ok=True)
    os.makedirs(config.SCREENSHOT_DIR, exist_ok=True)
    os.makedirs(config.LOG_DIR, exist_ok=True)

    # Confirm Chrome + seller session
    try:
        browser.browser_ws_url()
    except RuntimeError as e:
        print(f'ERROR: {e}')
        sys.exit(1)

    jobs = asyncio.Queue()
    for sku in pending:
        jobs.put_nowait(sku)
    reporter = StatusReporter(q)

    async def worker_loop(name):
        while not jobs.empty():
            try:
                sku = await asyncio.wait_for(jobs.get(), timeout=5)
            except asyncio.TimeoutError:
                break
            product = catalog.get(sku)
            if product is None:
                await reporter.set(sku, status='failed', step='setup', error='not in catalog')
                jobs.task_done()
                continue
            await reporter.set(sku, status=tracker.STATUS_RUNNING, step='start')
            print(f'[{name}] START {sku}', flush=True)
            try:
                result = await asyncio.wait_for(
                    worker.process_one(sku, product, config.IMG_CACHE,
                                       config.SCREENSHOT_DIR, reporter),
                    timeout=args.job_timeout)
            except asyncio.TimeoutError:
                result = {'status': tracker.STATUS_FAILED, 'step': 'timeout',
                          'error': f'timed out after {args.job_timeout}s'}
            except Exception as e:
                tb = traceback.format_exc()
                result = {'status': tracker.STATUS_FAILED, 'step': 'unknown',
                          'error': str(e), 'traceback': tb}
                print(tb, flush=True)
            await reporter.set(sku, status=result['status'], result=result)
            tag = result['status']
            err = result.get('error') or (', '.join(result.get('errors') or [])[:120])
            print(f'[{name}] DONE {sku}: {tag} {err}', flush=True)
            jobs.task_done()

    await asyncio.gather(*[worker_loop(i + 1) for i in range(args.workers)])

    summary = q.summary()
    print('\n=== SUMMARY ===')
    for k, v in summary.items():
        print(f'  {k}: {v}')


def main():
    ap = argparse.ArgumentParser(description='Flipkart listing runner (saves drafts, no QC)')
    ap.add_argument('--workers', type=int, default=2, help='parallel worker tabs')
    ap.add_argument('--limit', type=int, default=None, help='max products this run')
    ap.add_argument('--sku', action='append', default=[], help='run one SKU; repeat for multiple SKUs')
    ap.add_argument('--skus', nargs='*', default=[], help='run a space-separated list of SKUs')
    ap.add_argument('--retry-failed', action='store_true', help='re-run failed/needs_review')
    ap.add_argument('--dry-run', action='store_true', help='print plan, touch nothing')
    ap.add_argument('--job-timeout', type=int, default=300, help='max seconds per product')
    args = ap.parse_args()
    asyncio.run(run(args))


if __name__ == '__main__':
    main()
