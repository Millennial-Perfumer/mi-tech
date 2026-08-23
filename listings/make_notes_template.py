"""Generate an empty notes override template for SKUs missing notes.

The template is intended to be filled AI-assisted from each product's
description (see NOTES.md), then merged by `shopify_to_catalog.py --notes`.

Usage:
    python3 listings/make_notes_template.py [--out notes_override.json]
"""

import argparse
import json
import os

from listings import shopify_to_catalog as sc

PKG = os.path.dirname(os.path.abspath(__file__))


def main():
    ap = argparse.ArgumentParser(description='Generate empty notes override template')
    ap.add_argument('--out', default=os.path.join(PKG, 'notes_override.json'))
    args = ap.parse_args()

    raw = sc.load_raw()
    prev = sc.load_previous_catalog()
    edges = raw['data']['products']['edges']

    template = {}
    for e in edges:
        sku = e['node']['variants']['edges'][0]['node']['sku']
        old = prev.get(sku, {})
        has_all = all(old.get(k) for k in ('top_notes', 'heart_notes', 'base_notes'))
        if not has_all:
            template[sku] = {'top_notes': '', 'heart_notes': '', 'base_notes': ''}

    with open(args.out, 'w') as f:
        json.dump(template, f, indent=2, ensure_ascii=False)
        f.write('\n')
    print(f'Wrote template for {len(template)} SKUs -> {args.out}')


if __name__ == '__main__':
    main()
