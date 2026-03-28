import React, { useState, useEffect } from 'react';
import { API_BASE, fetchWithAuth } from '../api';
import { MobileCard } from '../components/MobileCard';
import { BottomSheet } from '../components/BottomSheet';
import './Customers.css';

interface Customer {
  id: string | number;
  phone_number: string;
  first_name: string;
  last_name: string;
  city: string;
  total_orders: number;
}

interface Template {
  name: string;
  language: string;
  status: string;
  category: string;
}

export const Customers: React.FC = () => {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [isTemplatePickerOpen, setIsTemplatePickerOpen] = useState(false);

  const fetchCustomers = async () => {
    setIsLoading(true);
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/customers?limit=20`);
      const data = await res.json();
      if (data.success) {
        setCustomers(data.customers);
      }
    } catch (err) {
      console.error('Error fetching customers:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchTemplates = async () => {
    try {
      const res = await fetchWithAuth(`${API_BASE}/api/automation/templates`);
      const data = await res.json();
      if (data.success) {
        setTemplates(data.templates || []);
      }
    } catch (err) {
      console.error('Error fetching templates:', err);
    }
  };

  useEffect(() => {
    fetchCustomers();
    fetchTemplates();
  }, []);

  const handleSendWhatsApp = async (templateName: string) => {
    if (!selectedCustomer) return;

    try {
      const res = await fetchWithAuth(`${API_BASE}/api/automation/send-template`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          phone: selectedCustomer.phone_number,
          template_name: templateName,
          customer_id: selectedCustomer.id
        })
      });
      const data = await res.json();
      if (data.success) {
        alert('WhatsApp sent successfully!');
        setIsTemplatePickerOpen(false);
        setSelectedCustomer(null);
      } else {
        alert('Failed to send: ' + data.message);
      }
    } catch (err) {
      console.error('Error sending template:', err);
      alert('Network error sending WhatsApp.');
    }
  };

  return (
    <div className="customers-container">
      <h2 className="section-title">Customers</h2>
      <p className="section-subtitle">Manage your customer relationships</p>

      <div className="customer-list">
        {isLoading ? (
          <div className="loader">Loading customers...</div>
        ) : (
          customers.map(customer => (
            <MobileCard key={customer.id} className="customer-card">
              <div className="customer-info">
                <span className="customer-name">{customer.first_name} {customer.last_name}</span>
                <span className="customer-detail">{customer.phone_number} • {customer.city || 'Unknown City'}</span>
                <span className="customer-stat">{customer.total_orders} Orders</span>
              </div>
              <button
                className="whatsapp-btn"
                onClick={() => {
                  setSelectedCustomer(customer);
                  setIsTemplatePickerOpen(true);
                }}
              >
                💬 Send
              </button>
            </MobileCard>
          ))
        )}
      </div>

      <BottomSheet
        isOpen={isTemplatePickerOpen}
        onClose={() => setIsTemplatePickerOpen(false)}
        title="Select Template"
      >
        <div className="template-picker">
          <p className="picker-hint">Sending to: <strong>{selectedCustomer?.first_name}</strong></p>
          <div className="template-grid">
            {templates.length === 0 ? (
              <div className="empty-templates">No WhatsApp templates found.</div>
            ) : (
              templates.map(tpl => (
                <button
                  key={tpl.name}
                  className="template-card"
                  onClick={() => handleSendWhatsApp(tpl.name)}
                >
                  <span className="tpl-category">{tpl.category}</span>
                  <span className="tpl-name">{tpl.name.replace(/_/g, ' ')}</span>
                  <span className="tpl-lang">{tpl.language}</span>
                </button>
              ))
            )}
          </div>
        </div>
      </BottomSheet>
    </div>
  );
};
