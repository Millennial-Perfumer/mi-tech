import json
import os
from datetime import datetime, timezone

STATUS_RUNNING = 'running'
STATUS_COMPLETED = 'completed'
STATUS_FAILED = 'failed'
STATUS_REVIEW = 'needs_review'


class Tracker:
    def __init__(self, manifest, results_path):
        self.manifest = manifest
        self.results_path = results_path
        self.results = {}
        if os.path.exists(results_path):
            try:
                self.results = json.load(open(results_path))
            except Exception:
                self.results = {}

    def pending(self, include_review=False, include_failed=False):
        # 'running' is treated as pending so a crashed process can be resumed.
        out = []
        for sku in self.manifest:
            st = (self.results.get(sku) or {}).get('status')
            if st == STATUS_COMPLETED:
                continue
            if st == STATUS_REVIEW and not include_review:
                continue
            if st == STATUS_FAILED and not include_failed:
                continue
            out.append(sku)
        return out

    def mark(self, sku, step=None, **fields):
        rec = self.results.setdefault(sku, {})
        if step:
            rec['step'] = step
        rec.update(fields)
        rec['updated'] = datetime.now(timezone.utc).isoformat()
        self._save()

    def _save(self):
        tmp = self.results_path + '.tmp'
        with open(tmp, 'w') as f:
            json.dump(self.results, f, indent=2, ensure_ascii=False)
        os.replace(tmp, self.results_path)

    def summary(self):
        counts = {s: 0 for s in (STATUS_RUNNING, STATUS_COMPLETED, STATUS_FAILED, STATUS_REVIEW)}
        counts['pending'] = 0
        for sku in self.manifest:
            st = (self.results.get(sku) or {}).get('status', 'pending')
            counts[st] = counts.get(st, 0) + 1
        return counts