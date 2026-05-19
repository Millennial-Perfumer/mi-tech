import React, { useState, useEffect } from 'react';
import { useToast } from './ToastContext';
import { API_BASE } from './api';

interface LineItem {
  id: string;
  title: string;
  sku: string;
  quantity: number;
  price: string;
  discount: string;
}

interface Order {
  id: string;
  order_number: string;
  total_price: string;
  subtotal_price: string;
  total_tax: string;
  currency: string;
  financial_status: string;
  fulfillment_status: string;
  status: string;
  created_at: string;
  customer_name: string;
  customer_first_name: string;
  customer_last_name: string;
  customer_email: string;
  customer_phone: string;
  customer_address1: string;
  customer_address2: string;
  customer_city: string;
  customer_state: string;
  customer_zip: string;
  customer_country: string;
  line_items?: LineItem[];
  tracking_number?: string;
  shipping_company?: string;
  tracking_url?: string;
  feedback_status_id?: number;
  feedback_sent_at?: string;
}

interface AutomationMessage {
  id: number;
  template_name: string;
  phone_number: string;
  status: string;
  sent_at: string;
  delivered_at?: string;
  read_at?: string;
  error_message?: string;
}


interface OrderDetailsModalProps {
  isOpen: boolean;
  onClose: () => void;
  orderId: string | number;
  fetchWithAuth: (url: string, options?: RequestInit) => Promise<any>;
  userRole?: string;
  onOrderUpdated?: () => void;
}

const OrderDetailsModal: React.FC<OrderDetailsModalProps> = ({
  isOpen,
  onClose,
  orderId,
  fetchWithAuth,
  userRole,
  onOrderUpdated
}) => {
  const { success, error } = useToast();
  const [order, setOrder] = useState<Order | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [formData, setFormData] = useState<Partial<Order>>({});
  const [activeTab, setActiveTab] = useState<'details' | 'messages'>('details');
  const [messages, setMessages] = useState<AutomationMessage[]>([]);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);

  useEffect(() => {
    if (isOpen && orderId) {
      fetchOrderDetails();
    }
  }, [isOpen, orderId]);

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen && !isSaving && !isDownloading) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen, isSaving, isDownloading, onClose]);

  const fetchOrderDetails = async () => {
    setIsLoading(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders?id=${orderId}`);
      const data = await response.json();
      if (data.success) {
        setOrder(data.order);
        setFormData(data.order);
        // Also fetch messages
        fetchMessages();
      } else {
        error('Failed to fetch order details');
      }
    } catch (err) {
      error('An error occurred while fetching order details');
    } finally {
      setIsLoading(false);
    }
  };

  const fetchMessages = async () => {
    setIsLoadingMessages(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/automation/whatsapp/messages/order?order_id=${orderId}`);
      const data = await response.json();
      setMessages(data || []);
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      setIsLoadingMessages(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const updateData = {
        customer_first_name: formData.customer_first_name,
        customer_last_name: formData.customer_last_name,
        customer_email: formData.customer_email,
        customer_phone: formData.customer_phone,
        customer_address1: formData.customer_address1,
        customer_address2: formData.customer_address2,
        customer_city: formData.customer_city,
        customer_state: formData.customer_state,
        customer_zip: formData.customer_zip,
        customer_country: formData.customer_country,
      };

      const response = await fetchWithAuth(`${API_BASE}/api/orders?id=${orderId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updateData),
      });

      const result = await response.json();

      if (result.success) {
        success('Order updated and synced with Shopify');
        setIsEditing(false);
        fetchOrderDetails();
        if (onOrderUpdated) onOrderUpdated();
      } else {
        error(result.message || 'Failed to update order');
      }
    } catch (err) {
      error('An error occurred while saving changes');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDownloadInvoice = async () => {
    setIsDownloading(true);
    try {
      const response = await fetchWithAuth(`${API_BASE}/api/orders/invoice?id=${orderId}`);
      if (!response.ok) {
        throw new Error('Failed to download invoice');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `invoice-${order?.order_number || orderId}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      success('Invoice downloaded successfully');
    } catch (err) {
      console.error('Error downloading invoice:', err);
      error('Failed to download invoice. Please try again.');
    } finally {
      setIsDownloading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="premium-modal order-details-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '740px', width: '95%', padding: '1.25rem 1.5rem' }}>


        <div className="modal-header-icon" style={{ 
          background: 'linear-gradient(135deg, var(--accent-color), var(--status-active))',
          width: '45px',
          height: '45px',
          top: '-22px'
        }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 0 1-8 0"/>
          </svg>
        </div>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
          <div>
            <h2 style={{ fontSize: '1.15rem', fontWeight: 700, margin: 0 }}>Order Details</h2>


            <p style={{ color: 'var(--text-secondary)', margin: '0.25rem 0 0 0' }}>
              {isLoading ? 'Loading...' : `Order ${order?.order_number} • ${new Date(order?.created_at || '').toLocaleDateString()} ${new Date(order?.created_at || '').toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: true })}`}
            </p>
          </div>
          {!isLoading && order && (
             <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
                <span className={`badge-pill badge-pill-${order.financial_status === 'paid' ? 'success' : 'warning'}`}>
                  <span className="dot"></span> {order.financial_status?.toUpperCase()}
                </span>
                <span className={`badge-pill badge-pill-${order.status === 'CANCELLED' ? 'danger' : (order.fulfillment_status === 'fulfilled' ? 'gray' : 'yellow')}`}>
                  <span className="dot"></span> {order.status === 'CANCELLED' ? 'CANCELLED' : (order.fulfillment_status || 'UNFULFILLED').toUpperCase()}
                </span>
             </div>
          )}
        </div>

        {/* Tab Navigation */}
        <div className="modal-tabs" style={{ display: 'flex', gap: '1.5rem', marginBottom: '1rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.4rem' }}>
          <button 
            className={`tab-btn ${activeTab === 'details' ? 'active' : ''}`}
            onClick={() => setActiveTab('details')}
          >
            Details
          </button>
          <button 
            className={`tab-btn ${activeTab === 'messages' ? 'active' : ''}`}
            onClick={() => setActiveTab('messages')}
          >
            Message History
            {messages.length > 0 && <span className="tab-count">{messages.length}</span>}
          </button>
        </div>
        
        {isLoading ? (

          <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
            <div className="loading-spinner"></div>
            <p>Fetching full order history...</p>
          </div>
        ) : !order ? (
          <div style={{ padding: '2rem', textAlign: 'center' }}>Order not found.</div>
        ) : (
          activeTab === 'details' ? (
            <div className="modal-content-scroll" style={{ maxHeight: '60vh', overflowY: 'auto', paddingRight: '0.5rem' }}>
              <div className="form-row" style={{ display: 'grid', gridTemplateColumns: '1.1fr 0.9fr', gap: '1.25rem' }}>


                
                {/* Customer Info Section */}
                <div className="details-section">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.6rem' }}>
                    <h3 style={{ fontSize: '0.85rem', fontWeight: 600, margin: 0, color: 'var(--accent-color)' }}>Customer & Shipping</h3>


                    {userRole === 'admin' && (
                      <button 
                        className="btn-icon-minimal" 
                        onClick={() => setIsEditing(!isEditing)}
                        aria-label={isEditing ? "Cancel Edit" : "Edit Customer Details"}
                        title={isEditing ? "Cancel Edit" : "Edit Customer Details"}
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          {isEditing ? <path d="M18 6L6 18M6 6l12 12"/> : <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>}
                        </svg>
                      </button>
                    )}
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>

                    <div className="form-row" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                      <div className="input-group">
                        <label>First Name</label>
                        <input 
                          name="customer_first_name" 
                          value={formData.customer_first_name || ''} 
                          onChange={handleInputChange} 
                          disabled={!isEditing}
                          className={!isEditing ? 'input-readonly' : ''}
                        />
                      </div>
                      <div className="input-group">
                        <label>Last Name</label>
                        <input 
                          name="customer_last_name" 
                          value={formData.customer_last_name || ''} 
                          onChange={handleInputChange} 
                          disabled={!isEditing}
                          className={!isEditing ? 'input-readonly' : ''}
                        />
                      </div>
                    </div>

                    <div className="input-group">
                      <label>Email Address</label>
                      <input 
                        name="customer_email" 
                        value={formData.customer_email || ''} 
                        onChange={handleInputChange} 
                        disabled={!isEditing}
                        className={!isEditing ? 'input-readonly' : ''}
                      />
                    </div>

                    <div className="input-group">
                      <label>Phone Number</label>
                      <input 
                        name="customer_phone" 
                        value={formData.customer_phone || ''} 
                        onChange={handleInputChange} 
                        disabled={!isEditing}
                        className={!isEditing ? 'input-readonly' : ''}
                      />
                    </div>

                    <div className="input-group">
                      <label>Address Line 1</label>
                      <input 
                        name="customer_address1" 
                        value={formData.customer_address1 || ''} 
                        onChange={handleInputChange} 
                        disabled={!isEditing}
                        className={!isEditing ? 'input-readonly' : ''}
                      />
                    </div>

                    <div className="input-group">
                      <label>Address Line 2 (Optional)</label>
                      <input 
                        name="customer_address2" 
                        value={formData.customer_address2 || ''} 
                        onChange={handleInputChange} 
                        disabled={!isEditing}
                        className={!isEditing ? 'input-readonly' : ''}
                      />
                    </div>

                    <div className="form-row" style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: '1rem' }}>
                      <div className="input-group">
                        <label>City</label>
                        <input 
                          name="customer_city" 
                          value={formData.customer_city || ''} 
                          onChange={handleInputChange} 
                          disabled={!isEditing}
                          className={!isEditing ? 'input-readonly' : ''}
                        />
                      </div>
                      <div className="input-group">
                        <label>Zip/Pincode</label>
                        <input 
                          name="customer_zip" 
                          value={formData.customer_zip || ''} 
                          onChange={handleInputChange} 
                          disabled={!isEditing}
                          className={!isEditing ? 'input-readonly' : ''}
                        />
                      </div>
                    </div>

                    <div className="form-row" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                      <div className="input-group">
                        <label>State</label>
                        <input 
                          name="customer_state" 
                          value={formData.customer_state || ''} 
                          onChange={handleInputChange} 
                          disabled={!isEditing}
                          className={!isEditing ? 'input-readonly' : ''}
                        />
                      </div>
                      <div className="input-group">
                        <label>Country</label>
                        <input 
                          name="customer_country" 
                          value={formData.customer_country || ''} 
                          onChange={handleInputChange} 
                          disabled={!isEditing}
                          className={!isEditing ? 'input-readonly' : ''}
                        />
                      </div>
                    </div>
                  </div>
                </div>

                {/* Order Summary & Items */}
                <div className="details-section">
                  <h3 style={{ fontSize: '0.85rem', fontWeight: 600, marginBottom: '0.6rem', color: 'var(--accent-color)' }}>Order Items</h3>


                  <div className="items-container" style={{ background: 'var(--bg-input)', borderRadius: '12px', padding: '0.5rem', border: '1px solid var(--border-color)' }}>
                    {order.line_items && order.line_items.length > 0 ? (
                      order.line_items.map((item, idx) => (
                        <div key={item.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem', borderBottom: idx === order.line_items!.length - 1 ? 'none' : '1px solid var(--border-color)' }}>

                          <div style={{ flex: 1 }}>
                            <div style={{ fontWeight: 600, fontSize: '0.9rem' }}>{item.title}</div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>SKU: {item.sku || 'N/A'}</div>
                          </div>
                          <div style={{ textAlign: 'right', marginLeft: '1rem' }}>
                            <div style={{ fontWeight: 600 }}>₹{item.price}</div>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)' }}>Qty: {item.quantity}</div>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-tertiary)' }}>No items found.</div>
                    )}
                  </div>

                  <div style={{ marginTop: '1rem', padding: '1rem 0.5rem 0', borderTop: '2px dashed var(--border-color)' }}>

                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                        <span style={{ color: 'var(--text-secondary)' }}>Subtotal</span>
                        <span>₹{order.subtotal_price}</span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                        <span style={{ color: 'var(--text-secondary)' }}>Tax</span>
                        <span>₹{order.total_tax}</span>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 700, fontSize: '1.1rem', marginTop: '1rem', color: 'var(--text-primary)' }}>
                        <span>Total</span>
                        <span>₹{order.total_price}</span>
                    </div>
                  </div>

                  {order.tracking_number && (
                    <div style={{ marginTop: '1.25rem', padding: '0.75rem', background: 'var(--bg-input)', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
                      <div style={{ fontSize: '0.7rem', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '0.4rem' }}>Shipment Tracking</div>

                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div>
                          <div style={{ fontWeight: 600 }}>{order.shipping_company}</div>
                          <div style={{ fontSize: '0.875rem' }}>{order.tracking_number}</div>
                        </div>
                        {order.tracking_url && (
                          <a href={order.tracking_url} target="_blank" rel="noreferrer" className="btn-icon-minimal" aria-label="Open tracking URL">
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3"/>
                            </svg>
                          </a>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="modal-content-scroll" style={{ maxHeight: '60vh', overflowY: 'auto' }}>


              <div className="messages-history" style={{ paddingRight: '0.5rem' }}>
                {isLoadingMessages ? (
                  <div style={{ padding: '2rem', textAlign: 'center' }}>
                    <div className="loading-spinner"></div>
                    <p>Loading message history...</p>
                  </div>
                ) : messages.length === 0 ? (
                  <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-tertiary)', background: 'var(--bg-input)', borderRadius: '12px' }}>
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round" style={{ marginBottom: '1rem', opacity: 0.5 }}>
                      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                    </svg>
                    <p>No messages have been sent for this order yet.</p>
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>

                    {messages.map((msg) => (
                      <div key={msg.id} className="message-card premium-hover">
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                          <div>
                            <div style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '0.25rem' }}>{msg.template_name}</div>
                            <div style={{ fontSize: '0.8rem', color: 'var(--text-tertiary)' }}>Sent to: {msg.phone_number}</div>
                          </div>
                          <span className={`status-badge status-${msg.status}`}>
                            {msg.status.toUpperCase()}
                          </span>
                        </div>
                        
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.75rem', marginTop: '0.75rem', paddingTop: '0.75rem', borderTop: '1px solid var(--border-color)', fontSize: '0.75rem' }}>

                          <div>
                            <label style={{ display: 'block', color: 'var(--text-tertiary)', marginBottom: '0.25rem' }}>Sent At</label>
                            <div style={{ color: 'var(--text-secondary)' }}>{new Date(msg.sent_at).toLocaleString()}</div>
                          </div>
                          {msg.delivered_at && (
                            <div>
                              <label style={{ display: 'block', color: 'var(--text-tertiary)', marginBottom: '0.25rem' }}>Delivered At</label>
                              <div style={{ color: 'var(--text-secondary)' }}>{new Date(msg.delivered_at).toLocaleString()}</div>
                            </div>
                          )}
                          {msg.read_at && (
                            <div>
                              <label style={{ display: 'block', color: 'var(--text-tertiary)', marginBottom: '0.25rem' }}>Read At</label>
                              <div style={{ color: 'var(--text-secondary)' }}>{new Date(msg.read_at).toLocaleString()}</div>
                            </div>
                          )}
                        </div>
                        
                        {msg.error_message && (
                          <div style={{ marginTop: '0.75rem', padding: '0.5rem 0.75rem', background: 'rgba(239, 68, 68, 0.1)', borderLeft: '3px solid var(--status-danger)', borderRadius: '4px', fontSize: '0.8rem', color: 'var(--status-danger)' }}>
                            {msg.error_message}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )
        )}


        <div className="modal-actions" style={{ marginTop: '1.25rem', paddingTop: '1rem', borderTop: '1px solid var(--border-color)' }}>


          <button className="btn-secondary" onClick={onClose} disabled={isSaving || isDownloading}>Close</button>
          {isEditing && (
            <button 
              className="btn-primary" 
              onClick={handleSave} 
              disabled={isSaving || isDownloading}
              style={{ minWidth: '160px' }}
            >
              {isSaving ? (
                <span style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <div className="loading-spinner" style={{ width: '14px', height: '14px', borderWidth: '2px' }}></div>
                  Saving to Shopify...
                </span>
              ) : 'Save Changes'}
            </button>
          )}
          {!isEditing && !isLoading && (
             <button 
               className="btn-primary" 
               onClick={handleDownloadInvoice}
               disabled={isDownloading || isSaving}
               style={{ minWidth: '180px' }}
             >
               {isDownloading ? (
                 <span style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', justifyContent: 'center' }}>
                   <div className="loading-spinner" style={{ width: '14px', height: '14px', borderWidth: '2px', margin: 0 }}></div>
                   Downloading...
                 </span>
               ) : 'Download GST Invoice'}
             </button>
          )}
        </div>
      </div>
      
      <style>{`
        .order-details-modal .input-group {
          display: flex;
          flex-direction: column;
          gap: 0.4rem;
        }
        .order-details-modal label {
          font-size: 0.75rem;
          font-weight: 600;
          color: var(--text-tertiary);
          text-transform: uppercase;
          letter-spacing: 0.025em;
        }
        .order-details-modal input {
          padding: 0.65rem;

          border-radius: 8px;
          border: 1px solid var(--border-color);
          background: var(--bg-input);
          color: var(--text-primary);
          font-size: 0.9rem;
          transition: all 0.2s;
        }
        .order-details-modal input:focus {
          outline: none;
          border-color: var(--accent-color);
          box-shadow: 0 0 0 2px var(--accent-subtle);
        }
        .order-details-modal input.input-readonly {
          background: transparent;
          border-color: transparent;
          padding-left: 0;
          cursor: default;
          font-weight: 500;
        }
        .order-details-modal input:disabled {
          opacity: 1;
        }
        .loading-spinner {
          display: inline-block;
          width: 24px;
          height: 24px;
          border: 3px solid rgba(255,255,255,0.1);
          border-radius: 50%;
          border-top-color: var(--accent-color);
          animation: spin 1s ease-in-out infinite;
          margin-bottom: 1rem;
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
        
        .tab-btn {
          background: none;
          border: none;
          padding: 0.5rem 0;
          color: var(--text-tertiary);
          font-weight: 600;
          font-size: 0.9rem;
          cursor: pointer;
          position: relative;
          transition: color 0.2s;
        }
        .tab-btn:hover {
          color: var(--text-secondary);
        }
        .tab-btn.active {
          color: var(--accent-color);
        }
        .tab-btn.active::after {
          content: '';
          position: absolute;
          bottom: -0.5rem;
          left: 0;
          width: 100%;
          height: 2px;
          background: var(--accent-color);
          box-shadow: 0 0 10px var(--accent-subtle);
        }
        .tab-count {
          background: var(--bg-input);
          color: var(--text-tertiary);
          font-size: 0.7rem;
          padding: 0.1rem 0.4rem;
          border-radius: 6px;
          margin-left: 0.5rem;
          border: 1px solid var(--border-color);
        }
        .tab-btn.active .tab-count {
          background: var(--accent-subtle);
          color: var(--accent-color);
          border-color: var(--accent-subtle);
        }
        
        .message-card {
          background: var(--bg-input);
          border: 1px solid var(--border-color);
          border-radius: 12px;
          padding: 1rem;
          transition: all 0.2s;
        }

        .message-card:hover {
          border-color: var(--accent-subtle);
          transform: translateY(-2px);
          box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        
        .status-badge {
          font-size: 0.7rem;
          font-weight: 700;
          padding: 0.25rem 0.5rem;
          border-radius: 4px;
        }
        .status-sent { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
        .status-delivered { background: rgba(16, 185, 129, 0.1); color: #10b981; }
        .status-read { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
        .status-failed { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
      `}</style>
    </div>
  );
};

export default OrderDetailsModal;
