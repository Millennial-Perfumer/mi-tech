"""Injected JavaScript used by the listing automation.

All JS that runs inside the Flipkart page lives here, kept separate from the
Python orchestration in `listing_form.py`.
"""

import json

from . import config

JS_COMMON = r'''
function triggerReact(inp, val) {
    if (!inp) return false;
    inp.focus();
    const proto = inp.tagName === 'TEXTAREA' ? HTMLTextAreaElement.prototype : HTMLInputElement.prototype;
    const desc = Object.getOwnPropertyDescriptor(proto, 'value');
    if (desc && desc.set) desc.set.call(inp, String(val));
    else inp.value = String(val);
    ['input', 'change', 'blur', 'focusout'].forEach(e => inp.dispatchEvent(new Event(e, { bubbles: true })));
    return true;
}
function clickRadio(fullId) {
    const el = document.getElementById(fullId);
    if (el) { el.click(); el.dispatchEvent(new Event('change', { bubbles: true })); return true; }
    return false;
}
function clickTab(part) {
    const b = Array.from(document.querySelectorAll('button[id*="tabitem"]')).find(x => x.id.includes('tab_' + part));
    if (b) { b.click(); return true; }
    return false;
}
function panel() {
    return document.querySelector('[role="tabpanel"]:not([aria-hidden="true"])')
        || document.querySelector('[class*="tabpanel"]:not([aria-hidden="true"])')
        || document.body;
}
function setText(attrName, val) {
    const root = panel();
    const allEls = Array.from(root.querySelectorAll('input:not([type="radio"]):not([type="checkbox"]):not([type="hidden"]), textarea'));
    for (const el of allEls) {
        const k = Object.keys(el).find(x => x.startsWith('__reactFiber'));
        if (!k) continue;
        let c = el[k]; let matched = false;
        while (c) { const d = c.memoizedProps && c.memoizedProps.definition; if (d && d.attributeName === attrName) { matched = true; break; } c = c.return; }
        if (!matched) continue;
        if (!(el.offsetParent !== null || el.getBoundingClientRect().width > 0)) continue;
        if ((el.getAttribute('placeholder') || '').toLowerCase() === 'select') return false;
        triggerReact(el, val);
        return true;
    }
    return false;
}
function setTag(attrName, val) {
    const root = panel();
    const allEls = Array.from(root.querySelectorAll('input[type="text"], input:not([type]), textarea'));
    for (const el of allEls) {
        const k = Object.keys(el).find(x => x.startsWith('__reactFiber'));
        if (!k) continue;
        let c = el[k]; let matched = false;
        while (c) { const d = c.memoizedProps && c.memoizedProps.definition; if (d && d.attributeName === attrName) { matched = true; break; } c = c.return; }
        if (!matched) continue;
        if (!(el.offsetParent !== null || el.getBoundingClientRect().width > 0)) continue;
        el.focus();
        const proto = el.tagName === 'TEXTAREA' ? HTMLTextAreaElement.prototype : HTMLInputElement.prototype;
        const desc = Object.getOwnPropertyDescriptor(proto, 'value');
        if (desc && desc.set) desc.set.call(el, String(val));
        else el.value = String(val);
        el.dispatchEvent(new Event('input', { bubbles: true }));
        el.dispatchEvent(new Event('change', { bubbles: true }));
        el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', code: 'Enter', keyCode: 13, which: 13, bubbles: true }));
        el.dispatchEvent(new KeyboardEvent('keypress', { key: 'Enter', keyCode: 13, which: 13, bubbles: true }));
        el.dispatchEvent(new KeyboardEvent('keyup', { key: 'Enter', keyCode: 13, which: 13, bubbles: true }));
        el.blur();
        return true;
    }
    return false;
}
'''


def _radio(attr, display):
    return f"clickRadio('checkMarkOption_{attr}_0_value_{display}')"


def build_fill_selling(product):
    c = config.CONSTANTS
    return f'''
    (() => {{
        const data = {json.dumps(product)};
        {JS_COMMON}
        const r = {{}};
        r.sku_id = setText('sku_id', data.sku_id);
        r.listing_status = {_radio('listing_status', c['listing_status'])};
        r.mrp = setText('mrp', data.mrp);
        r.flipkart_selling_price = setText('flipkart_selling_price', data.selling_price);
        r.minimum_order_quantity = {_radio('minimum_order_quantity', c['minimum_order_quantity'])};
        r.service_profile = {_radio('service_profile', c['service_profile'])};
        r.procurement_type = {_radio('procurement_type', c['procurement_type'])};
        r.shipping_days = setText('shipping_days', data.procurement_sla);
        r.stock_size = setText('stock_size', data.stock);
        r.shipping_provider = {_radio('shipping_provider', c['shipping_provider'])};
        r.local_shipping_fee = setText('local_shipping_fee_from_buyer', data.local_handling_fee);
        r.zonal_shipping_fee = setText('zonal_shipping_fee_from_buyer', data.zonal_handling_fee);
        r.national_shipping_fee = setText('national_shipping_fee_from_buyer', data.national_handling_fee);
        r.length = setText('length', data.package_length_cm);
        r.breadth = setText('breadth', data.package_breadth_cm);
        r.height = setText('height', data.package_height_cm);
        r.weight = setText('weight', data.package_weight_kg);
        r.hsn = setText('hsn', data.hsn_code);
        r.tax_code = {_radio('tax_code', c['tax_code'])};
        r.country_of_origin = {_radio('country_of_origin', c['country_of_origin'])};
        r.manufacturer_details = setText('manufacturer_details', data.manufacturer_details);
        r.packer_details = setText('packer_details', data.packer_details);
        r.earliest_mfg_date = setText('earliest_mfg_date', data.manufacturing_date);
        r.shelf_life = setText('shelf_life', data.max_shelf_life_months);
        return r;
    }})()
    '''


def build_fill_product(product):
    c = config.CONSTANTS
    family = product.get('fragrance_family') or []
    return f'''
    (() => {{
        const data = {json.dumps(product)};
        {JS_COMMON}
        const r = {{}};
        r.model_name = setText('model_name', data.model_name);
        r.ideal_for = {_radio('ideal_for', product['ideal_for'])};
        r.quantity = setText('quantity', data.quantity_value);
        r.quantity_unit = clickRadio('checkMarkOption_quantity_0_qualifier_ml');
        r.fragrance_classification = {_radio('fragrance_classification', c['fragrance_classification'])};
        {json.dumps(family)}.forEach(fam => {{
            const cb = document.getElementById(fam);
            if (cb && !cb.checked) {{ cb.click(); cb.dispatchEvent(new Event('change', {{ bubbles: true }})); }}
        }});
        return r;
    }})()
    '''


def build_fill_optional(product):
    c = config.CONSTANTS
    radios = {
        'fragrance_segment': c['fragrance_segment'],
        'anti_perspirant': 'No',
        'limited_edition': 'No',
        'is_pocket_perfume': 'No',
        'gas': 'No',
        'gift_pack': 'No',
        'organic': 'No',
    }
    lines = '\n'.join(f"r.{attr} = {_radio(attr, disp)};" for attr, disp in radios.items())
    return f'''
    (() => {{
        const data = {json.dumps(product)};
        {JS_COMMON}
        const r = {{}};
        {lines}
        r.max_shelf_life = setText('max_shelf_life', data.max_shelf_life_months);
        r.max_shelf_life_qualifier = clickRadio('checkMarkOption_max_shelf_life_0_qualifier_Months');
        r.sales_package = setTag('sales_package', {json.dumps(c['sales_package'])});
        r.model_number = setText('model_number', data.sku_code);
        r.description = setText('description', data.description);
        r.keywords = setTag('keywords', data.search_keywords);
        r.key_features = setTag('key_features', data.key_features);
        r.top_note = setTag('top_note', data.top_notes);
        r.heart_note = setTag('heart_note', data.heart_notes);
        r.base_note = setTag('base_note', data.base_notes);
        r.other_traits = setTag('other_traits', {json.dumps(c['other_traits'])});
        return r;
    }})()
    '''


JS_CLICK_TAB = (
    "Array.from(document.querySelectorAll('button[id*=\"tabitem\"]'))"
    ".find(b => b.id.includes('tab_{part}'))?.click(); 'ok'"
)


def upload_icon(idx):
    return f'''
    (() => {{
        const t = document.getElementById('thumbnail_{idx}');
        if (!t) return 'no-thumb';
        const icon = t.querySelector('.fa-upload');
        if (!icon) return 'already-has-image';
        icon.click();
        return true;
    }})()
    '''


JS_READ_ERRORS = '''
    (() => {
        const seen = new Set(); const out = [];
        document.querySelectorAll('[class*="error" i],[class*="Error"],[role="alert"]').forEach(el => {
            const t = (el.innerText||'').replace(/\\s+/g,' ').trim();
            if (t && t.length < 200 && !seen.has(t)) { seen.add(t); out.push(t); }
        });
        return out;
    })()
    '''


JS_CLOSE_MODALS = '''
    (() => {
        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', code: 'Escape', bubbles: true }));
        document.body.click();
    })()
    '''


JS_SAVE_CLICK = '''
    (() => {
        const b = Array.from(document.querySelectorAll('button')).find(x => /save & go back/i.test(x.innerText||''));
        if (b && !b.disabled) { b.click(); return true; }
        return false;
    })()
    '''


JS_SAVE_CONFIRM = '''
    (() => {
        const modal = Array.from(document.querySelectorAll('[role="dialog"],[class*="modal" i],[class*="Modal"]'))
            .find(x => x.offsetParent !== null || x.getBoundingClientRect().width > 0);
        if (!modal) return false;
        const btn = Array.from(modal.querySelectorAll('button')).find(b =>
            /yes|save|proceed|confirm|continue|ok/i.test((b.innerText||'').trim())
            && !/cancel|close|view issues|issues/i.test((b.innerText||'').trim()));
        if (btn) { btn.click(); return (btn.innerText||'').trim(); }
        return false;
    })()
    '''