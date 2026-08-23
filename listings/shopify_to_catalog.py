"""Build `products_catalog_clean.json` from the raw Shopify API export.

The Shopify GraphQL export (`shopify_api_raw.json`) stores each product's long
description in the `custom.product_description` metafield as Lexical rich-text
JSON.

Top/heart/base notes are NOT extracted here — descriptions are too varied to
parse reliably. Notes are filled separately (AI-assisted) and merged in via the
previous catalog or the `--notes` file. See `NOTES.md` for the workflow.

The script merges with the previous catalog so curated fields that have no
source in the export (`fragrance_family`, images, model_name overrides, notes)
are preserved.

Usage:
    python3 listings/shopify_to_catalog.py
    python3 listings/shopify_to_catalog.py --write
    python3 listings/shopify_to_catalog.py --write --notes notes_override.json
"""

import argparse
import json
import os
import re

PKG = os.path.dirname(os.path.abspath(__file__))
RAW = os.path.join(PKG, 'shopify_api_raw.json')
CATALOG = os.path.join(PKG, 'products_catalog_clean.json')
PHOTOS = '/Users/siddiqs_office/Documents/Business/millennial perfumer/Product Photos'

# --- static / config values (from flipkart.md + flipkart/config.py) ---------

STATIC = {
    'brand': 'Millennial Perfumer',
    'stock': 15,
    'procurement_sla': 2,
    'package_length_cm': 12,
    'package_breadth_cm': 10,
    'package_height_cm': 19,
    'package_weight_kg': 0.4,
    'hsn_code': '33029019',
    'tax_code': 'GST_18',
    'fragrance_classification': 'Extrait De Parfum',
    'country_of_origin': 'India',
    'manufacturer_details': ('Parfum Traders, No. 9/21, 1st floor, Sadiq Basha Nagar, '
                             '2nd Street, Virugambakkam, Chennai-600092, Contact: '
                             '+917904769823, millennialperfumer.cc@gmail.com'),
    'packer_details': ('Parfum Traders, No. 9/21, 1st floor, Sadiq Basha Nagar, '
                       '2nd Street, Virugambakkam, Chennai-600092, Contact: '
                       '+917904769823, millennialperfumer.cc@gmail.com'),
    'gas': 'No',
    'is_pocket_perfume': 'No',
    'max_shelf_life_months': 48,
    'min_order_quantity': 1,
    'fulfillment_by': 'SELLER',
    'procurement_type': 'EXPRESS',
    'shipping_provider': 'FLIPKART',
    'local_handling_fee': 0,
    'zonal_handling_fee': 0,
    'national_handling_fee': 0,
    'manufacturing_date': '2026-08-15',
    'fragrance_segment': 'Budget',
    'gift_pack': 'No',
}

KEY_FEATURES = ('✅ Rich, long-lasting Extrait de Parfum with a higher fragrance '
                'concentration (almost. 35%) for an intense, luxurious experience.')


# --- Lexical rich-text parsing ----------------------------------------------

def _walk_children(children, out):
    for c in children or []:
        if c.get('type') == 'text':
            out.append(c.get('value', ''))
        else:
            _walk_children(c.get('children'), out)
    return out


def _node_text(node):
    return ' '.join(_walk_children(node.get('children'), []))


def _plain_text(doc):
    paras = [_node_text(k) for k in (doc.get('children') or [])]
    return '\n'.join(p for p in paras if p.strip())


def _normalize(text):
    return re.sub(r'\s+', ' ', text).strip()


# --- note override (AI-filled) ----------------------------------------------

def load_notes(notes_path):
    """Load {sku: {'top_notes': str, 'heart_notes': str, 'base_notes': str}}.

    These are filled AI-assisted from the description (see NOTES.md) and merged
    over whatever the previous catalog already carried.
    """
    if not notes_path or not os.path.exists(notes_path):
        return {}
    with open(notes_path) as f:
        return json.load(f)


# --- title / model / gender -------------------------------------------------

def clean_title(title):
    return re.sub(r'\s+', ' ', title).strip()


def derive_model_name(title):
    t = clean_title(title)
    t = re.sub(r'^Millennial Perfumer\s*[™-]\s*', '', t)
    t = re.sub(r'\s*\|\s*.*$', '', t)
    return f'{t} | Long Lasting Fragrance |'


def derive_ideal_for(title):
    t = clean_title(title)
    low = t.lower()
    if 'women' in low and 'men' in low:
        return 'Men & Women'
    if 'women' in low:
        return 'Women'
    return 'Men'


def derive_sku_id(sku, title):
    t = clean_title(title)
    t = re.sub(r'^Millennial Perfumer\s*[™-]+\s*', '', t)
    t = re.sub(r'^[\s|–—-]+', '', t)
    name = re.split(r'\s*[|–—-]\s*', t)[0]
    name = re.sub(r'[^A-Za-z0-9]+', '', name).upper()
    return f'{sku}-{name}-50ML'


# --- image mapping ----------------------------------------------------------

def image_paths(sku):
    out = []
    for folder in os.listdir(PHOTOS):
        if folder.startswith(f'{sku} '):
            base = os.path.join(PHOTOS, folder)
            for i in range(1, 9):
                p = os.path.join(base, f'{i}.png')
                if os.path.exists(p):
                    out.append(p)
            break
    return out


# --- main -------------------------------------------------------------------

def load_raw():
    with open(RAW) as f:
        return json.load(f)


def load_previous_catalog():
    if not os.path.exists(CATALOG):
        return {}
    with open(CATALOG) as f:
        return {p['sku_code']: p for p in json.load(f)}


def build_product(node, prev, notes=None):
    sku = node['variants']['edges'][0]['node']['sku']
    variant = node['variants']['edges'][0]['node']
    title = node['title']
    notes = notes or {}

    desc_doc = None
    for mf in node['metafields']['edges']:
        if mf['node']['key'] == 'product_description' and mf['node'].get('value'):
            try:
                desc_doc = json.loads(mf['node']['value'])
            except (TypeError, ValueError):
                desc_doc = None
            break

    description = _plain_text(desc_doc) if desc_doc else (node.get('description') or '')
    note = notes.get(sku, {})

    old = prev.get(sku, {})
    mrp = variant.get('compareAtPrice')
    price = variant.get('price')
    images = old.get('images') or image_paths(sku)

    top = note.get('top_notes') or old.get('top_notes') or ''
    heart = note.get('heart_notes') or old.get('heart_notes') or ''
    base = note.get('base_notes') or old.get('base_notes') or ''

    keywords = (f'{clean_title(title)}, Millennial Perfumer, Extrait de Parfum, '
                f'Long lasting perfume, {top}, {base}')

    rec = dict(STATIC)
    rec.update({
        'sku_code': sku,
        'sku_id': derive_sku_id(sku, title),
        'model_name': old.get('model_name') or derive_model_name(title),
        'mrp': old.get('mrp') or 1299,
        'selling_price': old.get('selling_price') or 499,
        'quantity_value': 50,
        'quantity_unit': 'ml',
        'ideal_for': old.get('ideal_for') or derive_ideal_for(title),
        'fragrance_family': old.get('fragrance_family') or [],
        'top_notes': top,
        'heart_notes': heart,
        'base_notes': base,
        'key_features': KEY_FEATURES,
        'search_keywords': keywords,
        'description': _normalize(description),
        'images': images,
    })
    return rec


def main():
    ap = argparse.ArgumentParser(description='Build products_catalog_clean.json from Shopify export')
    ap.add_argument('--write', action='store_true', help='write the catalog; default is dry-run')
    ap.add_argument('--notes', help='path to AI-filled notes override JSON')
    args = ap.parse_args()

    raw = load_raw()
    prev = load_previous_catalog()
    notes = load_notes(args.notes)
    edges = raw['data']['products']['edges']

    records = []
    for e in edges:
        rec = build_product(e['node'], prev, notes)
        records.append(rec)

    records.sort(key=lambda r: r['sku_code'])
    missing_notes = [r['sku_code'] for r in records
                     if not r['top_notes'] or not r['heart_notes'] or not r['base_notes']]

    print(f'Products: {len(records)}')
    print(f'Using previous catalog: {len(prev)} records')
    print(f'Notes overrides loaded: {len(notes)}')
    print(f'Missing any note: {missing_notes or "none"}')

    if args.write:
        with open(CATALOG, 'w') as f:
            json.dump(records, f, indent=2, ensure_ascii=False)
            f.write('\n')
        print(f'Wrote {CATALOG}')
    else:
        print('Dry run (no write). Pass --write to save.')


if __name__ == '__main__':
    main()
