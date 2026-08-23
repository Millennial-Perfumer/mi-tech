import hashlib
import os
import subprocess

from . import config


def _cache_key(path):
    st = os.stat(path)
    return hashlib.md5(f'{path}:{st.st_size}:{int(st.st_mtime)}'.encode()).hexdigest()[:16]


def compress_image(path, cache_dir=None):
    """Downscale + re-encode an image to a Flipkart-safe size (JPEG, <= ~350KB).

    Uses macOS `sips`. Returns the cached compressed path.
    """
    if not os.path.exists(path):
        raise FileNotFoundError(path)
    cache_dir = cache_dir or config.IMG_CACHE
    os.makedirs(cache_dir, exist_ok=True)
    out = os.path.join(cache_dir, _cache_key(path) + '.jpg')
    if os.path.exists(out):
        return out
    cmd = [
        'sips', '-Z', str(config.MAX_IMAGE_DIM),
        '-s', 'format', 'jpeg',
        '-s', 'formatOptions', str(config.IMAGE_QUALITY),
        path, '--out', out,
    ]
    subprocess.run(cmd, check=True, capture_output=True)
    return out


def prepare_images(product, cache_dir=None):
    """Return compressed paths for all 8 images of a product."""
    return [compress_image(p, cache_dir) for p in product['images'] if os.path.exists(p)]