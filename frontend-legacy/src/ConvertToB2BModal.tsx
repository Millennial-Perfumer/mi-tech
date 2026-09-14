import React, { useState, useEffect } from 'react';
import { useToast } from './ToastContext';
import { API_BASE } from './api';

const GST_STATE_MAP: Record<string, string> = {
  "01": "Jammu and Kashmir",
  "02": "Himachal Pradesh",
  "03": "Punjab",
  "04": "Chandigarh",
  "05": "Uttarakhand",
  "06": "Haryana",
  "07": "Delhi",
  "08": "Rajasthan",
  "09": "Uttar Pradesh",
  "10": "Bihar",
  "11": "Sikkim",
  "12": "Arunachal Pradesh",
  "13": "Nagaland",
  "14": "Manipur",
  "15": "Mizoram",
  "16": "Tripura",
  "17": "Meghalaya",
  "18": "Assam",
  "19": "West Bengal",
  "20": "Jharkhand",
  "21": "Odisha",
  "22": "Chhattisgarh",
  "23": "Madhya Pradesh",
  "24": "Gujarat",
  "26": "Dadra and Nagar Haveli and Daman and Diu",
  "27": "Maharashtra",
  "29": "Karnataka",
  "30": "Goa",
  "31": "Lakshadweep",
  "32": "Kerala",
  "33": "Tamil Nadu",
  "34": "Puducherry",
  "35": "Andaman and Nicobar Islands",
  "36": "Telangana",
  "37": "Andhra Pradesh",
  "38": "Ladakh"
};

interface B2BItem {
  id?: number;
  product_id?: number;
  item_details: string;
  sku?: string;
  hsn_code?: string;
  quantity: number;
  rate: number;
  amount: number;
  gst_rate?: number;
}

interface B2BCustomer {
  id?: number;
  legal_name: string;
  trade_name?: string;
  gstin: string;
  pan?: string;
  email?: string;
  phone?: string;
  billing_address: string;
  shipping_address?: string;
  state: string;
  state_code: string;
}

interface LineItem {
  id: string;
  title: string;
  sku: string;
  quantity: number;
  price: string | number;
  discount?: string | number;
}

interface OrderData {
  id: string | number;
  order_number: string;
  total_price: string | number;
  subtotal_price?: string | number;
  total_tax?: string | number;
  currency?: string;
  financial_status?: string;
  source_id?: string;
  created_at: string;
  customer_name?: string;
  customer_first_name?: string;
  customer_last_name?: string;
  customer_email?: string;
  customer_phone?: string;
  customer_address1?: string;
  customer_address2?: string;
  customer_city?: string;
  customer_state?: string;
  customer_zip?: string;
  customer_country?: string;
  line_items?: LineItem[];
}

interface ConvertToB2BModalProps {
  isOpen: boolean;
  onClose: () => void;
  orderId: string | number;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  onSuccess?: () => void;
  onNavigateToB2B?: () => void;
}

export const ConvertToB2BModal: React.FC<ConvertToB2BModalProps> = ({
  isOpen,
  onClose,
  orderId,
  fetchWithAuth,
  onSuccess,
  onNavigateToB2B,
}) => {
  const { success: toastSuccess, error: toastError } = useToast();
  const [isLoadingOrder, setIsLoadingOrder] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [existingCustomers, setExistingCustomers] = useState<B2BCustomer[]>([]);
  const [customerSearch, setCustomerSearch] = useState('');
  const [showCustomerDropdown, setShowCustomerDropdown] = useState(false);

  // Form State
  const [orderNumber, setOrderNumber] = useState('');
  const [sourceId, setSourceId] = useState('Shopify');
  const [invoiceDate, setInvoiceDate] = useState(new Date().toISOString().split('T')[0]);
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().split('T')[0]);
  const [paymentStatus, setPaymentStatus] = useState('PAID');
  const [paymentMethod, setPaymentMethod] = useState('Bank Transfer / Online');

  // Customer / GSTIN details
  const [customerId, setCustomerId] = useState<number | undefined>(undefined);
  const [customerName, setCustomerName] = useState('');
  const [customerGstin, setCustomerGstin] = useState('');
  const [customerPan, setCustomerPan] = useState('');
  const [customerEmail, setCustomerEmail] = useState('');
  const [customerPhone, setCustomerPhone] = useState('');
  const [customerAddress, setCustomerAddress] = useState('');
  const [customerShippingAddress, setCustomerShippingAddress] = useState('');
  const [customerState, setCustomerState] = useState('');
  const [customerStateCode, setCustomerStateCode] = useState('');

  // Line items & charges
  const [items, setItems] = useState<B2BItem[]>([]);
  const [discountPercent, setDiscountPercent] = useState<number>(0);
  const [transportationCharge, setTransportationCharge] = useState<number>(0);

  // Calculated totals
  const [subtotalPrice, setSubtotalPrice] = useState<number>(0);
  const [discountAmount, setDiscountAmount] = useState<number>(0);
  const [cgstRate, setCgstRate] = useState<number>(0);
  const [cgstAmount, setCgstAmount] = useState<number>(0);
  const [sgstRate, setSgstRate] = useState<number>(0);
  const [sgstAmount, setSgstAmount] = useState<number>(0);
  const [igstRate, setIgstRate] = useState<number>(0);
  const [igstAmount, setIgstAmount] = useState<number>(0);
  const [totalPrice, setTotalPrice] = useState<number>(0);

  // Helper to match state name to GST state code
  const getStateCodeFromName = (stateName: string): string => {
    if (!stateName) return '33'; // Default TN
    const normalized = stateName.trim().toLowerCase();
    for (const [code, name] of Object.entries(GST_STATE_MAP)) {
      if (name.toLowerCase() === normalized || normalized.includes(name.toLowerCase()) || name.toLowerCase().includes(normalized)) {
        return code;
      }
    }
    return '33';
  };

  // Auto-fill state and PAN from GSTIN when typed
  const handleGstinChange = (gstinVal: string) => {
    const upper = gstinVal.toUpperCase().trim();
    setCustomerGstin(upper);

    if (upper.length >= 2) {
      const code = upper.substring(0, 2);
      if (GST_STATE_MAP[code]) {
        setCustomerStateCode(code);
        setCustomerState(GST_STATE_MAP[code]);
      }
    }

    if (upper.length >= 12) {
      const panPart = upper.substring(2, 12);
      setCustomerPan(panPart);
    }
  };

  // Fetch Existing B2B Customers for autocomplete
  useEffect(() => {
    if (!isOpen) return;
    const loadB2BCustomers = async () => {
      try {
        const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers`);
        if (res.ok) {
          const data = await res.json();
          setExistingCustomers(data || []);
        }
      } catch (err) {
        console.error('Failed to load B2B customers', err);
      }
    };
    loadB2BCustomers();
  }, [isOpen, fetchWithAuth]);

  // Fetch full Order details to pre-populate
  useEffect(() => {
    if (!isOpen || !orderId) return;
    setIsLoadingOrder(true);

    const loadOrder = async () => {
      try {
        const res = await fetchWithAuth(`${API_BASE}/api/orders?id=${orderId}`);
        if (!res.ok) throw new Error('Failed to load order');
        const data = await res.json();
        const ord: OrderData = data.order || data;

        setOrderNumber(ord.order_number || '');
        setSourceId(ord.source_id || 'Shopify');
        
        if (ord.created_at) {
          const d = new Date(ord.created_at).toISOString().split('T')[0];
          setInvoiceDate(d);
          setPaymentDate(d);
        }

        // Customer details
        const fullCustomerName = ord.customer_name || 
          ([ord.customer_first_name, ord.customer_last_name].filter(Boolean).join(' ') || 'B2B Client');
        setCustomerName(fullCustomerName);
        setCustomerEmail(ord.customer_email || '');
        setCustomerPhone(ord.customer_phone || '');

        // Address construction
        const addressParts = [
          ord.customer_address1,
          ord.customer_address2,
          ord.customer_city,
          ord.customer_state,
          ord.customer_zip,
          ord.customer_country
        ].filter(Boolean);
        const fullAddress = addressParts.join(', ');
        setCustomerAddress(fullAddress);
        setCustomerShippingAddress(fullAddress);

        const stateName = ord.customer_state || 'Tamil Nadu';
        setCustomerState(stateName);
        setCustomerStateCode(getStateCodeFromName(stateName));

        // Line Items
        const lItems: LineItem[] = ord.line_items || [];
        if (lItems.length > 0) {
          const b2bItems: B2BItem[] = lItems.map(item => {
            const qty = item.quantity || 1;
            // The item price in orders is the actual selling unit price (taxable rate)
            const price = Number(item.price) || 0;
            const rate = parseFloat(price.toFixed(2));
            const amount = parseFloat((rate * qty).toFixed(2));
            return {
              item_details: item.title || 'Fragrance Product',
              sku: item.sku || '',
              hsn_code: '33029019',
              quantity: qty,
              rate: rate,
              amount: amount,
              gst_rate: 18
            };
          });
          setItems(b2bItems);
        } else {
          // Fallback if no line items
          const fallbackTotal = Number(ord.total_price) || 0;
          const fallbackRate = parseFloat(fallbackTotal.toFixed(2));
          setItems([{
            item_details: `${ord.source_id || 'Retail'} Order #${ord.order_number}`,
            sku: '',
            hsn_code: '33029019',
            quantity: 1,
            rate: fallbackRate,
            amount: fallbackRate,
            gst_rate: 18
          }]);
        }
      } catch (err) {
        console.error('Error loading order for B2B conversion:', err);
        toastError('Failed to load order information.');
      } finally {
        setIsLoadingOrder(false);
      }
    };

    loadOrder();
  }, [isOpen, orderId, fetchWithAuth]);

  // Dynamic recalculation of taxes & totals
  useEffect(() => {
    let sub = 0;
    items.forEach(item => {
      const lineAmt = (item.quantity || 0) * (item.rate || 0);
      sub += lineAmt;
    });
    setSubtotalPrice(parseFloat(sub.toFixed(2)));

    let disc = 0;
    if (discountPercent > 0) {
      disc = parseFloat(((sub * discountPercent) / 100).toFixed(2));
    }
    setDiscountAmount(disc);

    const taxable = Math.max(0, sub - disc);
    const discountRatio = sub > 0 ? taxable / sub : 1;

    let cgst = 0;
    let sgst = 0;
    let igst = 0;
    let activeCgstRate = 0;
    let activeSgstRate = 0;
    let activeIgstRate = 0;

    const isSameState = customerStateCode === '33'; // TN Seller default prefix

    items.forEach(item => {
      const itemSubtotal = (item.quantity || 0) * (item.rate || 0);
      const itemTaxable = itemSubtotal * discountRatio;
      const gRate = item.gst_rate !== undefined ? item.gst_rate : 18;

      if (isSameState) {
        const halfRate = gRate / 2;
        cgst += (itemTaxable * halfRate) / 100;
        sgst += (itemTaxable * halfRate) / 100;
        activeCgstRate = halfRate;
        activeSgstRate = halfRate;
      } else {
        igst += (itemTaxable * gRate) / 100;
        activeIgstRate = gRate;
      }
    });

    setCgstRate(isSameState ? activeCgstRate : 0);
    setCgstAmount(parseFloat(cgst.toFixed(2)));
    setSgstRate(isSameState ? activeSgstRate : 0);
    setSgstAmount(parseFloat(sgst.toFixed(2)));
    setIgstRate(!isSameState ? activeIgstRate : 0);
    setIgstAmount(parseFloat(igst.toFixed(2)));

    const finalTot = taxable + cgst + sgst + igst + (Number(transportationCharge) || 0);
    setTotalPrice(parseFloat(finalTot.toFixed(2)));
  }, [items, discountPercent, transportationCharge, customerStateCode]);

  // Select an existing customer
  const handleSelectCustomer = (cust: B2BCustomer) => {
    setCustomerId(cust.id);
    setCustomerName(cust.legal_name || cust.trade_name || '');
    setCustomerGstin(cust.gstin || '');
    setCustomerPan(cust.pan || '');
    setCustomerEmail(cust.email || '');
    setCustomerPhone(cust.phone || '');
    setCustomerAddress(cust.billing_address || '');
    setCustomerShippingAddress(cust.shipping_address || cust.billing_address || '');
    setCustomerState(cust.state || '');
    setCustomerStateCode(cust.state_code || '');
    setShowCustomerDropdown(false);
  };

  // Submit and Issue B2B Invoice
  const handleSaveAndIssue = async (asDraftOnly: boolean = false) => {
    if (!customerGstin.trim()) {
      toastError('Please enter the Customer GSTIN for this B2B invoice.');
      return;
    }
    if (customerGstin.length !== 15) {
      toastError('Invalid GSTIN length. GSTIN must be exactly 15 characters (e.g. 33AAAAA0000A1Z5).');
      return;
    }
    if (!customerName.trim()) {
      toastError('Please enter the Customer / Company Name.');
      return;
    }

    setIsSubmitting(true);
    try {
      const payload = {
        origin_order_id: String(orderId),
        invoice_date: new Date(invoiceDate).toISOString(),
        order_number: `ORD-${orderNumber}`,
        customer_id: customerId,
        customer_gstin: customerGstin.toUpperCase().trim(),
        customer_name: customerName.trim(),
        customer_email: customerEmail.trim(),
        customer_phone: customerPhone.trim(),
        customer_state: customerState || 'Tamil Nadu',
        customer_state_code: customerStateCode || '33',
        customer_address: customerAddress || '',
        customer_shipping_address: customerShippingAddress || customerAddress || '',
        subtotal_price: subtotalPrice,
        discount_percent: discountPercent,
        discount_amount: discountAmount,
        cgst_rate: cgstRate,
        cgst_amount: cgstAmount,
        sgst_rate: sgstRate,
        sgst_amount: sgstAmount,
        igst_rate: igstRate,
        igst_amount: igstAmount,
        transportation_charge: Number(transportationCharge) || 0,
        total_price: totalPrice,
        status: 'DRAFT',
        payment_status: paymentStatus,
        paid_amount: paymentStatus === 'PAID' ? totalPrice : 0,
        balance_amount: paymentStatus === 'PAID' ? 0 : totalPrice,
        payment_date: paymentStatus === 'PAID' ? new Date(paymentDate).toISOString() : undefined,
        payment_method: paymentMethod,
        customer_notes: `Converted from ${sourceId} Order #${orderNumber}.\nThanks for your business.`,
        items: items.map(it => ({
          item_details: it.item_details,
          sku: it.sku || '',
          hsn_code: it.hsn_code || '33029019',
          quantity: Number(it.quantity) || 1,
          rate: Number(it.rate) || 0,
          amount: parseFloat(((Number(it.quantity) || 1) * (Number(it.rate) || 0)).toFixed(2)),
          gst_rate: it.gst_rate !== undefined ? it.gst_rate : 18
        }))
      };

      // 1. Create B2B Invoice
      const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || 'Failed to create B2B invoice');
      }

      const savedInvoice = await res.json();

      // 2. Issue the invoice directly (Default flow as requested)
      if (!asDraftOnly && savedInvoice.id) {
        const issueRes = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/issue?id=${savedInvoice.id}`, {
          method: 'POST'
        });
        if (!issueRes.ok) {
          const issueErr = await issueRes.text();
          console.warn('Invoice saved as draft, activation notice:', issueErr);
        }
      }

      toastSuccess(`Successfully converted Order #${orderNumber} into B2B Tax Invoice!`);
      if (onSuccess) onSuccess();
      if (onNavigateToB2B && !asDraftOnly) {
        // Optional navigation callback
      }
      onClose();
    } catch (err: any) {
      console.error('B2B Conversion error:', err);
      toastError(err.message || 'Failed to convert order to B2B invoice');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose} style={{ zIndex: 1050 }}>
      <div 
        className="premium-modal" 
        onClick={e => e.stopPropagation()} 
        style={{ maxWidth: '820px', width: '95%', maxHeight: '90vh', overflowY: 'auto', padding: '1.75rem 2rem', position: 'relative' }}
      >
        {/* Close Button */}
        <button
          type="button"
          onClick={onClose}
          style={{ position: 'absolute', top: '1.25rem', right: '1.25rem', color: 'var(--text-tertiary)', background: 'none', border: 'none', cursor: 'pointer', padding: '0.5rem', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
          className="hover-bg"
          aria-label="Close modal"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>

        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.25rem' }}>
          <div style={{ 
            width: '42px', 
            height: '42px', 
            borderRadius: '12px', 
            background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)', 
            color: 'white', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(16, 185, 129, 0.3)'
          }}>
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 800, margin: 0, color: 'var(--text-primary)' }}>Convert to B2B Tax Invoice</h2>
              <span className="badge-pill badge-pill-info" style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem' }}>
                {sourceId} #{orderNumber}
              </span>
            </div>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', margin: '0.2rem 0 0 0' }}>
              Input customer GST details to auto-generate and issue an official B2B GST tax invoice.
            </p>
          </div>
        </div>

        {isLoadingOrder ? (
          <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
            <div className="loading-spinner" style={{ width: '32px', height: '32px' }}></div>
            <p style={{ marginTop: '1rem', fontSize: '0.9rem' }}>Pre-populating order and line items...</p>
          </div>
        ) : (
          <form onSubmit={(e) => { e.preventDefault(); handleSaveAndIssue(false); }}>
            {/* Section 1: Customer & GSTIN Details */}
            <div style={{ 
              background: 'var(--bg-input)', 
              borderRadius: '12px', 
              padding: '1.25rem', 
              border: '1px solid var(--border-color)',
              marginBottom: '1.25rem' 
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
                <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  🏢 1. Business & GSTIN Information
                </span>
                
                {/* Autocomplete selector for existing B2B clients */}
                {existingCustomers.length > 0 && (
                  <div style={{ position: 'relative' }}>
                    <button
                      type="button"
                      onClick={() => setShowCustomerDropdown(!showCustomerDropdown)}
                      style={{ 
                        background: 'var(--surface-color)', 
                        border: '1px solid var(--border-color)', 
                        color: 'var(--text-primary)', 
                        fontSize: '0.75rem', 
                        padding: '0.3rem 0.6rem', 
                        borderRadius: '6px', 
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.35rem'
                      }}
                    >
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                      Select Existing Customer ({existingCustomers.length})
                    </button>

                    {showCustomerDropdown && (
                      <div style={{
                        position: 'absolute',
                        right: 0,
                        top: '100%',
                        marginTop: '4px',
                        background: 'var(--bg-card)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '8px',
                        width: '280px',
                        maxHeight: '200px',
                        overflowY: 'auto',
                        zIndex: 1100,
                        boxShadow: '0 8px 24px rgba(0,0,0,0.25)',
                        padding: '4px'
                      }}>
                        <input
                          type="text"
                          placeholder="Search customer name or GST..."
                          value={customerSearch}
                          onChange={(e) => setCustomerSearch(e.target.value)}
                          style={{
                            width: '100%',
                            padding: '6px 8px',
                            fontSize: '0.75rem',
                            background: 'var(--bg-input)',
                            border: '1px solid var(--border-color)',
                            borderRadius: '4px',
                            color: 'var(--text-primary)',
                            marginBottom: '4px'
                          }}
                        />
                        {existingCustomers
                          .filter(c => 
                            c.legal_name?.toLowerCase().includes(customerSearch.toLowerCase()) || 
                            c.gstin?.toLowerCase().includes(customerSearch.toLowerCase())
                          )
                          .map(cust => (
                            <div
                              key={cust.id}
                              onClick={() => handleSelectCustomer(cust)}
                              style={{
                                padding: '6px 8px',
                                fontSize: '0.75rem',
                                cursor: 'pointer',
                                borderRadius: '4px',
                                borderBottom: '1px solid var(--border-color)'
                              }}
                              onMouseEnter={(e) => e.currentTarget.style.background = 'var(--bg-hover)'}
                              onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
                            >
                              <div style={{ fontWeight: 600 }}>{cust.legal_name || cust.trade_name}</div>
                              <div style={{ fontSize: '0.7rem', color: 'var(--text-tertiary)' }}>{cust.gstin} • {cust.state}</div>
                            </div>
                          ))}
                      </div>
                    )}
                  </div>
                )}
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '0.75rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Customer GSTIN <span style={{ color: 'var(--status-danger)' }}>*</span>
                  </label>
                  <input
                    type="text"
                    required
                    maxLength={15}
                    placeholder="e.g. 33AAAAA0000A1Z5"
                    value={customerGstin}
                    onChange={(e) => handleGstinChange(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '0.65rem 0.85rem',
                      background: 'var(--surface-color)',
                      border: customerGstin.length === 15 ? '1px solid #10b981' : '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontSize: '0.9rem',
                      fontWeight: 700,
                      letterSpacing: '0.05em',
                      textTransform: 'uppercase'
                    }}
                  />
                  {customerGstin && customerGstin.length !== 15 && (
                    <span style={{ fontSize: '0.7rem', color: 'var(--status-warning)', marginTop: '2px', display: 'block' }}>
                      {15 - customerGstin.length} chars remaining
                    </span>
                  )}
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Company / Trade Name <span style={{ color: 'var(--status-danger)' }}>*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. ABC Enterprises Pvt Ltd"
                    value={customerName}
                    onChange={(e) => setCustomerName(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '0.65rem 0.85rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontSize: '0.9rem',
                      fontWeight: 600
                    }}
                  />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem', marginBottom: '0.75rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    PAN Number
                  </label>
                  <input
                    type="text"
                    maxLength={10}
                    placeholder="Auto-extracted"
                    value={customerPan}
                    onChange={(e) => setCustomerPan(e.target.value.toUpperCase())}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                      textTransform: 'uppercase'
                    }}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    State & Code
                  </label>
                  <div style={{ display: 'flex', gap: '4px' }}>
                    <input
                      type="text"
                      readOnly
                      value={customerStateCode}
                      style={{
                        width: '45px',
                        textAlign: 'center',
                        padding: '0.55rem 0.25rem',
                        background: 'var(--surface-color)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '8px',
                        color: 'var(--accent-color)',
                        fontWeight: 700,
                        fontSize: '0.85rem'
                      }}
                    />
                    <input
                      type="text"
                      value={customerState}
                      onChange={(e) => {
                        setCustomerState(e.target.value);
                        setCustomerStateCode(getStateCodeFromName(e.target.value));
                      }}
                      style={{
                        flex: 1,
                        padding: '0.55rem 0.75rem',
                        background: 'var(--surface-color)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '8px',
                        color: 'var(--text-primary)',
                        fontSize: '0.85rem'
                      }}
                    />
                  </div>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Tax Jurisdiction
                  </label>
                  <div style={{ 
                    padding: '0.55rem 0.75rem', 
                    borderRadius: '8px', 
                    background: customerStateCode === '33' ? 'rgba(16, 185, 129, 0.1)' : 'rgba(59, 130, 246, 0.1)',
                    color: customerStateCode === '33' ? '#10b981' : '#3b82f6',
                    fontSize: '0.8rem',
                    fontWeight: 700,
                    textAlign: 'center'
                  }}>
                    {customerStateCode === '33' ? 'INTRA-STATE (CGST + SGST)' : 'INTER-STATE (IGST)'}
                  </div>
                </div>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Billing Address
                </label>
                <input
                  type="text"
                  value={customerAddress}
                  onChange={(e) => setCustomerAddress(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.55rem 0.75rem',
                    background: 'var(--surface-color)',
                    border: '1px solid var(--border-color)',
                    borderRadius: '8px',
                    color: 'var(--text-primary)',
                    fontSize: '0.85rem'
                  }}
                />
              </div>
            </div>

            {/* Section 2: Invoice & Payment Settings */}
            <div style={{ 
              background: 'var(--bg-input)', 
              borderRadius: '12px', 
              padding: '1.25rem', 
              border: '1px solid var(--border-color)',
              marginBottom: '1.25rem' 
            }}>
              <span style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.75rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
                📅 2. Invoice Date & Payment Status (Defaults to PAID)
              </span>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Invoice Date
                  </label>
                  <input
                    type="date"
                    required
                    value={invoiceDate}
                    onChange={(e) => setInvoiceDate(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem'
                    }}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Payment Status
                  </label>
                  <select
                    value={paymentStatus}
                    onChange={(e) => setPaymentStatus(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                      fontWeight: 600
                    }}
                  >
                    <option value="PAID">PAID (Prepaid / Received)</option>
                    <option value="UNPAID">UNPAID (Payment Pending)</option>
                  </select>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Payment Method
                  </label>
                  <input
                    type="text"
                    value={paymentMethod}
                    onChange={(e) => setPaymentMethod(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem'
                    }}
                  />
                </div>
              </div>
            </div>

            {/* Section 3: Line Items Table */}
            <div style={{ 
              background: 'var(--bg-input)', 
              borderRadius: '12px', 
              padding: '1.25rem', 
              border: '1px solid var(--border-color)',
              marginBottom: '1.25rem' 
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--accent-color)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  📦 3. Line Items & Tax Rates
                </span>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>
                  {items.length} item(s) loaded
                </span>
              </div>

              <div style={{ overflowX: 'auto', marginBottom: '0.75rem' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                      <th style={{ textAlign: 'left', padding: '6px' }}>Item Description</th>
                      <th style={{ textAlign: 'center', padding: '6px', width: '80px' }}>HSN</th>
                      <th style={{ textAlign: 'center', padding: '6px', width: '60px' }}>Qty</th>
                      <th style={{ textAlign: 'right', padding: '6px', width: '90px' }}>Rate (₹)</th>
                      <th style={{ textAlign: 'center', padding: '6px', width: '70px' }}>GST %</th>
                      <th style={{ textAlign: 'right', padding: '6px', width: '90px' }}>Amount (₹)</th>
                    </tr>
                  </thead>
                  <tbody>
                    {items.map((item, idx) => (
                      <tr key={idx} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                        <td style={{ padding: '6px' }}>
                          <input
                            type="text"
                            value={item.item_details}
                            onChange={(e) => {
                              const newItems = [...items];
                              newItems[idx].item_details = e.target.value;
                              setItems(newItems);
                            }}
                            style={{
                              width: '100%',
                              background: 'transparent',
                              border: 'none',
                              color: 'var(--text-primary)',
                              fontSize: '0.8rem',
                              fontWeight: 600
                            }}
                          />
                        </td>
                        <td style={{ padding: '6px', textAlign: 'center' }}>
                          <input
                            type="text"
                            value={item.hsn_code}
                            onChange={(e) => {
                              const newItems = [...items];
                              newItems[idx].hsn_code = e.target.value;
                              setItems(newItems);
                            }}
                            style={{
                              width: '100%',
                              textAlign: 'center',
                              background: 'transparent',
                              border: '1px solid var(--border-color)',
                              borderRadius: '4px',
                              color: 'var(--text-secondary)',
                              fontSize: '0.75rem',
                              padding: '2px 4px'
                            }}
                          />
                        </td>
                        <td style={{ padding: '6px', textAlign: 'center' }}>
                          <input
                            type="number"
                            min={1}
                            value={item.quantity}
                            onChange={(e) => {
                              const newItems = [...items];
                              newItems[idx].quantity = Number(e.target.value) || 1;
                              newItems[idx].amount = parseFloat((newItems[idx].quantity * newItems[idx].rate).toFixed(2));
                              setItems(newItems);
                            }}
                            style={{
                              width: '100%',
                              textAlign: 'center',
                              background: 'transparent',
                              border: '1px solid var(--border-color)',
                              borderRadius: '4px',
                              color: 'var(--text-primary)',
                              fontSize: '0.8rem',
                              padding: '2px 4px'
                            }}
                          />
                        </td>
                        <td style={{ padding: '6px', textAlign: 'right' }}>
                          <input
                            type="number"
                            step="0.01"
                            value={item.rate}
                            onChange={(e) => {
                              const newItems = [...items];
                              newItems[idx].rate = Number(e.target.value) || 0;
                              newItems[idx].amount = parseFloat((newItems[idx].quantity * newItems[idx].rate).toFixed(2));
                              setItems(newItems);
                            }}
                            style={{
                              width: '100%',
                              textAlign: 'right',
                              background: 'transparent',
                              border: '1px solid var(--border-color)',
                              borderRadius: '4px',
                              color: 'var(--text-primary)',
                              fontSize: '0.8rem',
                              padding: '2px 4px'
                            }}
                          />
                        </td>
                        <td style={{ padding: '6px', textAlign: 'center' }}>
                          <select
                            value={item.gst_rate !== undefined ? item.gst_rate : 18}
                            onChange={(e) => {
                              const newItems = [...items];
                              newItems[idx].gst_rate = Number(e.target.value);
                              setItems(newItems);
                            }}
                            style={{
                              background: 'transparent',
                              border: '1px solid var(--border-color)',
                              borderRadius: '4px',
                              color: 'var(--text-primary)',
                              fontSize: '0.75rem',
                              padding: '2px 4px'
                            }}
                          >
                            <option value={18}>18%</option>
                            <option value={12}>12%</option>
                            <option value={5}>5%</option>
                            <option value={28}>28%</option>
                            <option value={0}>0%</option>
                          </select>
                        </td>
                        <td style={{ padding: '6px', textAlign: 'right', fontWeight: 600 }}>
                          ₹{((item.quantity || 0) * (item.rate || 0)).toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Discount & Shipping Charges */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginTop: '0.75rem', marginBottom: '0.75rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
                    Discount (%)
                  </label>
                  <input
                    type="number"
                    min={0}
                    max={100}
                    value={discountPercent}
                    onChange={(e) => setDiscountPercent(Number(e.target.value) || 0)}
                    style={{
                      width: '100%',
                      padding: '0.45rem 0.65rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '6px',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem'
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.25rem' }}>
                    Transportation / Shipping Charge (₹)
                  </label>
                  <input
                    type="number"
                    min={0}
                    value={transportationCharge}
                    onChange={(e) => setTransportationCharge(Number(e.target.value) || 0)}
                    style={{
                      width: '100%',
                      padding: '0.45rem 0.65rem',
                      background: 'var(--surface-color)',
                      border: '1px solid var(--border-color)',
                      borderRadius: '6px',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem'
                    }}
                  />
                </div>
              </div>

              {/* Tax Summary Breakdown */}
              <div style={{ 
                borderTop: '1px dashed var(--border-color)', 
                paddingTop: '0.75rem',
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: '1rem',
                fontSize: '0.825rem'
              }}>
                <div style={{ color: 'var(--text-secondary)' }}>
                  <div>Subtotal: <strong>₹{subtotalPrice.toFixed(2)}</strong> {discountAmount > 0 && <span style={{ color: '#ef4444' }}>(-₹{discountAmount.toFixed(2)})</span>}</div>
                  {customerStateCode === '33' ? (
                    <div>CGST ({cgstRate}%): ₹{cgstAmount.toFixed(2)} | SGST ({sgstRate}%): ₹{sgstAmount.toFixed(2)}</div>
                  ) : (
                    <div>IGST ({igstRate}%): ₹{igstAmount.toFixed(2)}</div>
                  )}
                  {transportationCharge > 0 && <div>Shipping: ₹{Number(transportationCharge).toFixed(2)}</div>}
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: '1.1rem', fontWeight: 800, color: 'var(--accent-color)' }}>
                    Total B2B Invoice: ₹{totalPrice.toFixed(2)}
                  </div>
                  <div style={{ fontSize: '0.75rem', color: '#10b981', fontWeight: 600 }}>
                    Status: {paymentStatus}
                  </div>
                </div>
              </div>
            </div>

            {/* Modal Actions */}
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between', 
              alignItems: 'center', 
              borderTop: '1px solid var(--border-color)', 
              paddingTop: '1.25rem' 
            }}>
              <button
                type="button"
                className="btn-secondary"
                onClick={onClose}
                disabled={isSubmitting}
                style={{ padding: '0.6rem 1.2rem', borderRadius: '8px' }}
              >
                Cancel
              </button>

              <div style={{ display: 'flex', gap: '0.75rem' }}>
                <button
                  type="button"
                  onClick={() => handleSaveAndIssue(true)}
                  disabled={isSubmitting}
                  className="btn-secondary"
                  style={{ padding: '0.6rem 1.2rem', borderRadius: '8px', fontSize: '0.85rem' }}
                >
                  Save as Draft
                </button>

                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="btn-primary"
                  style={{ 
                    padding: '0.6rem 1.5rem', 
                    borderRadius: '8px', 
                    fontSize: '0.85rem', 
                    fontWeight: 700,
                    background: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
                    border: 'none',
                    color: 'white',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    boxShadow: '0 4px 14px rgba(16, 185, 129, 0.35)'
                  }}
                >
                  {isSubmitting ? (
                    <>
                      <div className="loading-spinner" style={{ width: '14px', height: '14px', borderWidth: '2px', marginBottom: 0 }}></div>
                      Issuing B2B Invoice...
                    </>
                  ) : (
                    <>
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
                      Save & Issue B2B Invoice
                    </>
                  )}
                </button>
              </div>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};
