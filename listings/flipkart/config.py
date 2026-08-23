import os

PKG = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(PKG)
REPO = os.path.dirname(ROOT)

CATALOG = os.path.join(ROOT, 'products_catalog_clean.json')
SCHEMA = os.path.join(PKG, 'field_schema_clean.json')
MANIFEST = os.path.join(PKG, 'manifest.json')
RESULTS = os.path.join(PKG, 'results.json')
IMG_CACHE = os.path.join(PKG, '.imgcache')
SCREENSHOT_DIR = os.path.join(PKG, 'screenshots')
LOG_DIR = os.path.join(PKG, 'logs')

DEBUG_PORT = 9222
BASE_URL = 'https://seller.flipkart.com/index.html#dashboard/addListings/single?brand=Millennial%20Perfumer&vertical=perfume&context=CPUI'

ALREADY_LISTED = ['F001', 'F002', 'F003', 'F004', 'F005']

MAX_IMAGE_DIM = 1200
IMAGE_QUALITY = 82
WAIT_AFTER_CREATE = 8
WAIT_TAB_MOUNT = 2.5
WAIT_IMAGE_COMMIT = 6.5
WAIT_SAVE = 5

CONSTANTS = {
    'listing_status': 'Active',
    'minimum_order_quantity': '1',
    'service_profile': 'Seller',
    'procurement_type': 'express',
    'shipping_provider': 'Flipkart',
    'tax_code': 'GST_18',
    'country_of_origin': 'India',
    'fragrance_classification': 'Extrait De Parfum',
    'fragrance_segment': 'Budget',
    'ideal_for': None,
    'sales_package': '1 Perfume Bottle (50 ml)',
    'other_traits': 'Extrait De Parfum concentration (~35%), Long lasting',
}