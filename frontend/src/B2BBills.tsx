import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
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

function parseAddressString(addressStr: string) {
	if (!addressStr) return { street: '', city: '', pincode: '' };
	const parts = addressStr.split(',').map(s => s.trim());
	let pincode = '';
	let city = '';
	let streetParts = [...parts];

	const lastPart = parts[parts.length - 1] || '';
	const pinMatch = lastPart.match(/\b\d{6}\b/);
	if (pinMatch) {
		pincode = pinMatch[0];
		streetParts.pop();
	}

	if (streetParts.length >= 2) {
		const lastIndex = streetParts.length - 1;
		const possibleState = streetParts[lastIndex].toLowerCase();
		// If the second to last part looks like state candidate, pop it too
		const isState = Object.values(GST_STATE_MAP).some(s => s.toLowerCase() === possibleState);
		if (isState) {
			streetParts.pop();
		}
	}

	if (streetParts.length >= 1) {
		city = streetParts.pop() || '';
	}

	return {
		street: streetParts.join(', '),
		city: city,
		pincode: pincode
	};
}

interface B2BBillsProps {
	fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
	userRole?: string;
}

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

interface B2BInvoice {
	id?: number;
	invoice_number?: string;
	invoice_sequence?: number;
	financial_year?: string;
	order_number?: string;
	invoice_date: string;
	terms?: string;
	due_date?: string;
	salesperson?: string;
	subject?: string;
	customer_id?: number;
	customer_gstin: string;
	customer_name: string;
	customer_email?: string;
	customer_phone?: string;
	customer_state: string;
	customer_state_code: string;
	customer_address: string;
	seller_gstin?: string;
	seller_name?: string;
	seller_state?: string;
	seller_state_code?: string;
	seller_address?: string;
	subtotal_price: number;
	discount_percent: number;
	discount_amount: number;
	cgst_rate: number;
	cgst_amount: number;
	sgst_rate: number;
	sgst_amount: number;
	igst_rate: number;
	igst_amount: number;
	tds_tcs_type: string;
	tds_tcs_rate: number;
	tds_tcs_amount: number;
	transportation_charge: number;
	total_price: number;
	status: string;
	payment_status: string;
	paid_amount: number;
	balance_amount: number;
	payment_date?: string;
	payment_method?: string;
	customer_notes?: string;
	items: B2BItem[];
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
	notes?: string;
}

interface B2BPaymentTerm {
	id?: number;
	name: string;
	due_days: number;
}

interface B2BInventoryItem {
	id: number;
	mi_sku: string;
	title: string;
	description: string;
	current_stock: number;
	price?: number;
	hsn_code?: string;
}

const DEFAULT_CUSTOMER_NOTES = `Thanks for your business.

Payment Terms: Full payment is required before the due date mentioned on the invoice. 

No Refunds & Returns: Due to the nature of our products, we do not accept returns or provide refunds once the item has been opened or used. If the product remains sealed and unused, you may contact us within 7 days for return eligibility, subject to approval.

Damaged or Incorrect Items: If you receive a damaged or incorrect product, please contact us within 48 hours of delivery with photographic evidence for a replacement or resolution.

Shipping & Delivery: We aim to deliver orders promptly, but delays due to courier services, customs, or unforeseen circumstances are beyond our control. Tracking details will be provided once your order is shipped.

Intellectual Property: All branding, packaging, and product names are trademarks of Millennial Perfumer™ and may not be reproduced without permission.`;

export function B2BBills({ fetchWithAuth, userRole = 'read' }: B2BBillsProps) {
	// Navigation State
	const [activeSubTab, setActiveSubTab] = useState<'invoices' | 'customers'>('invoices');
	const [viewMode, setViewMode] = useState<'list' | 'create' | 'edit' | 'preview'>('list');

	// Inventory products state
	const [inventoryProducts, setInventoryProducts] = useState<B2BInventoryItem[]>([]);

	// Invoices List state
	const [invoices, setInvoices] = useState<B2BInvoice[]>([]);
	const [invoiceSearch, setInvoiceSearch] = useState('');
	const [invoiceStatusFilter, setInvoiceStatusFilter] = useState('');
	const [selectedInvoice, setSelectedInvoice] = useState<B2BInvoice | null>(null);

	// Customers List state
	const [customers, setCustomers] = useState<B2BCustomer[]>([]);
	const [customerSearch, setCustomerSearch] = useState('');
	const [showCustomerModal, setShowCustomerModal] = useState(false);
	const [editingCustomer, setEditingCustomer] = useState<B2BCustomer | null>(null);
	const [sameAsBilling, setSameAsBilling] = useState(true);

	// Address breakdown states
	const [billingStreet, setBillingStreet] = useState('');
	const [billingCity, setBillingCity] = useState('');
	const [billingPincode, setBillingPincode] = useState('');

	const [shippingStreet, setShippingStreet] = useState('');
	const [shippingCity, setShippingCity] = useState('');
	const [shippingPincode, setShippingPincode] = useState('');

	// Payment Terms state
	const [paymentTerms, setPaymentTerms] = useState<B2BPaymentTerm[]>([]);
	const [showPaymentTermModal, setShowPaymentTermModal] = useState(false);
	const [newTermName, setNewTermName] = useState('');
	const [newTermDays, setNewTermDays] = useState(0);

	useEffect(() => {
		if (editingCustomer) {
			const parsedBilling = parseAddressString(editingCustomer.billing_address);
			setBillingStreet(parsedBilling.street);
			setBillingCity(parsedBilling.city);
			setBillingPincode(parsedBilling.pincode);

			const parsedShipping = parseAddressString(editingCustomer.shipping_address || '');
			setShippingStreet(parsedShipping.street);
			setShippingCity(parsedShipping.city);
			setShippingPincode(parsedShipping.pincode);
		} else {
			setBillingStreet('');
			setBillingCity('');
			setBillingPincode('');
			setShippingStreet('');
			setShippingCity('');
			setShippingPincode('');
		}
	}, [editingCustomer]);

	// Payment Log modal
	const [showPaymentModal, setShowPaymentModal] = useState(false);
	const [paymentInvoice, setPaymentInvoice] = useState<B2BInvoice | null>(null);
	const [paymentAmount, setPaymentAmount] = useState(0);
	const [paymentMethod, setPaymentMethod] = useState('Bank Transfer');

	// Creator / Editor Form State
	const [formInvoice, setFormInvoice] = useState<B2BInvoice>({
		invoice_date: new Date().toISOString().split('T')[0],
		customer_gstin: '',
		customer_name: '',
		customer_state: '',
		customer_state_code: '',
		customer_address: '',
		subtotal_price: 0,
		discount_percent: 0,
		discount_amount: 0,
		cgst_rate: 0,
		cgst_amount: 0,
		sgst_rate: 0,
		sgst_amount: 0,
		igst_rate: 0,
		igst_amount: 0,
		tds_tcs_type: 'NONE',
		tds_tcs_rate: 0,
		tds_tcs_amount: 0,
		transportation_charge: 0,
		total_price: 0,
		status: 'DRAFT',
		payment_status: 'UNPAID',
		paid_amount: 0,
		balance_amount: 0,
		customer_notes: DEFAULT_CUSTOMER_NOTES,
		items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
	});

	// Load Data
	useEffect(() => {
		loadInvoices();
		loadCustomers();
		loadInventoryProducts();
		loadPaymentTerms();
	}, []);

	const loadInvoices = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices`);
			if (res.ok) {
				const data = await res.json();
				setInvoices(data || []);
			}
		} catch (err) {
			console.error('Failed to load invoices:', err);
		}
	};

	const loadInventoryProducts = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/inventory`);
			if (res.ok) {
				const data = await res.json();
				setInventoryProducts(data || []);
			}
		} catch (err) {
			console.error('Failed to load inventory products:', err);
		}
	};

	const loadCustomers = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers`);
			if (res.ok) {
				const data = await res.json();
				setCustomers(data || []);
			}
		} catch (err) {
			console.error('Failed to load customers:', err);
		}
	};

	const loadPaymentTerms = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/payment-terms`);
			if (res.ok) {
				const data = await res.json();
				setPaymentTerms(data || []);
			}
		} catch (err) {
			console.error('Failed to load payment terms:', err);
		}
	};

	const handleTermsChange = (termName: string, currentInvoiceDate?: string) => {
		const invoiceDate = currentInvoiceDate || formInvoice.invoice_date || new Date().toISOString().split('T')[0];
		const matchedTerm = paymentTerms.find(t => t.name === termName);

		let computedDueDate = '';
		if (matchedTerm) {
			const date = new Date(invoiceDate);
			date.setDate(date.getDate() + matchedTerm.due_days);
			computedDueDate = date.toISOString().split('T')[0];
		}

		setFormInvoice({
			...formInvoice,
			terms: termName,
			due_date: computedDueDate
		});
	};

	const handleSavePaymentTerm = async () => {
		if (!newTermName.trim()) {
			alert('Term name is required');
			return;
		}
		if (newTermDays < 0) {
			alert('Due Days cannot be negative');
			return;
		}

		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/payment-terms`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: newTermName.trim(), due_days: Number(newTermDays) })
			});
			if (res.ok) {
				const term = await res.json();
				setNewTermName('');
				setNewTermDays(0);
				setShowPaymentTermModal(false);

				// Reload the terms list
				const refreshRes = await fetchWithAuth(`${API_BASE}/api/b2b/payment-terms`);
				if (refreshRes.ok) {
					const data = await refreshRes.json();
					setPaymentTerms(data || []);

					// Automatically select the newly created term and compute its due date
					const matchedTerm = (data || []).find((t: any) => t.name === term.name);
					const invoiceDate = formInvoice.invoice_date || new Date().toISOString().split('T')[0];
					let computedDueDate = '';
					if (matchedTerm) {
						const date = new Date(invoiceDate);
						date.setDate(date.getDate() + matchedTerm.due_days);
						computedDueDate = date.toISOString().split('T')[0];
					}
					setFormInvoice({
						...formInvoice,
						terms: term.name,
						due_date: computedDueDate
					});
				}
			} else {
				const text = await res.text();
				alert('Failed to save payment term: ' + text);
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving payment term');
		}
	};

	// Save customer
	const handleSaveCustomer = async (cust: B2BCustomer) => {
		try {
			const finalCust = { ...cust };

			// Assemble Billing Address
			const billingParts = [];
			if (billingStreet.trim()) billingParts.push(billingStreet.trim());
			if (billingCity.trim()) billingParts.push(billingCity.trim());
			if (cust.state.trim()) billingParts.push(cust.state.trim());
			if (billingPincode.trim()) billingParts.push(billingPincode.trim());

			finalCust.billing_address = billingParts.join(', ');

			if (sameAsBilling) {
				finalCust.shipping_address = finalCust.billing_address;
			} else {
				// Assemble Shipping Address
				const shippingParts = [];
				if (shippingStreet.trim()) shippingParts.push(shippingStreet.trim());
				if (shippingCity.trim()) shippingParts.push(shippingCity.trim());
				if (cust.state.trim()) shippingParts.push(cust.state.trim());
				if (shippingPincode.trim()) shippingParts.push(shippingPincode.trim());

				finalCust.shipping_address = shippingParts.join(', ');
			}
			const method = finalCust.id ? 'PUT' : 'POST';
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(finalCust)
			});
			if (res.ok) {
				setShowCustomerModal(false);
				setEditingCustomer(null);
				loadCustomers();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save customer');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving customer');
		}
	};

	// Delete Customer
	const handleDeleteCustomer = async (id: number) => {
		if (!confirm('Are you sure you want to delete this customer?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				loadCustomers();
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Save B2B Invoice (Draft / Active)
	const handleSaveInvoice = async (asDraft: boolean) => {
		try {
			const inv = { ...formInvoice };
			const method = inv.id ? 'PUT' : 'POST';

			// Save Invoice first
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(inv)
			});

			if (res.ok) {
				const savedInv = await res.json();

				// If clicking "Save and Send" (Activate), trigger Issue API
				if (!asDraft) {
					const issueRes = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/issue?id=${savedInv.id}`, {
						method: 'POST'
					});
					if (!issueRes.ok) {
						const text = await issueRes.text();
						alert('Saved as draft, but activation failed: ' + text);
					}
				}

				setViewMode('list');
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save invoice');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving invoice');
		}
	};

	// Delete Draft Invoice
	const handleDeleteInvoice = async (id: number) => {
		if (!confirm('Are you sure you want to delete this draft invoice?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Issue Draft Invoice
	const handleIssueInvoice = async (id: number) => {
		if (!confirm('Are you sure you want to activate/issue this invoice? This locks modifications and generates the sequential B2B invoice number.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/issue?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Cancel issued invoice
	const handleCancelInvoice = async (id: number) => {
		if (!confirm('Are you sure you want to CANCEL this issued invoice? This removes it from active revenue and tax summaries historically.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/cancel?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Save Payment details
	const handleSavePayment = async () => {
		if (!paymentInvoice) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/payment`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					id: paymentInvoice.id,
					paid_amount: paymentAmount,
					payment_method: paymentMethod
				})
			});
			if (res.ok) {
				setShowPaymentModal(false);
				setPaymentInvoice(null);
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Dynamic Totals calculation helper in form
	const recalculateTotals = (updatedInvoice: B2BInvoice) => {
		let subtotal = 0;
		updatedInvoice.items.forEach(item => {
			item.amount = (item.quantity || 0) * (item.rate || 0);
			subtotal += item.amount;
		});
		updatedInvoice.subtotal_price = subtotal;

		if (updatedInvoice.discount_percent > 0) {
			updatedInvoice.discount_amount = (subtotal * updatedInvoice.discount_percent) / 100;
		} else {
			updatedInvoice.discount_amount = 0;
		}

		const taxable = subtotal - updatedInvoice.discount_amount;
		const discountRatio = subtotal > 0 ? taxable / subtotal : 1;

		// Reset tax values
		updatedInvoice.cgst_rate = 0;
		updatedInvoice.cgst_amount = 0;
		updatedInvoice.sgst_rate = 0;
		updatedInvoice.sgst_amount = 0;
		updatedInvoice.igst_rate = 0;
		updatedInvoice.igst_amount = 0;

		const isSameState = updatedInvoice.customer_state_code === '33'; // TN Seller default matching prefix '33'

		updatedInvoice.items.forEach(item => {
			const itemSubtotal = (item.quantity || 0) * (item.rate || 0);
			const itemTaxable = itemSubtotal * discountRatio;
			const itemGstRate = item.gst_rate !== undefined ? item.gst_rate : 18;

			if (isSameState) {
				const cgstRate = itemGstRate / 2;
				const sgstRate = itemGstRate / 2;
				updatedInvoice.cgst_amount += (itemTaxable * cgstRate) / 100;
				updatedInvoice.sgst_amount += (itemTaxable * sgstRate) / 100;
				// Maintain active rates for summary presentation
				updatedInvoice.cgst_rate = cgstRate;
				updatedInvoice.sgst_rate = sgstRate;
			} else {
				const igstRate = itemGstRate;
				updatedInvoice.igst_amount += (itemTaxable * igstRate) / 100;
				updatedInvoice.igst_rate = igstRate;
			}
		});

		const totalTax = updatedInvoice.cgst_amount + updatedInvoice.sgst_amount + updatedInvoice.igst_amount;

		updatedInvoice.tds_tcs_amount = 0;
		if (updatedInvoice.tds_tcs_type !== 'NONE') {
			updatedInvoice.tds_tcs_amount = (taxable * (updatedInvoice.tds_tcs_rate || 0)) / 100;
		}

		let finalTotal = taxable + totalTax + (Number(updatedInvoice.transportation_charge) || 0);
		if (updatedInvoice.tds_tcs_type === 'TCS') {
			finalTotal += updatedInvoice.tds_tcs_amount;
		} else if (updatedInvoice.tds_tcs_type === 'TDS') {
			finalTotal -= updatedInvoice.tds_tcs_amount;
		}

		updatedInvoice.total_price = finalTotal;
		updatedInvoice.balance_amount = finalTotal - (updatedInvoice.paid_amount || 0);

		setFormInvoice({ ...updatedInvoice });
	};

	const triggerPrint = () => {
		window.print();
	};

	const filteredInvoices = invoices.filter(inv => {
		const matchSearch = inv.customer_name.toLowerCase().includes(invoiceSearch.toLowerCase()) ||
			(inv.invoice_number && inv.invoice_number.toLowerCase().includes(invoiceSearch.toLowerCase()));
		const matchStatus = invoiceStatusFilter ? inv.status === invoiceStatusFilter : true;
		return matchSearch && matchStatus;
	});

	return (
		<>
			<div className="b2b-billing-container glass-card" style={{ padding: '32px', margin: '12px 0', borderRadius: '20px', background: 'var(--surface-color)', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-md)' }}>
				{/* Premium Layout Styles */}
				<style>{`
				@media print {
					body * {
						visibility: hidden;
					}
					.print-invoice-area, .print-invoice-area * {
						visibility: visible;
					}
					.print-invoice-area {
						position: absolute;
						left: 0;
						top: 0;
						width: 100%;
						background: white !important;
						color: black !important;
						padding: 40px !important;
					}
					.no-print {
						display: none !important;
					}
				}
				.b2b-billing-container {
					animation: fadeInUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
				}
				.b2b-input {
					background-color: var(--bg-input) !important;
					color: var(--text-primary) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 10px !important;
					padding: 0.65rem 0.9rem !important;
					font-size: 0.9rem !important;
					outline: none;
					transition: all 0.2s ease-in-out;
					width: 100%;
				}
				.b2b-input:focus {
					border-color: var(--accent-color) !important;
					background-color: var(--surface-color) !important;
					box-shadow: 0 0 0 3px var(--accent-subtle) !important;
				}
				.b2b-input::placeholder {
					color: var(--text-tertiary) !important;
					opacity: 0.8;
				}
				.b2b-btn {
					padding: 0.65rem 1.2rem;
					font-weight: 600;
					font-size: 0.9rem;
					border-radius: 10px;
					border: none;
					cursor: pointer;
					transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
					display: inline-flex;
					align-items: center;
					justify-content: center;
					gap: 6px;
				}
				.b2b-btn:hover {
					transform: translateY(-1px);
					box-shadow: var(--shadow-sm);
				}
				.b2b-btn-primary {
					background-color: var(--accent-color);
					color: white;
				}
				.b2b-btn-primary:hover {
					background-color: var(--accent-hover);
					box-shadow: 0 4px 12px var(--accent-subtle);
				}
				.b2b-btn-secondary {
					background-color: var(--bg-input);
					color: var(--text-primary);
					border: 1px solid var(--border-color);
				}
				.b2b-btn-secondary:hover {
					background-color: var(--bg-hover);
					border-color: var(--border-strong);
				}
				.b2b-btn-danger {
					background-color: var(--status-danger-bg);
					color: var(--status-danger);
					border: 1px solid rgba(239, 68, 68, 0.15);
				}
				.b2b-btn-danger:hover {
					background-color: var(--status-danger);
					color: white;
				}
				.b2b-btn-success {
					background-color: var(--status-active-bg);
					color: var(--status-active);
					border: 1px solid rgba(16, 185, 129, 0.15);
				}
				.b2b-btn-success:hover {
					background-color: var(--status-active);
					color: white;
				}
				.b2b-subtab-btn {
					background: transparent;
					border: 1px solid var(--border-color);
					color: var(--text-secondary);
					padding: 8px 18px;
					border-radius: 20px;
					font-weight: 600;
					font-size: 0.85rem;
					cursor: pointer;
					transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
				}
				.b2b-subtab-btn:hover {
					background: var(--bg-hover);
					border-color: var(--accent-color);
					color: var(--text-primary);
				}
				.b2b-subtab-btn.active {
					border-color: var(--accent-color);
					background: var(--accent-subtle);
					color: var(--accent-color);
					box-shadow: 0 4px 12px rgba(16, 185, 129, 0.08);
				}
				.b2b-form-section {
					background: var(--bg-hover);
					padding: 20px;
					border-radius: 12px;
					border: 1px solid var(--border-color);
					margin-bottom: 24px;
				}
				.b2b-form-section-title {
					font-size: 0.75rem;
					font-weight: 700;
					color: var(--text-secondary);
					text-transform: uppercase;
					letter-spacing: 0.05em;
					margin-bottom: 16px;
					border-bottom: 1px solid var(--border-color);
					padding-bottom: 8px;
				}
				.search-wrapper {
					position: relative;
					display: flex;
					align-items: center;
				}
				.search-icon {
					position: absolute;
					left: 12px;
					color: var(--text-tertiary);
					pointer-events: none;
				}
				.search-input {
					padding-left: 36px !important;
				}
				.clear-search-btn {
					position: absolute;
					right: 10px;
					background: transparent;
					border: none;
					color: var(--text-tertiary);
					cursor: pointer;
					display: flex;
					align-items: center;
					justify-content: center;
					padding: 4px;
					border-radius: 50%;
					transition: all 0.2s;
				}
				.clear-search-btn:hover {
					color: var(--text-primary);
					background: var(--bg-hover);
				}
				.form-label {
					font-size: 0.75rem;
					font-weight: 700;
					color: var(--text-secondary);
					text-transform: uppercase;
					letter-spacing: 0.05em;
					margin-bottom: 6px;
				}
				.b2b-table {
					width: 100%;
					border-collapse: separate;
					border-spacing: 0;
				}
				.b2b-table th {
					padding: 1.25rem 1.5rem;
					font-size: 0.75rem;
					font-weight: 800;
					color: var(--text-tertiary);
					text-transform: uppercase;
					letter-spacing: 1px;
					border-bottom: 1px solid var(--border-color);
					background: var(--bg-hover);
				}
				.b2b-table td {
					padding: 1.25rem 1.5rem;
					font-size: 0.9rem;
					border-bottom: 1px solid var(--border-color);
					transition: background-color 0.2s ease;
					color: var(--text-primary);
				}
				.b2b-table tr:last-child td {
					border-bottom: none;
				}
				.b2b-table tr:hover td {
					background-color: var(--bg-hover);
				}
				
				/* Redesigned Form UI Styles */
				.form-header-container {
					display: flex;
					justify-content: space-between;
					align-items: center;
					margin-bottom: 28px;
					border-bottom: 1px solid var(--border-color);
					padding-bottom: 20px;
				}
				.form-header-title {
					font-size: 1.6rem;
					font-weight: 800;
					color: var(--text-primary);
					margin: 0;
					letter-spacing: -0.025em;
					display: flex;
					align-items: center;
					gap: 12px;
				}
				.form-header-subtitle {
					font-size: 0.875rem;
					color: var(--text-secondary);
					margin: 4px 0 0 0;
				}
				.b2b-form-section {
					background: var(--surface-color) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 16px !important;
					padding: 24px !important;
					margin-bottom: 28px !important;
					box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.03), 0 2px 4px -1px rgba(0, 0, 0, 0.02) !important;
					transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
				}
				.b2b-form-section:hover {
					border-color: rgba(16, 185, 129, 0.3) !important;
					box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05), 0 4px 6px -2px rgba(0, 0, 0, 0.03) !important;
				}
				.b2b-form-section-title {
					font-size: 0.8rem !important;
					font-weight: 700 !important;
					color: var(--accent-color) !important;
					text-transform: uppercase !important;
					letter-spacing: 0.08em !important;
					margin-bottom: 20px !important;
					padding-bottom: 6px !important;
					border-bottom: 2px solid var(--accent-subtle) !important;
					display: inline-block !important;
				}
				.client-info-card {
					background: linear-gradient(135deg, var(--bg-hover) 0%, var(--surface-color) 100%) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 12px !important;
					padding: 18px 22px !important;
					display: flex;
					flex-direction: column;
					gap: 10px;
					position: relative;
					overflow: hidden;
					animation: fadeInScale 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
				}
				.client-info-card::before {
					content: '';
					position: absolute;
					left: 0;
					top: 0;
					bottom: 0;
					width: 4px;
					background: linear-gradient(180deg, var(--accent-color), #10b981);
				}
				.client-badge {
					display: inline-flex;
					align-items: center;
					padding: 4px 8px;
					background: var(--accent-subtle);
					color: var(--accent-color);
					border-radius: 6px;
					font-size: 0.75rem;
					font-weight: 700;
				}
				.items-table-header th {
					font-weight: 700 !important;
					text-transform: uppercase !important;
					font-size: 0.75rem !important;
					letter-spacing: 0.05em !important;
					color: var(--text-secondary) !important;
					border-bottom: 2px solid var(--border-color) !important;
					padding: 12px 16px !important;
					background: transparent !important;
				}
				.gst-select-input {
					padding: 0.65rem 1.5rem 0.65rem 0.75rem !important;
				}
				.items-table-row {
					transition: background-color 0.2s;
				}
				.items-table-row:hover {
					background-color: var(--bg-hover);
				}
				.delete-row-btn {
					background: transparent !important;
					border: none !important;
					color: var(--text-tertiary) !important;
					cursor: pointer;
					padding: 8px !important;
					border-radius: 50% !important;
					display: inline-flex !important;
					align-items: center;
					justify-content: center;
					transition: all 0.2s !important;
				}
				.delete-row-btn:hover {
					color: var(--status-danger) !important;
					background-color: var(--status-danger-bg) !important;
					transform: scale(1.1);
				}
				.add-row-btn {
					display: inline-flex;
					align-items: center;
					gap: 8px;
					padding: 8px 16px;
					font-weight: 600;
					font-size: 0.85rem;
					border-radius: 8px;
					border: 1px dashed var(--accent-color);
					background: transparent;
					color: var(--accent-color);
					cursor: pointer;
					transition: all 0.2s;
				}
				.add-row-btn:hover {
					background: var(--accent-subtle);
					border-style: solid;
					transform: translateY(-1px);
				}
				.summary-panel {
					background: linear-gradient(180deg, var(--bg-hover) 0%, var(--surface-color) 100%) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 16px !important;
					padding: 24px !important;
					box-shadow: var(--shadow-sm) !important;
				}
				.summary-row {
					display: flex;
					justify-content: space-between;
					align-items: center;
					margin-bottom: 14px;
					font-size: 0.9rem;
				}
				.summary-total-box {
					background: linear-gradient(135deg, var(--accent-subtle) 0%, rgba(16, 185, 129, 0.05) 100%) !important;
					border: 1px solid rgba(16, 185, 129, 0.15) !important;
					border-radius: 12px !important;
					padding: 16px !important;
					display: flex;
					justify-content: space-between;
					align-items: center;
					margin-top: 16px;
				}
				.days-input-left {
					border-top-right-radius: 0px !important;
					border-bottom-right-radius: 0px !important;
					border-right: none !important;
				}
				.days-label-right {
					background: var(--bg-hover) !important;
					border: 1px solid var(--border-color) !important;
					border-left: none !important;
					padding: 0.65rem 0.9rem !important;
					border-top-right-radius: 10px !important;
					border-bottom-right-radius: 10px !important;
					border-top-left-radius: 0px !important;
					border-bottom-left-radius: 0px !important;
					color: var(--text-secondary) !important;
					font-size: 0.9rem !important;
					white-space: nowrap !important;
					height: 42px !important;
					display: inline-flex !important;
					align-items: center !important;
					box-sizing: border-box !important;
				}
				@keyframes fadeInScale {
					from {
						opacity: 0;
						transform: scale(0.97);
					}
					to {
						opacity: 1;
						transform: scale(1);
					}
				}
			`}</style>

				{/* Sub Tabs */}
				{viewMode === 'list' && (
					<div className="sub-tabs-container no-print" style={{ display: 'flex', gap: '8px', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px' }}>
						<button className={`b2b-subtab-btn ${activeSubTab === 'invoices' ? 'active' : ''}`} onClick={() => setActiveSubTab('invoices')}>Invoices</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'customers' ? 'active' : ''}`} onClick={() => setActiveSubTab('customers')}>B2B Clients</button>
					</div>
				)}

				{/* LIST VIEW */}
				{viewMode === 'list' && activeSubTab === 'invoices' && (
					<div className="no-print">
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
							<div style={{ display: 'flex', gap: '10px' }}>
								<div className="search-wrapper" style={{ width: '240px', flex: 'none' }}>
									<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
										<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
									</svg>
									<input
										type="text"
										placeholder="Search Invoices..."
										className="b2b-input search-input"
										value={invoiceSearch}
										onChange={(e) => setInvoiceSearch(e.target.value)}
									/>
									{invoiceSearch && (
										<button className="clear-search-btn" onClick={() => setInvoiceSearch('')}>✕</button>
									)}
								</div>
								<select
									value={invoiceStatusFilter}
									onChange={(e) => setInvoiceStatusFilter(e.target.value)}
									className="b2b-input"
									style={{ width: '160px' }}
								>
									<option value="">All Statuses</option>
									<option value="DRAFT">Draft</option>
									<option value="ISSUED">Issued</option>
									<option value="CANCELLED">Cancelled</option>
								</select>
							</div>
							{userRole === 'admin' && (
								<button
									onClick={() => {
										setFormInvoice({
											invoice_date: new Date().toISOString().split('T')[0],
											customer_gstin: '',
											customer_name: '',
											customer_state: '',
											customer_state_code: '',
											customer_address: '',
											subtotal_price: 0,
											discount_percent: 0,
											discount_amount: 0,
											cgst_rate: 0,
											cgst_amount: 0,
											sgst_rate: 0,
											sgst_amount: 0,
											igst_rate: 0,
											igst_amount: 0,
											tds_tcs_type: 'NONE',
											tds_tcs_rate: 0,
											tds_tcs_amount: 0,
											transportation_charge: 0,
											total_price: 0,
											status: 'DRAFT',
											payment_status: 'UNPAID',
											paid_amount: 0,
											balance_amount: 0,
											customer_notes: DEFAULT_CUSTOMER_NOTES,
											items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
										});
										setViewMode('create');
									}}
									className="b2b-btn b2b-btn-primary"
								>
									+ Create Invoice
								</button>
							)}
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Invoice#</th>
										<th>Client</th>
										<th>Date</th>
										<th>Total Amount</th>
										<th>GST Split</th>
										<th>Status</th>
										<th>Payment Status</th>
										<th style={{ textAlign: 'right' }}>Actions</th>
									</tr>
								</thead>
								<tbody>
									{filteredInvoices.map((inv) => (
										<tr key={inv.id}>
											<td>{inv.invoice_number || <span style={{ opacity: 0.5 }}>Draft</span>}</td>
											<td>
												<strong>{inv.customer_name}</strong>
												<div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '2px' }}>{inv.customer_gstin}</div>
											</td>
											<td>{inv.invoice_date ? inv.invoice_date.split('T')[0] : ''}</td>
											<td style={{ fontWeight: 'bold' }}>₹{inv.total_price.toFixed(2)}</td>
											<td style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
												{inv.cgst_amount > 0 && <div>CGST: ₹{inv.cgst_amount.toFixed(2)} ({(inv.cgst_rate)}%)</div>}
												{inv.sgst_amount > 0 && <div>SGST: ₹{inv.sgst_amount.toFixed(2)} ({(inv.sgst_rate)}%)</div>}
												{inv.igst_amount > 0 && <div>IGST: ₹{inv.igst_amount.toFixed(2)} ({(inv.igst_rate)}%)</div>}
											</td>
											<td>
												<span style={{
													padding: '6px 10px',
													borderRadius: '12px',
													fontSize: '11px',
													fontWeight: 600,
													background: inv.status === 'ISSUED' ? 'var(--status-active-bg)' : inv.status === 'CANCELLED' ? 'var(--status-danger-bg)' : 'var(--status-warning-bg)',
													color: inv.status === 'ISSUED' ? 'var(--status-active)' : inv.status === 'CANCELLED' ? 'var(--status-danger)' : 'var(--status-warning)'
												}}>
													{inv.status}
												</span>
											</td>
											<td>
												<span style={{
													padding: '6px 10px',
													borderRadius: '12px',
													fontSize: '11px',
													fontWeight: 600,
													background: inv.payment_status === 'PAID' ? 'var(--status-active-bg)' : inv.payment_status === 'PARTIAL' ? 'var(--status-warning-bg)' : 'var(--status-danger-bg)',
													color: inv.payment_status === 'PAID' ? 'var(--status-active)' : inv.payment_status === 'PARTIAL' ? 'var(--status-warning)' : 'var(--status-danger)'
												}}>
													{inv.payment_status}
												</span>
												<div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px' }}>Bal: ₹{inv.balance_amount.toFixed(2)}</div>
											</td>
											<td style={{ textAlign: 'right' }}>
												<div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
													<button
														onClick={() => {
															setSelectedInvoice(inv);
															setViewMode('preview');
														}}
														className="b2b-btn b2b-btn-secondary"
														style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
													>
														View
													</button>

													{inv.status === 'DRAFT' && userRole === 'admin' && (
														<>
															<button
																onClick={() => {
																	setFormInvoice({ ...inv });
																	setViewMode('edit');
																}}
																className="b2b-btn b2b-btn-secondary"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Edit
															</button>
															<button
																onClick={() => handleIssueInvoice(inv.id!)}
																className="b2b-btn b2b-btn-success"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Issue
															</button>
															<button
																onClick={() => handleDeleteInvoice(inv.id!)}
																className="b2b-btn b2b-btn-danger"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Delete
															</button>
														</>
													)}

													{inv.status === 'ISSUED' && userRole === 'admin' && (
														<>
															<button
																onClick={() => {
																	setPaymentInvoice(inv);
																	setPaymentAmount(inv.balance_amount);
																	setShowPaymentModal(true);
																}}
																className="b2b-btn b2b-btn-success"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto', backgroundColor: 'rgba(16, 185, 129, 0.15)' }}
															>
																Payment
															</button>
															<button
																onClick={() => handleCancelInvoice(inv.id!)}
																className="b2b-btn b2b-btn-danger"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Cancel
															</button>
														</>
													)}
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* CUSTOMER REGISTRY VIEW */}
				{viewMode === 'list' && activeSubTab === 'customers' && (
					<div>
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
							<div className="search-wrapper" style={{ width: '280px', flex: 'none' }}>
								<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
									<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
								</svg>
								<input
									type="text"
									placeholder="Search Clients..."
									className="b2b-input search-input"
									value={customerSearch}
									onChange={(e) => setCustomerSearch(e.target.value)}
								/>
								{customerSearch && (
									<button className="clear-search-btn" onClick={() => setCustomerSearch('')}>✕</button>
								)}
							</div>
							{userRole === 'admin' && (
								<button
									onClick={() => {
										setEditingCustomer({
											legal_name: '',
											gstin: '',
											billing_address: '',
											shipping_address: '',
											state: '',
											state_code: ''
										});
										setSameAsBilling(true);
										setShowCustomerModal(true);
									}}
									className="b2b-btn b2b-btn-primary"
								>
									+ Add Client
								</button>
							)}
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Legal Name</th>
										<th>GSTIN</th>
										<th>PAN</th>
										<th>State (Code)</th>
										<th>Billing Address</th>
										<th style={{ textAlign: 'right' }}>Actions</th>
									</tr>
								</thead>
								<tbody>
									{customers.filter(c => c.legal_name.toLowerCase().includes(customerSearch.toLowerCase()) || c.gstin.toLowerCase().includes(customerSearch.toLowerCase())).map((cust) => (
										<tr key={cust.id}>
											<td><strong>{cust.legal_name}</strong> {cust.trade_name && <span style={{ display: 'block', fontSize: '11px', opacity: 0.7 }}>Trade: {cust.trade_name}</span>}</td>
											<td>{cust.gstin}</td>
											<td>{cust.pan}</td>
											<td>{cust.state} ({cust.state_code})</td>
											<td style={{ maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{cust.billing_address}</td>
											<td style={{ textAlign: 'right' }}>
												<div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
													<button
														onClick={() => {
															setEditingCustomer({ ...cust });
															setSameAsBilling(!cust.shipping_address || cust.shipping_address === cust.billing_address);
															setShowCustomerModal(true);
														}}
														className="b2b-btn b2b-btn-secondary"
														style={{ padding: '6px', minHeight: 'auto', borderRadius: '8px' }}
														title="Edit Client"
													>
														<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
															<path d="M12 20h9" /><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
														</svg>
													</button>
													<button
														onClick={() => handleDeleteCustomer(cust.id!)}
														className="b2b-btn b2b-btn-danger"
														style={{ padding: '6px', minHeight: 'auto', borderRadius: '8px' }}
														title="Delete Client"
													>
														<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
															<polyline points="3 6 5 6 21 6" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /><line x1="10" y1="11" x2="10" y2="17" /><line x1="14" y1="11" x2="14" y2="17" />
														</svg>
													</button>
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* BILL CREATOR & EDITOR VIEW */}
				{(viewMode === 'create' || viewMode === 'edit') && (
					<div className="b2b-form-area no-print" style={{ color: 'var(--text-primary)' }}>
						<div className="form-header-container">
							<div>
								<h3 className="form-header-title">
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--accent-color)' }}>
										<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
										<polyline points="14 2 14 8 20 8" />
										<line x1="16" y1="13" x2="8" y2="13" />
										<line x1="16" y1="17" x2="8" y2="17" />
										<polyline points="10 9 9 9 8 9" />
									</svg>
									{viewMode === 'create' ? 'Create B2B Bill' : 'Edit B2B Bill'}
								</h3>
								<p className="form-header-subtitle">Fill in the details below to generate a GST-compliant invoice.</p>
							</div>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary" style={{ height: '40px' }}>
								&larr; Back to List
							</button>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Client Information</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', alignItems: 'start' }}>
								{/* Customer Selector */}
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Select B2B Customer*</label>
									<select
										className="b2b-input"
										value={formInvoice.customer_id || ''}
										onChange={(e) => {
											const cId = Number(e.target.value);
											const selected = customers.find(c => c.id === cId);
											if (selected) {
												const updated = {
													...formInvoice,
													customer_id: selected.id,
													customer_gstin: selected.gstin,
													customer_name: selected.legal_name,
													customer_state: selected.state,
													customer_state_code: selected.state_code,
													customer_address: selected.billing_address
												};
												recalculateTotals(updated);
											}
										}}
										style={{ height: '46px' }}
									>
										<option value="">-- Choose Customer --</option>
										{customers.map(c => (
											<option key={c.id} value={c.id}>{c.legal_name} ({c.gstin})</option>
										))}
									</select>
								</div>

								{/* Client Quick Details Display */}
								{formInvoice.customer_gstin ? (
									<div className="client-info-card">
										<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
											<strong style={{ fontSize: '1rem', color: 'var(--text-primary)' }}>{formInvoice.customer_name}</strong>
											<span className="client-badge">GST ACTIVE</span>
										</div>
										<div style={{ display: 'grid', gridTemplateColumns: '80px 1fr', gap: '6px', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
											<span><strong>GSTIN:</strong></span>
											<span style={{ color: 'var(--text-primary)', fontFamily: 'monospace', fontWeight: 600 }}>{formInvoice.customer_gstin}</span>

											<span><strong>State:</strong></span>
											<span>{formInvoice.customer_state} ({formInvoice.customer_state_code})</span>

											<span><strong>Address:</strong></span>
											<span style={{ textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }} title={formInvoice.customer_address}>{formInvoice.customer_address}</span>
										</div>
									</div>
								) : (
									<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100px', border: '1px dashed var(--border-color)', borderRadius: '12px', background: 'var(--bg-hover)', color: 'var(--text-tertiary)', fontSize: '0.85rem' }}>
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginBottom: '8px', opacity: 0.6 }}>
											<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
										</svg>
										No client selected. Please choose a client to load billing details.
									</div>
								)}
							</div>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Billing Details</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', marginBottom: '20px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Invoice Date*</label>
									<input
										type="date"
										className="b2b-input"
										value={formInvoice.invoice_date ? formInvoice.invoice_date.split('T')[0] : ''}
										onChange={(e) => {
											const newDate = e.target.value;
											let computedDueDate = formInvoice.due_date;
											if (formInvoice.terms) {
												const matchedTerm = paymentTerms.find(t => t.name === formInvoice.terms);
												if (matchedTerm) {
													const date = new Date(newDate);
													date.setDate(date.getDate() + matchedTerm.due_days);
													computedDueDate = date.toISOString().split('T')[0];
												}
											}
											setFormInvoice({
												...formInvoice,
												invoice_date: newDate,
												due_date: computedDueDate
											});
										}}
									/>
								</div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Due Date</label>
									<input
										type="date"
										className="b2b-input"
										value={formInvoice.due_date ? formInvoice.due_date.split('T')[0] : ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, due_date: e.target.value })}
									/>
								</div>
							</div>

							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '20px', marginBottom: '20px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Order Number</label>
									<input
										type="text"
										className="b2b-input"
										placeholder="e.g. PO-9982"
										value={formInvoice.order_number || ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, order_number: e.target.value })}
									/>
								</div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Salesperson</label>
									<input
										type="text"
										className="b2b-input"
										placeholder="e.g. John Doe"
										value={formInvoice.salesperson || ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, salesperson: e.target.value })}
									/>
								</div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Terms</label>
									<select
										className="b2b-input"
										value={formInvoice.terms || ''}
										onChange={(e) => {
											const val = e.target.value;
											if (val === '__CREATE_NEW__') {
												setShowPaymentTermModal(true);
											} else {
												handleTermsChange(val);
											}
										}}
										style={{ height: '46px' }}
									>
										<option value="">-- Select Terms --</option>
										{paymentTerms.map(t => (
											<option key={t.id || t.name} value={t.name}>{t.name}</option>
										))}
										<option value="__CREATE_NEW__" style={{ fontWeight: 'bold', color: 'var(--accent-color)' }}>+ New Payment Term</option>
									</select>
								</div>
							</div>

							<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
								<label className="form-label">Subject</label>
								<input
									type="text"
									className="b2b-input"
									placeholder="e.g. Supply of Raw materials / Fragrance ingredients"
									value={formInvoice.subject || ''}
									onChange={(e) => setFormInvoice({ ...formInvoice, subject: e.target.value })}
								/>
							</div>
						</div>

						{/* Item Table */}
						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Items Table</div>
							<div className="table-responsive" style={{ overflowX: 'auto', marginBottom: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
								<table style={{ minWidth: '1220px', width: '100%', borderCollapse: 'collapse' }}>
									<thead>
										<tr className="items-table-header" style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', background: 'var(--bg-hover)' }}>
											<th style={{ padding: '14px 16px', minWidth: '400px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Item Details</th>
											<th style={{ padding: '14px 16px', minWidth: '140px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>HSN Code</th>
											<th style={{ padding: '14px 16px', minWidth: '120px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>GST (%)</th>
											<th style={{ padding: '14px 16px', minWidth: '110px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Quantity</th>
											<th style={{ padding: '14px 16px', minWidth: '130px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Rate (₹)</th>
											<th style={{ padding: '14px 16px', minWidth: '130px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>GST (₹)</th>
											<th style={{ padding: '14px 16px', minWidth: '140px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Amount</th>
											<th style={{ padding: '14px 16px', minWidth: '50px', textAlign: 'right' }}></th>
										</tr>
									</thead>
									<tbody>
										{formInvoice.items.map((item, idx) => (
											<tr key={idx} className="items-table-row" style={{ borderBottom: '1px solid var(--border-color)' }}>
												<td style={{ padding: '12px 16px', minWidth: '400px' }}>
													<input
														type="text"
														list={`products-list-${idx}`}
														className="b2b-input"
														placeholder="Search warehouse products or type custom details..."
														value={item.item_details}
														onChange={(e) => {
															const val = e.target.value;
															const newItems = [...formInvoice.items];
															newItems[idx].item_details = val;

															// Check if matching a product option value or title
															const foundProduct = inventoryProducts.find(p =>
																p.title === val ||
																`${p.title} (${p.mi_sku})` === val
															);
															if (foundProduct) {
																newItems[idx].product_id = foundProduct.id;
																newItems[idx].item_details = foundProduct.title;
																newItems[idx].sku = foundProduct.mi_sku;
																newItems[idx].rate = foundProduct.price || 0;
																newItems[idx].hsn_code = foundProduct.hsn_code || '33029019';
															}
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													/>
													<datalist id={`products-list-${idx}`}>
														{inventoryProducts.map(p => (
															<option key={p.id} value={`${p.title} (${p.mi_sku})`} />
														))}
													</datalist>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '140px' }}>
													<input
														type="text"
														className="b2b-input"
														placeholder="33029019"
														value={item.hsn_code || ''}
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].hsn_code = e.target.value;
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														maxLength={8}
														style={{ background: 'var(--surface-color)', height: '42px', fontFamily: 'monospace' }}
													/>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '120px' }}>
													<select
														className="b2b-input gst-select-input"
														value={item.gst_rate !== undefined ? item.gst_rate : 18}
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].gst_rate = Number(e.target.value);
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													>
														<option value={0}>0%</option>
														<option value={5}>5%</option>
														<option value={12}>12%</option>
														<option value={18}>18%</option>
														<option value={28}>28%</option>
													</select>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '110px' }}>
													<input
														type="number"
														value={item.quantity}
														min="1"
														className="b2b-input"
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].quantity = Number(e.target.value);
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													/>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '130px' }}>
													<input
														type="number"
														value={item.rate}
														className="b2b-input"
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].rate = Number(e.target.value);
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													/>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '130px', fontWeight: '500', color: 'var(--text-secondary)', whiteSpace: 'nowrap', fontSize: '0.9rem' }}>
													₹{((item.quantity * item.rate * (item.gst_rate !== undefined ? item.gst_rate : 18)) / 100).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
												</td>
												<td style={{ padding: '12px 16px', minWidth: '140px', fontWeight: '700', color: 'var(--text-primary)', whiteSpace: 'nowrap', fontSize: '0.95rem' }}>₹{(item.quantity * item.rate).toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
												<td style={{ padding: '12px 16px', minWidth: '50px', textAlign: 'right' }}>
													{formInvoice.items.length > 1 && (
														<button
															onClick={() => {
																const newItems = formInvoice.items.filter((_, i) => i !== idx);
																recalculateTotals({ ...formInvoice, items: newItems });
															}}
															className="delete-row-btn"
															title="Delete Row"
														>
															<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																<polyline points="3 6 5 6 21 6" />
																<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
																<line x1="10" y1="11" x2="10" y2="17" />
																<line x1="14" y1="11" x2="14" y2="17" />
															</svg>
														</button>
													)}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>

							<button
								onClick={() => {
									const newItems = [...formInvoice.items, { item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }];
									setFormInvoice({ ...formInvoice, items: newItems });
								}}
								className="add-row-btn"
							>
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
								</svg>
								Add Row
							</button>
						</div>

						{/* Calculations section */}
						<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginTop: '24px' }}>
							{/* Notes & Extra charges selection */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Customer Notes</label>
									<textarea
										placeholder="Thanks for your business. Let us know if you need any changes."
										value={formInvoice.customer_notes || ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, customer_notes: e.target.value })}
										className="b2b-input"
										style={{ minHeight: '110px', fontFamily: 'inherit' }}
									/>
								</div>

								<div style={{ background: 'var(--bg-hover)', padding: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
									<div style={{ fontWeight: 700, fontSize: '0.85rem', color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '14px', letterSpacing: '0.05em' }}>Tax Adjustments</div>
									<div style={{ display: 'flex', gap: '20px', alignItems: 'center', flexWrap: 'wrap' }}>

										<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
											<label className="form-label" style={{ fontSize: '0.7rem' }}>TDS/TCS Type</label>
											<select
												value={formInvoice.tds_tcs_type}
												onChange={(e) => {
													const type = e.target.value;
													const rate = type === 'NONE' ? 0 : formInvoice.tds_tcs_rate;
													recalculateTotals({ ...formInvoice, tds_tcs_type: type, tds_tcs_rate: rate });
												}}
												className="b2b-input"
												style={{ width: '130px', height: '40px' }}
											>
												<option value="NONE">None</option>
												<option value="TDS">TDS</option>
												<option value="TCS">TCS</option>
											</select>
										</div>
										{formInvoice.tds_tcs_type !== 'NONE' && (
											<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
												<label className="form-label" style={{ fontSize: '0.7rem' }}>Rate (%)</label>
												<input
													type="number"
													value={formInvoice.tds_tcs_rate}
													onChange={(e) => recalculateTotals({ ...formInvoice, tds_tcs_rate: Number(e.target.value) })}
													className="b2b-input"
													style={{ width: '100px', height: '40px' }}
												/>
											</div>
										)}
									</div>
								</div>
							</div>

							{/* Pricing Breakdown */}
							<div className="summary-panel">
								<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
									<span>Sub Total:</span>
									<strong style={{ color: 'var(--text-primary)' }}>₹{formInvoice.subtotal_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</strong>
								</div>

								<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
									<span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>Discount (%):</span>
									<input
										type="number"
										value={formInvoice.discount_percent}
										min="0"
										max="100"
										className="b2b-input"
										onChange={(e) => recalculateTotals({ ...formInvoice, discount_percent: Number(e.target.value) })}
										style={{ width: '70px', padding: '6px 10px !important', textAlign: 'right', background: 'var(--surface-color)', height: '32px' }}
									/>
								</div>

								{formInvoice.discount_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>Discount Amount:</span>
										<span style={{ color: 'var(--status-danger)', fontWeight: 600 }}>- ₹{formInvoice.discount_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								{formInvoice.cgst_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>CGST ({formInvoice.cgst_rate}%):</span>
										<span>₹{formInvoice.cgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}
								{formInvoice.sgst_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>SGST ({formInvoice.sgst_rate}%):</span>
										<span>₹{formInvoice.sgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}
								{formInvoice.igst_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>IGST ({formInvoice.igst_rate}%):</span>
										<span>₹{formInvoice.igst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								<div className="summary-row" style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
									<span>Transportation:</span>
									<input
										type="number"
										value={formInvoice.transportation_charge}
										className="b2b-input"
										onChange={(e) => recalculateTotals({ ...formInvoice, transportation_charge: Number(e.target.value) })}
										style={{ width: '110px', padding: '6px 10px !important', textAlign: 'right', background: 'var(--surface-color)', height: '32px' }}
									/>
								</div>

								{formInvoice.tds_tcs_type !== 'NONE' && (
									<div className="summary-row" style={{ color: 'var(--status-warning)', fontWeight: 600 }}>
										<span>{formInvoice.tds_tcs_type} ({formInvoice.tds_tcs_rate}%):</span>
										<span>{formInvoice.tds_tcs_type === 'TDS' ? '-' : '+'} ₹{formInvoice.tds_tcs_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', margin: '16px 0' }} />

								<div className="summary-total-box">
									<span style={{ color: 'var(--text-primary)', fontWeight: 700, fontSize: '1rem' }}>Total Amount:</span>
									<span style={{ color: 'var(--accent-color)', fontSize: '1.4rem', fontWeight: 800 }}>₹{formInvoice.total_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
								</div>
							</div>
						</div>

						{/* Form Buttons */}
						<div style={{ display: 'flex', justifyContent: 'flex-end', gap: '16px', marginTop: '40px', borderTop: '1px solid var(--border-color)', paddingTop: '28px' }}>
							<button
								onClick={() => setViewMode('list')}
								className="b2b-btn b2b-btn-secondary"
								style={{ minWidth: '110px', height: '46px' }}
							>
								Cancel
							</button>
							<button
								onClick={() => handleSaveInvoice(true)}
								className="b2b-btn b2b-btn-secondary"
								style={{ minWidth: '140px', height: '46px' }}
							>
								Save as Draft
							</button>
							<button
								onClick={() => handleSaveInvoice(false)}
								className="b2b-btn b2b-btn-primary"
								style={{ minWidth: '180px', height: '46px', gap: '8px' }}
							>
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<line x1="22" y1="2" x2="11" y2="13" /><polygon points="22 2 15 22 11 13 2 9 22 2" />
								</svg>
								Save & Issue Bill
							</button>
						</div>
					</div>
				)}

				{/* PRINT / PREVIEW LAYOUT VIEW */}
				{viewMode === 'preview' && selectedInvoice && (
					<div>
						<div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">&larr; Back to List</button>
							<button onClick={triggerPrint} className="b2b-btn b2b-btn-primary">Print / Download PDF</button>
						</div>

						{/* INVOICE PAGE DESIGN */}
						<div className="print-invoice-area" style={{ background: 'var(--surface-color)', padding: '40px', borderRadius: '16px', border: '1px solid var(--border-color)', color: 'var(--text-primary)', boxShadow: 'var(--shadow-sm)' }}>
							<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '36px' }}>
								<div>
									<h1 style={{ margin: '0 0 8px 0', textTransform: 'uppercase', color: 'var(--accent-color)', fontWeight: 800, fontSize: '2rem' }}>TAX INVOICE</h1>
									<div style={{ fontSize: '14px', color: 'var(--text-secondary)' }}><strong>FY:</strong> {selectedInvoice.financial_year || 'N/A'}</div>
								</div>
								<div style={{ textAlign: 'right' }}>
									<h2 style={{ margin: '0 0 6px 0', fontWeight: 800 }}>{selectedInvoice.seller_name}</h2>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedInvoice.seller_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', maxWidth: '280px', lineHeight: 1.4 }}>{selectedInvoice.seller_address}</div>
								</div>
							</div>

							<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', marginBottom: '24px' }} />

							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginBottom: '36px' }}>
								<div>
									<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Bill To</h4>
									<h3 style={{ margin: '0 0 6px 0', fontWeight: 700 }}>{selectedInvoice.customer_name}</h3>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedInvoice.customer_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedInvoice.customer_address}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}><strong>State:</strong> {selectedInvoice.customer_state} ({selectedInvoice.customer_state_code})</div>
								</div>
								<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.5fr', gap: '10px 16px', fontSize: '13px', color: 'var(--text-secondary)' }}>
									<strong>Invoice Number:</strong>
									<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedInvoice.invoice_number || 'DRAFT'}</span>

									<strong>Invoice Date:</strong>
									<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.invoice_date ? selectedInvoice.invoice_date.split('T')[0] : ''}</span>

									{selectedInvoice.due_date && (
										<>
											<strong>Due Date:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.due_date.split('T')[0]}</span>
										</>
									)}

									{selectedInvoice.order_number && (
										<>
											<strong>Order Number:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.order_number}</span>
										</>
									)}

									{selectedInvoice.salesperson && (
										<>
											<strong>Salesperson:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.salesperson}</span>
										</>
									)}
								</div>
							</div>

							{selectedInvoice.subject && (
								<div style={{ background: 'var(--bg-input)', padding: '12px 16px', borderRadius: '8px', marginBottom: '32px', borderLeft: '4px solid var(--accent-color)', color: 'var(--text-primary)' }}>
									<strong>Subject:</strong> {selectedInvoice.subject}
								</div>
							)}

							<table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '36px' }}>
								<thead>
									<tr style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', fontSize: '13px', color: 'var(--text-secondary)' }}>
										<th style={{ padding: '12px 8px', fontWeight: 600 }}>Item Details</th>
										<th style={{ padding: '12px 8px', width: '100px', textAlign: 'right', fontWeight: 600 }}>Qty</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Rate (₹)</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Amount (₹)</th>
									</tr>
								</thead>
								<tbody>
									{selectedInvoice.items.map((item, idx) => (
										<tr key={idx} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '14px', color: 'var(--text-primary)' }}>
											<td style={{ padding: '12px 8px' }}>
												<strong>{item.item_details}</strong>
												{item.hsn_code && <span style={{ display: 'block', fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '2px' }}>HSN: {item.hsn_code}</span>}
											</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>{item.quantity}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>₹{item.rate.toFixed(2)}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right', fontWeight: 'bold' }}>₹{item.amount.toFixed(2)}</td>
										</tr>
									))}
								</tbody>
							</table>

							{/* Calculations breakdowns */}
							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px' }}>
								<div>
									{selectedInvoice.customer_notes && (
										<div>
											<h5 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', fontSize: '11px', textTransform: 'uppercase' }}>Customer Notes</h5>
											<div style={{ color: 'var(--text-secondary)', fontSize: '13px', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>{selectedInvoice.customer_notes}</div>
										</div>
									)}
								</div>
								<div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
									<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
										<span>Sub Total:</span>
										<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.subtotal_price.toFixed(2)}</span>
									</div>

									{selectedInvoice.discount_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>Discount ({selectedInvoice.discount_percent}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>- ₹{selectedInvoice.discount_amount.toFixed(2)}</span>
										</div>
									)}

									{selectedInvoice.cgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>CGST ({selectedInvoice.cgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.cgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedInvoice.sgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>SGST ({selectedInvoice.sgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.sgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedInvoice.igst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>IGST ({selectedInvoice.igst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.igst_amount.toFixed(2)}</span>
										</div>
									)}

									{selectedInvoice.transportation_charge > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>Transportation Charges:</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.transportation_charge.toFixed(2)}</span>
										</div>
									)}

									{selectedInvoice.tds_tcs_type !== 'NONE' && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', color: 'var(--status-warning)', fontWeight: 600 }}>
											<span>{selectedInvoice.tds_tcs_type} ({selectedInvoice.tds_tcs_rate}%):</span>
											<span>{selectedInvoice.tds_tcs_type === 'TDS' ? '-' : '+'} ₹{selectedInvoice.tds_tcs_amount.toFixed(2)}</span>
										</div>
									)}

									<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', margin: '12px 0' }} />

									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '18px', fontWeight: 800, color: 'var(--text-primary)' }}>
										<span>Grand Total (₹):</span>
										<span>₹{selectedInvoice.total_price.toFixed(2)}</span>
									</div>

									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px', marginTop: '8px' }}>
										<span>Paid Amount:</span>
										<span>₹{selectedInvoice.paid_amount.toFixed(2)}</span>
									</div>
									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px', fontWeight: 700, marginTop: '4px' }}>
										<span>Balance Due:</span>
										<span style={{ color: selectedInvoice.balance_amount > 0 ? 'var(--status-warning)' : 'var(--status-active)' }}>₹{selectedInvoice.balance_amount.toFixed(2)}</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				)}
			</div>

			{/* CUSTOMER REGISTRATION DIALOG MODAL */}
			{showCustomerModal && editingCustomer && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowCustomerModal(false);
					setEditingCustomer(null);
				}} style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'center', overflowY: 'auto', padding: '2rem 1.5rem', zIndex: 3000 }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '800px', width: '100%', margin: '0 auto', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15)', borderRadius: '24px', padding: '2.5rem', background: 'var(--surface-color)', position: 'relative' }}>

						{/* Modal Header */}
						<div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '2rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1.25rem' }}>
							<div style={{
								width: '48px',
								height: '48px',
								borderRadius: '12px',
								background: 'linear-gradient(135deg, var(--accent-color), #10b981)',
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								color: 'white',
								boxShadow: '0 8px 16px rgba(16, 185, 129, 0.15)',
								flexShrink: 0
							}}>
								<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<path d="M17 21v-2a4 4 0 0 0-3-3.87" />
									<path d="M16 3.13a4 4 0 0 1 0 7.75" />
									<circle cx="9" cy="7" r="4" />
									<path d="M17 11a5.5 5.5 0 0 0-4.5 4.5" />
									<path d="M2 21v-2a4 4 0 0 1 4-4h6a4 4 0 0 1 4 4v2" />
								</svg>
							</div>
							<div>
								<h2 style={{ margin: 0, fontSize: '1.45rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>
									{editingCustomer.id ? 'Edit Client Details' : 'Register New Client'}
								</h2>
								<p style={{ margin: '3px 0 0 0', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
									Provide the official GSTIN, billing address, and shipping address details.
								</p>
							</div>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
							{/* Left Column: Business Details */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
								<div className="sync-form-group">
									<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Legal Business Name*</label>
									<input
										type="text"
										className="b2b-input"
										value={editingCustomer.legal_name}
										onChange={(e) => setEditingCustomer({ ...editingCustomer, legal_name: e.target.value })}
										placeholder="e.g. Acme Corporation Pvt Ltd"
									/>
								</div>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Trade Name (Optional)</label>
									<input
										type="text"
										className="b2b-input"
										value={editingCustomer.trade_name || ''}
										onChange={(e) => setEditingCustomer({ ...editingCustomer, trade_name: e.target.value })}
										placeholder="e.g. Acme Stores"
									/>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '16px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Client GSTIN*</label>
										<input
											type="text"
											placeholder="e.g. 33ABCDE1234F1Z5"
											maxLength={15}
											className="b2b-input"
											value={editingCustomer.gstin}
											onChange={(e) => {
												const val = e.target.value.toUpperCase();
												const updated = { ...editingCustomer, gstin: val };
												if (val.length >= 2) {
													const code = val.substring(0, 2);
													const detectedState = GST_STATE_MAP[code];
													if (detectedState) {
														updated.state_code = code;
														updated.state = detectedState;
													}
												}
												if (val.length >= 12) {
													const pan = val.substring(2, 12);
													updated.pan = pan;
												}
												setEditingCustomer(updated);
											}}
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>PAN</label>
										<input
											type="text"
											placeholder="Auto-extracted"
											maxLength={10}
											className="b2b-input"
											value={editingCustomer.pan || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, pan: e.target.value.toUpperCase() })}
										/>
									</div>
								</div>
							</div>

							{/* Right Column: Contact & GST Treatment */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Email</label>
										<input
											type="email"
											className="b2b-input"
											value={editingCustomer.email || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, email: e.target.value })}
											placeholder="billing@acme.com"
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Phone</label>
										<input
											type="text"
											className="b2b-input"
											value={editingCustomer.phone || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, phone: e.target.value })}
											placeholder="e.g. +91 9876543210"
										/>
									</div>
								</div>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>GST Treatment</label>
									<select className="b2b-input" style={{ appearance: 'auto' }} defaultValue="regular">
										<option value="regular">Registered Business - Regular</option>
										<option value="composition">Registered Business - Composition</option>
										<option value="unregistered">Unregistered Business</option>
										<option value="consumer">Consumer</option>
										<option value="overseas">Overseas</option>
									</select>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: '16px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>State / Place of Supply</label>
										<input
											type="text"
											className="b2b-input"
											value={editingCustomer.state || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, state: e.target.value })}
											placeholder="e.g. Tamil Nadu"
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>State Code</label>
										<input
											type="text"
											className="b2b-input"
											value={editingCustomer.state_code || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, state_code: e.target.value })}
											placeholder="e.g. 33"
											maxLength={2}
										/>
									</div>
								</div>
							</div>
						</div>

						{/* Addresses Section */}
						<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', borderTop: '1px solid var(--border-color)', marginTop: '24px', paddingTop: '20px' }}>
							{/* Billing Address Card */}
							<div style={{ background: 'var(--bg-input)', padding: '16px', borderRadius: '16px', border: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '12px' }}>
								<h3 style={{ margin: '0 0 4px 0', fontSize: '0.8rem', fontWeight: 800, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'flex', alignItems: 'center', gap: '6px' }}>
									<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ color: 'var(--accent-color)' }}>
										<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
										<circle cx="12" cy="10" r="3" />
									</svg>
									Billing Address
								</h3>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Street Address</label>
									<input
										type="text"
										className="b2b-input"
										value={billingStreet}
										onChange={(e) => setBillingStreet(e.target.value)}
										placeholder="Street, Area, Building"
									/>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>City</label>
										<input
											type="text"
											className="b2b-input"
											value={billingCity}
											onChange={(e) => setBillingCity(e.target.value)}
											placeholder="e.g. Chennai"
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Pin Code</label>
										<input
											type="text"
											className="b2b-input"
											value={billingPincode}
											onChange={(e) => setBillingPincode(e.target.value)}
											placeholder="6-digit PIN"
											maxLength={6}
										/>
									</div>
								</div>
							</div>

							{/* Shipping Address Card */}
							<div style={{ background: 'var(--bg-input)', padding: '16px', borderRadius: '16px', border: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '12px', opacity: sameAsBilling ? 0.85 : 1, transition: 'opacity 0.2s ease' }}>
								<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4px' }}>
									<h3 style={{ margin: 0, fontSize: '0.8rem', fontWeight: 800, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'flex', alignItems: 'center', gap: '6px' }}>
										<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ color: sameAsBilling ? 'var(--text-tertiary)' : 'var(--accent-color)' }}>
											<rect width="16" height="13" x="2" y="6" rx="2" />
											<path d="M16 2h4l3 4v13a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h4" />
										</svg>
										Shipping Address
									</h3>
									<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
										<input
											type="checkbox"
											id="sameAsBilling"
											checked={sameAsBilling}
											onChange={(e) => setSameAsBilling(e.target.checked)}
											style={{ width: 'auto', cursor: 'pointer' }}
										/>
										<label htmlFor="sameAsBilling" style={{ fontWeight: 700, fontSize: '0.75rem', cursor: 'pointer', color: 'var(--text-secondary)' }}>Same as Billing</label>
									</div>
								</div>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Street Address</label>
									<input
										type="text"
										className="b2b-input"
										value={sameAsBilling ? billingStreet : shippingStreet}
										onChange={(e) => setShippingStreet(e.target.value)}
										placeholder="Street, Area, Building"
										disabled={sameAsBilling}
										style={{ opacity: sameAsBilling ? 0.6 : 1 }}
									/>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>City</label>
										<input
											type="text"
											className="b2b-input"
											value={sameAsBilling ? billingCity : shippingCity}
											onChange={(e) => setShippingCity(e.target.value)}
											placeholder="e.g. Chennai"
											disabled={sameAsBilling}
											style={{ opacity: sameAsBilling ? 0.6 : 1 }}
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Pin Code</label>
										<input
											type="text"
											className="b2b-input"
											value={sameAsBilling ? billingPincode : shippingPincode}
											onChange={(e) => setShippingPincode(e.target.value)}
											placeholder="6-digit PIN"
											maxLength={6}
											disabled={sameAsBilling}
											style={{ opacity: sameAsBilling ? 0.6 : 1 }}
										/>
									</div>
								</div>
							</div>
						</div>

						<div style={{ marginTop: '20px', borderTop: '1px solid var(--border-color)', paddingTop: '16px' }}>
							<div className="sync-form-group">
								<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Notes & Other Details</label>
								<textarea
									className="b2b-input"
									rows={3}
									style={{ resize: 'vertical', minHeight: '80px', fontFamily: 'inherit' }}
									value={editingCustomer.notes || ''}
									onChange={(e) => setEditingCustomer({ ...editingCustomer, notes: e.target.value })}
									placeholder="Add internal remarks, secondary contacts, payment terms, or shipping guidelines for this client..."
								/>
							</div>
						</div>

						<div className="modal-actions" style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '24px', borderTop: '1px solid var(--border-color)', paddingTop: '20px' }}>
							<button onClick={() => {
								setShowCustomerModal(false);
								setEditingCustomer(null);
							}} className="b2b-btn b2b-btn-secondary" style={{ minWidth: '120px' }}>Cancel</button>
							<button onClick={() => handleSaveCustomer(editingCustomer)} className="b2b-btn b2b-btn-primary" style={{ minWidth: '140px' }}>Save Client</button>
						</div>
					</div>
				</div>,
				document.body
			)}

			{/* PAYMENT ENTRY MODAL */}
			{showPaymentModal && paymentInvoice && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowPaymentModal(false);
					setPaymentInvoice(null);
				}} style={{ alignItems: 'flex-start', overflowY: 'auto', padding: '2rem 1.5rem' }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '420px', width: '100%', margin: '0 auto' }}>
						<h2>Record Payment</h2>
						<p>Record a client transaction for this invoice below.</p>

						<div style={{ fontSize: '0.85rem', background: 'var(--bg-input)', padding: '14px', borderRadius: '10px', marginBottom: '20px', lineHeight: 1.6, border: '1px solid var(--border-color)' }}>
							<strong>Invoice:</strong> {paymentInvoice.invoice_number}<br />
							<strong>Total Price:</strong> ₹{paymentInvoice.total_price.toFixed(2)}<br />
							<strong>Current Balance:</strong> ₹{paymentInvoice.balance_amount.toFixed(2)}
						</div>

						<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
							<div className="sync-form-group">
								<label>Paid Amount (₹)</label>
								<input
									type="number"
									className="b2b-input"
									value={paymentAmount}
									onChange={(e) => setPaymentAmount(Number(e.target.value))}
									max={paymentInvoice.total_price}
								/>
							</div>

							<div className="sync-form-group">
								<label>Payment Method</label>
								<select
									value={paymentMethod}
									onChange={(e) => setPaymentMethod(e.target.value)}
									className="b2b-input"
								>
									<option value="Bank Transfer">Bank Transfer</option>
									<option value="Cash">Cash</option>
									<option value="UPI">UPI</option>
									<option value="Cheque">Cheque</option>
								</select>
							</div>
						</div>

						<div className="modal-actions" style={{ marginTop: '28px' }}>
							<button onClick={() => {
								setShowPaymentModal(false);
								setPaymentInvoice(null);
							}} className="btn-secondary">Cancel</button>
							<button onClick={handleSavePayment} className="btn-primary">Record Payment</button>
						</div>
					</div>
				</div>,
				document.body
			)}

			{/* NEW PAYMENT TERM MODAL */}
			{showPaymentTermModal && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowPaymentTermModal(false);
					setNewTermName('');
					setNewTermDays(0);
				}} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', overflowY: 'auto', padding: '2rem 1.5rem', zIndex: 3100 }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '450px', width: '100%', margin: '0 auto', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15)', borderRadius: '24px', padding: '2.5rem', background: 'var(--surface-color)', position: 'relative' }}>

						{/* Modal Header */}
						<div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '2rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1.25rem' }}>
							<div style={{
								width: '48px',
								height: '48px',
								borderRadius: '12px',
								background: 'linear-gradient(135deg, var(--accent-color), #0ea5e9)',
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								color: 'white',
								boxShadow: '0 8px 16px rgba(14, 165, 233, 0.15)',
								flexShrink: 0
							}}>
								<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
									<line x1="16" y1="2" x2="16" y2="6" />
									<line x1="8" y1="2" x2="8" y2="6" />
									<line x1="3" y1="10" x2="21" y2="10" />
								</svg>
							</div>
							<div>
								<h2 style={{ margin: 0, fontSize: '1.45rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>New Payment Term</h2>
								<p style={{ margin: '3px 0 0 0', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>Configure payment timelines for B2B billing.</p>
							</div>
						</div>

						{/* Yellow Banner Notification */}
						<div style={{ background: 'var(--accent-subtle)', padding: '14px', borderRadius: '12px', border: '1px solid rgba(14, 165, 233, 0.15)', color: 'var(--text-secondary)', fontSize: '0.85rem', marginBottom: '20px', display: 'flex', alignItems: 'center', gap: '8px', lineHeight: 1.4 }}>
							<span>💡</span>
							<span>The Payment Terms can now be configured and managed from <strong>Settings &rarr; Setup & Configuration &rarr; Payment Terms</strong>.</span>
						</div>

						<div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
							<div className="sync-form-group">
								<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Term Name*</label>
								<input
									type="text"
									className="b2b-input"
									value={newTermName}
									onChange={(e) => setNewTermName(e.target.value)}
									placeholder="e.g. Net 45"
								/>
							</div>

							<div className="sync-form-group">
								<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Due After*</label>
								<div style={{ display: 'flex', alignItems: 'center' }}>
									<input
										type="number"
										className="b2b-input days-input-left"
										value={newTermDays}
										onChange={(e) => setNewTermDays(Number(e.target.value))}
										min="0"
									/>
									<span className="days-label-right">Days</span>
								</div>
							</div>
						</div>

						<div className="modal-actions" style={{ marginTop: '28px' }}>
							<button onClick={() => {
								setShowPaymentTermModal(false);
								setNewTermName('');
								setNewTermDays(0);
							}} className="b2b-btn b2b-btn-secondary" style={{ flex: 1 }}>Cancel</button>
							<button onClick={handleSavePaymentTerm} className="b2b-btn b2b-btn-primary" style={{ flex: 1 }}>Save</button>
						</div>
					</div>
				</div>,
				document.body
			)}
		</>
	);
}
