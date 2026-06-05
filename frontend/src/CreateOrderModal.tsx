import React, { useState, useEffect, useRef } from 'react';

// Premium Icons for Form Inputs
const UserIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
    <circle cx="12" cy="7" r="4" />
  </svg>
);

const PhoneIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" />
  </svg>
);

const EmailIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
    <polyline points="22,6 12,13 2,6" />
  </svg>
);

const MapPinIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
    <circle cx="12" cy="10" r="3" />
  </svg>
);

const CityIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <path d="M3 21h18M3 7V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v2M5 21V7m14 14V7M9 7h6M9 11h6M9 15h6" />
  </svg>
);

const HashIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <line x1="4" y1="9" x2="20" y2="9" />
    <line x1="4" y1="15" x2="20" y2="15" />
    <line x1="10" y1="3" x2="8" y2="21" />
    <line x1="16" y1="3" x2="14" y2="21" />
  </svg>
);

const SearchIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.6 }}>
    <circle cx="11" cy="11" r="8" />
    <line x1="21" y1="21" x2="16.65" y2="16.65" />
  </svg>
);

interface CreateOrderModalProps {
  isOpen: boolean;
  onClose: () => void;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
  API_BASE: string;
  onSuccess: () => void;
}

interface LineItem {
  mi_sku: string;
  title: string;
  quantity: number;
  price: number; // item rate
  discount: number; // absolute discount in Rupees
  current_stock: number;
}

export const CreateOrderModal: React.FC<CreateOrderModalProps> = ({
  isOpen,
  onClose,
  fetchWithAuth,
  API_BASE,
  onSuccess
}) => {
  // Customer Search & Info States
  const [customerSearch, setCustomerSearch] = useState('');
  const [customerResults, setCustomerResults] = useState<any[]>([]);
  const [showCustDropdown, setShowCustDropdown] = useState(false);

  const [customerName, setCustomerName] = useState('');
  const [customerPhone, setCustomerPhone] = useState('');
  const [customerEmail, setCustomerEmail] = useState('');
  const [customerAddress1, setCustomerAddress1] = useState('');
  const [customerAddress2, setCustomerAddress2] = useState('');
  const [customerCity, setCustomerCity] = useState('');
  const [customerState, setCustomerState] = useState('Tamil Nadu');
  const [customerZip, setCustomerZip] = useState('');
  const [customerCountry, setCustomerCountry] = useState('India');

  const [financialStatus, setFinancialStatus] = useState('paid');
  const [fulfillmentStatus, setFulfillmentStatus] = useState('fulfilled');

  // Product Search & Items States
  const [productSearch, setProductSearch] = useState('');
  const [productResults, setProductResults] = useState<any[]>([]);
  const [showProdDropdown, setShowProdDropdown] = useState(false);
  const [lineItems, setLineItems] = useState<LineItem[]>([]);
  const [orderDiscountPercent, setOrderDiscountPercent] = useState<number>(0);

  // General States
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  // Refs for closing dropdowns on click outside
  const custDropdownRef = useRef<HTMLDivElement>(null);
  const prodDropdownRef = useRef<HTMLDivElement>(null);

  // Debounced search for customers
  useEffect(() => {
    if (!customerSearch.trim()) {
      setCustomerResults([]);
      setShowCustDropdown(false);
      return;
    }
    const timer = setTimeout(async () => {
      try {
        const res = await fetchWithAuth(
          `${API_BASE}/api/customers?search=${encodeURIComponent(customerSearch)}&page=1&pageSize=5`
        );
        if (res.ok) {
          const data = await res.json();
          if (data.success && data.customers) {
            setCustomerResults(data.customers);
            setShowCustDropdown(true);
          }
        }
      } catch (err) {
        console.error('Error searching customers:', err);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [customerSearch]);

  // Debounced search for products
  useEffect(() => {
    if (!productSearch.trim()) {
      setProductResults([]);
      setShowProdDropdown(false);
      return;
    }
    const timer = setTimeout(async () => {
      try {
        const res = await fetchWithAuth(
          `${API_BASE}/api/inventory?search=${encodeURIComponent(productSearch)}`
        );
        if (res.ok) {
          const data = await res.json();
          if (Array.isArray(data)) {
            setProductResults(data);
            setShowProdDropdown(true);
          }
        }
      } catch (err) {
        console.error('Error searching products:', err);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [productSearch]);

  // Handle click outside to close dropdowns
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (custDropdownRef.current && !custDropdownRef.current.contains(event.target as Node)) {
        setShowCustDropdown(false);
      }
      if (prodDropdownRef.current && !prodDropdownRef.current.contains(event.target as Node)) {
        setShowProdDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (!isOpen) return null;

  const selectCustomer = (c: any) => {
    const fullName = [c.first_name, c.last_name].filter(Boolean).join(' ');
    setCustomerName(fullName);
    setCustomerPhone(c.phone_number || '');
    setCustomerEmail(c.email || '');
    setCustomerAddress1(c.address1 || '');
    setCustomerAddress2(c.address2 || '');
    setCustomerCity(c.city || '');
    setCustomerState(c.state || 'Tamil Nadu');
    setCustomerZip(c.zip_code || '');
    setCustomerCountry(c.country || 'India');
    setShowCustDropdown(false);
    setCustomerSearch('');
  };

  const addProduct = (p: any) => {
    const existingIndex = lineItems.findIndex(item => item.mi_sku === p.mi_sku);
    if (existingIndex > -1) {
      const updated = [...lineItems];
      updated[existingIndex].quantity += 1;
      setLineItems(updated);
    } else {
      setLineItems([
        ...lineItems,
        {
          mi_sku: p.mi_sku,
          title: p.title,
          quantity: 1,
          price: 699.0, // Default POS price if not configured, let admin edit it
          discount: 0,
          current_stock: p.current_stock
        }
      ]);
    }
    setShowProdDropdown(false);
    setProductSearch('');
  };

  const updateLineItem = (index: number, field: keyof LineItem, value: any) => {
    const updated = [...lineItems];
    updated[index] = {
      ...updated[index],
      [field]: value
    };
    setLineItems(updated);
  };

  const removeLineItem = (index: number) => {
    setLineItems(lineItems.filter((_, i) => i !== index));
  };

  // Indian States List
  const states = [
    '', 'Tamil Nadu', 'Maharashtra', 'Karnataka', 'Delhi', 'Telangana', 'Andhra Pradesh', 
    'Kerala', 'Gujarat', 'Uttar Pradesh', 'West Bengal', 'Rajasthan', 'Madhya Pradesh', 
    'Punjab', 'Haryana', 'Bihar', 'Odisha', 'Assam', 'Goa'
  ];

  // Calculations
  const subtotal = lineItems.reduce((sum, item) => sum + (item.quantity * item.price - item.discount), 0);
  const orderDiscountVal = subtotal * (orderDiscountPercent / 100);
  const total = Math.max(0, subtotal - orderDiscountVal);

  // Inclusive GST Calculation (18% inclusive GST)
  const gstRate = 0.18;
  const taxableValue = total / (1 + gstRate);
  const gstAmount = total - taxableValue;

  const isTamilNadu = customerState.toLowerCase().trim() === 'tamil nadu';
  const cgst = isTamilNadu ? gstAmount / 2 : 0;
  const sgst = isTamilNadu ? gstAmount / 2 : 0;
  const igst = !isTamilNadu ? gstAmount : 0;

  const totalDiscountPayload = lineItems.reduce((sum, item) => sum + item.discount, 0) + orderDiscountVal;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (lineItems.length === 0) {
      setErrorMsg('At least one product is required');
      return;
    }
    if (!customerName.trim()) {
      setErrorMsg('Customer name is required');
      return;
    }
    if (!customerPhone.trim()) {
      setErrorMsg('Customer phone number is required');
      return;
    }

    setIsLoading(true);
    setErrorMsg('');

    const payload = {
      terminal_code: 'POS1',
      total_price: parseFloat(total.toFixed(2)),
      total_discount: parseFloat(totalDiscountPayload.toFixed(2)),
      financial_status: financialStatus,
      fulfillment_status: fulfillmentStatus,
      customer_name: customerName,
      customer_phone: customerPhone,
      customer_email: customerEmail,
      customer_address1: customerAddress1,
      customer_address2: customerAddress2,
      customer_city: customerCity,
      customer_state: customerState,
      customer_zip: customerZip,
      customer_country: customerCountry,
      line_items: lineItems.map(item => ({
        mi_sku: item.mi_sku,
        title: item.title,
        quantity: item.quantity,
        price: item.price,
        discount: item.discount
      }))
    };

    try {
      const res = await fetchWithAuth(`${API_BASE}/api/orders`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        onSuccess();
        onClose();
      } else {
        const text = await res.text();
        setErrorMsg(text || 'Failed to create order');
      }
    } catch (err) {
      console.error(err);
      setErrorMsg('Network error creating order');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="modal-overlay" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(5px)', backgroundColor: 'rgba(0, 0, 0, 0.4)' }}>
      <div className="premium-modal wide" style={{ display: 'flex', flexDirection: 'column', maxHeight: '90vh', width: '95%', maxWidth: '1100px', padding: '1.75rem 2.25rem', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-lg)' }}>
        
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.75rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div className="modal-header-icon" style={{ background: 'linear-gradient(135deg, #6366f1, #4f46e5)', width: '42px', height: '42px', borderRadius: '12px', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', marginBottom: 0 }}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <circle cx="9" cy="21" r="1"></circle>
                <circle cx="20" cy="21" r="1"></circle>
                <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path>
              </svg>
            </div>
            <div>
              <h2 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 700 }}>POS Manual Order Creation</h2>
              <p style={{ margin: 0, fontSize: '0.8rem', color: 'var(--text-secondary)' }}>Create order directly and deduct stock from main inventory</p>
            </div>
          </div>
          <button 
            type="button" 
            onClick={onClose} 
            style={{ background: 'var(--bg-input)', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', padding: '6px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)' }}
            onMouseEnter={e => (e.currentTarget.style.backgroundColor = 'var(--border-strong)')}
            onMouseLeave={e => (e.currentTarget.style.backgroundColor = 'var(--bg-input)')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        {errorMsg && (
          <div style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)', color: '#f87171', padding: '0.75rem 1rem', borderRadius: '10px', marginBottom: '1rem', border: '1px solid rgba(239, 68, 68, 0.2)', fontSize: '0.85rem' }}>
            {errorMsg}
          </div>
        )}

        <form onSubmit={handleSubmit} style={{ display: 'grid', gridTemplateColumns: '1.1fr 1.3fr', gap: '2rem', overflowY: 'auto', flex: 1, paddingRight: '4px' }}>
          
          {/* Left Column: Customer Information */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <h3 style={{ fontSize: '0.95rem', margin: '0 0 0.25rem 0', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--text-secondary)' }}>Customer Details</h3>
            
            {/* Customer Search Box */}
            <div ref={custDropdownRef} style={{ position: 'relative' }}>
              <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Search Existing Customer</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                  <SearchIcon />
                </div>
                <input
                  type="text"
                  value={customerSearch}
                  onChange={e => setCustomerSearch(e.target.value)}
                  placeholder="Search by name or phone..."
                  style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                />
              </div>
              {showCustDropdown && customerResults.length > 0 && (
                <div style={{ position: 'absolute', top: '100%', left: 0, right: 0, backgroundColor: 'var(--surface-color)', border: '1px solid var(--border-color)', borderRadius: '12px', zIndex: 10, boxShadow: 'var(--shadow-lg)', marginTop: '4px', maxHeight: '180px', overflowY: 'auto' }}>
                  {customerResults.map(c => (
                    <div
                      key={c.id}
                      onClick={() => selectCustomer(c)}
                      style={{ padding: '0.6rem 0.85rem', cursor: 'pointer', borderBottom: '1px solid var(--border-color)', fontSize: '0.825rem', transition: 'background-color 0.2s' }}
                      onMouseEnter={e => (e.currentTarget.style.backgroundColor = 'var(--bg-hover)')}
                      onMouseLeave={e => (e.currentTarget.style.backgroundColor = 'transparent')}
                    >
                      <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{[c.first_name, c.last_name].filter(Boolean).join(' ')}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '2px' }}>
                        📞 {c.phone_number} {c.city ? `• 📍 ${c.city}` : ''}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Inputs Grid */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Full Name *</label>
                <div style={{ position: 'relative' }}>
                  <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                    <UserIcon />
                  </div>
                  <input
                    type="text"
                    required
                    value={customerName}
                    onChange={e => setCustomerName(e.target.value)}
                    placeholder="Aamir Siddiqui"
                    style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                  />
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Phone Number *</label>
                <div style={{ position: 'relative' }}>
                  <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                    <PhoneIcon />
                  </div>
                  <input
                    type="text"
                    required
                    value={customerPhone}
                    onChange={e => setCustomerPhone(e.target.value)}
                    placeholder="9876543210"
                    style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                  />
                </div>
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Email Address</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                  <EmailIcon />
                </div>
                <input
                  type="email"
                  value={customerEmail}
                  onChange={e => setCustomerEmail(e.target.value)}
                  placeholder="aamir@example.com"
                  style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                />
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Address Line 1</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                  <MapPinIcon />
                </div>
                <input
                  type="text"
                  value={customerAddress1}
                  onChange={e => setCustomerAddress1(e.target.value)}
                  placeholder="123 Anna Nagar"
                  style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                />
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Address Line 2</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                  <MapPinIcon />
                </div>
                <input
                  type="text"
                  value={customerAddress2}
                  onChange={e => setCustomerAddress2(e.target.value)}
                  placeholder="Suite 401"
                  style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                />
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.3fr 0.9fr', gap: '0.75rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>City</label>
                <div style={{ position: 'relative' }}>
                  <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                    <CityIcon />
                  </div>
                  <input
                    type="text"
                    value={customerCity}
                    onChange={e => setCustomerCity(e.target.value)}
                    placeholder="Chennai"
                    style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                  />
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>State</label>
                <div style={{ position: 'relative' }}>
                  <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center', pointerEvents: 'none' }}>
                    <MapPinIcon />
                  </div>
                  <select
                    value={customerState}
                    onChange={e => setCustomerState(e.target.value)}
                    style={{ width: '100%', padding: '0.55rem 2.25rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem', height: '38px', appearance: 'none', WebkitAppearance: 'none', cursor: 'pointer' }}
                  >
                    {states.map(s => (
                      <option key={s} value={s}>{s || 'Select State'}</option>
                    ))}
                  </select>
                  <div style={{ position: 'absolute', right: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-secondary)', pointerEvents: 'none', display: 'flex', alignItems: 'center' }}>
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                      <polyline points="6 9 12 15 18 9" />
                    </svg>
                  </div>
                </div>
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>ZIP Code</label>
                <div style={{ position: 'relative' }}>
                  <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                    <HashIcon />
                  </div>
                  <input
                    type="text"
                    value={customerZip}
                    onChange={e => setCustomerZip(e.target.value)}
                    placeholder="600001"
                    style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                  />
                </div>
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem', marginTop: '0.5rem' }}>
              <div>
                <span style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.5rem', fontWeight: 600, color: 'var(--text-secondary)' }}>💳 Payment Status</span>
                <div style={{ display: 'flex', background: 'var(--bg-input)', padding: '4px', borderRadius: '10px', border: '1px solid var(--border-color)', gap: '4px' }}>
                  <button
                    type="button"
                    onClick={() => setFinancialStatus('paid')}
                    style={{
                      flex: 1,
                      padding: '0.5rem',
                      borderRadius: '8px',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: '0.825rem',
                      fontWeight: 600,
                      transition: 'all var(--transition-fast)',
                      background: financialStatus === 'paid' ? 'linear-gradient(135deg, #10b981 0%, #059669 100%)' : 'transparent',
                      color: financialStatus === 'paid' ? 'white' : 'var(--text-secondary)',
                      boxShadow: financialStatus === 'paid' ? '0 4px 10px rgba(16,185,129,0.2)' : 'none'
                    }}
                  >
                    Paid
                  </button>
                  <button
                    type="button"
                    onClick={() => setFinancialStatus('pending')}
                    style={{
                      flex: 1,
                      padding: '0.5rem',
                      borderRadius: '8px',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: '0.825rem',
                      fontWeight: 600,
                      transition: 'all var(--transition-fast)',
                      background: financialStatus === 'pending' ? 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' : 'transparent',
                      color: financialStatus === 'pending' ? 'white' : 'var(--text-secondary)',
                      boxShadow: financialStatus === 'pending' ? '0 4px 10px rgba(245,158,11,0.2)' : 'none'
                    }}
                  >
                    Pending
                  </button>
                </div>
              </div>
              <div>
                <span style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.5rem', fontWeight: 600, color: 'var(--text-secondary)' }}>📦 Fulfillment Status</span>
                <div style={{ display: 'flex', background: 'var(--bg-input)', padding: '4px', borderRadius: '10px', border: '1px solid var(--border-color)', gap: '4px' }}>
                  <button
                    type="button"
                    onClick={() => setFulfillmentStatus('fulfilled')}
                    style={{
                      flex: 1,
                      padding: '0.5rem',
                      borderRadius: '8px',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: '0.825rem',
                      fontWeight: 600,
                      transition: 'all var(--transition-fast)',
                      background: fulfillmentStatus === 'fulfilled' ? 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)' : 'transparent',
                      color: fulfillmentStatus === 'fulfilled' ? 'white' : 'var(--text-secondary)',
                      boxShadow: fulfillmentStatus === 'fulfilled' ? '0 4px 10px rgba(99,102,241,0.2)' : 'none'
                    }}
                  >
                    Fulfilled
                  </button>
                  <button
                    type="button"
                    onClick={() => setFulfillmentStatus('unfulfilled')}
                    style={{
                      flex: 1,
                      padding: '0.5rem',
                      borderRadius: '8px',
                      border: 'none',
                      cursor: 'pointer',
                      fontSize: '0.825rem',
                      fontWeight: 600,
                      transition: 'all var(--transition-fast)',
                      background: fulfillmentStatus === 'unfulfilled' ? 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)' : 'transparent',
                      color: fulfillmentStatus === 'unfulfilled' ? 'white' : 'var(--text-secondary)',
                      boxShadow: fulfillmentStatus === 'unfulfilled' ? '0 4px 10px rgba(239,68,68,0.2)' : 'none'
                    }}
                  >
                    Unfulfilled
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Product Selector & List & Summary */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', borderLeft: '1px solid var(--border-color)', paddingLeft: '2rem' }}>
            <h3 style={{ fontSize: '0.95rem', margin: '0 0 0.25rem 0', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--text-secondary)' }}>Order Items</h3>

            {/* Product Search Box */}
            <div ref={prodDropdownRef} style={{ position: 'relative' }}>
              <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>Add Product</label>
              <div style={{ position: 'relative' }}>
                <div style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center' }}>
                  <SearchIcon />
                </div>
                <input
                  type="text"
                  value={productSearch}
                  onChange={e => setProductSearch(e.target.value)}
                  placeholder="Type SKU or Title (e.g. Oud, mi-01)..."
                  style={{ width: '100%', padding: '0.55rem 0.75rem 0.55rem 2.25rem', borderRadius: '10px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                />
              </div>
              {showProdDropdown && productResults.length > 0 && (
                <div style={{ position: 'absolute', top: '100%', left: 0, right: 0, backgroundColor: 'var(--surface-color)', border: '1px solid var(--border-color)', borderRadius: '12px', zIndex: 10, boxShadow: 'var(--shadow-lg)', marginTop: '4px', maxHeight: '180px', overflowY: 'auto' }}>
                  {productResults.map(p => (
                    <div
                      key={p.id}
                      onClick={() => addProduct(p)}
                      style={{ padding: '0.6rem 0.85rem', cursor: 'pointer', borderBottom: '1px solid var(--border-color)', fontSize: '0.825rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', transition: 'background-color var(--transition-fast)' }}
                      onMouseEnter={e => (e.currentTarget.style.backgroundColor = 'var(--bg-hover)')}
                      onMouseLeave={e => (e.currentTarget.style.backgroundColor = 'transparent')}
                    >
                      <div>
                        <span style={{ fontWeight: 600, color: 'var(--accent-color)' }}>{p.mi_sku}</span> <span style={{ color: 'var(--text-primary)' }}>- {p.title}</span>
                      </div>
                      <div style={{ fontSize: '0.75rem', color: p.current_stock > 5 ? 'var(--status-active)' : p.current_stock > 0 ? '#f59e0b' : '#ef4444', fontWeight: 600, backgroundColor: p.current_stock > 5 ? 'var(--status-active-bg)' : p.current_stock > 0 ? 'var(--status-warning-bg)' : 'var(--status-danger-bg)', padding: '2px 8px', borderRadius: '6px' }}>
                        Stock: {p.current_stock}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Line Items Table */}
            <div style={{ flex: 1, overflowY: 'auto', minHeight: '140px', border: '1px solid var(--border-color)', borderRadius: '12px', backgroundColor: 'var(--bg-color)', boxShadow: 'inset 0 2px 4px rgba(0,0,0,0.02)' }}>
              {lineItems.length === 0 ? (
                <div style={{ padding: '2.5rem 1rem', textAlign: 'center', color: 'var(--text-secondary)', fontSize: '0.85rem', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}>
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.4 }}>
                    <circle cx="9" cy="21" r="1" />
                    <circle cx="20" cy="21" r="1" />
                    <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" />
                  </svg>
                  <span>No products added yet. Use the search bar above to add products.</span>
                </div>
              ) : (
                <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-hover)' }}>
                      <th style={{ padding: '0.65rem 0.75rem', textAlign: 'left', fontWeight: 600, color: 'var(--text-secondary)' }}>Item</th>
                      <th style={{ padding: '0.65rem 0.5rem', textAlign: 'center', width: '75px', fontWeight: 600, color: 'var(--text-secondary)' }}>Qty</th>
                      <th style={{ padding: '0.65rem 0.5rem', textAlign: 'right', width: '90px', fontWeight: 600, color: 'var(--text-secondary)' }}>Rate (₹)</th>
                      <th style={{ padding: '0.65rem 0.5rem', textAlign: 'right', width: '85px', fontWeight: 600, color: 'var(--text-secondary)' }}>Disc (₹)</th>
                      <th style={{ padding: '0.65rem 0.75rem', textAlign: 'right', width: '95px', fontWeight: 600, color: 'var(--text-secondary)' }}>Total</th>
                      <th style={{ padding: '0.65rem 0.5rem', width: '40px' }}></th>
                    </tr>
                  </thead>
                  <tbody>
                    {lineItems.map((item, idx) => {
                      const itemTotal = item.quantity * item.price - item.discount;
                      const stockWarning = item.quantity > item.current_stock;
                      return (
                        <tr key={item.mi_sku} style={{ borderBottom: '1px solid var(--border-color)', transition: 'background-color var(--transition-fast)' }} onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-hover)'} onMouseLeave={e => e.currentTarget.style.backgroundColor = 'transparent'}>
                          <td style={{ padding: '0.65rem 0.75rem' }}>
                            <div style={{ fontWeight: 700, color: 'var(--text-primary)' }}>{item.mi_sku}</div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', maxWidth: '140px', marginTop: '2px' }} title={item.title}>
                              {item.title}
                            </div>
                            {stockWarning && (
                              <span style={{ display: 'inline-block', fontSize: '0.65rem', backgroundColor: 'var(--status-danger-bg)', color: 'var(--status-danger)', padding: '1px 6px', borderRadius: '4px', marginTop: '4px', fontWeight: 600 }}>
                                ⚠️ Exceeds stock ({item.current_stock})
                              </span>
                            )}
                          </td>
                          <td style={{ padding: '0.65rem 0.5rem', textAlign: 'center' }}>
                            <input
                              type="number"
                              min="1"
                              value={item.quantity}
                              onChange={e => updateLineItem(idx, 'quantity', parseInt(e.target.value) || 1)}
                              style={{ width: '48px', padding: '0.35rem', borderRadius: '6px', border: '1px solid var(--border-color)', backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)', textAlign: 'center', fontSize: '0.8rem', outline: 'none' }}
                            />
                          </td>
                          <td style={{ padding: '0.65rem 0.5rem', textAlign: 'right' }}>
                            <input
                              type="number"
                              min="0"
                              value={item.price}
                              onChange={e => updateLineItem(idx, 'price', parseFloat(e.target.value) || 0)}
                              style={{ width: '75px', padding: '0.35rem', borderRadius: '6px', border: '1px solid var(--border-color)', backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)', textAlign: 'right', fontSize: '0.8rem', outline: 'none' }}
                            />
                          </td>
                          <td style={{ padding: '0.65rem 0.5rem', textAlign: 'right' }}>
                            <input
                              type="number"
                              min="0"
                              value={item.discount}
                              onChange={e => updateLineItem(idx, 'discount', parseFloat(e.target.value) || 0)}
                              style={{ width: '65px', padding: '0.35rem', borderRadius: '6px', border: '1px solid var(--border-color)', backgroundColor: 'var(--surface-color)', color: 'var(--text-primary)', textAlign: 'right', fontSize: '0.8rem', outline: 'none' }}
                            />
                          </td>
                          <td style={{ padding: '0.65rem 0.75rem', textAlign: 'right', fontWeight: 700, color: 'var(--text-primary)' }}>
                            ₹{itemTotal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                          </td>
                          <td style={{ padding: '0.65rem 0.5rem', textAlign: 'center' }}>
                            <button
                              type="button"
                              onClick={() => removeLineItem(idx)}
                              style={{ background: 'transparent', border: 'none', color: 'var(--status-danger)', cursor: 'pointer', padding: '4px', borderRadius: '6px', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'all var(--transition-fast)' }}
                              onMouseEnter={e => {
                                e.currentTarget.style.backgroundColor = 'var(--status-danger-bg)';
                              }}
                              onMouseLeave={e => {
                                e.currentTarget.style.backgroundColor = 'transparent';
                              }}
                            >
                              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                                <polyline points="3 6 5 6 21 6"></polyline>
                                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                              </svg>
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              )}
            </div>

            {/* Discount & Order Summary Card */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1.25fr', gap: '1.25rem', backgroundColor: 'var(--surface-color)', padding: '1rem 1.25rem', borderRadius: '12px', border: '1px solid var(--border-color)', fontSize: '0.825rem', boxShadow: 'var(--shadow-sm)' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.725rem', marginBottom: '0.35rem', fontWeight: 600, color: 'var(--text-secondary)' }}>🎟️ Order Discount (%)</label>
                <div style={{ position: 'relative' }}>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    value={orderDiscountPercent || ''}
                    onChange={e => setOrderDiscountPercent(parseFloat(e.target.value) || 0)}
                    placeholder="e.g. 10"
                    style={{ width: '100%', padding: '0.45rem 0.75rem', borderRadius: '8px', border: '1px solid var(--border-color)', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)', fontSize: '0.85rem', outline: 'none' }}
                  />
                </div>
              </div>

              {/* Calculations Summary */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', borderLeft: '1px solid var(--border-color)', paddingLeft: '1.25rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
                  <span>Subtotal:</span>
                  <span style={{ fontWeight: 600 }}>₹{subtotal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                </div>
                {orderDiscountPercent > 0 && (
                  <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--status-danger)' }}>
                    <span>Order Discount ({orderDiscountPercent}%):</span>
                    <span style={{ fontWeight: 600 }}>-₹{orderDiscountVal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                  </div>
                )}
                
                <div style={{ borderTop: '1px dashed var(--border-color)', margin: '4px 0' }}></div>
                
                <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                  {isTamilNadu ? (
                    <>
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <span>CGST (9%):</span>
                        <span>₹{cgst.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <span>SGST (9%):</span>
                        <span>₹{sgst.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                      </div>
                    </>
                  ) : (
                    <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                      <span>IGST (18%):</span>
                      <span>₹{igst.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                    </div>
                  )}
                </div>
                
                <div style={{ borderTop: '1px dashed var(--border-color)', margin: '4px 0' }}></div>
                
                <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 800, fontSize: '0.95rem', color: 'var(--accent-color)', padding: '2px 0' }}>
                  <span>Total Payable:</span>
                  <span>₹{total.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                </div>
              </div>
            </div>

            {/* Form Actions */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
              <button
                type="button"
                className="btn-secondary"
                onClick={onClose}
                disabled={isLoading}
                style={{ padding: '0.55rem 1.5rem', borderRadius: '10px', fontSize: '0.85rem', fontWeight: 600, cursor: 'pointer', transition: 'all var(--transition-fast)' }}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="btn-primary"
                disabled={isLoading || lineItems.length === 0}
                style={{ padding: '0.55rem 1.75rem', borderRadius: '10px', fontSize: '0.85rem', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '0.5rem', background: 'linear-gradient(135deg, var(--accent-color) 0%, var(--accent-hover) 100%)', border: 'none', color: 'white', cursor: 'pointer', boxShadow: '0 4px 12px rgba(16,185,129,0.2)', transition: 'all var(--transition-fast)' }}
              >
                {isLoading ? (
                  <>
                    <svg className="animate-spin" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" style={{ animation: 'spin 1s linear infinite' }}>
                      <circle cx="12" cy="12" r="10" strokeDasharray="32" strokeDashoffset="8"></circle>
                    </svg>
                    Creating...
                  </>
                ) : (
                  <>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                    Create Order
                  </>
                )}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};
